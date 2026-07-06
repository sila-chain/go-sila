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

package history

import (
	"testing"

	"github.com/sila-chain/go-sila/common"
	"github.com/sila-chain/go-sila/params"
)

func TestNewPolicy(t *testing.T) {
	// KeepAll: no target.
	p, err := NewPolicy(KeepAll, params.SilaMainnetGenesisHash)
	if err != nil {
		t.Fatalf("KeepAll: %v", err)
	}
	if p.Mode != KeepAll || p.Target != nil {
		t.Errorf("KeepAll: unexpected policy %+v", p)
	}

	// PostMerge: resolves known sila-mainnet prune point.
	p, err = NewPolicy(KeepPostMerge, params.SilaMainnetGenesisHash)
	if err != nil {
		t.Fatalf("PostMerge: %v", err)
	}
	if p.Target == nil || p.Target.BlockNumber != 15537393 {
		t.Errorf("PostMerge: unexpected target %+v", p.Target)
	}

	// PostSilaPrague: resolves known sila-mainnet prune point.
	p, err = NewPolicy(KeepPostSilaPrague, params.SilaMainnetGenesisHash)
	if err != nil {
		t.Fatalf("PostSilaPrague: %v", err)
	}
	if p.Target == nil || p.Target.BlockNumber != 22431084 {
		t.Errorf("PostSilaPrague: unexpected target %+v", p.Target)
	}

	// PostMerge on unknown network: error.
	if _, err = NewPolicy(KeepPostMerge, common.HexToHash("0xdeadbeef")); err == nil {
		t.Fatal("PostMerge unknown network: expected error")
	}
}
