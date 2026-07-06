// Copyright 2026 The go-sila Authors
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

package misc

import (
	"github.com/sila-chain/go-sila/core/tracing"
	"github.com/sila-chain/go-sila/core/vm"
	"github.com/sila-chain/go-sila/crypto"
	"github.com/sila-chain/go-sila/params"
)

// ApplySIP7997 inserts the deterministic deployment factory into the state as an
// irregular state transition, as specified by SIP-7997. The factory is a keyless
// CREATE2 factory that, once present at the canonical address on every EVM chain,
// allows contracts to be deployed at identical addresses across chains.
func ApplySIP7997(statedb vm.StateDB) {
	// The account must hold the canonical factory runtime code. If its code hash
	// already matches, the chain satisfies SIP-7997 and nothing needs to change.
	wantHash := crypto.Keccak256Hash(params.DeterministicFactoryCode)
	if statedb.GetCodeHash(params.DeterministicFactoryAddress) == wantHash {
		return
	}
	if !statedb.Exist(params.DeterministicFactoryAddress) {
		statedb.CreateAccount(params.DeterministicFactoryAddress)
	}
	statedb.CreateContract(params.DeterministicFactoryAddress)
	statedb.SetCode(params.DeterministicFactoryAddress, params.DeterministicFactoryCode, tracing.CodeChangeUnspecified)

	// Preserve a pre-existing nonce; only bump the default zero nonce to 1.
	if statedb.GetNonce(params.DeterministicFactoryAddress) == 0 {
		statedb.SetNonce(params.DeterministicFactoryAddress, 1, tracing.NonceChangeNewContract)
	}
}
