// Copyright 2025 The go-sila Authors
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

package rawdb

import (
	"github.com/sila-chain/go-sila/common"
	"github.com/sila-chain/go-sila/sildb"
)

func ReadVerkleTransitionState(db sildb.KeyValueReader, hash common.Hash) ([]byte, error) {
	return db.Get(transitionStateKey(hash))
}

func WriteVerkleTransitionState(db sildb.KeyValueWriter, hash common.Hash, state []byte) error {
	return db.Put(transitionStateKey(hash), state)
}
