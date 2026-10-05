// Copyright 2017 The go-sila Authors
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

package vm

import (
	"errors"

	"github.com/sila-chain/go-sila/common"
	"github.com/sila-chain/go-sila/common/math"
	"github.com/sila-chain/go-sila/params"
)

// memoryGasCost calculates the quadratic gas for memory expansion. It does so
// only for the memory region that is expanded, not the total memory.
func memoryGasCost(mem *Memory, newMemSize uint64) (uint64, error) {
	if newMemSize == 0 {
		return 0, nil
	}
	// The maximum that will fit in a uint64 is max_word_count - 1. Anything above
	// that will result in an overflow. Additionally, a newMemSize which results in
	// a newMemSizeWords larger than 0xFFFFFFFF will cause the square operation to
	// overflow. The constant 0x1FFFFFFFE0 is the highest number that can be used
	// without overflowing the gas calculation.
	if newMemSize > 0x1FFFFFFFE0 {
		return 0, ErrGasUintOverflow
	}
	newMemSizeWords := toWordSize(newMemSize)
	newMemSize = newMemSizeWords * 32

	if newMemSize > uint64(mem.Len()) {
		square := newMemSizeWords * newMemSizeWords
		linCoef := newMemSizeWords * params.MemoryGas
		quadCoef := square / params.QuadCoeffDiv
		newTotalFee := linCoef + quadCoef

		fee := newTotalFee - mem.lastGasCost
		mem.lastGasCost = newTotalFee

		return fee, nil
	}
	return 0, nil
}

// memoryCopierGas creates the gas functions for the following opcodes, and takes
// the stack position of the operand which determines the size of the data to copy
// as argument:
// CALLDATACOPY (stack position 2)
// CODECOPY (stack position 2)
// MCOPY (stack position 2)
// EXTCODECOPY (stack position 3)
// RETURNDATACOPY (stack position 2)
func memoryCopierGas(stackpos int) gasFunc {
	return func(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (GasCosts, error) {
		// Gas for expanding the memory
		gas, err := memoryGasCost(mem, memorySize)
		if err != nil {
			return GasCosts{}, err
		}
		// And gas for copying data, charged per word at param.CopyGas
		words, overflow := stack.back(stackpos).Uint64WithOverflow()
		if overflow {
			return GasCosts{}, ErrGasUintOverflow
		}

		if words, overflow = math.SafeMul(toWordSize(words), params.CopyGas); overflow {
			return GasCosts{}, ErrGasUintOverflow
		}

		if gas, overflow = math.SafeAdd(gas, words); overflow {
			return GasCosts{}, ErrGasUintOverflow
		}
		return GasCosts{ExecutionGas: gas}, nil
	}
}

var (
	gasCallDataCopy   = memoryCopierGas(2)
	gasCodeCopy       = memoryCopierGas(2)
	gasMcopy          = memoryCopierGas(2)
	gasExtCodeCopy    = memoryCopierGas(3)
	gasReturnDataCopy = memoryCopierGas(2)
)

func gasSStore(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (GasCosts, error) {
	if sivm.readOnly {
		return GasCosts{}, ErrWriteProtection
	}
	var (
		y, x              = stack.back(1), stack.back(0)
		current, original = sivm.StateDB.GetStateAndCommittedState(contract.Address(), x.Bytes32())
	)
	// The legacy gas metering only takes into consideration the current state
	// Legacy rules should be applied if we are in Petersburg (removal of SIP-1283)
	// OR SilaConstantinople is not active
	if sivm.chainRules.IsPetersburg || !sivm.chainRules.IsSilaConstantinople {
		// This checks for 3 scenarios and calculates gas accordingly:
		//
		// 1. From a zero-value address to a non-zero value         (NEW VALUE)
		// 2. From a non-zero value address to a zero-value address (DELETE)
		// 3. From a non-zero to a non-zero                         (CHANGE)
		switch {
		case current == (common.Hash{}) && y.Sign() != 0: // 0 => non 0
			return GasCosts{ExecutionGas: params.SstoreSetGas}, nil
		case current != (common.Hash{}) && y.Sign() == 0: // non 0 => 0
			sivm.StateDB.AddRefund(params.SstoreRefundGas)
			return GasCosts{ExecutionGas: params.SstoreClearGas}, nil
		default: // non 0 => non 0 (or 0 => 0)
			return GasCosts{ExecutionGas: params.SstoreResetGas}, nil
		}
	}

	// The new gas metering is based on net gas costs (SIP-1283):
	//
	// (1.) If current value equals new value (this is a no-op), 200 gas is deducted.
	// (2.) If current value does not equal new value
	//	(2.1.) If original value equals current value (this storage slot has not been changed by the current execution context)
	//		(2.1.1.) If original value is 0, 20000 gas is deducted.
	//		(2.1.2.) Otherwise, 5000 gas is deducted. If new value is 0, add 15000 gas to refund counter.
	//	(2.2.) If original value does not equal current value (this storage slot is dirty), 200 gas is deducted. Apply both of the following clauses.
	//		(2.2.1.) If original value is not 0
	//			(2.2.1.1.) If current value is 0 (also means that new value is not 0), remove 15000 gas from refund counter. We can prove that refund counter will never go below 0.
	//			(2.2.1.2.) If new value is 0 (also means that current value is not 0), add 15000 gas to refund counter.
	//		(2.2.2.) If original value equals new value (this storage slot is reset)
	//			(2.2.2.1.) If original value is 0, add 19800 gas to refund counter.
	//			(2.2.2.2.) Otherwise, add 4800 gas to refund counter.
	value := common.Hash(y.Bytes32())
	if current == value { // noop (1)
		return GasCosts{ExecutionGas: params.NetSstoreNoopGas}, nil
	}
	if original == current {
		if original == (common.Hash{}) { // create slot (2.1.1)
			return GasCosts{ExecutionGas: params.NetSstoreInitGas}, nil
		}
		if value == (common.Hash{}) { // delete slot (2.1.2b)
			sivm.StateDB.AddRefund(params.NetSstoreClearRefund)
		}
		return GasCosts{ExecutionGas: params.NetSstoreCleanGas}, nil // write existing slot (2.1.2)
	}
	if original != (common.Hash{}) {
		if current == (common.Hash{}) { // recreate slot (2.2.1.1)
			sivm.StateDB.SubRefund(params.NetSstoreClearRefund)
		} else if value == (common.Hash{}) { // delete slot (2.2.1.2)
			sivm.StateDB.AddRefund(params.NetSstoreClearRefund)
		}
	}
	if original == value {
		if original == (common.Hash{}) { // reset to original inexistent slot (2.2.2.1)
			sivm.StateDB.AddRefund(params.NetSstoreResetClearRefund)
		} else { // reset to original existing slot (2.2.2.2)
			sivm.StateDB.AddRefund(params.NetSstoreResetRefund)
		}
	}
	return GasCosts{ExecutionGas: params.NetSstoreDirtyGas}, nil
}

// Here come the SIP2200 rules:
//
//	(0.) If *gasleft* is less than or equal to 2300, fail the current call.
//	(1.) If current value equals new value (this is a no-op), SLOAD_GAS is deducted.
//	(2.) If current value does not equal new value:
//		(2.1.) If original value equals current value (this storage slot has not been changed by the current execution context):
//			(2.1.1.) If original value is 0, SSTORE_SET_GAS (20K) gas is deducted.
//			(2.1.2.) Otherwise, SSTORE_RESET_GAS gas is deducted. If new value is 0, add SSTORE_CLEARS_SCHEDULE to refund counter.
//		(2.2.) If original value does not equal current value (this storage slot is dirty), SLOAD_GAS gas is deducted. Apply both of the following clauses:
//			(2.2.1.) If original value is not 0:
//				(2.2.1.1.) If current value is 0 (also means that new value is not 0), subtract SSTORE_CLEARS_SCHEDULE gas from refund counter.
//				(2.2.1.2.) If new value is 0 (also means that current value is not 0), add SSTORE_CLEARS_SCHEDULE gas to refund counter.
//			(2.2.2.) If original value equals new value (this storage slot is reset):
//				(2.2.2.1.) If original value is 0, add SSTORE_SET_GAS - SLOAD_GAS to refund counter.
//				(2.2.2.2.) Otherwise, add SSTORE_RESET_GAS - SLOAD_GAS gas to refund counter.
func gasSStoreSIP2200(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (GasCosts, error) {
	if sivm.readOnly {
		return GasCosts{}, ErrWriteProtection
	}
	// If we fail the minimum gas availability invariant, fail (0)
	if contract.Gas.ExecutionGas <= params.SstoreSentryGasSIP2200 {
		return GasCosts{}, errors.New("not enough gas for reentrancy sentry")
	}
	// Gas sentry honoured, do the actual gas calculation based on the stored value
	var (
		y, x              = stack.back(1), stack.back(0)
		current, original = sivm.StateDB.GetStateAndCommittedState(contract.Address(), x.Bytes32())
	)
	value := common.Hash(y.Bytes32())

	if current == value { // noop (1)
		return GasCosts{ExecutionGas: params.SloadGasSIP2200}, nil
	}
	if original == current {
		if original == (common.Hash{}) { // create slot (2.1.1)
			return GasCosts{ExecutionGas: params.SstoreSetGasSIP2200}, nil
		}
		if value == (common.Hash{}) { // delete slot (2.1.2b)
			sivm.StateDB.AddRefund(params.SstoreClearsScheduleRefundSIP2200)
		}
		return GasCosts{ExecutionGas: params.SstoreResetGasSIP2200}, nil // write existing slot (2.1.2)
	}
	if original != (common.Hash{}) {
		if current == (common.Hash{}) { // recreate slot (2.2.1.1)
			sivm.StateDB.SubRefund(params.SstoreClearsScheduleRefundSIP2200)
		} else if value == (common.Hash{}) { // delete slot (2.2.1.2)
			sivm.StateDB.AddRefund(params.SstoreClearsScheduleRefundSIP2200)
		}
	}
	if original == value {
		if original == (common.Hash{}) { // reset to original inexistent slot (2.2.2.1)
			sivm.StateDB.AddRefund(params.SstoreSetGasSIP2200 - params.SloadGasSIP2200)
		} else { // reset to original existing slot (2.2.2.2)
			sivm.StateDB.AddRefund(params.SstoreResetGasSIP2200 - params.SloadGasSIP2200)
		}
	}
	return GasCosts{ExecutionGas: params.SloadGasSIP2200}, nil // dirty update (2.2)
}

func makeGasLog(n uint64) gasFunc {
	return func(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (GasCosts, error) {
		requestedSize, overflow := stack.back(1).Uint64WithOverflow()
		if overflow {
			return GasCosts{}, ErrGasUintOverflow
		}

		gas, err := memoryGasCost(mem, memorySize)
		if err != nil {
			return GasCosts{}, err
		}

		if gas, overflow = math.SafeAdd(gas, params.LogGas); overflow {
			return GasCosts{}, ErrGasUintOverflow
		}
		if gas, overflow = math.SafeAdd(gas, n*params.LogTopicGas); overflow {
			return GasCosts{}, ErrGasUintOverflow
		}

		var memorySizeGas uint64
		if memorySizeGas, overflow = math.SafeMul(requestedSize, params.LogDataGas); overflow {
			return GasCosts{}, ErrGasUintOverflow
		}
		if gas, overflow = math.SafeAdd(gas, memorySizeGas); overflow {
			return GasCosts{}, ErrGasUintOverflow
		}
		return GasCosts{ExecutionGas: gas}, nil
	}
}

func gasKeccak256(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (GasCosts, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return GasCosts{}, err
	}
	wordGas, overflow := stack.back(1).Uint64WithOverflow()
	if overflow {
		return GasCosts{}, ErrGasUintOverflow
	}
	if wordGas, overflow = math.SafeMul(toWordSize(wordGas), params.Keccak256WordGas); overflow {
		return GasCosts{}, ErrGasUintOverflow
	}
	if gas, overflow = math.SafeAdd(gas, wordGas); overflow {
		return GasCosts{}, ErrGasUintOverflow
	}
	return GasCosts{ExecutionGas: gas}, nil
}

// pureMemoryGascost is used by several operations, which aside from their
// static cost have a dynamic cost which is solely based on the memory
// expansion
func pureMemoryGascost(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (GasCosts, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return GasCosts{}, err
	}
	return GasCosts{ExecutionGas: gas}, nil
}

var (
	gasReturn  = pureMemoryGascost
	gasRevert  = pureMemoryGascost
	gasMLoad   = pureMemoryGascost
	gasMStore8 = pureMemoryGascost
	gasMStore  = pureMemoryGascost
)

func gasCreate(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (GasCosts, error) {
	if sivm.readOnly {
		return GasCosts{}, ErrWriteProtection
	}
	return pureMemoryGascost(sivm, contract, stack, mem, memorySize)
}

func gasCreate2(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (GasCosts, error) {
	if sivm.readOnly {
		return GasCosts{}, ErrWriteProtection
	}
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return GasCosts{}, err
	}
	wordGas, overflow := stack.back(2).Uint64WithOverflow()
	if overflow {
		return GasCosts{}, ErrGasUintOverflow
	}
	if wordGas, overflow = math.SafeMul(toWordSize(wordGas), params.Keccak256WordGas); overflow {
		return GasCosts{}, ErrGasUintOverflow
	}
	if gas, overflow = math.SafeAdd(gas, wordGas); overflow {
		return GasCosts{}, ErrGasUintOverflow
	}
	return GasCosts{ExecutionGas: gas}, nil
}

func gasCreateSip3860(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (GasCosts, error) {
	if sivm.readOnly {
		return GasCosts{}, ErrWriteProtection
	}
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return GasCosts{}, err
	}
	size, overflow := stack.back(2).Uint64WithOverflow()
	if overflow {
		return GasCosts{}, ErrGasUintOverflow
	}
	if err := CheckMaxInitCodeSize(&sivm.chainRules, size); err != nil {
		return GasCosts{}, err
	}
	// Since size <= the protocol-defined maximum initcode size limit, these multiplication cannot overflow
	moreGas := params.InitCodeWordGas * ((size + 31) / 32)
	if gas, overflow = math.SafeAdd(gas, moreGas); overflow {
		return GasCosts{}, ErrGasUintOverflow
	}
	return GasCosts{ExecutionGas: gas}, nil
}

func gasCreate2Sip3860(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (GasCosts, error) {
	if sivm.readOnly {
		return GasCosts{}, ErrWriteProtection
	}
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return GasCosts{}, err
	}
	size, overflow := stack.back(2).Uint64WithOverflow()
	if overflow {
		return GasCosts{}, ErrGasUintOverflow
	}
	if err := CheckMaxInitCodeSize(&sivm.chainRules, size); err != nil {
		return GasCosts{}, err
	}
	// Since size <= the protocol-defined maximum initcode size limit, these multiplication cannot overflow
	moreGas := (params.InitCodeWordGas + params.Keccak256WordGas) * ((size + 31) / 32)
	if gas, overflow = math.SafeAdd(gas, moreGas); overflow {
		return GasCosts{}, ErrGasUintOverflow
	}
	return GasCosts{ExecutionGas: gas}, nil
}

func gasExpFrontier(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (GasCosts, error) {
	expByteLen := uint64((stack.back(1).BitLen() + 7) / 8)

	var (
		gas      = expByteLen * params.ExpByteFrontier // no overflow check required. Max is 256 * ExpByte gas
		overflow bool
	)
	if gas, overflow = math.SafeAdd(gas, params.ExpGas); overflow {
		return GasCosts{}, ErrGasUintOverflow
	}
	return GasCosts{ExecutionGas: gas}, nil
}

func gasExpSIP158(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (GasCosts, error) {
	expByteLen := uint64((stack.back(1).BitLen() + 7) / 8)

	var (
		gas      = expByteLen * params.ExpByteSIP158 // no overflow check required. Max is 256 * ExpByte gas
		overflow bool
	)
	if gas, overflow = math.SafeAdd(gas, params.ExpGas); overflow {
		return GasCosts{}, ErrGasUintOverflow
	}
	return GasCosts{ExecutionGas: gas}, nil
}

var (
	gasCall         = makeCallVariantGasCost(gasCallIntrinsic)
	gasCallCode     = makeCallVariantGasCost(gasCallCodeIntrinsic)
	gasDelegateCall = makeCallVariantGasCost(gasDelegateCallIntrinsic)
	gasStaticCall   = makeCallVariantGasCost(gasStaticCallIntrinsic)
)

func makeCallVariantGasCost(intrinsicFunc intrinsicGasFunc) gasFunc {
	return func(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (GasCosts, error) {
		intrinsic, err := intrinsicFunc(sivm, contract, stack, mem, memorySize)
		if err != nil {
			return GasCosts{}, err
		}
		sivm.callGasTemp, err = callGas(sivm.chainRules.IsSIP150, contract.Gas.ExecutionGas, intrinsic, stack.back(0))
		if err != nil {
			return GasCosts{}, err
		}
		gas, overflow := math.SafeAdd(intrinsic, sivm.callGasTemp)
		if overflow {
			return GasCosts{}, ErrGasUintOverflow
		}
		return GasCosts{ExecutionGas: gas}, nil
	}
}

func gasCallIntrinsic(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var (
		gas            uint64
		transfersValue = !stack.back(2).IsZero()
		address        = common.Address(stack.back(1).Bytes20())
	)
	if sivm.readOnly && transfersValue {
		return 0, ErrWriteProtection
	}
	// Stateless check
	memoryGas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	var transferGas uint64
	if transfersValue && !sivm.chainRules.IsSIP4762 {
		transferGas = params.CallValueTransferGas
	}
	var overflow bool
	if gas, overflow = math.SafeAdd(memoryGas, transferGas); overflow {
		return 0, ErrGasUintOverflow
	}
	// Terminate the gas measurement if the leftover gas is not sufficient,
	// it can effectively prevent accessing the states in the following steps.
	if contract.Gas.ExecutionGas < gas {
		return 0, ErrOutOfGas
	}
	// Stateful check
	var stateGas uint64
	if sivm.chainRules.IsSIP158 {
		if transfersValue && sivm.StateDB.Empty(address) {
			stateGas += params.CallNewAccountGas
		}
	} else if !sivm.StateDB.Exist(address) {
		stateGas += params.CallNewAccountGas
	}
	if gas, overflow = math.SafeAdd(gas, stateGas); overflow {
		return 0, ErrGasUintOverflow
	}
	return gas, nil
}

func gasCallCodeIntrinsic(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	memoryGas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	var (
		gas      uint64
		overflow bool
	)
	if stack.back(2).Sign() != 0 && !sivm.chainRules.IsSIP4762 {
		gas += params.CallValueTransferGas
	}
	if gas, overflow = math.SafeAdd(gas, memoryGas); overflow {
		return 0, ErrGasUintOverflow
	}
	return gas, nil
}

// gasCallCodeIntrinsic8038 mirrors gasCallCodeIntrinsic but charges the
// re-priced CALL_VALUE (ACCOUNT_WRITE + CALL_STIPEND) on value transfers per
// SIP-8038. CALLCODE executes in the caller's context, so it never creates a
// new account and has no state-gas component.
func gasCallCodeIntrinsic8038(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	memoryGas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	var (
		gas      uint64
		overflow bool
	)
	if stack.back(2).Sign() != 0 && !sivm.chainRules.IsSIP4762 {
		gas += params.CallValueTransferSilaAmsterdam
	}
	if gas, overflow = math.SafeAdd(gas, memoryGas); overflow {
		return 0, ErrGasUintOverflow
	}
	return gas, nil
}

func gasDelegateCallIntrinsic(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	return gas, nil
}

func gasStaticCallIntrinsic(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	return gas, nil
}

func gasSelfdestruct(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (GasCosts, error) {
	if sivm.readOnly {
		return GasCosts{}, ErrWriteProtection
	}
	var gas uint64
	// SIP150 homestead gas reprice fork:
	if sivm.chainRules.IsSIP150 {
		gas = params.SelfdestructGasSIP150
		var address = common.Address(stack.back(0).Bytes20())
		if sivm.chainRules.IsSIP158 {
			// if empty and transfers value
			if sivm.StateDB.Empty(address) && sivm.StateDB.GetBalance(contract.Address()).Sign() != 0 {
				gas += params.CreateBySelfdestructGas
			}
		} else if !sivm.StateDB.Exist(address) {
			gas += params.CreateBySelfdestructGas
		}
	}

	if !sivm.StateDB.HasSelfDestructed(contract.Address()) {
		sivm.StateDB.AddRefund(params.SelfdestructRefundGas)
	}
	return GasCosts{ExecutionGas: gas}, nil
}

func gasCreateSip8037(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (GasCosts, error) {
	if sivm.readOnly {
		return GasCosts{}, ErrWriteProtection
	}
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return GasCosts{}, err
	}
	size, overflow := stack.back(2).Uint64WithOverflow()
	if overflow {
		return GasCosts{}, ErrGasUintOverflow
	}
	if err := CheckMaxInitCodeSize(&sivm.chainRules, size); err != nil {
		return GasCosts{}, err
	}
	// Since size <= MaxInitCodeSizeSilaAmsterdam, these multiplications cannot overflow
	words := (size + 31) / 32
	wordGas := params.InitCodeWordGas * words

	// The account-creation state gas is not part of the opcode cost: it is
	// charged conditionally at the destination access, in the creating frame,
	// right before the 63/64ths split (see opCreate).
	return GasCosts{ExecutionGas: gas + wordGas}, nil
}

func gasCreate2Sip8037(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (GasCosts, error) {
	if sivm.readOnly {
		return GasCosts{}, ErrWriteProtection
	}
	gas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return GasCosts{}, err
	}
	size, overflow := stack.back(2).Uint64WithOverflow()
	if overflow {
		return GasCosts{}, ErrGasUintOverflow
	}
	if err := CheckMaxInitCodeSize(&sivm.chainRules, size); err != nil {
		return GasCosts{}, err
	}
	// Since size <= MaxInitCodeSizeSilaAmsterdam, these multiplications cannot overflow
	words := (size + 31) / 32

	// CREATE2 charges both InitCodeWordGas (SIP-3860) and Keccak256WordGas
	// (for address hashing).
	wordGas := (params.InitCodeWordGas + params.Keccak256WordGas) * words

	// The account-creation state gas is not part of the opcode cost: it is
	// charged conditionally at the destination access, in the creating frame,
	// right before the 63/64ths split (see opCreate2).
	return GasCosts{ExecutionGas: gas + wordGas}, nil
}

// executionGasCall8038 is the intrinsic execution-gas calculator for CALL in
// SilaAmsterdam. It computes memory expansion plus the re-priced CALL_VALUE
// (ACCOUNT_WRITE + CALL_STIPEND) on value transfers, but excludes new account
// creation, which is handled as state gas by stateGasCall8037.
func executionGasCall8038(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	var (
		gas            uint64
		transfersValue = !stack.back(2).IsZero()
	)
	if sivm.readOnly && transfersValue {
		return 0, ErrWriteProtection
	}
	memoryGas, err := memoryGasCost(mem, memorySize)
	if err != nil {
		return 0, err
	}
	var transferGas uint64
	if transfersValue && !sivm.chainRules.IsSIP4762 {
		transferGas = params.CallValueTransferSilaAmsterdam
	}
	var overflow bool
	if gas, overflow = math.SafeAdd(memoryGas, transferGas); overflow {
		return 0, ErrGasUintOverflow
	}
	return gas, nil
}

// stateGasCall8037 is the stateful gas calculator for CALL in SilaAmsterdam (SIP-8037).
// It only returns the state-dependent gas (account creation as state gas).
// Memory gas, transfer gas, and callGas are handled by gasCallStateless and
// makeCallVariantGasCall.
func stateGasCall8037(sivm *Sivm, contract *Contract, stack *Stack) (uint64, error) {
	var (
		gas            uint64
		transfersValue = !stack.back(2).IsZero()
		address        = common.Address(stack.back(1).Bytes20())
	)
	// Important: use StateDB.Empty instead of !StateDB.Exist. An account may exist
	// in the current state yet still be considered non-existent by SIP-161 if its
	// nonce, balance, and code are all zero. Such accounts can appear temporarily
	// during execution (e.g. via SELFDESTRUCT) and are removed at tx end.
	//
	// Funding such an account makes it permanent state growth and must be charged.
	if transfersValue && sivm.StateDB.Empty(address) {
		gas += params.AccountCreationSize * sivm.Context.CostPerStateByte
	}
	return gas, nil
}

// gasSelfdestruct8037And8038 implements the SELFDESTRUCT gas charging under
// the SIP8037 and SIP-8038.
func gasSelfdestruct8037And8038(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (GasCosts, error) {
	if sivm.readOnly {
		return GasCosts{}, ErrWriteProtection
	}
	var (
		gas     GasCosts
		address = common.Address(stack.peek().Bytes20())
	)
	if !sivm.StateDB.AddressInAccessList(address) {
		// If the caller cannot afford the cost, this change will be rolled back.
		sivm.StateDB.AddAddressToAccessList(address)
		gas.ExecutionGas = params.ColdAccountAccessSilaAmsterdam
	}
	// Check we have enough execution gas before we add the address to the BAL.
	if contract.Gas.ExecutionGas < gas.ExecutionGas {
		return gas, ErrOutOfGas
	}
	// Important: use StateDB.Empty instead of !StateDB.Exist. An account may exist
	// in the current state yet still be considered non-existent by SIP-161 if its
	// nonce, balance, and code are all zero. Such accounts can appear temporarily
	// during execution (e.g. via SELFDESTRUCT) and are removed at tx end.
	//
	// Funding such an account makes it permanent state growth and must be charged.
	if sivm.StateDB.Empty(address) && sivm.StateDB.GetBalance(contract.Address()).Sign() != 0 {
		gas.ExecutionGas += params.AccountWriteSilaAmsterdam
		gas.StateGas += params.AccountCreationSize * sivm.Context.CostPerStateByte
	}
	return gas, nil
}

// gasSStore8037And8038 implements the SSTORE gas charging under SIP-8037 and
// SIP-8038.
func gasSStore8037And8038(sivm *Sivm, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (GasCosts, error) {
	if sivm.readOnly {
		return GasCosts{}, ErrWriteProtection
	}
	// If we fail the minimum gas availability invariant, fail (0).
	if contract.Gas.ExecutionGas <= params.SstoreSentryGasSIP2200 {
		return GasCosts{}, errors.New("not enough gas for reentrancy sentry")
	}
	var (
		y, x     = stack.back(1), stack.peek()
		slot     = common.Hash(x.Bytes32())
		stateSet = params.StorageCreationSize * sivm.Context.CostPerStateByte
	)
	// Check slot presence in the access list
	access := params.WarmStorageAccessSilaAmsterdam
	_, slotPresent := sivm.StateDB.SlotInAccessList(contract.Address(), slot)
	if !slotPresent {
		access = params.ColdStorageAccessSilaAmsterdam
	}
	// Check access cost affordability before reading slot
	if contract.Gas.ExecutionGas < access {
		return GasCosts{}, errors.New("not enough gas for slot access")
	}
	if !slotPresent {
		sivm.StateDB.AddSlotToAccessList(contract.Address(), slot)
	}
	// Read the slot value for gas cost measurement
	var (
		value             = common.Hash(y.Bytes32())
		current, original = sivm.StateDB.GetStateAndCommittedState(contract.Address(), slot)
	)
	if current == value { // noop (1)
		return GasCosts{ExecutionGas: access}, nil
	}
	if original == current { // first change of the slot (2.1)
		if original == (common.Hash{}) { // create slot (2.1.1)
			return GasCosts{
				ExecutionGas: access + params.StorageWriteSilaAmsterdam,
				StateGas:     stateSet,
			}, nil
		}
		if value == (common.Hash{}) { // delete slot (2.1.2b)
			sivm.StateDB.AddRefund(params.StorageClearRefundSilaAmsterdam)
		}
		return GasCosts{ExecutionGas: access + params.StorageWriteSilaAmsterdam}, nil // write existing slot (2.1.2)
	}
	if original != (common.Hash{}) {
		if current == (common.Hash{}) { // recreate slot (2.2.1.1)
			sivm.StateDB.SubRefund(params.StorageClearRefundSilaAmsterdam)
		} else if value == (common.Hash{}) { // delete slot (2.2.1.2)
			sivm.StateDB.AddRefund(params.StorageClearRefundSilaAmsterdam)
		}
	}
	if original == value { // reset to original value (2.2.2)
		if original == (common.Hash{}) { // reset to original inexistent slot (2.2.2.1)
			contract.Gas.RefundState(stateSet)
		}
		sivm.StateDB.AddRefund(params.StorageWriteSilaAmsterdam)
	}
	return GasCosts{ExecutionGas: access}, nil // dirty update (2.2)
}
