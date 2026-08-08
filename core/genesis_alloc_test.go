// Copyright 2017 The go-xcosh Authors
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

package core

import (
	"testing"

	"github.com/xcosh/go-xcosh/params"
)

func TestDefaultGenesisBlock(t *testing.T) {
	block := GenesisToBlock(params.DefaultGenesisBlock(), nil)
	if block.Hash() != params.MainnetGenesisHash {
		t.Errorf("wrong mainnet genesis hash, got %v, want %v", block.Hash(), params.MainnetGenesisHash)
	}
	block = GenesisToBlock(params.DefaultSepoliaGenesisBlock(), nil)
	if block.Hash() != params.SepoliaGenesisHash {
		t.Errorf("wrong testnet genesis hash, got %v, want %v", block.Hash(), params.SepoliaGenesisHash)
	}
}
