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

package eth

import (
	"github.com/xcosh/go-xcosh/common"
	"github.com/xcosh/go-xcosh/common/hexutil"
)

// XcoshAPI provides an API to access Xcosh full node-related information.
type XcoshAPI struct {
	e *Xcosh
}

// NewXcoshAPI creates a new Xcosh protocol API for full nodes.
func NewXcoshAPI(e *Xcosh) *XcoshAPI {
	return &XcoshAPI{e}
}

// Etherbase is the address that mining rewards will be sent to.
func (api *XcoshAPI) Etherbase() (common.Address, error) {
	return api.e.Etherbase()
}

// Coinbase is the address that mining rewards will be sent to (alias for Etherbase).
func (api *XcoshAPI) Coinbase() (common.Address, error) {
	return api.Etherbase()
}

// Hashrate returns the POW hashrate.
func (api *XcoshAPI) Hashrate() hexutil.Uint64 {
	return hexutil.Uint64(api.e.Miner().Hashrate())
}

// Mining returns an indication if this node is currently mining.
func (api *XcoshAPI) Mining() bool {
	return api.e.IsMining()
}
