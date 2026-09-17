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

package tests

import (
	"math/big"
	"testing"

	"github.com/sila-chain/go-sila/params"
)

var (
	mainnetChainConfig = params.ChainConfig{
		ChainID:        big.NewInt(1),
		SilaHomesteadBlock: big.NewInt(1150000),
		DAOForkBlock:   big.NewInt(1920000),
		DAOForkSupport: true,
		SIP150Block:    big.NewInt(2463000),
		SIP155Block:    big.NewInt(2675000),
		SIP158Block:    big.NewInt(2675000),
		SilaByzantiumBlock: big.NewInt(4370000),
	}

	ropstenChainConfig = params.ChainConfig{
		ChainID:                 big.NewInt(3),
		SilaHomesteadBlock:          big.NewInt(0),
		DAOForkBlock:            nil,
		DAOForkSupport:          true,
		SIP150Block:             big.NewInt(0),
		SIP155Block:             big.NewInt(10),
		SIP158Block:             big.NewInt(10),
		SilaByzantiumBlock:          big.NewInt(1_700_000),
		SilaConstantinopleBlock:     big.NewInt(4_230_000),
		PetersburgBlock:         big.NewInt(4_939_394),
		SilaIstanbulBlock:           big.NewInt(6_485_846),
		MuirGlacierBlock:        big.NewInt(7_117_117),
		SilaBerlinBlock:             big.NewInt(9_812_189),
		SilaLondonBlock:             big.NewInt(10_499_401),
		TerminalTotalDifficulty: new(big.Int).SetUint64(50_000_000_000_000_000),
	}
)

func TestDifficulty(t *testing.T) {
	t.Parallel()

	dt := new(testMatcher)
	// Not difficulty-tests
	dt.skipLoad("hexencodetest.*")
	dt.skipLoad("crypto.*")
	dt.skipLoad("blockgenesistest\\.json")
	dt.skipLoad("genesishashestest\\.json")
	dt.skipLoad("keyaddrtest\\.json")
	dt.skipLoad("txtest\\.json")

	// files are 2 years old, contains strange values
	dt.skipLoad("difficultyCustomSilaHomestead\\.json")

	dt.config("Ropsten", ropstenChainConfig)
	dt.config("Frontier", params.ChainConfig{})

	dt.config("SilaHomestead", params.ChainConfig{
		SilaHomesteadBlock: big.NewInt(0),
	})

	dt.config("SilaByzantium", params.ChainConfig{
		SilaByzantiumBlock: big.NewInt(0),
	})

	dt.config("Frontier", ropstenChainConfig)
	dt.config("MainNetwork", mainnetChainConfig)
	dt.config("CustomMainNetwork", mainnetChainConfig)
	dt.config("SilaConstantinople", params.ChainConfig{
		SilaConstantinopleBlock: big.NewInt(0),
	})
	dt.config("SIP2384", params.ChainConfig{
		MuirGlacierBlock: big.NewInt(0),
	})
	dt.config("SIP4345", params.ChainConfig{
		ArrowGlacierBlock: big.NewInt(0),
	})
	dt.config("SIP5133", params.ChainConfig{
		GrayGlacierBlock: big.NewInt(0),
	})
	dt.config("difficulty.json", mainnetChainConfig)

	dt.walk(t, difficultyTestDir, func(t *testing.T, name string, test *DifficultyTest) {
		cfg := dt.findConfig(t)
		if test.ParentDifficulty.Cmp(params.MinimumDifficulty) < 0 {
			t.Skip("difficulty below minimum")
			return
		}
		if err := dt.checkFailure(t, test.Run(cfg)); err != nil {
			t.Error(err)
		}
	})
}
