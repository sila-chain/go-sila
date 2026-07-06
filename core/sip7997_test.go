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

// Tests for SIP-7997: the deterministic deployment factory inserted as an
// irregular state transition at the Amsterdam activation block.

package core

import (
	"bytes"
	"testing"

	"github.com/sila-chain/go-sila/common"
	"github.com/sila-chain/go-sila/consensus/misc"
	"github.com/sila-chain/go-sila/core/types"
	"github.com/sila-chain/go-sila/core/vm"
	"github.com/sila-chain/go-sila/crypto"
	"github.com/sila-chain/go-sila/params"
	"github.com/holiman/uint256"
)

// TestApplySIP7997 verifies the irregular state transition seeds the factory
// account with the canonical code and nonce.
func TestApplySIP7997(t *testing.T) {
	sdb := mkState(nil)
	misc.ApplySIP7997(sdb)

	if got := sdb.GetCode(params.DeterministicFactoryAddress); !bytes.Equal(got, params.DeterministicFactoryCode) {
		t.Fatalf("factory code mismatch:\n got %x\nwant %x", got, params.DeterministicFactoryCode)
	}
	if got := sdb.GetNonce(params.DeterministicFactoryAddress); got != 1 {
		t.Fatalf("factory nonce = %d, want %d", got, 1)
	}
}

// TestApplySIP7997Existing checks that a chain which already hosts the factory
// (for example via its keyless creation transaction) is left untouched, so the
// transition never rewrites an existing nonce.
func TestApplySIP7997Existing(t *testing.T) {
	sdb := mkState(types.GenesisAlloc{
		params.DeterministicFactoryAddress: {Code: params.DeterministicFactoryCode, Nonce: 5},
	})
	misc.ApplySIP7997(sdb)

	if got := sdb.GetNonce(params.DeterministicFactoryAddress); got != 5 {
		t.Fatalf("existing factory nonce overwritten: got %d, want 5", got)
	}
}

// TestApplySIP7997WrongCode checks that an account occupying the factory address
// with the wrong code is force-overwritten with the canonical runtime code, while
// a pre-existing non-zero nonce is preserved.
func TestApplySIP7997WrongCode(t *testing.T) {
	sdb := mkState(types.GenesisAlloc{
		params.DeterministicFactoryAddress: {Code: []byte{0x60, 0x00}, Nonce: 7},
	})
	misc.ApplySIP7997(sdb)

	if got := sdb.GetCode(params.DeterministicFactoryAddress); !bytes.Equal(got, params.DeterministicFactoryCode) {
		t.Fatalf("factory code not overwritten:\n got %x\nwant %x", got, params.DeterministicFactoryCode)
	}
	if got := sdb.GetNonce(params.DeterministicFactoryAddress); got != 7 {
		t.Fatalf("factory nonce = %d, want %d (existing nonce must be preserved)", got, 7)
	}
}

// TestSIP7997FactoryDeploys exercises the inserted factory bytecode: calling it
// with a salt followed by init code must CREATE2-deploy the contract at the
// canonical deterministic address and return that address (20 bytes, unpadded).
func TestSIP7997FactoryDeploys(t *testing.T) {
	sdb := mkState(nil)
	misc.ApplySIP7997(sdb)

	var (
		caller = common.Address{0xca}
		salt   [32]byte
		// initcode returning the single-byte runtime 0xfe:
		//   PUSH1 0xfe PUSH1 0x00 MSTORE8 PUSH1 0x01 PUSH1 0x00 RETURN
		initcode = common.FromHex("60fe60005360016000f3")
	)
	salt[31] = 0x42

	input := append(append([]byte{}, salt[:]...), initcode...)

	ret, _, err := amsterdamCoreEVM(sdb).Call(caller, params.DeterministicFactoryAddress, input, vm.NewGasBudget(10_000_000, 0), new(uint256.Int))
	if err != nil {
		t.Fatalf("factory call failed: %v", err)
	}

	want := crypto.CreateAddress2(params.DeterministicFactoryAddress, salt, crypto.Keccak256(initcode))
	if len(ret) != 20 {
		t.Fatalf("factory returned %d bytes, want 20", len(ret))
	}
	if got := common.BytesToAddress(ret); got != want {
		t.Fatalf("factory returned address %x, want %x", got, want)
	}
	if code := sdb.GetCode(want); !bytes.Equal(code, []byte{0xfe}) {
		t.Fatalf("deployed runtime code = %x, want fe", code)
	}
}
