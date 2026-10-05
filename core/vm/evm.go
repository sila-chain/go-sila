// Copyright 2014 The go-sila Authors
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
	"math/big"
	"sync/atomic"

	"github.com/holiman/uint256"
	"github.com/sila-chain/go-sila/common"
	"github.com/sila-chain/go-sila/core/state"
	"github.com/sila-chain/go-sila/core/tracing"
	"github.com/sila-chain/go-sila/core/types"
	"github.com/sila-chain/go-sila/crypto"
	"github.com/sila-chain/go-sila/log"
	"github.com/sila-chain/go-sila/params"
)

type (
	// CanTransferFunc is the signature of a transfer guard function
	CanTransferFunc func(StateDB, common.Address, *uint256.Int) bool
	// TransferFunc is the signature of a transfer function
	TransferFunc func(StateDB, common.Address, common.Address, *uint256.Int, *params.Rules)
	// GetHashFunc returns the n'th block hash in the blockchain
	// and is used by the BLOCKHASH Sivm op code.
	GetHashFunc func(uint64) common.Hash
)

func (sivm *Sivm) precompile(addr common.Address) (PrecompiledContract, bool) {
	p, ok := sivm.precompiles[addr]
	return p, ok
}

// BlockContext provides the Sivm with auxiliary information. Once provided
// it shouldn't be modified.
type BlockContext struct {
	// CanTransfer returns whether the account contains
	// sufficient sila to transfer the value
	CanTransfer CanTransferFunc
	// Transfer transfers sila from one account to the other
	Transfer TransferFunc
	// GetHash returns the hash corresponding to n
	GetHash GetHashFunc

	// Block information
	Coinbase    common.Address // Provides information for COINBASE
	GasLimit    uint64         // Provides information for GASLIMIT
	BlockNumber *big.Int       // Provides information for NUMBER
	Time        uint64         // Provides information for TIME
	Difficulty  *big.Int       // Provides information for DIFFICULTY
	BaseFee     *big.Int       // Provides information for BASEFEE (0 if vm runs with NoBaseFee flag and 0 gas price)
	BlobBaseFee *big.Int       // Provides information for BLOBBASEFEE (0 if vm runs with NoBaseFee flag and 0 blob gas price)
	Random      *common.Hash   // Provides information for PREVRANDAO
	SlotNum     uint64         // Provides information for SLOTNUM

	CostPerStateByte uint64 // CostPerByte for new state after SIP-8037
}

// TxContext provides the Sivm with information about a transaction.
// All fields can change between transactions.
type TxContext struct {
	// Message information
	Origin       common.Address      // Provides information for ORIGIN
	GasPrice     *uint256.Int        // Provides information for GASPRICE (and is used to zero the basefee if NoBaseFee is set)
	BlobHashes   []common.Hash       // Provides information for BLOBHASH
	AccessEvents *state.AccessEvents // Capture all state accesses for this tx
}

// Sivm is the Sila Virtual Machine base object and provides
// the necessary tools to run a contract on the given state with
// the provided context. It should be noted that any error
// generated through any of the calls should be considered a
// revert-state-and-consume-all-gas operation, no checks on
// specific errors should ever be performed. The interpreter makes
// sure that any errors generated are to be considered faulty code.
//
// The Sivm should never be reused and is not thread safe.
type Sivm struct {
	// Context provides auxiliary blockchain related information
	Context BlockContext
	TxContext

	// StateDB gives access to the underlying state
	StateDB StateDB

	// table holds the opcode specific handlers
	table *JumpTable

	// depth is the current call stack
	depth int

	// chainConfig contains information about the current chain
	chainConfig *params.ChainConfig

	// chain rules contains the chain rules for the current epoch
	chainRules params.Rules

	// virtual machine configuration options used to initialise the sivm
	Config Config

	// abort is used to abort the Sivm calling operations
	abort atomic.Bool

	// callGasTemp holds the gas available for the current call. This is needed because the
	// available gas is calculated in gasCall* according to the 63/64 rule and later
	// applied in opCall*.
	callGasTemp uint64

	// precompiles holds the precompiled contracts for the current epoch
	precompiles map[common.Address]PrecompiledContract

	// jumpDests stores results of JUMPDEST analysis.
	jumpDests JumpDestCache

	// precompileCache stores outputs of pure precompile runs, may be nil.
	precompileCache *PrecompileCache

	readOnly   bool   // Whether to throw on stateful modifications
	returnData []byte // Last CALL's return data for subsequent reuse

	arena *stackArena
}

// NewSivm constructs an Sivm instance with the supplied block context, state
// database and several configs. It meant to be used throughout the entire
// state transition of a block, with the transaction context switched as
// needed by calling sivm.SetTxContext.
func NewSivm(blockCtx BlockContext, statedb StateDB, chainConfig *params.ChainConfig, config Config) *Sivm {
	sivm := &Sivm{
		Context:     blockCtx,
		StateDB:     statedb,
		Config:      config,
		chainConfig: chainConfig,
		chainRules:  chainConfig.Rules(blockCtx.BlockNumber, blockCtx.Random != nil, blockCtx.Time),
		jumpDests:   newMapJumpDests(),
		arena:       newArena(),
	}
	sivm.precompiles = *activePrecompiledContracts(sivm.chainRules)

	switch {
	case sivm.chainRules.IsBogota:
		sivm.table = &bogotaInstructionSet
	case sivm.chainRules.IsSilaAmsterdam:
		sivm.table = &silaAmsterdamInstructionSet
	case sivm.chainRules.IsSilaOsaka:
		sivm.table = &osakaInstructionSet
	case sivm.chainRules.IsUBT:
		// TODO replace with proper instruction set when fork is specified
		sivm.table = &verkleInstructionSet
	case sivm.chainRules.IsSilaPrague:
		sivm.table = &pragueInstructionSet
	case sivm.chainRules.IsSilaCancun:
		sivm.table = &cancunInstructionSet
	case sivm.chainRules.IsSilaShanghai:
		sivm.table = &shanghaiInstructionSet
	case sivm.chainRules.IsMerge:
		sivm.table = &mergeInstructionSet
	case sivm.chainRules.IsSilaLondon:
		sivm.table = &londonInstructionSet
	case sivm.chainRules.IsSilaBerlin:
		sivm.table = &berlinInstructionSet
	case sivm.chainRules.IsSilaIstanbul:
		sivm.table = &istanbulInstructionSet
	case sivm.chainRules.IsSilaConstantinople:
		sivm.table = &constantinopleInstructionSet
	case sivm.chainRules.IsSilaByzantium:
		sivm.table = &byzantiumInstructionSet
	case sivm.chainRules.IsSIP158:
		sivm.table = &spuriousDragonInstructionSet
	case sivm.chainRules.IsSIP150:
		sivm.table = &tangerineWhistleInstructionSet
	case sivm.chainRules.IsSilaHomestead:
		sivm.table = &homesteadInstructionSet
	default:
		sivm.table = &frontierInstructionSet
	}
	var extraSips []int
	if len(sivm.Config.ExtraSips) > 0 {
		// Deep-copy jumptable to prevent modification of opcodes in other tables
		sivm.table = copyJumpTable(sivm.table)
	}
	for _, sip := range sivm.Config.ExtraSips {
		if err := EnableSIP(sip, sivm.table); err != nil {
			// Disable it, so caller can check if it's activated or not
			log.Error("SIP activation failed", "sip", sip, "error", err)
		} else {
			extraSips = append(extraSips, sip)
		}
	}
	sivm.Config.ExtraSips = extraSips
	return sivm
}

// SetPrecompiles sets the precompiled contracts for the Sivm.
// This method is only used through RPC calls.
// It is not thread-safe.
func (sivm *Sivm) SetPrecompiles(precompiles PrecompiledContracts) {
	sivm.precompiles = precompiles
	// Overridden precompiles no longer match the address keyed result cache.
	sivm.precompileCache = nil
}

// SetJumpDestCache configures the analysis cache.
func (sivm *Sivm) SetJumpDestCache(jumpDests JumpDestCache) {
	sivm.jumpDests = jumpDests
}

// SetPrecompileCache configures the precompile result cache.
func (sivm *Sivm) SetPrecompileCache(cache *PrecompileCache) {
	sivm.precompileCache = cache
}

// SetStateDB configures the state for interaction.
func (sivm *Sivm) SetStateDB(statedb *state.StateDB) {
	sivm.StateDB = statedb
}

// SetTxContext resets the Sivm with a new transaction context.
// This is not threadsafe and should only be done very cautiously.
func (sivm *Sivm) SetTxContext(txCtx TxContext) {
	if sivm.chainRules.IsSIP4762 {
		txCtx.AccessEvents = state.NewAccessEvents()
	}
	sivm.TxContext = txCtx
}

// Cancel cancels any running Sivm operation. This may be called concurrently and
// it's safe to be called multiple times.
func (sivm *Sivm) Cancel() {
	sivm.abort.Store(true)
}

// Release returns some memory allocated by the Sivm, should be called after the Sivm was used
// for the last time. Not necessary, but an improvement.
func (sivm *Sivm) Release() {
	returnStack(sivm.arena)
}

// Cancelled returns true if Cancel has been called
func (sivm *Sivm) Cancelled() bool {
	return sivm.abort.Load()
}

func isSystemCall(caller common.Address) bool {
	return caller == params.SystemAddress
}

// Call executes the contract associated with the addr with the given input as
// parameters. It also handles any necessary value transfer required and takse
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
func (sivm *Sivm) Call(caller common.Address, addr common.Address, input []byte, gas GasBudget, value *uint256.Int) (ret []byte, result GasBudget, err error) {
	// Capture the tracer start/end events in debug mode
	if sivm.Config.Tracer != nil {
		sivm.captureBegin(sivm.depth, CALL, caller, addr, input, gas, value.ToBig())
		defer func(startGas GasBudget) {
			sivm.captureEnd(sivm.depth, startGas, result, ret, err)
		}(gas)
	}
	// Fail if we're trying to execute above the call depth limit
	if sivm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}
	syscall := isSystemCall(caller)

	// Fail if we're trying to transfer more than the available balance.
	if !syscall && !value.IsZero() && !sivm.Context.CanTransfer(sivm.StateDB, caller, value) {
		return nil, gas, ErrInsufficientBalance
	}
	snapshot := sivm.StateDB.Snapshot()
	p, isPrecompile := sivm.precompile(addr)
	if !sivm.StateDB.Exist(addr) {
		if !isPrecompile && sivm.chainRules.IsSIP4762 && !isSystemCall(caller) {
			// Add proof of absence to witness
			// At this point, the read costs have already been charged, either because this
			// is a direct tx call, in which case it's covered by the intrinsic gas, or because
			// of a CALL instruction, in which case BASIC_DATA has been added to the access
			// list in write mode. If there is enough gas paying for the addition of the code
			// hash leaf to the access list, then account creation will proceed unimpaired.
			// Thus, only pay for the creation of the code hash leaf here.
			wgas := sivm.AccessEvents.CodeHashGas(addr, true, gas.ExecutionGas, false)
			if _, ok := gas.ChargeExecution(wgas); !ok {
				sivm.StateDB.RevertToSnapshot(snapshot)
				return nil, gas.ExitHalt(), ErrOutOfGas
			}
		}

		if !isPrecompile && sivm.chainRules.IsSIP158 && value.IsZero() {
			// Calling a non-existing account, don't do anything.
			return nil, gas, nil
		}
		sivm.StateDB.CreateAccount(addr)
	}
	// Perform the value transfer only in non-syscall mode.
	// Calling this is required even for zero-value transfers,
	// to ensure the state clearing mechanism is applied.
	if !syscall {
		sivm.Context.Transfer(sivm.StateDB, caller, addr, value, &sivm.chainRules)
	}

	if isPrecompile {
		ret, gas, err = RunPrecompiledContract(sivm.StateDB, p, addr, input, gas, sivm.Config.Tracer, sivm.chainRules, sivm.precompileCache)
	} else {
		// Initialise a new contract and set the code that is to be used by the Sivm.
		code := sivm.resolveCode(addr)
		if len(code) == 0 {
			ret, err = nil, nil // gas is unchanged
		} else {
			// The contract is a scoped environment for this execution context only.
			contract := NewContract(caller, addr, value, gas, sivm.jumpDests)
			contract.IsSystemCall = isSystemCall(caller)
			contract.SetCallCode(sivm.resolveCodeHash(addr), code)
			ret, err = sivm.Run(contract, input, false)
			gas = contract.Gas
		}
	}

	// Calculate the remaining gas at the end of frame
	exitGas := gas.Exit(err)
	if err != nil {
		sivm.StateDB.RevertToSnapshot(snapshot)
		sivm.traceFrameExit(gas, exitGas, err)
	}
	return ret, exitGas, err
}

// CallCode executes the contract associated with the addr with the given input
// as parameters. It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
//
// CallCode differs from Call in the sense that it executes the given address'
// code with the caller as context.
func (sivm *Sivm) CallCode(caller common.Address, addr common.Address, input []byte, gas GasBudget, value *uint256.Int) (ret []byte, result GasBudget, err error) {
	// Invoke tracer hooks that signal entering/exiting a call frame
	if sivm.Config.Tracer != nil {
		sivm.captureBegin(sivm.depth, CALLCODE, caller, addr, input, gas, value.ToBig())
		defer func(startGas GasBudget) {
			sivm.captureEnd(sivm.depth, startGas, result, ret, err)
		}(gas)
	}
	// Fail if we're trying to execute above the call depth limit
	if sivm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}
	// Fail if we're trying to transfer more than the available balance
	if !sivm.Context.CanTransfer(sivm.StateDB, caller, value) {
		return nil, gas, ErrInsufficientBalance
	}
	snapshot := sivm.StateDB.Snapshot()

	// It is allowed to call precompiles, even via delegatecall
	if p, isPrecompile := sivm.precompile(addr); isPrecompile {
		ret, gas, err = RunPrecompiledContract(sivm.StateDB, p, addr, input, gas, sivm.Config.Tracer, sivm.chainRules, sivm.precompileCache)
	} else {
		// Initialise a new contract and set the code that is to be used by the Sivm.
		// The contract is a scoped environment for this execution context only.
		contract := NewContract(caller, caller, value, gas, sivm.jumpDests)
		contract.SetCallCode(sivm.resolveCodeHash(addr), sivm.resolveCode(addr))
		ret, err = sivm.Run(contract, input, false)
		gas = contract.Gas
	}

	// Calculate the remaining gas at the end of frame
	exitGas := gas.Exit(err)
	if err != nil {
		sivm.StateDB.RevertToSnapshot(snapshot)
		sivm.traceFrameExit(gas, exitGas, err)
	}
	return ret, exitGas, err
}

// DelegateCall executes the contract associated with the addr with the given input
// as parameters. It reverses the state in case of an execution error.
//
// DelegateCall differs from CallCode in the sense that it executes the given address'
// code with the caller as context and the caller is set to the caller of the caller.
func (sivm *Sivm) DelegateCall(originCaller common.Address, caller common.Address, addr common.Address, input []byte, gas GasBudget, value *uint256.Int) (ret []byte, result GasBudget, err error) {
	// Invoke tracer hooks that signal entering/exiting a call frame
	if sivm.Config.Tracer != nil {
		// DELEGATECALL inherits value from parent call
		sivm.captureBegin(sivm.depth, DELEGATECALL, caller, addr, input, gas, value.ToBig())
		defer func(startGas GasBudget) {
			sivm.captureEnd(sivm.depth, startGas, result, ret, err)
		}(gas)
	}
	// Fail if we're trying to execute above the call depth limit
	if sivm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}
	snapshot := sivm.StateDB.Snapshot()

	// It is allowed to call precompiles, even via delegatecall
	if p, isPrecompile := sivm.precompile(addr); isPrecompile {
		ret, gas, err = RunPrecompiledContract(sivm.StateDB, p, addr, input, gas, sivm.Config.Tracer, sivm.chainRules, sivm.precompileCache)
	} else {
		contract := NewContract(originCaller, caller, value, gas, sivm.jumpDests)
		contract.SetCallCode(sivm.resolveCodeHash(addr), sivm.resolveCode(addr))
		ret, err = sivm.Run(contract, input, false)
		gas = contract.Gas
	}

	// Calculate the remaining gas at the end of frame
	exitGas := gas.Exit(err)
	if err != nil {
		sivm.StateDB.RevertToSnapshot(snapshot)
		sivm.traceFrameExit(gas, exitGas, err)
	}
	return ret, exitGas, err
}

// StaticCall executes the contract associated with the addr with the given input
// as parameters while disallowing any modifications to the state during the call.
// Opcodes that attempt to perform such modifications will result in exceptions
// instead of performing the modifications.
func (sivm *Sivm) StaticCall(caller common.Address, addr common.Address, input []byte, gas GasBudget) (ret []byte, result GasBudget, err error) {
	// Invoke tracer hooks that signal entering/exiting a call frame
	if sivm.Config.Tracer != nil {
		sivm.captureBegin(sivm.depth, STATICCALL, caller, addr, input, gas, nil)
		defer func(startGas GasBudget) {
			sivm.captureEnd(sivm.depth, startGas, result, ret, err)
		}(gas)
	}
	// Fail if we're trying to execute above the call depth limit
	if sivm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}
	// We take a snapshot here. This is a bit counter-intuitive, and could probably be skipped.
	// However, even a staticcall is considered a 'touch'. On SilaMainnet, static calls were introduced
	// after all empty accounts were deleted, so this is not required. However, if we omit this,
	// then certain tests start failing; stRevertTest/RevertPrecompiledTouchExactOOG.json.
	// We could change this, but for now it's left for legacy reasons
	snapshot := sivm.StateDB.Snapshot()

	// We do an AddBalance of zero here, just in order to trigger a touch.
	// This doesn't matter on SilaMainnet, where all empties are gone at the time of SilaByzantium,
	// but is the correct thing to do and matters on other networks, in tests, and potential
	// future scenarios
	sivm.StateDB.AddBalance(addr, new(uint256.Int), tracing.BalanceChangeTouchAccount)

	if p, isPrecompile := sivm.precompile(addr); isPrecompile {
		ret, gas, err = RunPrecompiledContract(sivm.StateDB, p, addr, input, gas, sivm.Config.Tracer, sivm.chainRules, sivm.precompileCache)
	} else {
		contract := NewContract(caller, addr, new(uint256.Int), gas, sivm.jumpDests)
		contract.SetCallCode(sivm.resolveCodeHash(addr), sivm.resolveCode(addr))
		ret, err = sivm.Run(contract, input, true)
		gas = contract.Gas
	}

	// Calculate the remaining gas at the end of frame
	exitGas := gas.Exit(err)
	if err != nil {
		sivm.StateDB.RevertToSnapshot(snapshot)
		sivm.traceFrameExit(gas, exitGas, err)
	}
	return ret, exitGas, err
}

// traceFrameExit reports the budget change a failing frame applies on its way out:
// a halt burns the gas left, a revert refills the state-gas its rolled back state
// creations had paid for. Pre-SIP-8037 a revert moves nothing and stays silent.
func (sivm *Sivm) traceFrameExit(gas, exitGas GasBudget, err error) {
	if !sivm.Config.Tracer.HasGasHook() {
		return
	}
	if err != ErrExecutionReverted {
		sivm.Config.Tracer.EmitGasChange(gas.AsTracing(), exitGas.AsTracing(), tracing.GasChangeCallFailedExecution)
	} else if gas != exitGas {
		sivm.Config.Tracer.EmitGasChange(gas.AsTracing(), exitGas.AsTracing(), tracing.GasChangeRefundRevertedState)
	}
}

// createFramePreCheck the precondition before executing the contract deployment,
// halts the create frame if fails with any check below.
func (sivm *Sivm) createFramePreCheck(caller common.Address, value *uint256.Int) error {
	if sivm.depth > int(params.CallCreateDepth) {
		return ErrDepth
	}
	if !sivm.Context.CanTransfer(sivm.StateDB, caller, value) {
		return ErrInsufficientBalance
	}
	nonce := sivm.StateDB.GetNonce(caller)
	if nonce+1 < nonce {
		return ErrNonceUintOverflow
	}
	return nil
}

// chargeAccountCreation runs the create-frame precheck and charges the
// account-creation state gas since SilaAmsterdam, before the 63/64ths split.
//
// The charge only applies if the destination is empty, skipping pre-funded
// deployment destinations. Note, a destination colliding on storage alone
// (zero nonce, zero balance, empty code) is still empty and is charged.
//
// If halt is true, the caller must terminate with the returned error:
//   - a failed precheck halts the create frame only and parent frame continues,
//   - an insufficient charge halts the parent frame with ErrOutOfGas.
func (sivm *Sivm) chargeAccountCreation(scope *ScopeContext, contractAddr common.Address, value *uint256.Int) (charged, halt bool, err error) {
	if !sivm.chainRules.IsSilaAmsterdam {
		return false, false, nil
	}
	if err := sivm.createFramePreCheck(scope.Contract.Address(), value); err != nil {
		scope.Stack.get().Clear()
		sivm.returnData = nil
		return false, true, nil
	}
	if !sivm.StateDB.Empty(contractAddr) {
		return false, false, nil
	}
	cost := params.AccountCreationSize * sivm.Context.CostPerStateByte
	if !scope.Contract.chargeState(cost, sivm.Config.Tracer, tracing.GasChangeAccountCreation) {
		return false, true, ErrOutOfGas
	}
	return true, false, nil
}

// create creates a new contract using code as deployment code.
func (sivm *Sivm) create(caller common.Address, code []byte, gas GasBudget, value *uint256.Int, address common.Address, typ OpCode) (ret []byte, createAddress common.Address, result GasBudget, err error) {
	// Since SilaAmsterdam, the precheck has been folded into the parent frame
	// due to account-creation determination, so skip the duplicate check here.
	if !sivm.chainRules.IsSilaAmsterdam {
		err = sivm.createFramePreCheck(caller, value)
	}
	if sivm.Config.Tracer != nil {
		sivm.captureBegin(sivm.depth, typ, caller, address, code, gas, value.ToBig())
		defer func(startGas GasBudget) {
			sivm.captureEnd(sivm.depth, startGas, result, ret, err)
		}(gas)
	}
	if err != nil {
		return nil, common.Address{}, gas, err
	}
	// Increment the caller's nonce after passing all validations
	sivm.StateDB.SetNonce(caller, sivm.StateDB.GetNonce(caller)+1, tracing.NonceChangeContractCreator)

	// Charge the contract creation init gas in verkle mode
	if sivm.chainRules.IsSIP4762 {
		statelessGas := sivm.AccessEvents.ContractCreatePreCheckGas(address, gas.ExecutionGas)
		prior, ok := gas.Charge(GasCosts{ExecutionGas: statelessGas})
		if !ok {
			return nil, common.Address{}, gas.ExitHalt(), ErrOutOfGas
		}
		if sivm.Config.Tracer.HasGasHook() {
			sivm.Config.Tracer.EmitGasChange(prior.AsTracing(), gas.AsTracing(), tracing.GasChangeWitnessContractCollisionCheck)
		}
	}

	// We add this to the access list _before_ taking a snapshot. Even if the
	// creation fails, the access-list change should not be rolled back.
	if sivm.chainRules.IsSIP2929 {
		sivm.StateDB.AddAddressToAccessList(address)
	}
	// Ensure there's no existing contract already at the designated address.
	// Account is regarded as existent if either of these conditions is met:
	// - the nonce is non-zero
	// - the code is non-empty
	contractHash := sivm.StateDB.GetCodeHash(address)
	if sivm.StateDB.GetNonce(address) != 0 ||
		(contractHash != (common.Hash{}) && contractHash != types.EmptyCodeHash) { // non-empty code
		halt := gas.ExitHalt()
		if sivm.Config.Tracer.HasGasHook() {
			sivm.Config.Tracer.EmitGasChange(gas.AsTracing(), halt.AsTracing(), tracing.GasChangeCallFailedExecution)
		}
		// SIP-8037 collision rule: the state reservoir is fully preserved on
		// address collision while execution gas is burnt.
		return nil, common.Address{}, halt, ErrContractAddressCollision
	}
	// Create a new account on the state only if the object was not present.
	// It might be possible the contract code is deployed to a pre-existent
	// account with non-zero balance.
	snapshot := sivm.StateDB.Snapshot()
	if !sivm.StateDB.Exist(address) {
		sivm.StateDB.CreateAccount(address)
	}
	// CreateContract means that regardless of whether the account previously existed
	// in the state trie or not, it _now_ becomes created as a _contract_ account.
	// This is performed _prior_ to executing the initcode,  since the initcode
	// acts inside that account.
	sivm.StateDB.CreateContract(address)

	if sivm.chainRules.IsSIP158 {
		sivm.StateDB.SetNonce(address, 1, tracing.NonceChangeNewContract)
	}
	// Charge the contract creation init gas in verkle mode
	if sivm.chainRules.IsSIP4762 {
		consumed, wanted := sivm.AccessEvents.ContractCreateInitGas(address, gas.ExecutionGas)
		if consumed < wanted {
			return nil, common.Address{}, gas.ExitHalt(), ErrOutOfGas
		}
		prior, _ := gas.Charge(GasCosts{ExecutionGas: consumed})
		if sivm.Config.Tracer.HasGasHook() {
			sivm.Config.Tracer.EmitGasChange(prior.AsTracing(), gas.AsTracing(), tracing.GasChangeWitnessContractInit)
		}
	}
	sivm.Context.Transfer(sivm.StateDB, caller, address, value, &sivm.chainRules)

	// Initialise a new contract and set the code that is to be used by the Sivm.
	// The contract is a scoped environment for this execution context only.
	contract := NewContract(caller, address, value, gas, sivm.jumpDests)

	// Explicitly set the code to a null hash to prevent caching of jump analysis
	// for the initialization code.
	contract.SetCallCode(common.Hash{}, code)
	contract.IsDeployment = true

	ret, err = sivm.initNewContract(contract, address)

	// Special case: ErrCodeStoreOutOfGas pre-SilaHomestead does NOT roll back
	// state and gas is preserved (i.e., treated as success).
	if err != nil && (sivm.chainRules.IsSilaHomestead || err != ErrCodeStoreOutOfGas) {
		sivm.StateDB.RevertToSnapshot(snapshot)

		exit := contract.Gas.Exit(err)
		sivm.traceFrameExit(contract.Gas, exit, err)
		return ret, address, exit, err
	}
	// Either success, or pre-SilaHomestead ErrCodeStoreOutOfGas (gas preserved).
	// Both packaged as a success-form GasBudget.
	return ret, address, contract.Gas.ExitSuccess(), err
}

// initNewContract runs a new contract's creation code, performs checks on the
// resulting code that is to be deployed, and consumes necessary gas.
func (sivm *Sivm) initNewContract(contract *Contract, address common.Address) ([]byte, error) {
	ret, err := sivm.Run(contract, nil, false)
	if err != nil {
		return ret, err
	}
	// Check prefix before gas calculation.
	// Reject code starting with 0xEF if SIP-3541 is enabled.
	if len(ret) >= 1 && ret[0] == 0xEF && sivm.chainRules.IsSilaLondon {
		return ret, ErrInvalidCode
	}
	if sivm.chainRules.IsSIP4762 {
		consumed, wanted := sivm.AccessEvents.CodeChunksRangeGas(address, 0, uint64(len(ret)), uint64(len(ret)), true, contract.Gas.ExecutionGas)
		contract.chargeExecution(consumed, sivm.Config.Tracer, tracing.GasChangeWitnessCodeChunk)
		if len(ret) > 0 && (consumed < wanted) {
			return ret, ErrCodeStoreOutOfGas
		}
		if err := CheckMaxCodeSize(&sivm.chainRules, uint64(len(ret))); err != nil {
			return ret, err
		}
	} else if sivm.chainRules.IsSilaAmsterdam {
		// Check max code size BEFORE charging gas so over-max code
		// does not consume state gas (which would inflate tx_state).
		if err := CheckMaxCodeSize(&sivm.chainRules, uint64(len(ret))); err != nil {
			return ret, err
		}
		// Charge execution gas (hash cost) before state gas.
		executionCost := toWordSize(uint64(len(ret))) * params.Keccak256WordGas
		if !contract.chargeExecution(executionCost, sivm.Config.Tracer, tracing.GasChangeCallCodeStorage) {
			return ret, ErrCodeStoreOutOfGas
		}
		// Charge state gas (code-deposit) afterwards.
		stateCost := uint64(len(ret)) * sivm.Context.CostPerStateByte
		if !contract.chargeState(stateCost, sivm.Config.Tracer, tracing.GasChangeCallCodeStorage) {
			return ret, ErrCodeStoreOutOfGas
		}
	} else {
		createDataCost := uint64(len(ret)) * params.CreateDataGas
		if !contract.chargeExecution(createDataCost, sivm.Config.Tracer, tracing.GasChangeCallCodeStorage) {
			return ret, ErrCodeStoreOutOfGas
		}
		if err := CheckMaxCodeSize(&sivm.chainRules, uint64(len(ret))); err != nil {
			return ret, err
		}
	}
	if len(ret) > 0 {
		sivm.StateDB.SetCode(address, ret, tracing.CodeChangeContractCreation)
	}
	return ret, nil
}

// Create creates a new contract using code as deployment code.
func (sivm *Sivm) Create(caller common.Address, code []byte, gas GasBudget, value *uint256.Int) (ret []byte, contractAddr common.Address, result GasBudget, err error) {
	contractAddr = crypto.CreateAddress(caller, sivm.StateDB.GetNonce(caller))
	return sivm.create(caller, code, gas, value, contractAddr, CREATE)
}

// Create2 creates a new contract using code as deployment code.
//
// The different between Create2 with Create is Create2 uses keccak256(0xff ++ msg.sender ++ salt ++ keccak256(init_code))[12:]
// instead of the usual sender-and-nonce-hash as the address where the contract is initialized at.
func (sivm *Sivm) Create2(caller common.Address, code []byte, gas GasBudget, endowment *uint256.Int, salt *uint256.Int) (ret []byte, contractAddr common.Address, result GasBudget, err error) {
	inithash := crypto.Keccak256Hash(code)
	contractAddr = crypto.CreateAddress2(caller, salt.Bytes32(), inithash[:])
	return sivm.create(caller, code, gas, endowment, contractAddr, CREATE2)
}

// resolveCode returns the code associated with the provided account. After
// SilaPrague, it can also resolve code pointed to by a delegation designator.
func (sivm *Sivm) resolveCode(addr common.Address) []byte {
	code := sivm.StateDB.GetCode(addr)
	if !sivm.chainRules.IsSilaPrague {
		return code
	}
	if target, ok := types.ParseDelegation(code); ok {
		// Note we only follow one level of delegation.
		return sivm.StateDB.GetCode(target)
	}
	return code
}

// resolveCodeHash returns the code hash associated with the provided address.
// After SilaPrague, it can also resolve code hash of the account pointed to by a
// delegation designator. Although this is not accessible in the Sivm it is used
// internally to associate jumpdest analysis to code.
func (sivm *Sivm) resolveCodeHash(addr common.Address) common.Hash {
	if sivm.chainRules.IsSilaPrague {
		code := sivm.StateDB.GetCode(addr)
		if target, ok := types.ParseDelegation(code); ok {
			// Note we only follow one level of delegation.
			return sivm.StateDB.GetCodeHash(target)
		}
	}
	return sivm.StateDB.GetCodeHash(addr)
}

// ChainConfig returns the environment's chain configuration
func (sivm *Sivm) ChainConfig() *params.ChainConfig { return sivm.chainConfig }

func (sivm *Sivm) captureBegin(depth int, typ OpCode, from common.Address, to common.Address, input []byte, startGas GasBudget, value *big.Int) {
	tracer := sivm.Config.Tracer
	tracer.EmitEnter(depth, byte(typ), from, to, input, startGas.AsTracing(), value)
	if tracer.HasGasHook() {
		tracer.EmitGasChange(tracing.Gas{}, startGas.AsTracing(), tracing.GasChangeCallInitialBalance)
	}
}

func (sivm *Sivm) captureEnd(depth int, startGas GasBudget, leftOverGas GasBudget, ret []byte, err error) {
	tracer := sivm.Config.Tracer
	if !leftOverGas.IsZero() && tracer.HasGasHook() {
		tracer.EmitGasChange(leftOverGas.AsTracing(), tracing.Gas{}, tracing.GasChangeCallLeftOverReturned)
	}
	var reverted bool
	if err != nil {
		reverted = true
	}
	if !sivm.chainRules.IsSilaHomestead && errors.Is(err, ErrCodeStoreOutOfGas) {
		reverted = false
	}
	tracer.EmitExit(depth, ret, startGas.AsTracing(), leftOverGas.AsTracing(), VMErrorFromErr(err), reverted)
}

// GetVMContext provides context about the block being executed as well as state
// to the tracers.
func (sivm *Sivm) GetVMContext() *tracing.VMContext {
	return &tracing.VMContext{
		Coinbase:    sivm.Context.Coinbase,
		BlockNumber: sivm.Context.BlockNumber,
		Time:        sivm.Context.Time,
		Random:      sivm.Context.Random,
		BaseFee:     sivm.Context.BaseFee,
		StateDB:     sivm.StateDB,
	}
}

// GetRules returns the chain rules used throughout the Sivm execution.
func (sivm *Sivm) GetRules() params.Rules {
	return sivm.chainRules
}
