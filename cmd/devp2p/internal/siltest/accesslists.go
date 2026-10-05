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
	"bytes"
	"fmt"

	"github.com/sila-chain/go-sila/core/types"
	"github.com/sila-chain/go-sila/core/types/bal"
	"github.com/sila-chain/go-sila/crypto"
	"github.com/sila-chain/go-sila/internal/utesting"
	"github.com/sila-chain/go-sila/rlp"
)

// validateAccessListsResponse validates a BlockAccessLists response shared by
// the snap/2 (SIP-8189) and sil/71 (SIP-8159) test suites. Both protocols use
// the same positional response list where the RLP empty string (0x80) marks an
// unavailable BAL.
func (s *Suite) validateAccessListsResponse(t *utesting.T, tc *accessListsTest, reqID, resID uint64, accessLists rlp.RawList[rlp.RawValue]) error {
	if resID != reqID {
		return fmt.Errorf("request id mismatch: got %d, want %d", resID, reqID)
	}

	// Check list length bounds.
	got := accessLists.Len()
	if got < tc.minEntries || got > tc.maxEntries {
		return fmt.Errorf("response has %d entries, want between %d and %d", got, tc.minEntries, tc.maxEntries)
	}

	// Build a map of request-index -> block so we can verify BAL hashes.
	blocks := make(map[int]*types.Block)
	for i, h := range tc.hashes {
		for _, b := range s.chain.blocks {
			if b.Hash() == h {
				blocks[i] = b
				break
			}
		}
	}

	// Iterate the response, validating each entry positionally.
	var (
		idx int
		it  = accessLists.ContentIterator()
	)
	for it.Next() {
		raw := it.Value()
		block := blocks[idx]

		// Empty entry: per spec, indicates BAL is unavailable for that block.
		if bytes.Equal(raw, rlp.EmptyString) {
			if block != nil && block.Header().BlockAccessListHash != nil {
				if s.requireAvailableBALs {
					return fmt.Errorf("entry %d: local sila returned no BAL for known post-SilaAmsterdam block %x", idx, tc.hashes[idx])
				}
				// Not a failure — the server is allowed to legitimately not
				// have the BAL. But we log it so the test output is diagnosable.
				t.Logf("    entry %d: server returned empty for known post-SilaAmsterdam block %x", idx, tc.hashes[idx])
			}
			idx++
			continue
		}

		// Non-empty entry. A BAL is only legitimate for a block we know
		// locally whose header commits to one; for any other hash the only
		// valid response is the RLP empty string, so receiving data here
		// means the server fabricated it.
		if block == nil {
			return fmt.Errorf("entry %d: server returned BAL data for unknown hash %x", idx, tc.hashes[idx])
		}
		if block.Header().BlockAccessListHash == nil {
			return fmt.Errorf("entry %d: server returned BAL data for a block with no expected BAL (hash %x)", idx, tc.hashes[idx])
		}

		// Per SIP-8189: compute keccak256(rlp.encode(bal)) against the raw
		// bytes actually received on the wire, and compare to the header
		// commitment. Hashing raw bytes (rather than re-encoding after a
		// decode round-trip) catches peers that send non-canonical BAL
		// encodings.
		have := crypto.Keccak256Hash(raw)
		want := *block.Header().BlockAccessListHash
		if have != want {
			return fmt.Errorf("entry %d: BAL hash mismatch: have %x, want %x", idx, have, want)
		}

		// Decode and validate the BAL's internal structure: ordering of
		// accounts/slots/changes, code-size limits, and per-entry access-index
		// bounds, against the known block.
		var accessList bal.BlockAccessList
		if err := rlp.DecodeBytes(raw, &accessList); err != nil {
			return fmt.Errorf("entry %d: invalid BAL RLP: %v", idx, err)
		}
		if err := accessList.Validate(block.GasLimit(), len(block.Transactions())); err != nil {
			return fmt.Errorf("entry %d: BAL failed validation: %v", idx, err)
		}
		idx++
	}

	// Sanity: iterator consumed exactly the reported number of entries.
	if idx != got {
		return fmt.Errorf("iterator visited %d entries, expected %d", idx, got)
	}
	return nil
}
