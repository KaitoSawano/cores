// Copyright 2015 The go-xcosh Authors
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

package backends

import (
	"context"

	"github.com/xcosh/go-xcosh/common"
	"github.com/xcosh/go-xcosh/ethclient/simulated"
	"github.com/xcosh/go-xcosh/params/types/genesisT"
)

// SimulatedBackend is a simulated blockchain.
// Deprecated: use package github.com/xcosh/go-xcosh/ethclient/simulated instead.
type SimulatedBackend struct {
	*simulated.Backend
	simulated.Client
}

// Fork sets the head to a new block, which is based on the provided parentHash.
func (b *SimulatedBackend) Fork(ctx context.Context, parentHash common.Hash) error {
	return b.Backend.Fork(parentHash)
}

// NewSimulatedBackend creates a new binding backend using a simulated blockchain
// for testing purposes.
//
// A simulated backend always uses chainID 1337.
//
// Deprecated: please use simulated.Backend from package
// github.com/xcosh/go-xcosh/ethclient/simulated instead.
func NewSimulatedBackend(alloc genesisT.GenesisAlloc, gasLimit uint64) *SimulatedBackend {
	b := simulated.NewBackend(alloc, simulated.WithBlockGasLimit(gasLimit))
	return &SimulatedBackend{
		Backend: b,
		Client:  b.Client(),
	}
}
