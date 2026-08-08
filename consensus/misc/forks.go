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

package misc

import (
	"fmt"

	"github.com/xcosh/go-xcosh/common"
	"github.com/xcosh/go-xcosh/core/types"
	"github.com/xcosh/go-xcosh/params/types/ctypes"
)

// VerifyForkHashes verifies that blocks conforming to network hard-forks do have
// the correct hashes, to avoid clients going off on different chains. This is an
// optional feature.
func VerifyForkHashes(config ctypes.ChainConfigurator, header *types.Header, uncle bool) error {
	// We don't care about uncles
	if uncle {
		return nil
	}
	if wantHash := config.GetForkCanonHash(header.Number.Uint64()); wantHash == (common.Hash{}) || wantHash == header.Hash() {
		return nil
	} else {
		return fmt.Errorf("verify canonical block hash failed, block number %d: have %#x, want %#x", header.Number.Uint64(), header.Hash(), wantHash)
	}
}
