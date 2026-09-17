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

package core

import (
	"bytes"
	"encoding/json"
	"math/big"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/sila-chain/go-sila/common"
	"github.com/sila-chain/go-sila/consensus/silash"
	"github.com/sila-chain/go-sila/core/rawdb"
	"github.com/sila-chain/go-sila/core/types"
	"github.com/sila-chain/go-sila/params"
	"github.com/sila-chain/go-sila/sildb"
	"github.com/sila-chain/go-sila/triedb"
	"github.com/sila-chain/go-sila/triedb/pathdb"
)

func TestSetupGenesis(t *testing.T) {
	testSetupGenesis(t, rawdb.HashScheme)
	testSetupGenesis(t, rawdb.PathScheme)
}

func testSetupGenesis(t *testing.T, scheme string) {
	var (
		customghash = common.HexToHash("0x89c99d90b79719238d2645c7642f2c9295246e80775b38cfd162b696817fbd50")
		customg     = Genesis{
			Config: &params.ChainConfig{SilaHomesteadBlock: big.NewInt(3), Silash: &params.SilashConfig{}},
			Alloc: types.GenesisAlloc{
				{1}: {Balance: big.NewInt(1), Storage: map[common.Hash]common.Hash{{1}: {1}}},
			},
		}
		oldcustomg = customg
	)
	oldcustomg.Config = &params.ChainConfig{SilaHomesteadBlock: big.NewInt(2), Silash: &params.SilashConfig{}}

	tests := []struct {
		name           string
		fn             func(sildb.Database) (*params.ChainConfig, common.Hash, *params.ConfigCompatError, error)
		wantConfig     *params.ChainConfig
		wantHash       common.Hash
		wantErr        error
		wantCompactErr *params.ConfigCompatError
	}{
		{
			name: "genesis without ChainConfig",
			fn: func(db sildb.Database) (*params.ChainConfig, common.Hash, *params.ConfigCompatError, error) {
				return SetupGenesisBlock(db, triedb.NewDatabase(db, newDbConfig(scheme)), new(Genesis))
			},
			wantErr: errGenesisNoConfig,
		},
		{
			name: "no block in DB, genesis == nil",
			fn: func(db sildb.Database) (*params.ChainConfig, common.Hash, *params.ConfigCompatError, error) {
				return SetupGenesisBlock(db, triedb.NewDatabase(db, newDbConfig(scheme)), nil)
			},
			wantHash:   params.SilaMainnetGenesisHash,
			wantConfig: params.SilaMainnetChainConfig,
		},
		{
			name: "mainnet block in DB, genesis == nil",
			fn: func(db sildb.Database) (*params.ChainConfig, common.Hash, *params.ConfigCompatError, error) {
				DefaultGenesisBlock().MustCommit(db, triedb.NewDatabase(db, newDbConfig(scheme)))
				return SetupGenesisBlock(db, triedb.NewDatabase(db, newDbConfig(scheme)), nil)
			},
			wantHash:   params.SilaMainnetGenesisHash,
			wantConfig: params.SilaMainnetChainConfig,
		},
		{
			name: "custom block in DB, genesis == nil",
			fn: func(db sildb.Database) (*params.ChainConfig, common.Hash, *params.ConfigCompatError, error) {
				tdb := triedb.NewDatabase(db, newDbConfig(scheme))
				customg.Commit(db, tdb, nil)
				return SetupGenesisBlock(db, tdb, nil)
			},
			wantHash:   customghash,
			wantConfig: customg.Config,
		},
		{
			name: "custom block in DB, genesis == sepolia",
			fn: func(db sildb.Database) (*params.ChainConfig, common.Hash, *params.ConfigCompatError, error) {
				tdb := triedb.NewDatabase(db, newDbConfig(scheme))
				customg.Commit(db, tdb, nil)
				return SetupGenesisBlock(db, tdb, DefaultSilaSepoliaGenesisBlock())
			},
			wantErr: &GenesisMismatchError{Stored: customghash, New: params.SilaSepoliaGenesisHash},
		},
		{
			name: "custom block in DB, genesis == sila-hoodi",
			fn: func(db sildb.Database) (*params.ChainConfig, common.Hash, *params.ConfigCompatError, error) {
				tdb := triedb.NewDatabase(db, newDbConfig(scheme))
				customg.Commit(db, tdb, nil)
				return SetupGenesisBlock(db, tdb, DefaultSilaHoodiGenesisBlock())
			},
			wantErr: &GenesisMismatchError{Stored: customghash, New: params.SilaHoodiGenesisHash},
		},
		{
			name: "compatible config in DB",
			fn: func(db sildb.Database) (*params.ChainConfig, common.Hash, *params.ConfigCompatError, error) {
				tdb := triedb.NewDatabase(db, newDbConfig(scheme))
				oldcustomg.Commit(db, tdb, nil)
				return SetupGenesisBlock(db, tdb, &customg)
			},
			wantHash:   customghash,
			wantConfig: customg.Config,
		},
		{
			name: "incompatible config in DB",
			fn: func(db sildb.Database) (*params.ChainConfig, common.Hash, *params.ConfigCompatError, error) {
				// Commit the 'old' genesis block with SilaHomestead transition at #2.
				// Advance to block #4, past the homestead transition block of customg.
				tdb := triedb.NewDatabase(db, newDbConfig(scheme))
				oldcustomg.Commit(db, tdb, nil)

				bc, _ := NewBlockChain(db, &oldcustomg, silash.NewFullFaker(), DefaultConfig().WithStateScheme(scheme))
				defer bc.Stop()

				_, blocks, _ := GenerateChainWithGenesis(&oldcustomg, silash.NewFaker(), 4, nil)
				bc.InsertChain(blocks)

				// This should return a compatibility error.
				return SetupGenesisBlock(db, tdb, &customg)
			},
			wantHash:   customghash,
			wantConfig: customg.Config,
			wantCompactErr: &params.ConfigCompatError{
				What:          "SilaHomestead fork block",
				StoredBlock:   big.NewInt(2),
				NewBlock:      big.NewInt(3),
				RewindToBlock: 1,
			},
		},
	}

	for _, test := range tests {
		db := rawdb.NewMemoryDatabase()
		config, hash, compatErr, err := test.fn(db)
		// Check the return values.
		if !reflect.DeepEqual(err, test.wantErr) {
			spew := spew.ConfigState{DisablePointerAddresses: true, DisableCapacities: true}
			t.Errorf("%s: returned error %#v, want %#v", test.name, spew.NewFormatter(err), spew.NewFormatter(test.wantErr))
		}
		if !reflect.DeepEqual(compatErr, test.wantCompactErr) {
			spew := spew.ConfigState{DisablePointerAddresses: true, DisableCapacities: true}
			t.Errorf("%s: returned error %#v, want %#v", test.name, spew.NewFormatter(compatErr), spew.NewFormatter(test.wantCompactErr))
		}
		if !reflect.DeepEqual(config, test.wantConfig) {
			t.Errorf("%s:\nreturned %v\nwant     %v", test.name, config, test.wantConfig)
		}
		if hash != test.wantHash {
			t.Errorf("%s: returned hash %s, want %s", test.name, hash.Hex(), test.wantHash.Hex())
		} else if err == nil {
			// Check database content.
			stored := rawdb.ReadBlock(db, test.wantHash, 0)
			if stored.Hash() != test.wantHash {
				t.Errorf("%s: block in DB has hash %s, want %s", test.name, stored.Hash(), test.wantHash)
			}
		}
	}
}

// TestGenesisHashes checks the congruity of default genesis data to
// corresponding hardcoded genesis hash values.
func TestGenesisHashes(t *testing.T) {
	for i, c := range []struct {
		genesis *Genesis
		want    common.Hash
	}{
		{DefaultGenesisBlock(), params.SilaMainnetGenesisHash},
		{DefaultSilaSepoliaGenesisBlock(), params.SilaSepoliaGenesisHash},
		{DefaultSilaHoleskyGenesisBlock(), params.SilaHoleskyGenesisHash},
		{DefaultSilaHoodiGenesisBlock(), params.SilaHoodiGenesisHash},
	} {
		// Test via MustCommit
		db := rawdb.NewMemoryDatabase()
		if have := c.genesis.MustCommit(db, triedb.NewDatabase(db, triedb.HashDefaults)).Hash(); have != c.want {
			t.Errorf("case: %d a), want: %s, got: %s", i, c.want.Hex(), have.Hex())
		}
		// Test via ToBlock
		if have := c.genesis.ToBlock().Hash(); have != c.want {
			t.Errorf("case: %d b), want: %s, got: %s", i, c.want.Hex(), have.Hex())
		}
	}
}

func TestGenesisCommit(t *testing.T) {
	genesis := &Genesis{
		BaseFee: big.NewInt(params.InitialBaseFee),
		Config:  params.TestChainConfig,
		// difficulty is nil
	}

	db := rawdb.NewMemoryDatabase()
	genesisBlock := genesis.MustCommit(db, triedb.NewDatabase(db, triedb.HashDefaults))

	if genesis.Difficulty != nil {
		t.Fatalf("assumption wrong")
	}

	// This value should have been set as default in the ToBlock method.
	if genesisBlock.Difficulty().Cmp(params.GenesisDifficulty) != 0 {
		t.Errorf("assumption wrong: want: %d, got: %v", params.GenesisDifficulty, genesisBlock.Difficulty())
	}
}

func TestReadWriteGenesisAlloc(t *testing.T) {
	var (
		db    = rawdb.NewMemoryDatabase()
		alloc = &types.GenesisAlloc{
			{1}: {Balance: big.NewInt(1), Storage: map[common.Hash]common.Hash{{1}: {1}}},
			{2}: {Balance: big.NewInt(2), Storage: map[common.Hash]common.Hash{{2}: {2}}},
		}
		hash, _ = hashAlloc(alloc, false)
	)
	blob, _ := json.Marshal(alloc)
	rawdb.WriteGenesisStateSpec(db, hash, blob)

	var reload types.GenesisAlloc
	err := reload.UnmarshalJSON(rawdb.ReadGenesisStateSpec(db, hash))
	if err != nil {
		t.Fatalf("Failed to load genesis state %v", err)
	}
	if len(reload) != len(*alloc) {
		t.Fatal("Unexpected genesis allocation")
	}
	for addr, account := range reload {
		want, ok := (*alloc)[addr]
		if !ok {
			t.Fatal("Account is not found")
		}
		if !reflect.DeepEqual(want, account) {
			t.Fatal("Unexpected account")
		}
	}
}

func newDbConfig(scheme string) *triedb.Config {
	if scheme == rawdb.HashScheme {
		return triedb.HashDefaults
	}
	config := *pathdb.Defaults
	config.NoAsyncFlush = true
	return &triedb.Config{PathDB: &config}
}

func TestBinaryGenesisCommit(t *testing.T) {
	var ubtTime uint64 = 0
	ubtConfig := &params.ChainConfig{
		ChainID:                 big.NewInt(1),
		SilaHomesteadBlock:          big.NewInt(0),
		DAOForkBlock:            nil,
		DAOForkSupport:          false,
		SIP150Block:             big.NewInt(0),
		SIP155Block:             big.NewInt(0),
		SIP158Block:             big.NewInt(0),
		SilaByzantiumBlock:          big.NewInt(0),
		SilaConstantinopleBlock:     big.NewInt(0),
		PetersburgBlock:         big.NewInt(0),
		SilaIstanbulBlock:           big.NewInt(0),
		MuirGlacierBlock:        big.NewInt(0),
		SilaBerlinBlock:             big.NewInt(0),
		SilaLondonBlock:             big.NewInt(0),
		ArrowGlacierBlock:       big.NewInt(0),
		GrayGlacierBlock:        big.NewInt(0),
		MergeNetsplitBlock:      nil,
		SilaShanghaiTime:            &ubtTime,
		SilaCancunTime:              &ubtTime,
		SilaPragueTime:              &ubtTime,
		SilaOsakaTime:               &ubtTime,
		UBTTime:                 &ubtTime,
		TerminalTotalDifficulty: big.NewInt(0),
		EnableUBTAtGenesis:      true,
		Silash:                  nil,
		Clique:                  nil,
		BlobScheduleConfig: &params.BlobScheduleConfig{
			SilaCancun: params.DefaultSilaCancunBlobConfig,
			SilaPrague: params.DefaultSilaPragueBlobConfig,
		},
	}

	genesis := &Genesis{
		BaseFee:    big.NewInt(params.InitialBaseFee),
		Config:     ubtConfig,
		Timestamp:  ubtTime,
		Difficulty: big.NewInt(0),
		Alloc: types.GenesisAlloc{
			{1}: {Balance: big.NewInt(1), Storage: map[common.Hash]common.Hash{{1}: {1}}},
		},
	}

	expected := common.FromHex("0870fd587c41dc778019de8c5cb3193fe4ef1f417976461952d3712ba39163f5")
	got := genesis.ToBlock().Root().Bytes()
	if !bytes.Equal(got, expected) {
		t.Fatalf("invalid genesis state root, expected %x, got %x", expected, got)
	}

	db := rawdb.NewMemoryDatabase()

	config := *pathdb.Defaults
	config.NoAsyncFlush = true

	triedb := triedb.NewDatabase(db, &triedb.Config{
		IsUBT:             true,
		PathDB:            &config,
		BinTrieGroupDepth: triedb.DefaultBinTrieGroupDepth,
	})
	block := genesis.MustCommit(db, triedb)
	if !bytes.Equal(block.Root().Bytes(), expected) {
		t.Fatalf("invalid genesis state root, expected %x, got %x", expected, block.Root())
	}

	// Test that the trie is a unified binary trie
	if !triedb.IsUBT() {
		t.Fatalf("expected trie to be a unified binary trie")
	}
	vdb := rawdb.NewTable(db, string(rawdb.VerklePrefix))
	if !rawdb.HasAccountTrieNode(vdb, nil) {
		t.Fatal("could not find node")
	}
}