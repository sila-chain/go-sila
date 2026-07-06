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
	"testing"
	"time"

	"github.com/sila-chain/go-sila/p2p"
	"github.com/sila-chain/go-sila/p2p/enode"
	"github.com/sila-chain/go-sila/sil/protocols/sil"
	"github.com/sila-chain/go-sila/sil/protocols/snap"
	"github.com/sila-chain/go-sila/sil/silconfig"
)

// Tests that snap sync is disabled after a successful sync cycle.
func TestSnapSyncDisabling69(t *testing.T) { testSnapSyncDisabling(t, sil.ETH69, snap.SNAP1) }

// Tests that snap sync gets disabled as soon as a real block is successfully
// imported into the blockchain.
func testSnapSyncDisabling(t *testing.T, silVer uint, snapVer uint) {
	t.Parallel()

	// Create an empty handler and ensure it's in snap sync mode
	empty := newTestHandler(silconfig.SnapSync)
	defer empty.close()

	// Create a full handler and ensure snap sync ends up disabled
	full := newTestHandlerWithBlocks(1024, silconfig.SnapSync)
	defer full.close()

	// Sync up the two handlers via both `sil` and `snap`
	caps := []p2p.Cap{{Name: "sil", Version: silVer}, {Name: "snap", Version: snapVer}}

	emptyPipeSil, fullPipeSil := p2p.MsgPipe()
	defer emptyPipeSil.Close()
	defer fullPipeSil.Close()

	emptyPeerSil := sil.NewPeer(silVer, p2p.NewPeer(enode.ID{1}, "", caps), emptyPipeSil, empty.txpool, nil)
	fullPeerSil := sil.NewPeer(silVer, p2p.NewPeer(enode.ID{2}, "", caps), fullPipeSil, full.txpool, nil)
	defer emptyPeerSil.Close()
	defer fullPeerSil.Close()

	go empty.handler.runEthPeer(emptyPeerSil, func(peer *sil.Peer) error {
		return sil.Handle((*silHandler)(empty.handler), peer)
	})
	go full.handler.runEthPeer(fullPeerSil, func(peer *sil.Peer) error {
		return sil.Handle((*silHandler)(full.handler), peer)
	})

	emptyPipeSnap, fullPipeSnap := p2p.MsgPipe()
	defer emptyPipeSnap.Close()
	defer fullPipeSnap.Close()

	emptyPeerSnap := snap.NewPeer(snapVer, p2p.NewPeer(enode.ID{1}, "", caps), emptyPipeSnap)
	fullPeerSnap := snap.NewPeer(snapVer, p2p.NewPeer(enode.ID{2}, "", caps), fullPipeSnap)

	go empty.handler.runSnapExtension(emptyPeerSnap, func(peer *snap.Peer) error {
		return snap.Handle((*snapHandler)(empty.handler), peer)
	})
	go full.handler.runSnapExtension(fullPeerSnap, func(peer *snap.Peer) error {
		return snap.Handle((*snapHandler)(full.handler), peer)
	})
	// Wait a bit for the above handlers to start
	time.Sleep(250 * time.Millisecond)

	// Check that snap sync was disabled
	if err := empty.handler.downloader.BeaconSync(full.chain.CurrentBlock(), nil); err != nil {
		t.Fatal("sync failed:", err)
	}
	// Snap sync and mode switching happen asynchronously, poll for completion.
	timeout := time.NewTimer(15 * time.Second)
	tick := time.NewTicker(100 * time.Millisecond)
	defer timeout.Stop()
	defer tick.Stop()

	for {
		select {
		case <-timeout.C:
			t.Fatalf("snap sync not disabled after successful synchronisation")
		case <-tick.C:
			if empty.handler.synced.Load() && empty.handler.downloader.ConfigSyncMode() == silconfig.FullSync {
				return
			}
		}
	}
}
