// Copyright 2024 The go-sila Authors
// This file is part of the go-sila library.
//
// The go-sila library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-sila library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-sila library. If not, see <http://www.gnu.org/licenses/>.

package simulated

import (
	"math/big"

	"github.com/sila-chain/go-sila/sil/silconfig"
	"github.com/sila-chain/go-sila/node"
)

// WithBlockGasLimit configures the simulated backend to target a specific gas limit
// when producing blocks.
func WithBlockGasLimit(gaslimit uint64) func(nodeConf *node.Config, silConf *silconfig.Config) {
	return func(nodeConf *node.Config, silConf *silconfig.Config) {
		silConf.Genesis.GasLimit = gaslimit
		silConf.Miner.GasCeil = gaslimit
	}
}

// WithCallGasLimit configures the simulated backend to cap sil_calls to a specific
// gas limit when running client operations.
func WithCallGasLimit(gaslimit uint64) func(nodeConf *node.Config, silConf *silconfig.Config) {
	return func(nodeConf *node.Config, silConf *silconfig.Config) {
		silConf.RPCGasCap = gaslimit
	}
}

// WithMinerMinTip configures the simulated backend to require a specific minimum
// gas tip for a transaction to be included.
//
// 0 is not possible as a live Sila node would reject that due to DoS protection,
// so the simulated backend will replicate that behavior for consistency.
func WithMinerMinTip(tip *big.Int) func(nodeConf *node.Config, silConf *silconfig.Config) {
	if tip == nil || tip.Sign() <= 0 {
		panic("invalid miner minimum tip")
	}
	return func(nodeConf *node.Config, silConf *silconfig.Config) {
		silConf.Miner.GasPrice = tip
	}
}
