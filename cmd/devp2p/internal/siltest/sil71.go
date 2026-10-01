// Copyright 2026 The go-sila Authors
// This file is part of go-sila.
//
// go-sila is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-sila is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-sila. If not, see <http://www.gnu.org/licenses/>.

package siltest

import (
	"fmt"
	"math/rand"

	"github.com/sila-chain/go-sila/common"
	"github.com/sila-chain/go-sila/sila/protocols/sil"
	"github.com/sila-chain/go-sila/internal/utesting"
)

// Sil/71 (SIP-8159) adds BAL exchange to the sil protocol:
// GetBlockAccessLists (0x12) / BlockAccessLists (0x13).
//
// The tests in this file focus on the wire behavior introduced in sil/71.
// Tests for messages unchanged from earlier sil versions are covered by the
// main sil suite.

// TestSil71GetBlockAccessLists exercises GetBlockAccessLists requests defined
// in SIP-8159. Per the spec:
//
//   - BlockAccessLists entries correspond to request hashes in order.
//   - Unavailable BALs are returned as the RLP empty string (0x80) at the
//     matching position.
//   - The server may return fewer entries than requested when applying
//     response-size or implementation-defined limits, truncating from the tail.
//   - When a BAL is returned, its keccak256(rlp.encode(bal)) MUST match the
//     block-access-list-hash field of the corresponding block header.
func (s *Suite) TestSil71GetBlockAccessLists(t *utesting.T) {
	var (
		head     = s.chain.Head()
		headHash = head.Hash()
		preHash  = s.chain.blocks[s.chain.Len()-2].Hash()
		unknown  = common.HexToHash("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	)

	// Collect a window of recent canonical block hashes. Limit to at most 16
	// entries to keep the request small and well under any reasonable limit.
	var recent []common.Hash
	start := s.chain.Len() - 16
	if start < 1 {
		start = 1
	}
	for i := start; i < s.chain.Len(); i++ {
		recent = append(recent, s.chain.blocks[i].Hash())
	}

	tests := []accessListsTest{
		{
			desc: `An empty request. The server must respond with an empty list and must
not disconnect.`,
			hashes:     nil,
			minEntries: 0,
			maxEntries: 0,
		},
		{
			desc: `A request for a single random/unknown block hash. Per the spec the
server must respond and include an RLP empty string (0x80) at that position.`,
			hashes:     []common.Hash{unknown},
			minEntries: 1,
			maxEntries: 1,
		},
		{
			desc: `A request for multiple random/unknown block hashes. The server must
preserve request order and return an RLP empty string for each position.`,
			hashes: []common.Hash{
				unknown,
				common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
			},
			minEntries: 3,
			maxEntries: 3,
		},
		{
			desc: `A request for the chain head. The server must respond. If the node is
post-SilaAmsterdam and has the BAL for this block, the returned BAL must hash to
the block-access-list-hash in the header. Otherwise an empty entry is valid.`,
			hashes:     []common.Hash{headHash},
			minEntries: 1,
			maxEntries: 1,
		},
		{
			desc: `A request for the chain head and its parent. The server must return
exactly two entries, in request order.`,
			hashes:     []common.Hash{headHash, preHash},
			minEntries: 2,
			maxEntries: 2,
		},
		{
			desc: `A mixed request with known and unknown hashes. The server must
return entries in request order, with the RLP empty string at positions
corresponding to unknown hashes.`,
			hashes: []common.Hash{headHash, unknown, preHash, unknown},
			// We expect exactly 4 entries because the mixed response is small and
			// well under the recommended 2 MiB soft limit.
			minEntries: 4,
			maxEntries: 4,
		},
		{
			desc: `A request spanning the most recent canonical window. Implementations
may return empty entries for unavailable BALs or truncate from the tail, but
the entries that are returned must preserve request order.`,
			hashes:     recent,
			minEntries: 0,
			maxEntries: len(recent),
		},
		{
			desc: `A request containing the same hash repeated. The server must treat
each position independently and preserve request order.`,
			hashes:     []common.Hash{headHash, headHash, headHash},
			minEntries: 3,
			maxEntries: 3,
		},
	}

	for i, tc := range tests {
		if i > 0 {
			t.Log("\n")
		}
		t.Logf("-- Test %d", i)
		t.Log(tc.desc)
		t.Log("  request:")
		t.Logf("      hashes: %d", len(tc.hashes))
		if err := s.sil71GetBlockAccessLists(t, &tc); err != nil {
			t.Errorf("test %d failed: %v", i, err)
		}
	}
}

// sil71GetBlockAccessLists sends a GetBlockAccessLists request over sil/71 and
// validates the response against SIP-8159, using the response validation
// shared with the snap/2 suite.
func (s *Suite) sil71GetBlockAccessLists(t *utesting.T, tc *accessListsTest) error {
	conn, err := s.dialEth71()
	if err != nil {
		return fmt.Errorf("dial failed: %v", err)
	}
	defer conn.Close()
	if err = conn.peer(s.chain, nil); err != nil {
		return fmt.Errorf("peering failed: %v", err)
	}

	req := &sil.GetBlockAccessListsPacket{
		RequestId:                  uint64(rand.Int63()),
		GetBlockAccessListsRequest: tc.hashes,
	}
	if err := conn.Write(silProto, sil.GetBlockAccessListsMsg, req); err != nil {
		return fmt.Errorf("access list request failed: %v", err)
	}
	var res sil.BlockAccessListPacket
	if err := conn.ReadMsg(silProto, sil.BlockAccessListsMsg, &res); err != nil {
		return fmt.Errorf("access list response failed: %v", err)
	}
	return s.validateAccessListsResponse(t, tc, req.RequestId, res.RequestId, res.List)
}
