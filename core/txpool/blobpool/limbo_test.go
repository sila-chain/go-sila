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

package blobpool

import (
	"testing"

	"github.com/holiman/billy"
	"github.com/sila-chain/go-sila/common"
	"github.com/sila-chain/go-sila/core/types"
	"github.com/sila-chain/go-sila/crypto"
	"github.com/sila-chain/go-sila/params"
	"github.com/sila-chain/go-sila/rlp"
)

// TestLimboLegacyMigration checks whether limbo entries in the legacy limboBlob type
// are migrated to the blobTxForPool layout on startup instead of being dropped.
func TestLimboLegacyMigration(t *testing.T) {
	key, _ := crypto.GenerateKey()
	tx := makeMultiBlobTx(0, 10, 100, 100, 2, 0, key)

	dir := t.TempDir()

	// Write a single entry using the legacy on-disk layout.
	store, err := billy.Open(billy.Options{Path: dir}, newSlotterSIP7594(params.BlobTxMaxBlobs), nil)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	legacy := struct {
		TxHash common.Hash
		Block  uint64
		Tx     *types.Transaction
	}{tx.Hash(), 42, tx}
	data, err := rlp.EncodeToBytes(&legacy)
	if err != nil {
		t.Fatalf("failed to encode legacy entry: %v", err)
	}
	if _, err := store.Put(data); err != nil {
		t.Fatalf("failed to store legacy entry: %v", err)
	}
	store.Close()

	// Open the limbo, which should migrate the legacy entry.
	l, err := newLimbo(new(params.ChainConfig), dir)
	if err != nil {
		t.Fatalf("failed to open limbo: %v", err)
	}
	defer l.Close()

	// The migrated transaction must be tracked and reconstruct the original.
	ptx, err := l.pull(tx.Hash())
	if err != nil {
		t.Fatalf("failed to pull migrated tx: %v", err)
	}
	if got := ptx.ToTx().Hash(); got != tx.Hash() {
		t.Fatalf("migrated tx hash mismatch: got %x, want %x", got, tx.Hash())
	}
}
