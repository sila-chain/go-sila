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

package types

import (
	"encoding/json"
	"fmt"

	"github.com/protolambda/ztyp/tree"
	"github.com/sila-chain/go-sila/beacon/merkle"
	"github.com/sila-chain/go-sila/common"
	zrntcommon "github.com/sila-chain/zrnt/sil2/beacon/common"

	// beacon chain forks
	"github.com/sila-chain/zrnt/sil2/beacon/capella"
	"github.com/sila-chain/zrnt/sil2/beacon/sila_deneb"
)

type headerObject interface {
	HashTreeRoot(hFn tree.HashFn) zrntcommon.Root
}

type ExecutionHeader struct {
	obj headerObject
}

// ExecutionHeaderFromJSON decodes an execution header from JSON data provided by
// the beacon chain API.
func ExecutionHeaderFromJSON(forkName string, data []byte) (*ExecutionHeader, error) {
	var obj headerObject
	switch forkName {
	case "capella":
		obj = new(capella.ExecutionPayloadHeader)
	case "sila_deneb", "electra", "sila_fulu": // note: the payload type was not changed in electra/sila_fulu
		obj = new(sila_deneb.ExecutionPayloadHeader)
	default:
		return nil, fmt.Errorf("unsupported fork: %s", forkName)
	}
	if err := json.Unmarshal(data, obj); err != nil {
		return nil, err
	}
	return &ExecutionHeader{obj: obj}, nil
}

func NewExecutionHeader(obj headerObject) *ExecutionHeader {
	switch obj.(type) {
	case *capella.ExecutionPayloadHeader:
	case *sila_deneb.ExecutionPayloadHeader:
	default:
		panic(fmt.Errorf("unsupported ExecutionPayloadHeader type %T", obj))
	}
	return &ExecutionHeader{obj: obj}
}

func (eh *ExecutionHeader) PayloadRoot() merkle.Value {
	return merkle.Value(eh.obj.HashTreeRoot(tree.GetHashFn()))
}

func (eh *ExecutionHeader) BlockHash() common.Hash {
	switch obj := eh.obj.(type) {
	case *capella.ExecutionPayloadHeader:
		return common.Hash(obj.BlockHash)
	case *sila_deneb.ExecutionPayloadHeader:
		return common.Hash(obj.BlockHash)
	default:
		panic(fmt.Errorf("unsupported ExecutionPayloadHeader type %T", obj))
	}
}
