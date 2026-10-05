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

package core

import (
	"context"
	"math/big"
	"sync"
	"testing"

	"github.com/sila-chain/go-sila/common"
	"github.com/sila-chain/go-sila/consensus/beacon"
	"github.com/sila-chain/go-sila/consensus/silash"
	"github.com/sila-chain/go-sila/core/rawdb"
	"github.com/sila-chain/go-sila/core/tracing"
	"github.com/sila-chain/go-sila/core/types"
	"github.com/sila-chain/go-sila/core/vm"
	"github.com/sila-chain/go-sila/crypto"
	"github.com/sila-chain/go-sila/params"
)

// hookRecorder counts every live-tracer hook invocation, keyed by hook name.
type hookRecorder struct {
	mu     sync.Mutex
	counts map[string]int
}

func (r *hookRecorder) mark(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.counts == nil {
		r.counts = make(map[string]int)
	}
	r.counts[name]++
}

// reset drops every recorded hook, so that a subsequent census only reflects
// the invocations made after this point.
func (r *hookRecorder) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counts = nil
}

func (r *hookRecorder) total() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	var n int
	for _, c := range r.counts {
		n += c
	}
	return n
}

func (r *hookRecorder) fired() map[string]int {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]int, len(r.counts))
	for k, v := range r.counts {
		out[k] = v
	}
	return out
}

// hooks builds a live tracer covering the hooks a plain value-transfer block
// exercises, both at block scope and inside the Sivm.
func (r *hookRecorder) hooks() *tracing.Hooks {
	return &tracing.Hooks{
		OnBlockStart: func(tracing.BlockEvent) { r.mark("OnBlockStart") },
		OnBlockEnd:   func(error) { r.mark("OnBlockEnd") },
		OnTxStart: func(*tracing.VMContext, *types.Transaction, common.Address) {
			r.mark("OnTxStart")
		},
		OnTxEnd: func(*types.Receipt, error) { r.mark("OnTxEnd") },
		OnEnter: func(int, byte, common.Address, common.Address, []byte, uint64, *big.Int) {
			r.mark("OnEnter")
		},
		OnExit: func(int, []byte, uint64, error, bool) { r.mark("OnExit") },
		OnOpcode: func(uint64, byte, uint64, uint64, tracing.OpContext, []byte, int, error) {
			r.mark("OnOpcode")
		},
		OnBalanceChange: func(common.Address, *big.Int, *big.Int, tracing.BalanceChangeReason) {
			r.mark("OnBalanceChange")
		},
		OnStorageChange: func(common.Address, common.Hash, common.Hash, common.Hash) {
			r.mark("OnStorageChange")
		},
		OnLog: func(*types.Log) { r.mark("OnLog") },
	}
}

// tracedChain builds a chain whose node-wide vm.Config carries a hook-counting
// live tracer, as `sila --vmtrace` configures it, together with a single block
// carrying one value transfer, so that both the block-scope and the tx-scope
// hooks have something to report. The returned recorder starts empty; the
// hooks fired while the genesis is committed are discarded.
func tracedChain(t *testing.T, config *params.ChainConfig) (*BlockChain, *types.Block, common.Hash, *hookRecorder) {
	t.Helper()

	var (
		engine = beacon.New(silash.NewFaker())
		key, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		addr   = crypto.PubkeyToAddress(key.PublicKey)
		to     = common.HexToAddress("0x00000000000000000000000000000000deadbeef")
		funds  = new(big.Int).Mul(big.NewInt(1000), big.NewInt(params.Sila))
		signer = types.LatestSigner(config)
		gspec  = &Genesis{
			Config:  config,
			BaseFee: big.NewInt(params.InitialBaseFee),
			Alloc: types.GenesisAlloc{
				addr:                             {Balance: funds},
				params.BeaconRootsAddress:        {Nonce: 1, Code: params.BeaconRootsCode, Balance: common.Big0},
				params.HistoryStorageAddress:     {Nonce: 1, Code: params.HistoryStorageCode, Balance: common.Big0},
				params.WithdrawalQueueAddress:    {Nonce: 1, Code: params.WithdrawalQueueCode, Balance: common.Big0},
				params.ConsolidationQueueAddress: {Nonce: 1, Code: params.ConsolidationQueueCode, Balance: common.Big0},
				params.BuilderDepositAddress:     {Nonce: 1, Code: params.BuilderDepositCode, Balance: common.Big0},
				params.BuilderExitAddress:        {Nonce: 1, Code: params.BuilderExitCode, Balance: common.Big0},
			},
		}
	)
	_, blocks, _ := GenerateChainWithGenesis(gspec, engine, 1, func(i int, b *BlockGen) {
		tx, err := types.SignNewTx(key, signer, &types.DynamicFeeTx{
			ChainID:   config.ChainID,
			Nonce:     0,
			To:        &to,
			Gas:       100000,
			GasFeeCap: big.NewInt(params.InitialBaseFee * 2),
			GasTipCap: big.NewInt(1),
			Value:     big.NewInt(1000),
		})
		if err != nil {
			t.Fatalf("failed to sign tx: %v", err)
		}
		b.AddTx(tx)
	})
	if got := len(blocks[0].Transactions()); got != 1 {
		t.Fatalf("test block carries %d transactions, want 1", got)
	}
	recorder := new(hookRecorder)
	options := DefaultConfig().WithStateScheme(rawdb.HashScheme)
	options.VmConfig = vm.Config{Tracer: recorder.hooks()}

	chain, err := NewBlockChain(rawdb.NewMemoryDatabase(), gspec, engine, options)
	if err != nil {
		t.Fatalf("failed to create chain: %v", err)
	}
	t.Cleanup(chain.Stop)
	recorder.reset() // discard OnBlockchainInit/OnGenesisBlock

	return chain, blocks[0], gspec.ToBlock().Root(), recorder
}

// TestProcessBlockTracerOptIn asserts that ExecuteConfig.EnableTracer gates every
// tracing hook that block execution reaches through vm.Config: the
// OnBlockStart/OnBlockEnd envelope, the transaction and Sivm frame hooks, and the
// state hooks carried by the hooked StateDB.
func TestProcessBlockTracerOptIn(t *testing.T) {
	// mergedConfig is the pre-SilaAmsterdam chain the sequential processor runs.
	mergedConfig := func() *params.ChainConfig {
		config := *params.AllSilashProtocolChanges
		config.TerminalTotalDifficulty = common.Big0
		config.SilaShanghaiTime = u64(0)
		return &config
	}

	t.Run("opted out", func(t *testing.T) {
		// The exact config debug_executionWitness uses.
		chain, block, parentRoot, recorder := tracedChain(t, mergedConfig())
		_, err := chain.ProcessBlock(context.Background(), parentRoot, block, ExecuteConfig{
			WriteState:   false,
			EnableTracer: false,
			MakeWitness:  true,
		})
		if err != nil {
			t.Fatalf("failed to process block: %v", err)
		}
		if n := recorder.total(); n != 0 {
			t.Errorf("live tracer fired %d hook(s) with EnableTracer=false: %v", n, recorder.fired())
		}
	})

	t.Run("opted out without witness", func(t *testing.T) {
		// The opted-out combination debug_executionWitness does not reach: no
		// witness requested, on an SilaAmsterdam block carrying a block access list.
		// Detaching the tracer makes the execution eligible for the BAL-driven
		// parallel processor, which wires no tracing hooks at all, so the gate
		// must hold on that path too.
		config := *params.MergedTestChainConfig
		config.SilaAmsterdamTime = u64(0)

		chain, block, parentRoot, recorder := tracedChain(t, &config)
		if block.AccessList() == nil {
			t.Fatal("test block carries no access list, parallel execution unreachable")
		}
		_, err := chain.ProcessBlock(context.Background(), parentRoot, block, ExecuteConfig{
			WriteState:   false,
			EnableTracer: false,
			MakeWitness:  false,
		})
		if err != nil {
			t.Fatalf("failed to process block: %v", err)
		}
		if n := recorder.total(); n != 0 {
			t.Errorf("live tracer fired %d hook(s) with EnableTracer=false: %v", n, recorder.fired())
		}
	})

	t.Run("opted in", func(t *testing.T) {
		// Canonical import: the tracer must still see the whole block.
		chain, block, parentRoot, recorder := tracedChain(t, mergedConfig())
		_, err := chain.ProcessBlock(context.Background(), parentRoot, block, ExecuteConfig{
			WriteState:   true,
			WriteHead:    true,
			EnableTracer: true,
		})
		if err != nil {
			t.Fatalf("failed to process block: %v", err)
		}
		for _, hook := range []string{"OnBlockStart", "OnBlockEnd", "OnTxStart", "OnTxEnd", "OnEnter", "OnExit", "OnBalanceChange"} {
			if recorder.fired()[hook] == 0 {
				t.Errorf("%s did not fire with EnableTracer=true; fired: %v", hook, recorder.fired())
			}
		}
	})
}
