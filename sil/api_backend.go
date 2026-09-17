// Copyright 2015 The go-sila Authors
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

package sil

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/sila-chain/go-sila"
	"github.com/sila-chain/go-sila/accounts"
	"github.com/sila-chain/go-sila/common"
	"github.com/sila-chain/go-sila/consensus"
	"github.com/sila-chain/go-sila/consensus/misc/sip1559"
	"github.com/sila-chain/go-sila/consensus/misc/sip4844"
	"github.com/sila-chain/go-sila/core"
	"github.com/sila-chain/go-sila/core/filtermaps"
	"github.com/sila-chain/go-sila/core/history"
	"github.com/sila-chain/go-sila/core/rawdb"
	"github.com/sila-chain/go-sila/core/state"
	"github.com/sila-chain/go-sila/core/txpool"
	"github.com/sila-chain/go-sila/core/txpool/locals"
	"github.com/sila-chain/go-sila/core/types"
	"github.com/sila-chain/go-sila/core/vm"
	"github.com/sila-chain/go-sila/sil/gasprice"
	"github.com/sila-chain/go-sila/sil/tracers"
	"github.com/sila-chain/go-sila/sildb"
	"github.com/sila-chain/go-sila/event"
	"github.com/sila-chain/go-sila/internal/silapi"
	"github.com/sila-chain/go-sila/params"
	"github.com/sila-chain/go-sila/rpc"
)

// SilAPIBackend implements silapi.Backend and tracers.Backend for full nodes
type SilAPIBackend struct {
	extRPCEnabled       bool
	allowUnprotectedTxs bool
	sil                 *Sila
	gpo                 *gasprice.Oracle
}

// ChainConfig returns the active chain configuration.
func (b *SilAPIBackend) ChainConfig() *params.ChainConfig {
	return b.sil.blockchain.Config()
}

func (b *SilAPIBackend) CurrentBlock() *types.Header {
	return b.sil.blockchain.CurrentBlock()
}

func (b *SilAPIBackend) SetHead(number uint64) error {
	// Reject rewinding to a point before the snap-sync pivot. The earliest
	// recoverable state is the pivot block, so a target below it would simply
	// reset the chain to genesis, which is almost never the intent behind a
	// manual debug_setHead.
	if pivot := rawdb.ReadLastPivotNumber(b.sil.ChainDb()); pivot != nil && number < *pivot {
		return fmt.Errorf("rewind target %d is before the snap-sync pivot %d", number, *pivot)
	}
	// In path mode the deepest reachable state is bounded by the amount of state
	// histories retained. If the reverse diffs for the target have already been
	// pruned, its state is no longer recoverable.
	bc := b.sil.blockchain
	if bc.TrieDB().Scheme() == rawdb.PathScheme {
		if header := bc.GetHeaderByNumber(number); header != nil {
			if !bc.HasState(header.Root) && !bc.StateRecoverable(header.Root) {
				return errors.New("rewind target is not recoverable")
			}
		}
	}
	b.sil.handler.downloader.Cancel()
	return b.sil.blockchain.SetHead(number)
}

func (b *SilAPIBackend) HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if number == rpc.PendingBlockNumber {
		block, _, _ := b.sil.miner.Pending()
		if block == nil {
			return nil, errors.New("pending block is not available")
		}
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if number == rpc.LatestBlockNumber {
		return b.sil.blockchain.CurrentBlock(), nil
	}
	if number == rpc.FinalizedBlockNumber {
		block := b.sil.blockchain.CurrentFinalBlock()
		if block == nil {
			return nil, errors.New("finalized block not found")
		}
		return block, nil
	}
	if number == rpc.SafeBlockNumber {
		block := b.sil.blockchain.CurrentSafeBlock()
		if block == nil {
			return nil, errors.New("safe block not found")
		}
		return block, nil
	}
	var bn uint64
	if number == rpc.EarliestBlockNumber {
		bn = b.HistoryPruningCutoff()
	} else {
		bn = uint64(number)
	}
	return b.sil.blockchain.GetHeaderByNumber(bn), nil
}

func (b *SilAPIBackend) HeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*types.Header, error) {
	if blockNr, ok := blockNrOrHash.Number(); ok {
		return b.HeaderByNumber(ctx, blockNr)
	}
	if hash, ok := blockNrOrHash.Hash(); ok {
		header := b.sil.blockchain.GetHeaderByHash(hash)
		if header == nil {
			return nil, errors.New("header for hash not found")
		}
		if blockNrOrHash.RequireCanonical && b.sil.blockchain.GetCanonicalHash(header.Number.Uint64()) != hash {
			return nil, errors.New("hash is not currently canonical")
		}
		return header, nil
	}
	return nil, errors.New("invalid arguments; neither block nor hash specified")
}

func (b *SilAPIBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return b.sil.blockchain.GetHeaderByHash(hash), nil
}

func (b *SilAPIBackend) BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if number == rpc.PendingBlockNumber {
		block, _, _ := b.sil.miner.Pending()
		if block == nil {
			return nil, errors.New("pending block is not available")
		}
		return block, nil
	}
	// Otherwise resolve and return the block
	if number == rpc.LatestBlockNumber {
		header := b.sil.blockchain.CurrentBlock()
		return b.sil.blockchain.GetBlock(header.Hash(), header.Number.Uint64()), nil
	}
	if number == rpc.FinalizedBlockNumber {
		header := b.sil.blockchain.CurrentFinalBlock()
		if header == nil {
			return nil, errors.New("finalized block not found")
		}
		return b.sil.blockchain.GetBlock(header.Hash(), header.Number.Uint64()), nil
	}
	if number == rpc.SafeBlockNumber {
		header := b.sil.blockchain.CurrentSafeBlock()
		if header == nil {
			return nil, errors.New("safe block not found")
		}
		return b.sil.blockchain.GetBlock(header.Hash(), header.Number.Uint64()), nil
	}
	bn := uint64(number) // the resolved number
	if number == rpc.EarliestBlockNumber {
		bn = b.HistoryPruningCutoff()
	}
	block := b.sil.blockchain.GetBlockByNumber(bn)
	if block == nil && bn < b.HistoryPruningCutoff() {
		return nil, &history.PrunedHistoryError{}
	}
	return block, nil
}

func (b *SilAPIBackend) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	number := b.sil.blockchain.GetBlockNumber(hash)
	if number == nil {
		return nil, nil
	}
	block := b.sil.blockchain.GetBlock(hash, *number)
	if block == nil && *number < b.HistoryPruningCutoff() {
		return nil, &history.PrunedHistoryError{}
	}
	return block, nil
}

// GetBody returns body of a block. It does not resolve special block numbers.
func (b *SilAPIBackend) GetBody(ctx context.Context, hash common.Hash, number rpc.BlockNumber) (*types.Body, error) {
	if number < 0 || hash == (common.Hash{}) {
		return nil, errors.New("invalid arguments; expect hash and no special block numbers")
	}
	body := b.sil.blockchain.GetBody(hash)
	if body == nil {
		if uint64(number) < b.HistoryPruningCutoff() {
			return nil, &history.PrunedHistoryError{}
		}
		return nil, errors.New("block body not found")
	}
	return body, nil
}

func (b *SilAPIBackend) BlockByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*types.Block, error) {
	if blockNr, ok := blockNrOrHash.Number(); ok {
		return b.BlockByNumber(ctx, blockNr)
	}
	if hash, ok := blockNrOrHash.Hash(); ok {
		header := b.sil.blockchain.GetHeaderByHash(hash)
		if header == nil {
			// Return 'null' and no error if block is not found.
			// This behavior is required by RPC spec.
			return nil, nil
		}
		if blockNrOrHash.RequireCanonical && b.sil.blockchain.GetCanonicalHash(header.Number.Uint64()) != hash {
			return nil, errors.New("hash is not currently canonical")
		}
		block := b.sil.blockchain.GetBlock(hash, header.Number.Uint64())
		if block == nil {
			if header.Number.Uint64() < b.HistoryPruningCutoff() {
				return nil, &history.PrunedHistoryError{}
			}
			return nil, errors.New("header found, but block body is missing")
		}
		return block, nil
	}
	return nil, errors.New("invalid arguments; neither block nor hash specified")
}

func (b *SilAPIBackend) Pending() (*types.Block, types.Recsipts, *state.StateDB) {
	return b.sil.miner.Pending()
}

func (b *SilAPIBackend) StateAndHeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if number == rpc.PendingBlockNumber {
		block, _, state := b.sil.miner.Pending()
		if block == nil || state == nil {
			return nil, nil, errors.New("pending state is not available")
		}
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, number)
	if err != nil {
		return nil, nil, err
	}
	if header == nil {
		return nil, nil, errors.New("header not found")
	}
	stateDb, err := b.sil.BlockChain().StateAt(header)
	if err != nil {
		stateDb, err = b.sil.BlockChain().HistoricState(header)
		if err != nil {
			return nil, nil, err
		}
	}
	return stateDb, header, nil
}

func (b *SilAPIBackend) StateAndHeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*state.StateDB, *types.Header, error) {
	if blockNr, ok := blockNrOrHash.Number(); ok {
		return b.StateAndHeaderByNumber(ctx, blockNr)
	}
	if hash, ok := blockNrOrHash.Hash(); ok {
		header, err := b.HeaderByHash(ctx, hash)
		if err != nil {
			return nil, nil, err
		}
		if header == nil {
			return nil, nil, errors.New("header for hash not found")
		}
		if blockNrOrHash.RequireCanonical && b.sil.blockchain.GetCanonicalHash(header.Number.Uint64()) != hash {
			return nil, nil, errors.New("hash is not currently canonical")
		}
		stateDb, err := b.sil.BlockChain().StateAt(header)
		if err != nil {
			stateDb, err = b.sil.BlockChain().HistoricState(header)
			if err != nil {
				return nil, nil, err
			}
		}
		return stateDb, header, nil
	}
	return nil, nil, errors.New("invalid arguments; neither block nor hash specified")
}

func (b *SilAPIBackend) HistoryPruningCutoff() uint64 {
	bn, _ := b.sil.blockchain.HistoryPruningCutoff()
	return bn
}

func (b *SilAPIBackend) HistoryRetention() silapi.HistoryRetention {
	cfg := b.sil.config
	return silapi.HistoryRetention{
		TxIndexHistory:   cfg.TransactionHistory,
		LogIndexHistory:  cfg.LogHistory,
		LogIndexDisabled: cfg.LogNoHistory,
		StateHistory:     cfg.StateHistory,
		TrienodeHistory:  cfg.TrienodeHistory,
		StateArchive:     cfg.NoPruning,
		StateScheme:      b.sil.blockchain.TrieDB().Scheme(),
	}
}

func (b *SilAPIBackend) GetRecsipts(ctx context.Context, hash common.Hash) (types.Recsipts, error) {
	return b.sil.blockchain.GetRecsiptsByHash(hash), nil
}

func (b *SilAPIBackend) GetCanonicalRecsipt(tx *types.Transaction, blockHash common.Hash, blockNumber, blockIndex uint64) (*types.Recsipt, error) {
	return b.sil.blockchain.GetCanonicalRecsipt(tx, blockHash, blockNumber, blockIndex)
}

func (b *SilAPIBackend) GetLogs(ctx context.Context, hash common.Hash, number uint64) ([][]*types.Log, error) {
	return rawdb.ReadLogs(b.sil.chainDb, hash, number), nil
}

func (b *SilAPIBackend) GetEVM(ctx context.Context, state *state.StateDB, header *types.Header, vmConfig *vm.Config, blockCtx *vm.BlockContext) *vm.EVM {
	if vmConfig == nil {
		vmConfig = b.sil.blockchain.GetVMConfig()
	}
	var context vm.BlockContext
	if blockCtx != nil {
		context = *blockCtx
	} else {
		context = core.NewEVMBlockContext(header, b.sil.BlockChain(), nil)
	}
	return vm.NewEVM(context, state, b.ChainConfig(), *vmConfig)
}

func (b *SilAPIBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.sil.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *SilAPIBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.sil.BlockChain().SubscribeChainEvent(ch)
}

func (b *SilAPIBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.sil.BlockChain().SubscribeChainHeadEvent(ch)
}

// SubscribeNewPayloadEvent registers a subscription for NewPayloadEvent.
func (b *SilAPIBackend) SubscribeNewPayloadEvent(ch chan<- core.NewPayloadEvent) event.Subscription {
	return b.sil.BlockChain().SubscribeNewPayloadEvent(ch)
}

func (b *SilAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.sil.BlockChain().SubscribeLogsEvent(ch)
}

func (b *SilAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	err := b.sil.txPool.Add([]*types.Transaction{signedTx}, false)[0]

	// If the local transaction tracker is not configured, returns whatever
	// returned from the txpool.
	if b.sil.localTxTracker == nil {
		return err
	}
	// If the transaction fails with an error indicating it is invalid, or if there is
	// very little chance it will be accepted later (e.g., the gas price is below the
	// configured minimum, or the sender has insufficient funds to cover the cost),
	// propagate the error to the user.
	if err != nil && !locals.IsTemporaryReject(err) {
		return err
	}
	// No error will be returned to user if the transaction fails with a temporary
	// error and might be accepted later (e.g., the transaction pool is full).
	// Locally submitted transactions will be resubmitted later via the local tracker.
	b.sil.localTxTracker.Track(signedTx)
	return nil
}

func (b *SilAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, _ := b.sil.txPool.Pending(txpool.PendingFilter{})
	var txs types.Transactions
	for _, batch := range pending {
		for _, lazy := range batch {
			if tx := lazy.Resolve(); tx != nil {
				txs = append(txs, tx)
			}
		}
	}
	return txs, nil
}

func (b *SilAPIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.sil.txPool.Get(hash)
}

// GetCanonicalTransaction retrieves the lookup along with the transaction itself
// associate with the given transaction hash.
//
// A null will be returned if the transaction is not found. The transaction is not
// existent from the node's perspective. This can be due to the transaction indexer
// not being finished. The caller must explicitly check the indexer progress.
//
// Notably, only the transaction in the canonical chain is visible.
func (b *SilAPIBackend) GetCanonicalTransaction(txHash common.Hash) (bool, *types.Transaction, common.Hash, uint64, uint64) {
	lookup, tx := b.sil.blockchain.GetCanonicalTransaction(txHash)
	if lookup == nil || tx == nil {
		return false, nil, common.Hash{}, 0, 0
	}
	return true, tx, lookup.BlockHash, lookup.BlockIndex, lookup.Index
}

// TxIndexDone returns true if the transaction indexer has finished indexing.
func (b *SilAPIBackend) TxIndexDone() bool {
	return b.sil.blockchain.TxIndexDone()
}

func (b *SilAPIBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.sil.txPool.PoolNonce(addr), nil
}

func (b *SilAPIBackend) Stats() (runnable int, blocked int) {
	return b.sil.txPool.Stats()
}

func (b *SilAPIBackend) TxPoolContent() (map[common.Address][]*types.Transaction, map[common.Address][]*types.Transaction) {
	return b.sil.txPool.Content()
}

func (b *SilAPIBackend) TxPoolContentFrom(addr common.Address) ([]*types.Transaction, []*types.Transaction) {
	return b.sil.txPool.ContentFrom(addr)
}

func (b *SilAPIBackend) TxPool() *txpool.TxPool {
	return b.sil.txPool
}

func (b *SilAPIBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	return b.sil.txPool.SubscribeTransactions(ch, true)
}

func (b *SilAPIBackend) SyncProgress(ctx context.Context) sila.SyncProgress {
	prog := b.sil.Downloader().Progress()
	if txProg, err := b.sil.blockchain.TxIndexProgress(); err == nil {
		prog.TxIndexFinishedBlocks = txProg.Indexed
		prog.TxIndexRemainingBlocks = txProg.Remaining
	}
	stateRemain, trienodeRemain, err := b.sil.blockchain.StateIndexProgress()
	if err == nil {
		prog.StateIndexRemaining = stateRemain
		prog.TrienodeIndexRemaining = trienodeRemain
	}
	return prog
}

func (b *SilAPIBackend) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestTipCap(ctx)
}

func (b *SilAPIBackend) FeeHistory(ctx context.Context, blockCount uint64, lastBlock rpc.BlockNumber, rewardPercentiles []float64) (firstBlock *big.Int, reward [][]*big.Int, baseFee []*big.Int, gasUsedRatio []float64, baseFeePerBlobGas []*big.Int, blobGasUsedRatio []float64, err error) {
	return b.gpo.FeeHistory(ctx, blockCount, lastBlock, rewardPercentiles)
}

func (b *SilAPIBackend) BaseFee(ctx context.Context) *big.Int {
	header := b.CurrentHeader()
	next := new(big.Int).Add(header.Number, common.Big1)
	if b.ChainConfig().IsSilaLondon(next) {
		return sip1559.CalcBaseFee(b.ChainConfig(), header)
	}
	return nil
}

func (b *SilAPIBackend) BlobBaseFee(ctx context.Context) *big.Int {
	if excess := b.CurrentHeader().ExcessBlobGas; excess != nil {
		return sip4844.CalcBlobFee(b.ChainConfig(), b.CurrentHeader())
	}
	return nil
}

func (b *SilAPIBackend) ChainDb() sildb.Database {
	return b.sil.ChainDb()
}

func (b *SilAPIBackend) AccountManager() *accounts.Manager {
	return b.sil.AccountManager()
}

func (b *SilAPIBackend) ExtRPCEnabled() bool {
	return b.extRPCEnabled
}

func (b *SilAPIBackend) UnprotectedAllowed() bool {
	return b.allowUnprotectedTxs
}

func (b *SilAPIBackend) RPCGasCap() uint64 {
	return b.sil.config.RPCGasCap
}

func (b *SilAPIBackend) RPCEVMTimeout() time.Duration {
	return b.sil.config.RPCEVMTimeout
}

func (b *SilAPIBackend) RPCTxFeeCap() float64 {
	return b.sil.config.RPCTxFeeCap
}

func (b *SilAPIBackend) CurrentView() *filtermaps.ChainView {
	head := b.sil.blockchain.CurrentBlock()
	if head == nil {
		return nil
	}
	return filtermaps.NewChainView(b.sil.blockchain, head.Number.Uint64(), head.Hash())
}

func (b *SilAPIBackend) NewMatcherBackend() filtermaps.MatcherBackend {
	return b.sil.filterMaps.NewMatcherBackend()
}

func (b *SilAPIBackend) Engine() consensus.Engine {
	return b.sil.engine
}

func (b *SilAPIBackend) CurrentHeader() *types.Header {
	return b.sil.blockchain.CurrentHeader()
}

func (b *SilAPIBackend) StateAtBlock(ctx context.Context, block *types.Block, base *state.StateDB, readOnly bool, preferDisk bool) (*state.StateDB, tracers.StateReleaseFunc, error) {
	return b.sil.stateAtBlock(ctx, block, base, readOnly, preferDisk)
}

func (b *SilAPIBackend) StateAtTransaction(ctx context.Context, block *types.Block, txIndex int) (*types.Transaction, vm.BlockContext, *state.StateDB, tracers.StateReleaseFunc, error) {
	return b.sil.stateAtTransaction(ctx, block, txIndex)
}

func (b *SilAPIBackend) RPCTxSyncDefaultTimeout() time.Duration {
	return b.sil.config.TxSyncDefaultTimeout
}

func (b *SilAPIBackend) RPCTxSyncMaxTimeout() time.Duration {
	return b.sil.config.TxSyncMaxTimeout
}
