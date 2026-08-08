// Copyright 2018 The go-xcosh Authors
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

// This file contains a miner stress test based on the Ethash consensus engine.
package main

import (
	"crypto/ecdsa"
	"math/big"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"golang.org/x/exp/slog"

	"github.com/xcosh/go-xcosh/common"
	"github.com/xcosh/go-xcosh/common/fdlimit"
	"github.com/xcosh/go-xcosh/consensus/ethash"
	"github.com/xcosh/go-xcosh/core/txpool/legacypool"
	"github.com/xcosh/go-xcosh/core/types"
	"github.com/xcosh/go-xcosh/crypto"
	"github.com/xcosh/go-xcosh/eth"
	"github.com/xcosh/go-xcosh/eth/downloader"
	"github.com/xcosh/go-xcosh/eth/ethconfig"
	"github.com/xcosh/go-xcosh/miner"
	"github.com/xcosh/go-xcosh/node"
	"github.com/xcosh/go-xcosh/p2p"
	"github.com/xcosh/go-xcosh/p2p/enode"
	"github.com/xcosh/go-xcosh/params"
	"github.com/xcosh/go-xcosh/params/types/genesisT"
	"github.com/xcosh/go-xcosh/params/vars"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))
	fdlimit.Raise(2048)

	// Generate a batch of accounts to seal and fund with
	faucets := make([]*ecdsa.PrivateKey, 128)
	for i := 0; i < len(faucets); i++ {
		faucets[i], _ = crypto.GenerateKey()
	}
	// Pre-generate the ethash mining DAG so we don't race
	ethash.MakeDataset(1, ethash.CalcEpochLength(0, nil), ethconfig.Defaults.Ethash.DatasetDir)

	// Create an Ethash network
	genesis := makeGenesis(faucets)

	// Handle interrupts.
	interruptCh := make(chan os.Signal, 5)
	signal.Notify(interruptCh, os.Interrupt)

	var (
		stacks []*node.Node
		nodes  []*eth.Xcosh
		enodes []*enode.Node
	)
	for i := 0; i < 4; i++ {
		// Start the node and wait until it's up
		stack, ethBackend, err := makeMiner(genesis)
		if err != nil {
			panic(err)
		}
		defer stack.Close()

		for stack.Server().NodeInfo().Ports.Listener == 0 {
			time.Sleep(250 * time.Millisecond)
		}
		// Connect the node to all the previous ones
		for _, n := range enodes {
			stack.Server().AddPeer(n)
		}
		// Start tracking the node and its enode
		stacks = append(stacks, stack)
		nodes = append(nodes, ethBackend)
		enodes = append(enodes, stack.Server().Self())
	}

	// Iterate over all the nodes and start mining
	time.Sleep(3 * time.Second)
	for _, node := range nodes {
		if err := node.StartMining(1); err != nil {
			panic(err)
		}
	}
	time.Sleep(3 * time.Second)

	// Start injecting transactions from the faucets like crazy
	nonces := make([]uint64, len(faucets))
	for {
		// Stop when interrupted.
		select {
		case <-interruptCh:
			for _, node := range stacks {
				node.Close()
			}
			return
		default:
		}

		// Pick a random mining node
		index := rand.Intn(len(faucets))
		backend := nodes[index%len(nodes)]

		// Create a self transaction and inject into the pool
		tx, err := types.SignTx(types.NewTransaction(nonces[index], crypto.PubkeyToAddress(faucets[index].PublicKey), new(big.Int), 21000, big.NewInt(100000000000+rand.Int63n(65536)), nil), types.HomesteadSigner{}, faucets[index])
		if err != nil {
			panic(err)
		}
		if err := backend.TxPool().Add([]*types.Transaction{tx}, true, false); err != nil {
			panic(err)
		}
		nonces[index]++

		// Wait if we're too saturated
		if pend, _ := backend.TxPool().Stats(); pend > 2048 {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// makeGenesis creates a custom Ethash genesis block based on some pre-defined
// faucet accounts.
func makeGenesis(faucets []*ecdsa.PrivateKey) *genesisT.Genesis {
	genesis := params.DefaultGenesisBlock()
	genesis.Difficulty = vars.MinimumDifficulty
	genesis.GasLimit = 25000000

	genesis.SetChainID(big.NewInt(18))
	// genesis.Config.EIP150Hash = common.Hash{}

	genesis.Alloc = genesisT.GenesisAlloc{}
	for _, faucet := range faucets {
		genesis.Alloc[crypto.PubkeyToAddress(faucet.PublicKey)] = genesisT.GenesisAccount{
			Balance: new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil),
		}
	}
	return genesis
}

func makeMiner(genesis *genesisT.Genesis) (*node.Node, *eth.Xcosh, error) {
	// Define the basic configurations for the Xcosh node
	datadir, _ := os.MkdirTemp("", "")

	config := &node.Config{
		Name:    "geth",
		Version: params.Version,
		DataDir: datadir,
		P2P: p2p.Config{
			ListenAddr:  "0.0.0.0:0",
			NoDiscovery: true,
			MaxPeers:    25,
		},
		UseLightweightKDF: true,
	}
	// Create the node and configure a full Xcosh node on it
	stack, err := node.New(config)
	if err != nil {
		return nil, nil, err
	}
	ethBackend, err := eth.New(stack, &ethconfig.Config{
		Genesis:         genesis,
		NetworkId:       genesis.Config.GetChainID().Uint64(),
		SyncMode:        downloader.FullSync,
		DatabaseCache:   256,
		DatabaseHandles: 256,
		TxPool:          legacypool.DefaultConfig,
		GPO:             ethconfig.Defaults.GPO,
		Ethash:          ethconfig.Defaults.Ethash,
		Miner: miner.Config{
			Etherbase: common.Address{1},
			GasCeil:   genesis.GasLimit * 11 / 10,
			GasPrice:  big.NewInt(1),
			Recommit:  time.Second,
		},
	})
	if err != nil {
		return nil, nil, err
	}

	err = stack.Start()
	return stack, ethBackend, err
}
