// Copyright 2017 The go-sila Authors
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

// Package silash implements the silash proof-of-work consensus engine.
package silash

import (
	"time"

	"github.com/sila-chain/go-sila/consensus"
	"github.com/sila-chain/go-sila/core/types"
)

// Silash is a consensus engine based on proof-of-work implementing the silash
// algorithm.
type Silash struct {
	fakeFail  *uint64        // Block number which fails PoW check even in fake mode
	fakeDelay *time.Duration // Time delay to sleep for before returning from verify
	fakeFull  bool           // Accepts everything as valid
}

// NewFaker creates an silash consensus engine with a fake PoW scheme that accepts
// all blocks' seal as valid, though they still have to conform to the Sila
// consensus rules.
func NewFaker() *Silash {
	return new(Silash)
}

// NewFakeFailer creates a silash consensus engine with a fake PoW scheme that
// accepts all blocks as valid apart from the single one specified, though they
// still have to conform to the Sila consensus rules.
func NewFakeFailer(fail uint64) *Silash {
	return &Silash{
		fakeFail: &fail,
	}
}

// NewFakeDelayer creates a silash consensus engine with a fake PoW scheme that
// accepts all blocks as valid, but delays verifications by some time, though
// they still have to conform to the Sila consensus rules.
func NewFakeDelayer(delay time.Duration) *Silash {
	return &Silash{
		fakeDelay: &delay,
	}
}

// NewFullFaker creates an silash consensus engine with a full fake scheme that
// accepts all blocks as valid, without checking any consensus rules whatsoever.
func NewFullFaker() *Silash {
	return &Silash{
		fakeFull: true,
	}
}

// Close closes the exit channel to notify all backend threads exiting.
func (silash *Silash) Close() error {
	return nil
}

// Seal generates a new sealing request for the given input block and pushes
// the result into the given channel. For the silash engine, this method will
// just panic as sealing is not supported anymore.
func (silash *Silash) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	panic("silash (pow) sealing not supported any more")
}
