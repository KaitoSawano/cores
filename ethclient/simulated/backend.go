// Copyright 2023 The go-xcosh Authors
// This file is part of the go-xcosh library.
//
// The go-xcosh library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-xcosh library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-xcosh library. If not, see <http://www.gnu.org/licenses/>.

package simulated

import (
	"time"

	"github.com/xcosh/go-xcosh"
	"github.com/xcosh/go-xcosh/common"
	"github.com/xcosh/go-xcosh/eth"
	"github.com/xcosh/go-xcosh/eth/catalyst"
	"github.com/xcosh/go-xcosh/eth/downloader"
	"github.com/xcosh/go-xcosh/eth/ethconfig"
	"github.com/xcosh/go-xcosh/eth/filters"
	"github.com/xcosh/go-xcosh/ethclient"
	"github.com/xcosh/go-xcosh/node"
	"github.com/xcosh/go-xcosh/p2p"
	"github.com/xcosh/go-xcosh/params"
	"github.com/xcosh/go-xcosh/params/types/genesisT"
	"github.com/xcosh/go-xcosh/rpc"
)

// Client exposes the methods provided by the Xcosh RPC client.
type Client interface {
	xcosh.BlockNumberReader
	xcosh.ChainReader
	xcosh.ChainStateReader
	xcosh.ContractCaller
	xcosh.GasEstimator
	xcosh.GasPricer
	xcosh.GasPricer1559
	xcosh.FeeHistoryReader
	xcosh.LogFilterer
	xcosh.PendingStateReader
	xcosh.PendingContractCaller
	xcosh.TransactionReader
	xcosh.TransactionSender
	xcosh.ChainIDReader
}

// simClient wraps ethclient. This exists to prevent extracting ethclient.Client
// from the Client interface returned by Backend.
type simClient struct {
	*ethclient.Client
}

// Backend is a simulated blockchain. You can use it to test your contracts or
// other code that interacts with the Xcosh chain.
type Backend struct {
	eth    *eth.Xcosh
	beacon *catalyst.SimulatedBeacon
	client simClient
}

// NewBackend creates a new simulated blockchain that can be used as a backend for
// contract bindings in unit tests.
//
// A simulated backend always uses chainID 1337.
func NewBackend(alloc genesisT.GenesisAlloc, options ...func(nodeConf *node.Config, ethConf *ethconfig.Config)) *Backend {
	// Create the default configurations for the outer node shell and the Xcosh
	// service to mutate with the options afterwards
	nodeConf := node.DefaultConfig
	nodeConf.DataDir = ""
	nodeConf.P2P = p2p.Config{NoDiscovery: true}

	ethConf := ethconfig.Defaults
	ethConf.Genesis = &genesisT.Genesis{
		Config:   params.AllDevChainProtocolChanges,
		GasLimit: ethconfig.Defaults.Miner.GasCeil,
		Alloc:    alloc,
	}
	ethConf.SyncMode = downloader.FullSync
	ethConf.TxPool.NoLocals = true

	for _, option := range options {
		option(&nodeConf, &ethConf)
	}
	// Assemble the Xcosh stack to run the chain with
	stack, err := node.New(&nodeConf)
	if err != nil {
		panic(err) // this should never happen
	}
	sim, err := newWithNode(stack, &ethConf, 0)
	if err != nil {
		panic(err) // this should never happen
	}
	return sim
}

// newWithNode sets up a simulated backend on an existing node. The provided node
// must not be started and will be started by this method.
func newWithNode(stack *node.Node, conf *eth.Config, blockPeriod uint64) (*Backend, error) {
	backend, err := eth.New(stack, conf)
	if err != nil {
		return nil, err
	}
	// Register the filter system
	filterSystem := filters.NewFilterSystem(backend.APIBackend, filters.Config{})
	stack.RegisterAPIs([]rpc.API{{
		Namespace: "eth",
		Service:   filters.NewFilterAPI(filterSystem, false),
	}})
	// Start the node
	if err := stack.Start(); err != nil {
		return nil, err
	}
	// Set up the simulated beacon
	beacon, err := catalyst.NewSimulatedBeacon(blockPeriod, backend)
	if err != nil {
		return nil, err
	}
	// Reorg our chain back to genesis
	if err := beacon.Fork(backend.BlockChain().GetCanonicalHash(0)); err != nil {
		return nil, err
	}
	return &Backend{
		eth:    backend,
		beacon: beacon,
		client: simClient{ethclient.NewClient(stack.Attach())},
	}, nil
}

// Close shuts down the simBackend.
// The simulated backend can't be used afterwards.
func (n *Backend) Close() error {
	if n.client.Client != nil {
		n.client.Close()
		n.client = simClient{}
	}
	if n.beacon != nil {
		err := n.beacon.Stop()
		n.beacon = nil
		return err
	}
	return nil
}

// Commit seals a block and moves the chain forward to a new empty block.
func (n *Backend) Commit() common.Hash {
	return n.beacon.Commit()
}

// Rollback removes all pending transactions, reverting to the last committed state.
func (n *Backend) Rollback() {
	n.beacon.Rollback()
}

// Fork creates a side-chain that can be used to simulate reorgs.
//
// This function should be called with the ancestor block where the new side
// chain should be started. Transactions (old and new) can then be applied on
// top and Commit-ed.
//
// Note, the side-chain will only become canonical (and trigger the events) when
// it becomes longer. Until then CallContract will still operate on the current
// canonical chain.
//
// There is a % chance that the side chain becomes canonical at the same length
// to simulate live network behavior.
func (n *Backend) Fork(parentHash common.Hash) error {
	return n.beacon.Fork(parentHash)
}

// AdjustTime changes the block timestamp and creates a new block.
// It can only be called on empty blocks.
func (n *Backend) AdjustTime(adjustment time.Duration) error {
	return n.beacon.AdjustTime(adjustment)
}

// Client returns a client that accesses the simulated chain.
func (n *Backend) Client() Client {
	return n.client
}
