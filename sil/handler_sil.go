// Copyright 2020 The go-sila Authors
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
	"errors"
	"fmt"

	"github.com/sila-chain/go-sila/common"
	"github.com/sila-chain/go-sila/core"
	"github.com/sila-chain/go-sila/core/types"
	"github.com/sila-chain/go-sila/crypto/kzg4844"
	"github.com/sila-chain/go-sila/sil/protocols/sil"
	"github.com/sila-chain/go-sila/p2p/enode"
	"github.com/sila-chain/go-sila/params"
)

// silHandler implements the sil.Backend interface to handle the various network
// packets that are sent as replies or broadcasts.
type silHandler handler

func (h *silHandler) Chain() *core.BlockChain { return h.chain }
func (h *silHandler) TxPool() sil.TxPool      { return h.txpool }
func (h *silHandler) BlobPool() sil.BlobPool  { return h.blobpool }

// RunPeer is invoked when a peer joins on the `sil` protocol.
func (h *silHandler) RunPeer(peer *sil.Peer, hand sil.Handler) error {
	return (*handler)(h).runEthPeer(peer, hand)
}

// PeerInfo retrieves all known `sil` information about a peer.
func (h *silHandler) PeerInfo(id enode.ID) interface{} {
	if p := h.peers.peer(id.String()); p != nil {
		return p.info()
	}
	return nil
}

// AcceptTxs retrieves whether transaction processing is enabled on the node
// or if inbound transactions should simply be dropped.
func (h *silHandler) AcceptTxs() bool {
	return h.synced.Load()
}

// Handle is invoked from a peer's message handler when it receives a new remote
// message that the handler couldn't consume and serve itself.
func (h *silHandler) Handle(peer *sil.Peer, packet sil.Packet) error {
	// Consume any broadcasts and announces, forwarding the rest to the downloader
	switch packet := packet.(type) {
	case *sil.NewPooledTransactionHashesPacket72:
		hashes, err := h.txFetcher.Notify(peer.ID(), packet.Types, packet.Sizes, packet.Hashes)
		if err != nil {
			return err
		}
		if len(hashes) != 0 {
			return h.blobFetcher.Notify(peer.ID(), hashes, packet.Mask)
		}
		return nil

	case *sil.NewPooledTransactionHashesPacket71:
		_, err := h.txFetcher.Notify(peer.ID(), packet.Types, packet.Sizes, packet.Hashes)
		return err

	case *sil.TransactionsPacket:
		txs, err := packet.Items()
		if err != nil {
			return fmt.Errorf("Transactions: %v", err)
		}
		if err := handleTransactions(peer, txs, true); err != nil {
			return fmt.Errorf("Transactions: %v", err)
		}
		return h.txFetcher.Enqueue(peer.ID(), peer.Version(), txs, false)

	case *sil.PooledTransactionsPacket:
		txs, err := packet.List.Items()
		if err != nil {
			return fmt.Errorf("PooledTransactions: %v", err)
		}
		if err := handleTransactions(peer, txs, false); err != nil {
			return fmt.Errorf("PooledTransactions: %v", err)
		}
		return h.txFetcher.Enqueue(peer.ID(), peer.Version(), txs, true)

	case *sil.CellsResponse:
		outer, err := packet.Cells.Items()
		if err != nil {
			return fmt.Errorf("Cells: %v", err)
		}
		cells := make([][]kzg4844.Cell, len(outer))
		for i := range outer {
			if outer[i].Len() > params.BlobTxMaxBlobs*kzg4844.CellsPerBlob {
				return fmt.Errorf("Cells: cells per tx exceeded the possible maximum")
			}
			if cells[i], err = outer[i].Items(); err != nil {
				return fmt.Errorf("Cells: %v", err)
			}
		}
		return h.blobFetcher.Enqueue(peer.ID(), packet.Hashes, cells, packet.Mask)

	default:
		return fmt.Errorf("unexpected sil packet type: %T", packet)
	}
}

// handleTransactions marks all given transactions as known to the peer
// and performs basic validations.
func handleTransactions(peer *sil.Peer, list []*types.Transaction, directBroadcast bool) error {
	seen := make(map[common.Hash]struct{}, len(list))
	for _, tx := range list {
		if tx.Type() == types.BlobTxType {
			if directBroadcast {
				return errors.New("disallowed broadcast blob transaction")
			} else {
				// If we receive any blob transactions missing sidecars, or with
				// sidecars that don't correspond to the versioned hashes reported
				// in the header, disconnect from the sending peer.
				if tx.BlobTxSidecar() == nil {
					return errors.New("received sidecar-less blob transaction")
				}
				if err := tx.BlobTxSidecar().ValidateBlobCommitmentHashes(tx.BlobHashes()); err != nil {
					return err
				}
			}
		}

		// Check for duplicates.
		hash := tx.Hash()
		if _, exists := seen[hash]; exists {
			return fmt.Errorf("multiple copies of the same hash %v", hash)
		}
		seen[hash] = struct{}{}

		// Mark as known.
		peer.MarkTransaction(hash)
	}
	return nil
}
