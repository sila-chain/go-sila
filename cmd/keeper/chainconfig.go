// Copyright 2025 The go-sila Authors
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

package main

import (
	"fmt"

	"github.com/sila-chain/go-sila/params"
)

// getChainConfig returns the appropriate chain configuration based on the chainID.
// Returns an error for unsupported chain IDs.
func getChainConfig(chainID uint64) (*params.ChainConfig, error) {
	switch chainID {
	case 0, params.SilaMainnetChainConfig.ChainID.Uint64():
		return params.SilaMainnetChainConfig, nil
	case params.SilaSepoliaChainConfig.ChainID.Uint64():
		return params.SilaSepoliaChainConfig, nil
	case params.HoodiChainConfig.ChainID.Uint64():
		return params.HoodiChainConfig, nil
	default:
		return nil, fmt.Errorf("unsupported chain ID: %d", chainID)
	}
}
