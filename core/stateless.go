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

package core

import (
	"context"

	"github.com/sila-chain/go-sila/common"
	"github.com/sila-chain/go-sila/common/lru"
	"github.com/sila-chain/go-sila/consensus/beacon"
	"github.com/sila-chain/go-sila/consensus/silash"
	"github.com/sila-chain/go-sila/core/state"
	"github.com/sila-chain/go-sila/core/stateless"
	"github.com/sila-chain/go-sila/core/types"
	"github.com/sila-chain/go-sila/core/vm"
	"github.com/sila-chain/go-sila/log"
	"github.com/sila-chain/go-sila/params"
	"github.com/sila-chain/go-sila/trie"
	"github.com/sila-chain/go-sila/triedb"
)

// ExecuteStateless runs a stateless execution based on a witness, verifies
// everything it can locally and returns the state root and receipt root, that
// need the other side to explicitly check.
//
// This method is a bit of a sore thumb here, but:
//   - It cannot be placed in core/stateless, because state.New produces a circular dep
//   - It cannot be placed outside of core, because it needs to construct a dud headerchain
//
// TODO(karalabe): Would be nice to resolve both issues above somehow and move it.
func ExecuteStateless(ctx context.Context, config *params.ChainConfig, vmconfig vm.Config, block *types.Block, witness *stateless.Witness) (common.Hash, common.Hash, error) {
	// Sanity check if the supplied block accidentally contains a set root or
	// receipt hash. If so, be very loud, but still continue.
	if block.Root() != (common.Hash{}) {
		log.Error("stateless runner received state root it's expected to calculate (faulty consensus client)", "block", block.Number())
	}
	if block.ReceiptHash() != (common.Hash{}) {
		log.Error("stateless runner received receipt root it's expected to calculate (faulty consensus client)", "block", block.Number())
	}
	// Create and populate the state database to serve as the stateless backend
	memdb := witness.MakeHashDB()
	db, err := state.New(witness.Root(), state.NewDatabase(triedb.NewDatabase(memdb, triedb.HashDefaults), state.NewCodeDB(memdb)))
	if err != nil {
		return common.Hash{}, common.Hash{}, err
	}
	// Create a blockchain that is idle, but can be used to access headers through
	chain := &HeaderChain{
		config:      config,
		chainDb:     memdb,
		headerCache: lru.NewCache[common.Hash, *types.Header](256),
		engine:      beacon.New(silash.NewFaker()),
	}
	processor := NewStateProcessor(chain)
	validator := NewBlockValidator(config, nil) // No chain, we only validate the state, not the block

	// Run the stateless blocks processing and self-validate certain fields
	res, err := processor.Process(ctx, block, db, nil, vmconfig, nil)
	if err != nil {
		return common.Hash{}, common.Hash{}, err
	}
	if err = validator.ValidateState(block, db, res, true); err != nil {
		return common.Hash{}, common.Hash{}, err
	}
	// Almost everything validated, but receipt and state root needs to be returned
	receiptRoot := types.DeriveSha(res.Receipts, trie.NewStackTrie(nil))
	stateRoot := db.IntermediateRoot(config.IsSIP158(block.Number()))
	return stateRoot, receiptRoot, nil
}
