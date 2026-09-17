// Copyright 2023 The go-sila Authors
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
	"errors"
	"time"

	"github.com/sila-chain/go-sila"
	"github.com/sila-chain/go-sila/common"
	"github.com/sila-chain/go-sila/core"
	"github.com/sila-chain/go-sila/core/types"
	"github.com/sila-chain/go-sila/sil"
	"github.com/sila-chain/go-sila/sil/catalyst"
	"github.com/sila-chain/go-sila/sil/silconfig"
	"github.com/sila-chain/go-sila/sil/filters"
	"github.com/sila-chain/go-sila/silclient"
	"github.com/sila-chain/go-sila/node"
	"github.com/sila-chain/go-sila/p2p"
	"github.com/sila-chain/go-sila/params"
	"github.com/sila-chain/go-sila/rpc"
)

// Client exposes the methods provided by the Sila RPC client.
type Client interface {
	sila.BlockNumberReader
	sila.ChainReader
	sila.ChainStateReader
	sila.ContractCaller
	sila.GasEstimator
	sila.GasPricer
	sila.GasPricer1559
	sila.FeeHistoryReader
	sila.LogFilterer
	sila.PendingStateReader
	sila.PendingContractCaller
	sila.TransactionReader
	sila.TransactionSender
	sila.ChainIDReader
}

// simClient wraps silclient. This exists to prevent extracting silclient.Client
// from the Client interface returned by Backend.
type simClient struct {
	*silclient.Client
}

// Backend is a simulated blockchain. You can use it to test your contracts or
// other code that interacts with the Sila chain.
type Backend struct {
	node   *node.Node
	beacon *catalyst.SimulatedBeacon
	client simClient
}

// NewBackend creates a new simulated blockchain that can be used as a backend for
// contract bindings in unit tests.
//
// A simulated backend always uses chainID 1337.
func NewBackend(alloc types.GenesisAlloc, options ...func(nodeConf *node.Config, silConf *silconfig.Config)) *Backend {
	// Create the default configurations for the outer node shell and the Sila
	// service to mutate with the options afterwards
	nodeConf := node.DefaultConfig
	nodeConf.DataDir = ""
	nodeConf.P2P = p2p.Config{NoDiscovery: true}

	silConf := silconfig.Defaults
	silConf.Genesis = &core.Genesis{
		Config:   params.AllDevChainProtocolChanges,
		GasLimit: silconfig.Defaults.Miner.GasCeil,
		Alloc:    alloc,
	}
	silConf.SyncMode = silconfig.FullSync
	silConf.TxPool.NoLocals = true
	// Disable log indexing to force unindexed log search
	silConf.LogNoHistory = true

	for _, option := range options {
		option(&nodeConf, &silConf)
	}
	// Assemble the Sila stack to run the chain with
	stack, err := node.New(&nodeConf)
	if err != nil {
		panic(err) // this should never happen
	}
	sim, err := newWithNode(stack, &silConf, 0)
	if err != nil {
		panic(err) // this should never happen
	}
	return sim
}

// newWithNode sets up a simulated backend on an existing node. The provided node
// must not be started and will be started by this method.
func newWithNode(stack *node.Node, conf *sil.Config, blockPeriod uint64) (*Backend, error) {
	backend, err := sil.New(stack, conf)
	if err != nil {
		return nil, err
	}
	// Register the filter system
	filterSystem := filters.NewFilterSystem(backend.APIBackend, filters.Config{})
	stack.RegisterAPIs([]rpc.API{{
		Namespace: "sil",
		Service:   filters.NewFilterAPI(filterSystem),
	}})
	// Start the node
	if err := stack.Start(); err != nil {
		return nil, err
	}
	// Set up the simulated beacon
	beacon, err := catalyst.NewSimulatedBeacon(blockPeriod, common.Address{}, backend)
	if err != nil {
		return nil, err
	}
	// Reorg our chain back to genesis
	if err := beacon.Fork(backend.BlockChain().GetCanonicalHash(0)); err != nil {
		return nil, err
	}
	return &Backend{
		node:   stack,
		beacon: beacon,
		client: simClient{silclient.NewClient(stack.Attach())},
	}, nil
}

// Close shuts down the simBackend.
// The simulated backend can't be used afterwards.
func (n *Backend) Close() error {
	if n.client.Client != nil {
		n.client.Close()
		n.client = simClient{}
	}
	var err error
	if n.beacon != nil {
		err = n.beacon.Stop()
		n.beacon = nil
	}
	if n.node != nil {
		err = errors.Join(err, n.node.Close())
		n.node = nil
	}
	return err
}

// Commit seals a block and moves the chain forward to a new empty block.
func (n *Backend) Commit() common.Hash {
	return n.beacon.Commit()
}

// Rollback removes all pending transactions, reverting to the last committed state.
func (n *Backend) Rollback() {
	n.beacon.Rollback()
}

// Fork creates a side-chain that can be used to simulate reorgs.
//
// This function should be called with the ancestor block where the new side
// chain should be started. Transactions (old and new) can then be applied on
// top and Commit-ed.
//
// Note, the side-chain will only become canonical (and trigger the events) when
// it becomes longer. Until then CallContract will still operate on the current
// canonical chain.
//
// There is a % chance that the side chain becomes canonical at the same length
// to simulate live network behavior.
func (n *Backend) Fork(parentHash common.Hash) error {
	return n.beacon.Fork(parentHash)
}

// AdjustTime changes the block timestamp and creates a new block.
// It can only be called on empty blocks.
func (n *Backend) AdjustTime(adjustment time.Duration) error {
	return n.beacon.AdjustTime(adjustment)
}

// Client returns a client that accesses the simulated chain.
func (n *Backend) Client() Client {
	return n.client
}
