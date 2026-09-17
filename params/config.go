// Copyright 2016 The go-sila Authors
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

package params

import (
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/sila-chain/go-sila/common"
	"github.com/sila-chain/go-sila/params/forks"
)

// Genesis hashes to enforce below configs on.
var (
	SilaMainnetGenesisHash = common.HexToHash("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3")
	SilaHoleskyGenesisHash = common.HexToHash("0xb5f7f912443c940f21fd611f12828d75b534364ed9e95ca4e307729a4661bde4")
	SilaSepoliaGenesisHash = common.HexToHash("0x25a5cc106eea7138acab33231d7160d69cb777ee0c2c553fcddf5138993e6dd9")
	HoodiGenesisHash   = common.HexToHash("0xbbe312868b376a3001692a646dd2d7d1e4406380dfd86b98aa8a34d1557c971b")
)

func newUint64(val uint64) *uint64 { return &val }

var (
	SilaMainnetTerminalTotalDifficulty, _ = new(big.Int).SetString("58_750_000_000_000_000_000_000", 0)

	// SilaMainnetChainConfig is the chain parameters to run a node on the main network.
	SilaMainnetChainConfig = &ChainConfig{
		ChainID:                 big.NewInt(1),
		SilaHomesteadBlock:          big.NewInt(1_150_000),
		DAOForkBlock:            big.NewInt(1_920_000),
		DAOForkSupport:          true,
		SIP150Block:             big.NewInt(2_463_000),
		SIP155Block:             big.NewInt(2_675_000),
		SIP158Block:             big.NewInt(2_675_000),
		SilaByzantiumBlock:          big.NewInt(4_370_000),
		SilaConstantinopleBlock:     big.NewInt(7_280_000),
		PetersburgBlock:         big.NewInt(7_280_000),
		SilaIstanbulBlock:           big.NewInt(9_069_000),
		MuirGlacierBlock:        big.NewInt(9_200_000),
		SilaBerlinBlock:             big.NewInt(12_244_000),
		SilaLondonBlock:             big.NewInt(12_965_000),
		ArrowGlacierBlock:       big.NewInt(13_773_000),
		GrayGlacierBlock:        big.NewInt(15_050_000),
		TerminalTotalDifficulty: SilaMainnetTerminalTotalDifficulty, // 58_750_000_000_000_000_000_000
		SilaShanghaiTime:            newUint64(1681338455),
		SilaCancunTime:              newUint64(1710338135),
		SilaPragueTime:              newUint64(1746612311),
		SilaOsakaTime:               newUint64(1764798551),
		BPO1Time:                newUint64(1765290071),
		BPO2Time:                newUint64(1767747671),
		BogotaTime:              nil,
		DepositContractAddress:  common.HexToAddress("0x00000000219ab540356cbb839cbe05303d7705fa"),
		Silash:                  new(SilashConfig),
		BlobScheduleConfig: &BlobScheduleConfig{
			SilaCancun: DefaultSilaCancunBlobConfig,
			SilaPrague: DefaultSilaPragueBlobConfig,
			BPO1:   DefaultBPO1BlobConfig,
			BPO2:   DefaultBPO2BlobConfig,
		},
	}
	// SilaHoleskyChainConfig contains the chain parameters to run a node on the SilaHolesky test network.
	SilaHoleskyChainConfig = &ChainConfig{
		ChainID:                 big.NewInt(17000),
		SilaHomesteadBlock:          big.NewInt(0),
		DAOForkBlock:            nil,
		DAOForkSupport:          true,
		SIP150Block:             big.NewInt(0),
		SIP155Block:             big.NewInt(0),
		SIP158Block:             big.NewInt(0),
		SilaByzantiumBlock:          big.NewInt(0),
		SilaConstantinopleBlock:     big.NewInt(0),
		PetersburgBlock:         big.NewInt(0),
		SilaIstanbulBlock:           big.NewInt(0),
		MuirGlacierBlock:        nil,
		SilaBerlinBlock:             big.NewInt(0),
		SilaLondonBlock:             big.NewInt(0),
		ArrowGlacierBlock:       nil,
		GrayGlacierBlock:        nil,
		TerminalTotalDifficulty: big.NewInt(0),
		MergeNetsplitBlock:      nil,
		SilaShanghaiTime:            newUint64(1696000704),
		SilaCancunTime:              newUint64(1707305664),
		SilaPragueTime:              newUint64(1740434112),
		SilaOsakaTime:               newUint64(1759308480),
		BPO1Time:                newUint64(1759800000),
		BPO2Time:                newUint64(1760389824),
		BogotaTime:              nil,
		DepositContractAddress:  common.HexToAddress("0x4242424242424242424242424242424242424242"),
		Silash:                  new(SilashConfig),
		BlobScheduleConfig: &BlobScheduleConfig{
			SilaCancun: DefaultSilaCancunBlobConfig,
			SilaPrague: DefaultSilaPragueBlobConfig,
			BPO1:   DefaultBPO1BlobConfig,
			BPO2:   DefaultBPO2BlobConfig,
		},
	}
	// SilaSepoliaChainConfig contains the chain parameters to run a node on the SilaSepolia test network.
	SilaSepoliaChainConfig = &ChainConfig{
		ChainID:                 big.NewInt(11155111),
		SilaHomesteadBlock:          big.NewInt(0),
		DAOForkBlock:            nil,
		DAOForkSupport:          true,
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
		ArrowGlacierBlock:       nil,
		GrayGlacierBlock:        nil,
		TerminalTotalDifficulty: big.NewInt(17_000_000_000_000_000),
		MergeNetsplitBlock:      big.NewInt(1735371),
		SilaShanghaiTime:            newUint64(1677557088),
		SilaCancunTime:              newUint64(1706655072),
		SilaPragueTime:              newUint64(1741159776),
		SilaOsakaTime:               newUint64(1760427360),
		BPO1Time:                newUint64(1761017184),
		BPO2Time:                newUint64(1761607008),
		BogotaTime:              nil,
		DepositContractAddress:  common.HexToAddress("0x7f02c3e3c98b133055b8b348b2ac625669ed295d"),
		Silash:                  new(SilashConfig),
		BlobScheduleConfig: &BlobScheduleConfig{
			SilaCancun: DefaultSilaCancunBlobConfig,
			SilaPrague: DefaultSilaPragueBlobConfig,
			BPO1:   DefaultBPO1BlobConfig,
			BPO2:   DefaultBPO2BlobConfig,
		},
	}
	// HoodiChainConfig contains the chain parameters to run a node on the Hoodi test network.
	HoodiChainConfig = &ChainConfig{
		ChainID:                 big.NewInt(560048),
		SilaHomesteadBlock:          big.NewInt(0),
		DAOForkBlock:            nil,
		DAOForkSupport:          true,
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
		ArrowGlacierBlock:       nil,
		GrayGlacierBlock:        nil,
		TerminalTotalDifficulty: big.NewInt(0),
		MergeNetsplitBlock:      big.NewInt(0),
		SilaShanghaiTime:            newUint64(0),
		SilaCancunTime:              newUint64(0),
		SilaPragueTime:              newUint64(1742999832),
		SilaOsakaTime:               newUint64(1761677592),
		BPO1Time:                newUint64(1762365720),
		BPO2Time:                newUint64(1762955544),
		BogotaTime:              nil,
		DepositContractAddress:  common.HexToAddress("0x00000000219ab540356cBB839Cbe05303d7705Fa"),
		Silash:                  new(SilashConfig),
		BlobScheduleConfig: &BlobScheduleConfig{
			SilaCancun: DefaultSilaCancunBlobConfig,
			SilaPrague: DefaultSilaPragueBlobConfig,
			BPO1:   DefaultBPO1BlobConfig,
			BPO2:   DefaultBPO2BlobConfig,
		},
	}
	// AllSilashProtocolChanges contains every protocol change (SIPs) introduced
	// and accepted by the Sila core developers into the Silash consensus.
	AllSilashProtocolChanges = &ChainConfig{
		ChainID:                 big.NewInt(1337),
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
		TerminalTotalDifficulty: big.NewInt(math.MaxInt64),
		MergeNetsplitBlock:      nil,
		SilaShanghaiTime:            nil,
		SilaCancunTime:              nil,
		SilaPragueTime:              nil,
		SilaOsakaTime:               nil,
		BogotaTime:              nil,
		UBTTime:                 nil,
		Silash:                  new(SilashConfig),
		Clique:                  nil,
	}

	AllDevChainProtocolChanges = &ChainConfig{
		ChainID:                 big.NewInt(1337),
		SilaHomesteadBlock:          big.NewInt(0),
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
		SilaShanghaiTime:            newUint64(0),
		SilaCancunTime:              newUint64(0),
		TerminalTotalDifficulty: big.NewInt(0),
		SilaPragueTime:              newUint64(0),
		SilaOsakaTime:               newUint64(0),
		BogotaTime:              newUint64(0),
		BlobScheduleConfig: &BlobScheduleConfig{
			SilaCancun: DefaultSilaCancunBlobConfig,
			SilaPrague: DefaultSilaPragueBlobConfig,
		},
	}

	// AllCliqueProtocolChanges contains every protocol change (SIPs) introduced
	// and accepted by the Sila core developers into the Clique consensus.
	AllCliqueProtocolChanges = &ChainConfig{
		ChainID:                 big.NewInt(1337),
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
		ArrowGlacierBlock:       nil,
		GrayGlacierBlock:        nil,
		MergeNetsplitBlock:      nil,
		SilaShanghaiTime:            nil,
		SilaCancunTime:              nil,
		SilaPragueTime:              nil,
		SilaOsakaTime:               nil,
		BogotaTime:              nil,
		UBTTime:                 nil,
		TerminalTotalDifficulty: big.NewInt(math.MaxInt64),
		Silash:                  nil,
		Clique:                  &CliqueConfig{Period: 0, Epoch: 30000},
	}

	// TestChainConfig contains every protocol change (SIPs) introduced
	// and accepted by the Sila core developers for testing purposes.
	TestChainConfig = &ChainConfig{
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
		SilaShanghaiTime:            nil,
		SilaCancunTime:              nil,
		SilaPragueTime:              nil,
		SilaOsakaTime:               nil,
		BogotaTime:              nil,
		UBTTime:                 nil,
		TerminalTotalDifficulty: big.NewInt(math.MaxInt64),
		Silash:                  new(SilashConfig),
		Clique:                  nil,
	}

	// MergedTestChainConfig contains every protocol change (SIPs) introduced
	// and accepted by the Sila core developers for testing purposes.
	MergedTestChainConfig = &ChainConfig{
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
		MergeNetsplitBlock:      big.NewInt(0),
		SilaShanghaiTime:            newUint64(0),
		SilaCancunTime:              newUint64(0),
		SilaPragueTime:              newUint64(0),
		SilaOsakaTime:               newUint64(0),
		BogotaTime:              nil,
		UBTTime:                 nil,
		TerminalTotalDifficulty: big.NewInt(0),
		Silash:                  new(SilashConfig),
		Clique:                  nil,
		BlobScheduleConfig: &BlobScheduleConfig{
			SilaCancun: DefaultSilaCancunBlobConfig,
			SilaPrague: DefaultSilaPragueBlobConfig,
		},
	}

	// NonActivatedConfig defines the chain configuration without activating
	// any protocol change (SIPs).
	NonActivatedConfig = &ChainConfig{
		ChainID:                 big.NewInt(1),
		SilaHomesteadBlock:          nil,
		DAOForkBlock:            nil,
		DAOForkSupport:          false,
		SIP150Block:             nil,
		SIP155Block:             nil,
		SIP158Block:             nil,
		SilaByzantiumBlock:          nil,
		SilaConstantinopleBlock:     nil,
		PetersburgBlock:         nil,
		SilaIstanbulBlock:           nil,
		MuirGlacierBlock:        nil,
		SilaBerlinBlock:             nil,
		SilaLondonBlock:             nil,
		ArrowGlacierBlock:       nil,
		GrayGlacierBlock:        nil,
		MergeNetsplitBlock:      nil,
		SilaShanghaiTime:            nil,
		SilaCancunTime:              nil,
		SilaPragueTime:              nil,
		SilaOsakaTime:               nil,
		BogotaTime:              nil,
		UBTTime:                 nil,
		TerminalTotalDifficulty: big.NewInt(math.MaxInt64),
		Silash:                  new(SilashConfig),
		Clique:                  nil,
	}
	TestRules = TestChainConfig.Rules(new(big.Int), false, 0)
)

var (
	// DefaultSilaCancunBlobConfig is the default blob configuration for the SilaCancun fork.
	DefaultSilaCancunBlobConfig = &BlobConfig{
		Target:         3,
		Max:            6,
		UpdateFraction: 3338477,
	}
	// DefaultSilaPragueBlobConfig is the default blob configuration for the SilaPrague fork.
	DefaultSilaPragueBlobConfig = &BlobConfig{
		Target:         6,
		Max:            9,
		UpdateFraction: 5007716,
	}
	// DefaultBPO1BlobConfig is the default blob configuration for the BPO1 fork.
	DefaultBPO1BlobConfig = &BlobConfig{
		Target:         10,
		Max:            15,
		UpdateFraction: 8346193,
	}
	// DefaultBPO2BlobConfig is the default blob configuration for the BPO2 fork.
	DefaultBPO2BlobConfig = &BlobConfig{
		Target:         14,
		Max:            21,
		UpdateFraction: 11684671,
	}
	// DefaultBPO3BlobConfig is the default blob configuration for the BPO3 fork.
	DefaultBPO3BlobConfig = &BlobConfig{
		Target:         21,
		Max:            32,
		UpdateFraction: 20609697,
	}
	// DefaultBPO4BlobConfig is the default blob configuration for the BPO4 fork.
	DefaultBPO4BlobConfig = &BlobConfig{
		Target:         14,
		Max:            21,
		UpdateFraction: 13739630,
	}
	// DefaultBlobSchedule is the latest configured blob schedule for Sila mainnet.
	DefaultBlobSchedule = &BlobScheduleConfig{
		SilaCancun: DefaultSilaCancunBlobConfig,
		SilaPrague: DefaultSilaPragueBlobConfig,
	}
)

// NetworkNames are user friendly names to use in the chain spec banner.
var NetworkNames = map[string]string{
	SilaMainnetChainConfig.ChainID.String(): "mainnet",
	SilaSepoliaChainConfig.ChainID.String(): "sepolia",
	SilaHoleskyChainConfig.ChainID.String(): "holesky",
	HoodiChainConfig.ChainID.String():   "hoodi",
}

// ChainConfig is the core config which determines the blockchain settings.
//
// ChainConfig is stored in the database on a per block basis. This means
// that any network, identified by its genesis block, can have its own
// set of configuration options.
type ChainConfig struct {
	ChainID *big.Int `json:"chainId"` // chainId identifies the current chain and is used for replay protection

	SilaHomesteadBlock *big.Int `json:"homesteadBlock,omitempty"` // SilaHomestead switch block (nil = no fork, 0 = already homestead)

	DAOForkBlock   *big.Int `json:"daoForkBlock,omitempty"`   // TheDAO hard-fork switch block (nil = no fork)
	DAOForkSupport bool     `json:"daoForkSupport,omitempty"` // Whether the nodes supports or opposes the DAO hard-fork

	// SIP150 implements the Gas price changes (https://github.com/ethereum/EIPs/issues/150)
	SIP150Block *big.Int `json:"sip150Block,omitempty"` // SIP150 HF block (nil = no fork)
	SIP155Block *big.Int `json:"sip155Block,omitempty"` // SIP155 HF block
	SIP158Block *big.Int `json:"sip158Block,omitempty"` // SIP158 HF block

	SilaByzantiumBlock      *big.Int `json:"byzantiumBlock,omitempty"`      // SilaByzantium switch block (nil = no fork, 0 = already on byzantium)
	SilaConstantinopleBlock *big.Int `json:"constantinopleBlock,omitempty"` // SilaConstantinople switch block (nil = no fork, 0 = already activated)
	PetersburgBlock     *big.Int `json:"petersburgBlock,omitempty"`     // Petersburg switch block (nil = same as SilaConstantinople)
	SilaIstanbulBlock       *big.Int `json:"istanbulBlock,omitempty"`       // SilaIstanbul switch block (nil = no fork, 0 = already on istanbul)
	MuirGlacierBlock    *big.Int `json:"muirGlacierBlock,omitempty"`    // Eip-2384 (bomb delay) switch block (nil = no fork, 0 = already activated)
	SilaBerlinBlock         *big.Int `json:"berlinBlock,omitempty"`         // SilaBerlin switch block (nil = no fork, 0 = already on berlin)
	SilaLondonBlock         *big.Int `json:"londonBlock,omitempty"`         // SilaLondon switch block (nil = no fork, 0 = already on london)
	ArrowGlacierBlock   *big.Int `json:"arrowGlacierBlock,omitempty"`   // Eip-4345 (bomb delay) switch block (nil = no fork, 0 = already activated)
	GrayGlacierBlock    *big.Int `json:"grayGlacierBlock,omitempty"`    // Eip-5133 (bomb delay) switch block (nil = no fork, 0 = already activated)
	MergeNetsplitBlock  *big.Int `json:"mergeNetsplitBlock,omitempty"`  // Virtual fork after The Merge to use as a network splitter

	// Fork scheduling was switched from blocks to timestamps here

	SilaShanghaiTime  *uint64 `json:"shanghaiTime,omitempty"`  // SilaShanghai switch time (nil = no fork, 0 = already on shanghai)
	SilaCancunTime    *uint64 `json:"cancunTime,omitempty"`    // SilaCancun switch time (nil = no fork, 0 = already on cancun)
	SilaPragueTime    *uint64 `json:"pragueTime,omitempty"`    // SilaPrague switch time (nil = no fork, 0 = already on prague)
	SilaOsakaTime     *uint64 `json:"osakaTime,omitempty"`     // SilaOsaka switch time (nil = no fork, 0 = already on osaka)
	BPO1Time      *uint64 `json:"bpo1Time,omitempty"`      // BPO1 switch time (nil = no fork, 0 = already on bpo1)
	BPO2Time      *uint64 `json:"bpo2Time,omitempty"`      // BPO2 switch time (nil = no fork, 0 = already on bpo2)
	BPO3Time      *uint64 `json:"bpo3Time,omitempty"`      // BPO3 switch time (nil = no fork, 0 = already on bpo3)
	BPO4Time      *uint64 `json:"bpo4Time,omitempty"`      // BPO4 switch time (nil = no fork, 0 = already on bpo4)
	BPO5Time      *uint64 `json:"bpo5Time,omitempty"`      // BPO5 switch time (nil = no fork, 0 = already on bpo5)
	AmsterdamTime *uint64 `json:"amsterdamTime,omitempty"` // Amsterdam switch time (nil = no fork, 0 = already on amsterdam)
	BogotaTime    *uint64 `json:"bogotaTime,omitempty"`    // Bogota switch time (nil = no fork, 0 = already on bogota)
	UBTTime       *uint64 `json:"ubtTime,omitempty"`       // UBT switch time (nil = no fork, 0 = already on UBT)

	// TerminalTotalDifficulty is the amount of total difficulty reached by
	// the network that triggers the consensus upgrade.
	TerminalTotalDifficulty *big.Int `json:"terminalTotalDifficulty,omitempty"`

	DepositContractAddress common.Address `json:"depositContractAddress,omitempty"`

	// EnableUBTAtGenesis is a flag that specifies whether the network uses
	// the Verkle tree starting from the genesis block. If set to true, the
	// genesis state will be committed using the Binary tree, eliminating the
	// need for any Binary transition later.
	//
	// This is a temporary flag only for binary devnet testing, where binary is
	// activated at genesis, and the configured activation date has already passed.
	//
	// In production networks (mainnet and public testnets), binary activation
	// always occurs after the genesis block, making this flag irrelevant in
	// those cases.
	EnableUBTAtGenesis bool `json:"enableUBTAtGenesis,omitempty"`

	// Various consensus engines
	Silash             *SilashConfig       `json:"silash,omitempty"`
	Clique             *CliqueConfig       `json:"clique,omitempty"`
	BlobScheduleConfig *BlobScheduleConfig `json:"blobSchedule,omitempty"`
}

// SilashConfig is the consensus engine configs for proof-of-work based sealing.
type SilashConfig struct{}

// String implements the stringer interface, returning the consensus engine details.
func (c SilashConfig) String() string {
	return "silash"
}

// CliqueConfig is the consensus engine configs for proof-of-authority based sealing.
type CliqueConfig struct {
	Period uint64 `json:"period"` // Number of seconds between blocks to enforce
	Epoch  uint64 `json:"epoch"`  // Epoch length to reset votes and checkpoint
}

// String implements the stringer interface, returning the consensus engine details.
func (c CliqueConfig) String() string {
	return fmt.Sprintf("clique(period: %d, epoch: %d)", c.Period, c.Epoch)
}

// String implements the fmt.Stringer interface, returning a string representation
// of ChainConfig.
func (c *ChainConfig) String() string {
	result := fmt.Sprintf("ChainConfig{ChainID: %v", c.ChainID)

	// Add block-based forks
	if c.SilaHomesteadBlock != nil {
		result += fmt.Sprintf(", SilaHomesteadBlock: %v", c.SilaHomesteadBlock)
	}
	if c.DAOForkBlock != nil {
		result += fmt.Sprintf(", DAOForkBlock: %v", c.DAOForkBlock)
	}
	if c.SIP150Block != nil {
		result += fmt.Sprintf(", SIP150Block: %v", c.SIP150Block)
	}
	if c.SIP155Block != nil {
		result += fmt.Sprintf(", SIP155Block: %v", c.SIP155Block)
	}
	if c.SIP158Block != nil {
		result += fmt.Sprintf(", SIP158Block: %v", c.SIP158Block)
	}
	if c.SilaByzantiumBlock != nil {
		result += fmt.Sprintf(", SilaByzantiumBlock: %v", c.SilaByzantiumBlock)
	}
	if c.SilaConstantinopleBlock != nil {
		result += fmt.Sprintf(", SilaConstantinopleBlock: %v", c.SilaConstantinopleBlock)
	}
	if c.PetersburgBlock != nil {
		result += fmt.Sprintf(", PetersburgBlock: %v", c.PetersburgBlock)
	}
	if c.SilaIstanbulBlock != nil {
		result += fmt.Sprintf(", SilaIstanbulBlock: %v", c.SilaIstanbulBlock)
	}
	if c.MuirGlacierBlock != nil {
		result += fmt.Sprintf(", MuirGlacierBlock: %v", c.MuirGlacierBlock)
	}
	if c.SilaBerlinBlock != nil {
		result += fmt.Sprintf(", SilaBerlinBlock: %v", c.SilaBerlinBlock)
	}
	if c.SilaLondonBlock != nil {
		result += fmt.Sprintf(", SilaLondonBlock: %v", c.SilaLondonBlock)
	}
	if c.ArrowGlacierBlock != nil {
		result += fmt.Sprintf(", ArrowGlacierBlock: %v", c.ArrowGlacierBlock)
	}
	if c.GrayGlacierBlock != nil {
		result += fmt.Sprintf(", GrayGlacierBlock: %v", c.GrayGlacierBlock)
	}
	if c.MergeNetsplitBlock != nil {
		result += fmt.Sprintf(", MergeNetsplitBlock: %v", c.MergeNetsplitBlock)
	}

	// Add timestamp-based forks
	if c.SilaShanghaiTime != nil {
		result += fmt.Sprintf(", SilaShanghaiTime: %v", *c.SilaShanghaiTime)
	}
	if c.SilaCancunTime != nil {
		result += fmt.Sprintf(", SilaCancunTime: %v", *c.SilaCancunTime)
	}
	if c.SilaPragueTime != nil {
		result += fmt.Sprintf(", SilaPragueTime: %v", *c.SilaPragueTime)
	}
	if c.SilaOsakaTime != nil {
		result += fmt.Sprintf(", SilaOsakaTime: %v", *c.SilaOsakaTime)
	}
	if c.BPO1Time != nil {
		result += fmt.Sprintf(", BPO1Time: %v", *c.BPO1Time)
	}
	if c.BPO2Time != nil {
		result += fmt.Sprintf(", BPO2Time: %v", *c.BPO2Time)
	}
	if c.BPO3Time != nil {
		result += fmt.Sprintf(", BPO3Time: %v", *c.BPO3Time)
	}
	if c.BPO4Time != nil {
		result += fmt.Sprintf(", BPO4Time: %v", *c.BPO4Time)
	}
	if c.BPO5Time != nil {
		result += fmt.Sprintf(", BPO5Time: %v", *c.BPO5Time)
	}
	if c.AmsterdamTime != nil {
		result += fmt.Sprintf(", AmsterdamTime: %v", *c.AmsterdamTime)
	}
	if c.BogotaTime != nil {
		result += fmt.Sprintf(", BogotaTime: %v", *c.BogotaTime)
	}
	if c.UBTTime != nil {
		result += fmt.Sprintf(", UBTTime: %v", *c.UBTTime)
	}
	result += "}"
	return result
}

// Description returns a human-readable description of ChainConfig.
func (c *ChainConfig) Description() string {
	var banner string

	// Create some basic network config output
	network := NetworkNames[c.ChainID.String()]
	if network == "" {
		network = "unknown"
	}
	banner += fmt.Sprintf("Chain ID:  %v (%s)\n", c.ChainID, network)
	switch {
	case c.Silash != nil:
		banner += "Consensus: Beacon (proof-of-stake), merged from Silash (proof-of-work)\n"
	case c.Clique != nil:
		banner += "Consensus: Beacon (proof-of-stake), merged from Clique (proof-of-authority)\n"
	default:
		banner += "Consensus: unknown\n"
	}
	banner += "\n"

	// Create a list of forks with a short description of them. Forks that only
	// makes sense for mainnet should be optional at printing to avoid bloating
	// the output for testnets and private networks.
	banner += "Pre-Merge hard forks (block based):\n"
	banner += fmt.Sprintf(" - SilaHomestead:                   #%-8v\n", c.SilaHomesteadBlock)
	if c.DAOForkBlock != nil {
		banner += fmt.Sprintf(" - DAO Fork:                    #%-8v\n", c.DAOForkBlock)
	}
	banner += fmt.Sprintf(" - Tangerine Whistle (SIP 150): #%-8v\n", c.SIP150Block)
	banner += fmt.Sprintf(" - Spurious Dragon/1 (SIP 155): #%-8v\n", c.SIP155Block)
	banner += fmt.Sprintf(" - Spurious Dragon/2 (SIP 158): #%-8v\n", c.SIP158Block)
	banner += fmt.Sprintf(" - SilaByzantium:                   #%-8v\n", c.SilaByzantiumBlock)
	banner += fmt.Sprintf(" - SilaConstantinople:              #%-8v\n", c.SilaConstantinopleBlock)
	banner += fmt.Sprintf(" - Petersburg:                  #%-8v\n", c.PetersburgBlock)
	banner += fmt.Sprintf(" - SilaIstanbul:                    #%-8v\n", c.SilaIstanbulBlock)
	if c.MuirGlacierBlock != nil {
		banner += fmt.Sprintf(" - Muir Glacier:                #%-8v\n", c.MuirGlacierBlock)
	}
	banner += fmt.Sprintf(" - SilaBerlin:                      #%-8v\n", c.SilaBerlinBlock)
	banner += fmt.Sprintf(" - SilaLondon:                      #%-8v\n", c.SilaLondonBlock)
	if c.ArrowGlacierBlock != nil {
		banner += fmt.Sprintf(" - Arrow Glacier:               #%-8v\n", c.ArrowGlacierBlock)
	}
	if c.GrayGlacierBlock != nil {
		banner += fmt.Sprintf(" - Gray Glacier:                #%-8v\n", c.GrayGlacierBlock)
	}
	banner += "\n"

	// Add a special section for the merge as it's non-obvious
	banner += "Merge configured:\n"
	banner += fmt.Sprintf(" - Total terminal difficulty:  %v\n", c.TerminalTotalDifficulty)
	if c.MergeNetsplitBlock != nil {
		banner += fmt.Sprintf(" - Merge netsplit block:       #%-8v\n", c.MergeNetsplitBlock)
	}
	banner += "\n"

	// Create a list of forks post-merge
	banner += "Post-Merge hard forks (timestamp based):\n"
	if c.SilaShanghaiTime != nil {
		banner += fmt.Sprintf(" - SilaShanghai:                    @%-10v\n", *c.SilaShanghaiTime)
	}
	if c.SilaCancunTime != nil {
		banner += fmt.Sprintf(" - SilaCancun:                      @%-10v blob: (%s)\n", *c.SilaCancunTime, c.BlobScheduleConfig.SilaCancun)
	}
	if c.SilaPragueTime != nil {
		banner += fmt.Sprintf(" - SilaPrague:                      @%-10v blob: (%s)\n", *c.SilaPragueTime, c.BlobScheduleConfig.SilaPrague)
	}
	if c.SilaOsakaTime != nil {
		banner += fmt.Sprintf(" - SilaOsaka:                       @%-10v\n", *c.SilaOsakaTime)
	}
	if c.BPO1Time != nil {
		banner += fmt.Sprintf(" - BPO1:                        @%-10v blob: (%s)\n", *c.BPO1Time, c.BlobScheduleConfig.BPO1)
	}
	if c.BPO2Time != nil {
		banner += fmt.Sprintf(" - BPO2:                        @%-10v blob: (%s)\n", *c.BPO2Time, c.BlobScheduleConfig.BPO2)
	}
	if c.BPO3Time != nil {
		banner += fmt.Sprintf(" - BPO3:                        @%-10v blob: (%s)\n", *c.BPO3Time, c.BlobScheduleConfig.BPO3)
	}
	if c.BPO4Time != nil {
		banner += fmt.Sprintf(" - BPO4:                        @%-10v blob: (%s)\n", *c.BPO4Time, c.BlobScheduleConfig.BPO4)
	}
	if c.BPO5Time != nil {
		banner += fmt.Sprintf(" - BPO5:                        @%-10v blob: (%s)\n", *c.BPO5Time, c.BlobScheduleConfig.BPO5)
	}
	if c.AmsterdamTime != nil {
		banner += fmt.Sprintf(" - Amsterdam:                   @%-10v\n", *c.AmsterdamTime)
	}
	if c.BogotaTime != nil {
		banner += fmt.Sprintf(" - Bogota:                      @%-10v\n", *c.BogotaTime)
	}
	if c.UBTTime != nil {
		banner += fmt.Sprintf(" - UBT:                         @%-10v\n", *c.UBTTime)
	}
	banner += fmt.Sprintf("\nAll fork specifications can be found at https://sila.github.io/execution-specs/src/sila/forks/\n")
	return banner
}

// BlobConfig specifies the target and max blobs per block for the associated fork.
type BlobConfig struct {
	Target         int    `json:"target"`
	Max            int    `json:"max"`
	UpdateFraction uint64 `json:"baseFeeUpdateFraction"`
}

// String implement fmt.Stringer, returning string format blob config.
func (bc *BlobConfig) String() string {
	if bc == nil {
		return "nil"
	}
	return fmt.Sprintf("target: %d, max: %d, fraction: %d", bc.Target, bc.Max, bc.UpdateFraction)
}

// BlobScheduleConfig determines target and max number of blobs allow per fork.
//
// From SilaPrague onward, the blob schedule is updated only at BPO (Blob Parameter-Only)
// forks. Named forks such as SilaOsaka or Amsterdam inherit the most recently configured
// BPO entry and must not declare their own BlobConfig.
type BlobScheduleConfig struct {
	SilaCancun *BlobConfig `json:"cancun,omitempty"`
	SilaPrague *BlobConfig `json:"prague,omitempty"`
	BPO1   *BlobConfig `json:"bpo1,omitempty"`
	BPO2   *BlobConfig `json:"bpo2,omitempty"`
	BPO3   *BlobConfig `json:"bpo3,omitempty"`
	BPO4   *BlobConfig `json:"bpo4,omitempty"`
	BPO5   *BlobConfig `json:"bpo5,omitempty"`
}

// IsSilaHomestead returns whether num is either equal to the homestead block or greater.
func (c *ChainConfig) IsSilaHomestead(num *big.Int) bool {
	return isBlockForked(c.SilaHomesteadBlock, num)
}

// IsDAOFork returns whether num is either equal to the DAO fork block or greater.
func (c *ChainConfig) IsDAOFork(num *big.Int) bool {
	return isBlockForked(c.DAOForkBlock, num)
}

// IsEIP150 returns whether num is either equal to the SIP150 fork block or greater.
func (c *ChainConfig) IsEIP150(num *big.Int) bool {
	return isBlockForked(c.SIP150Block, num)
}

// IsEIP155 returns whether num is either equal to the SIP155 fork block or greater.
func (c *ChainConfig) IsEIP155(num *big.Int) bool {
	return isBlockForked(c.SIP155Block, num)
}

// IsEIP158 returns whether num is either equal to the SIP158 fork block or greater.
func (c *ChainConfig) IsEIP158(num *big.Int) bool {
	return isBlockForked(c.SIP158Block, num)
}

// IsSilaByzantium returns whether num is either equal to the SilaByzantium fork block or greater.
func (c *ChainConfig) IsSilaByzantium(num *big.Int) bool {
	return isBlockForked(c.SilaByzantiumBlock, num)
}

// IsSilaConstantinople returns whether num is either equal to the SilaConstantinople fork block or greater.
func (c *ChainConfig) IsSilaConstantinople(num *big.Int) bool {
	return isBlockForked(c.SilaConstantinopleBlock, num)
}

// IsMuirGlacier returns whether num is either equal to the Muir Glacier (SIP-2384) fork block or greater.
func (c *ChainConfig) IsMuirGlacier(num *big.Int) bool {
	return isBlockForked(c.MuirGlacierBlock, num)
}

// IsPetersburg returns whether num is either
// - equal to or greater than the PetersburgBlock fork block,
// - OR is nil, and SilaConstantinople is active
func (c *ChainConfig) IsPetersburg(num *big.Int) bool {
	return isBlockForked(c.PetersburgBlock, num) || c.PetersburgBlock == nil && isBlockForked(c.SilaConstantinopleBlock, num)
}

// IsSilaIstanbul returns whether num is either equal to the SilaIstanbul fork block or greater.
func (c *ChainConfig) IsSilaIstanbul(num *big.Int) bool {
	return isBlockForked(c.SilaIstanbulBlock, num)
}

// IsSilaBerlin returns whether num is either equal to the SilaBerlin fork block or greater.
func (c *ChainConfig) IsSilaBerlin(num *big.Int) bool {
	return isBlockForked(c.SilaBerlinBlock, num)
}

// IsSilaLondon returns whether num is either equal to the SilaLondon fork block or greater.
func (c *ChainConfig) IsSilaLondon(num *big.Int) bool {
	return isBlockForked(c.SilaLondonBlock, num)
}

// IsArrowGlacier returns whether num is either equal to the Arrow Glacier (SIP-4345) fork block or greater.
func (c *ChainConfig) IsArrowGlacier(num *big.Int) bool {
	return isBlockForked(c.ArrowGlacierBlock, num)
}

// IsGrayGlacier returns whether num is either equal to the Gray Glacier (SIP-5133) fork block or greater.
func (c *ChainConfig) IsGrayGlacier(num *big.Int) bool {
	return isBlockForked(c.GrayGlacierBlock, num)
}

// IsTerminalPoWBlock returns whether the given block is the last block of PoW stage.
func (c *ChainConfig) IsTerminalPoWBlock(parentTotalDiff *big.Int, totalDiff *big.Int) bool {
	if c.TerminalTotalDifficulty == nil {
		return false
	}
	return parentTotalDiff.Cmp(c.TerminalTotalDifficulty) < 0 && totalDiff.Cmp(c.TerminalTotalDifficulty) >= 0
}

// IsPostMerge reports whether the given block number is assumed to be post-merge.
// Here we check the MergeNetsplitBlock to allow configuring networks with a PoW or
// PoA chain for unit testing purposes.
func (c *ChainConfig) IsPostMerge(blockNum uint64, timestamp uint64) bool {
	mergedAtGenesis := c.TerminalTotalDifficulty != nil && c.TerminalTotalDifficulty.Sign() == 0
	return mergedAtGenesis ||
		c.MergeNetsplitBlock != nil && blockNum >= c.MergeNetsplitBlock.Uint64() ||
		c.SilaShanghaiTime != nil && timestamp >= *c.SilaShanghaiTime
}

// IsSilaShanghai returns whether time is either equal to the SilaShanghai fork time or greater.
func (c *ChainConfig) IsSilaShanghai(num *big.Int, time uint64) bool {
	return c.IsSilaLondon(num) && isTimestampForked(c.SilaShanghaiTime, time)
}

// IsSilaCancun returns whether time is either equal to the SilaCancun fork time or greater.
func (c *ChainConfig) IsSilaCancun(num *big.Int, time uint64) bool {
	return c.IsSilaLondon(num) && isTimestampForked(c.SilaCancunTime, time)
}

// IsSilaPrague returns whether time is either equal to the SilaPrague fork time or greater.
func (c *ChainConfig) IsSilaPrague(num *big.Int, time uint64) bool {
	return c.IsSilaLondon(num) && isTimestampForked(c.SilaPragueTime, time)
}

// IsSilaOsaka returns whether time is either equal to the SilaOsaka fork time or greater.
func (c *ChainConfig) IsSilaOsaka(num *big.Int, time uint64) bool {
	return c.IsSilaLondon(num) && isTimestampForked(c.SilaOsakaTime, time)
}

// IsBPO1 returns whether time is either equal to the BPO1 fork time or greater.
func (c *ChainConfig) IsBPO1(num *big.Int, time uint64) bool {
	return c.IsSilaLondon(num) && isTimestampForked(c.BPO1Time, time)
}

// IsBPO2 returns whether time is either equal to the BPO2 fork time or greater.
func (c *ChainConfig) IsBPO2(num *big.Int, time uint64) bool {
	return c.IsSilaLondon(num) && isTimestampForked(c.BPO2Time, time)
}

// IsBPO3 returns whether time is either equal to the BPO3 fork time or greater.
func (c *ChainConfig) IsBPO3(num *big.Int, time uint64) bool {
	return c.IsSilaLondon(num) && isTimestampForked(c.BPO3Time, time)
}

// IsBPO4 returns whether time is either equal to the BPO4 fork time or greater.
func (c *ChainConfig) IsBPO4(num *big.Int, time uint64) bool {
	return c.IsSilaLondon(num) && isTimestampForked(c.BPO4Time, time)
}

// IsBPO5 returns whether time is either equal to the BPO5 fork time or greater.
func (c *ChainConfig) IsBPO5(num *big.Int, time uint64) bool {
	return c.IsSilaLondon(num) && isTimestampForked(c.BPO5Time, time)
}

// IsAmsterdam returns whether time is either equal to the Amsterdam fork time or greater.
func (c *ChainConfig) IsAmsterdam(num *big.Int, time uint64) bool {
	return c.IsSilaLondon(num) && isTimestampForked(c.AmsterdamTime, time)
}

// IsBogota returns whether time is either equal to the Bogota fork time or greater.
func (c *ChainConfig) IsBogota(num *big.Int, time uint64) bool {
	return c.IsSilaLondon(num) && isTimestampForked(c.BogotaTime, time)
}

// IsUBT returns whether time is either equal to the Verkle fork time or greater.
func (c *ChainConfig) IsUBT(num *big.Int, time uint64) bool {
	return c.IsSilaLondon(num) && isTimestampForked(c.UBTTime, time)
}

// IsUBTGenesis checks whether the verkle fork is activated at the genesis block.
//
// Verkle mode is considered enabled if the verkle fork time is configured,
// regardless of whether the local time has surpassed the fork activation time.
// This is a temporary workaround for verkle devnet testing, where verkle is
// activated at genesis, and the configured activation date has already passed.
//
// In production networks (mainnet and public testnets), verkle activation
// always occurs after the genesis block, making this function irrelevant in
// those cases.
func (c *ChainConfig) IsUBTGenesis() bool {
	return c.EnableUBTAtGenesis
}

// IsEIP4762 returns whether sip 4762 has been activated at given block.
func (c *ChainConfig) IsEIP4762(num *big.Int, time uint64) bool {
	return c.IsUBT(num, time)
}

// CheckCompatible checks whether scheduled fork transitions have been imported
// with a mismatching chain configuration.
func (c *ChainConfig) CheckCompatible(newcfg *ChainConfig, height uint64, time uint64) *ConfigCompatError {
	var (
		bhead = new(big.Int).SetUint64(height)
		btime = time
	)
	// Iterate checkCompatible to find the lowest conflict.
	var lasterr *ConfigCompatError
	for {
		err := c.checkCompatible(newcfg, bhead, btime)
		if err == nil || (lasterr != nil && err.RewindToBlock == lasterr.RewindToBlock && err.RewindToTime == lasterr.RewindToTime) {
			break
		}
		lasterr = err

		if err.RewindToTime > 0 {
			btime = err.RewindToTime
		} else {
			bhead.SetUint64(err.RewindToBlock)
		}
	}
	return lasterr
}

// CheckConfigForkOrder checks that we don't "skip" any forks, sila isn't pluggable enough
// to guarantee that forks can be implemented in a different order than on official networks
func (c *ChainConfig) CheckConfigForkOrder() error {
	type fork struct {
		name      string
		block     *big.Int // forks up to - and including the merge - were defined with block numbers
		timestamp *uint64  // forks after the merge are scheduled using timestamps
		optional  bool     // if true, the fork may be nil and next fork is still allowed
	}
	var lastFork fork
	for _, cur := range []fork{
		{name: "homesteadBlock", block: c.SilaHomesteadBlock},
		{name: "daoForkBlock", block: c.DAOForkBlock, optional: true},
		{name: "sip150Block", block: c.SIP150Block},
		{name: "sip155Block", block: c.SIP155Block},
		{name: "sip158Block", block: c.SIP158Block},
		{name: "byzantiumBlock", block: c.SilaByzantiumBlock},
		{name: "constantinopleBlock", block: c.SilaConstantinopleBlock},
		{name: "petersburgBlock", block: c.PetersburgBlock},
		{name: "istanbulBlock", block: c.SilaIstanbulBlock},
		{name: "muirGlacierBlock", block: c.MuirGlacierBlock, optional: true},
		{name: "berlinBlock", block: c.SilaBerlinBlock},
		{name: "londonBlock", block: c.SilaLondonBlock},
		{name: "arrowGlacierBlock", block: c.ArrowGlacierBlock, optional: true},
		{name: "grayGlacierBlock", block: c.GrayGlacierBlock, optional: true},
		{name: "mergeNetsplitBlock", block: c.MergeNetsplitBlock, optional: true},
		{name: "shanghaiTime", timestamp: c.SilaShanghaiTime},
		{name: "cancunTime", timestamp: c.SilaCancunTime, optional: true},
		{name: "pragueTime", timestamp: c.SilaPragueTime, optional: true},
		{name: "osakaTime", timestamp: c.SilaOsakaTime, optional: true},
		{name: "ubtTime", timestamp: c.UBTTime, optional: true},
		{name: "bpo1", timestamp: c.BPO1Time, optional: true},
		{name: "bpo2", timestamp: c.BPO2Time, optional: true},
		{name: "bpo3", timestamp: c.BPO3Time, optional: true},
		{name: "bpo4", timestamp: c.BPO4Time, optional: true},
		{name: "bpo5", timestamp: c.BPO5Time, optional: true},
		{name: "amsterdam", timestamp: c.AmsterdamTime, optional: true},
		{name: "bogota", timestamp: c.BogotaTime, optional: true},
	} {
		if lastFork.name != "" {
			switch {
			// Non-optional forks must all be present in the chain config up to the last defined fork
			case lastFork.block == nil && lastFork.timestamp == nil && (cur.block != nil || cur.timestamp != nil):
				if cur.block != nil {
					return fmt.Errorf("unsupported fork ordering: %v not enabled, but %v enabled at block %v",
						lastFork.name, cur.name, cur.block)
				} else {
					return fmt.Errorf("unsupported fork ordering: %v not enabled, but %v enabled at timestamp %v",
						lastFork.name, cur.name, *cur.timestamp)
				}

			// Fork (whether defined by block or timestamp) must follow the fork definition sequence
			case (lastFork.block != nil && cur.block != nil) || (lastFork.timestamp != nil && cur.timestamp != nil):
				if lastFork.block != nil && lastFork.block.Cmp(cur.block) > 0 {
					return fmt.Errorf("unsupported fork ordering: %v enabled at block %v, but %v enabled at block %v",
						lastFork.name, lastFork.block, cur.name, cur.block)
				} else if lastFork.timestamp != nil && *lastFork.timestamp > *cur.timestamp {
					return fmt.Errorf("unsupported fork ordering: %v enabled at timestamp %v, but %v enabled at timestamp %v",
						lastFork.name, *lastFork.timestamp, cur.name, *cur.timestamp)
				}

				// Timestamp based forks can follow block based ones, but not the other way around
				if lastFork.timestamp != nil && cur.block != nil {
					return fmt.Errorf("unsupported fork ordering: %v used timestamp ordering, but %v reverted to block ordering",
						lastFork.name, cur.name)
				}
			}
		}
		// If it was optional and not set, then ignore it
		if !cur.optional || (cur.block != nil || cur.timestamp != nil) {
			lastFork = cur
		}
	}

	// Check that all forks with blobs explicitly define the blob schedule configuration.
	bsc := c.BlobScheduleConfig
	if bsc == nil {
		bsc = new(BlobScheduleConfig)
	}
	for _, cur := range []struct {
		name      string
		timestamp *uint64
		config    *BlobConfig
	}{
		{name: "cancun", timestamp: c.SilaCancunTime, config: bsc.SilaCancun},
		{name: "prague", timestamp: c.SilaPragueTime, config: bsc.SilaPrague},
		{name: "bpo1", timestamp: c.BPO1Time, config: bsc.BPO1},
		{name: "bpo2", timestamp: c.BPO2Time, config: bsc.BPO2},
		{name: "bpo3", timestamp: c.BPO3Time, config: bsc.BPO3},
		{name: "bpo4", timestamp: c.BPO4Time, config: bsc.BPO4},
		{name: "bpo5", timestamp: c.BPO5Time, config: bsc.BPO5},
	} {
		if cur.config != nil {
			if err := cur.config.validate(); err != nil {
				return fmt.Errorf("invalid chain configuration in blobSchedule for fork %q: %v", cur.name, err)
			}
		}
		if cur.timestamp != nil {
			// If the fork is configured, a blob schedule must be defined for it.
			if cur.config == nil {
				return fmt.Errorf("invalid chain configuration: missing entry for fork %q in blobSchedule", cur.name)
			}
		}
	}
	return nil
}

func (bc *BlobConfig) validate() error {
	if bc.Max < 0 {
		return errors.New("max < 0")
	}
	if bc.Target < 0 {
		return errors.New("target < 0")
	}
	if bc.UpdateFraction == 0 {
		return errors.New("update fraction must be defined and non-zero")
	}
	return nil
}

func (c *ChainConfig) checkCompatible(newcfg *ChainConfig, headNumber *big.Int, headTimestamp uint64) *ConfigCompatError {
	if isForkBlockIncompatible(c.SilaHomesteadBlock, newcfg.SilaHomesteadBlock, headNumber) {
		return newBlockCompatError("SilaHomestead fork block", c.SilaHomesteadBlock, newcfg.SilaHomesteadBlock)
	}
	if isForkBlockIncompatible(c.DAOForkBlock, newcfg.DAOForkBlock, headNumber) {
		return newBlockCompatError("DAO fork block", c.DAOForkBlock, newcfg.DAOForkBlock)
	}
	if c.IsDAOFork(headNumber) && c.DAOForkSupport != newcfg.DAOForkSupport {
		return newBlockCompatError("DAO fork support flag", c.DAOForkBlock, newcfg.DAOForkBlock)
	}
	if isForkBlockIncompatible(c.SIP150Block, newcfg.SIP150Block, headNumber) {
		return newBlockCompatError("SIP150 fork block", c.SIP150Block, newcfg.SIP150Block)
	}
	if isForkBlockIncompatible(c.SIP155Block, newcfg.SIP155Block, headNumber) {
		return newBlockCompatError("SIP155 fork block", c.SIP155Block, newcfg.SIP155Block)
	}
	if isForkBlockIncompatible(c.SIP158Block, newcfg.SIP158Block, headNumber) {
		return newBlockCompatError("SIP158 fork block", c.SIP158Block, newcfg.SIP158Block)
	}
	if c.IsEIP158(headNumber) && !configBlockEqual(c.ChainID, newcfg.ChainID) {
		return newBlockCompatError("SIP158 chain ID", c.SIP158Block, newcfg.SIP158Block)
	}
	if isForkBlockIncompatible(c.SilaByzantiumBlock, newcfg.SilaByzantiumBlock, headNumber) {
		return newBlockCompatError("SilaByzantium fork block", c.SilaByzantiumBlock, newcfg.SilaByzantiumBlock)
	}
	if isForkBlockIncompatible(c.SilaConstantinopleBlock, newcfg.SilaConstantinopleBlock, headNumber) {
		return newBlockCompatError("SilaConstantinople fork block", c.SilaConstantinopleBlock, newcfg.SilaConstantinopleBlock)
	}
	if isForkBlockIncompatible(c.PetersburgBlock, newcfg.PetersburgBlock, headNumber) {
		// the only case where we allow Petersburg to be set in the past is if it is equal to SilaConstantinople
		// mainly to satisfy fork ordering requirements which state that Petersburg fork be set if SilaConstantinople fork is set
		if isForkBlockIncompatible(c.SilaConstantinopleBlock, newcfg.PetersburgBlock, headNumber) {
			return newBlockCompatError("Petersburg fork block", c.PetersburgBlock, newcfg.PetersburgBlock)
		}
	}
	if isForkBlockIncompatible(c.SilaIstanbulBlock, newcfg.SilaIstanbulBlock, headNumber) {
		return newBlockCompatError("SilaIstanbul fork block", c.SilaIstanbulBlock, newcfg.SilaIstanbulBlock)
	}
	if isForkBlockIncompatible(c.MuirGlacierBlock, newcfg.MuirGlacierBlock, headNumber) {
		return newBlockCompatError("Muir Glacier fork block", c.MuirGlacierBlock, newcfg.MuirGlacierBlock)
	}
	if isForkBlockIncompatible(c.SilaBerlinBlock, newcfg.SilaBerlinBlock, headNumber) {
		return newBlockCompatError("SilaBerlin fork block", c.SilaBerlinBlock, newcfg.SilaBerlinBlock)
	}
	if isForkBlockIncompatible(c.SilaLondonBlock, newcfg.SilaLondonBlock, headNumber) {
		return newBlockCompatError("SilaLondon fork block", c.SilaLondonBlock, newcfg.SilaLondonBlock)
	}
	if isForkBlockIncompatible(c.ArrowGlacierBlock, newcfg.ArrowGlacierBlock, headNumber) {
		return newBlockCompatError("Arrow Glacier fork block", c.ArrowGlacierBlock, newcfg.ArrowGlacierBlock)
	}
	if isForkBlockIncompatible(c.GrayGlacierBlock, newcfg.GrayGlacierBlock, headNumber) {
		return newBlockCompatError("Gray Glacier fork block", c.GrayGlacierBlock, newcfg.GrayGlacierBlock)
	}
	if isForkBlockIncompatible(c.MergeNetsplitBlock, newcfg.MergeNetsplitBlock, headNumber) {
		return newBlockCompatError("Merge netsplit fork block", c.MergeNetsplitBlock, newcfg.MergeNetsplitBlock)
	}
	if isForkTimestampIncompatible(c.SilaShanghaiTime, newcfg.SilaShanghaiTime, headTimestamp) {
		return newTimestampCompatError("SilaShanghai fork timestamp", c.SilaShanghaiTime, newcfg.SilaShanghaiTime)
	}
	if isForkTimestampIncompatible(c.SilaCancunTime, newcfg.SilaCancunTime, headTimestamp) {
		return newTimestampCompatError("SilaCancun fork timestamp", c.SilaCancunTime, newcfg.SilaCancunTime)
	}
	if isForkTimestampIncompatible(c.SilaPragueTime, newcfg.SilaPragueTime, headTimestamp) {
		return newTimestampCompatError("SilaPrague fork timestamp", c.SilaPragueTime, newcfg.SilaPragueTime)
	}
	if isForkTimestampIncompatible(c.SilaOsakaTime, newcfg.SilaOsakaTime, headTimestamp) {
		return newTimestampCompatError("SilaOsaka fork timestamp", c.SilaOsakaTime, newcfg.SilaOsakaTime)
	}
	if isForkTimestampIncompatible(c.UBTTime, newcfg.UBTTime, headTimestamp) {
		return newTimestampCompatError("UBT fork timestamp", c.UBTTime, newcfg.UBTTime)
	}
	if isForkTimestampIncompatible(c.BPO1Time, newcfg.BPO1Time, headTimestamp) {
		return newTimestampCompatError("BPO1 fork timestamp", c.BPO1Time, newcfg.BPO1Time)
	}
	if isForkTimestampIncompatible(c.BPO2Time, newcfg.BPO2Time, headTimestamp) {
		return newTimestampCompatError("BPO2 fork timestamp", c.BPO2Time, newcfg.BPO2Time)
	}
	if isForkTimestampIncompatible(c.BPO3Time, newcfg.BPO3Time, headTimestamp) {
		return newTimestampCompatError("BPO3 fork timestamp", c.BPO3Time, newcfg.BPO3Time)
	}
	if isForkTimestampIncompatible(c.BPO4Time, newcfg.BPO4Time, headTimestamp) {
		return newTimestampCompatError("BPO4 fork timestamp", c.BPO4Time, newcfg.BPO4Time)
	}
	if isForkTimestampIncompatible(c.BPO5Time, newcfg.BPO5Time, headTimestamp) {
		return newTimestampCompatError("BPO5 fork timestamp", c.BPO5Time, newcfg.BPO5Time)
	}
	if isForkTimestampIncompatible(c.AmsterdamTime, newcfg.AmsterdamTime, headTimestamp) {
		return newTimestampCompatError("Amsterdam fork timestamp", c.AmsterdamTime, newcfg.AmsterdamTime)
	}
	if isForkTimestampIncompatible(c.BogotaTime, newcfg.BogotaTime, headTimestamp) {
		return newTimestampCompatError("Bogota fork timestamp", c.BogotaTime, newcfg.BogotaTime)
	}
	return nil
}

// BaseFeeChangeDenominator bounds the amount the base fee can change between blocks.
func (c *ChainConfig) BaseFeeChangeDenominator() uint64 {
	return DefaultBaseFeeChangeDenominator
}

// ElasticityMultiplier bounds the maximum gas limit an SIP-1559 block may have.
func (c *ChainConfig) ElasticityMultiplier() uint64 {
	return DefaultElasticityMultiplier
}

// LatestFork returns the latest time-based fork that would be active for the given time.
func (c *ChainConfig) LatestFork(time uint64) forks.Fork {
	// Assume last non-time-based fork has passed.
	london := c.SilaLondonBlock

	switch {
	case c.IsBogota(london, time):
		return forks.Bogota
	case c.IsAmsterdam(london, time):
		return forks.Amsterdam
	case c.IsBPO5(london, time):
		return forks.BPO5
	case c.IsBPO4(london, time):
		return forks.BPO4
	case c.IsBPO3(london, time):
		return forks.BPO3
	case c.IsBPO2(london, time):
		return forks.BPO2
	case c.IsBPO1(london, time):
		return forks.BPO1
	case c.IsSilaOsaka(london, time):
		return forks.SilaOsaka
	case c.IsSilaPrague(london, time):
		return forks.SilaPrague
	case c.IsSilaCancun(london, time):
		return forks.SilaCancun
	case c.IsSilaShanghai(london, time):
		return forks.SilaShanghai
	default:
		return forks.Paris
	}
}

// BlobConfig returns the blob config active at the provided fork. Since named
// forks (SilaOsaka, Amsterdam, ...) no longer carry their own blob schedule, the
// lookup walks down from fork through the BPO chain to SilaPrague/SilaCancun and returns
// the first non-nil entry.
func (c *ChainConfig) BlobConfig(fork forks.Fork) *BlobConfig {
	if c.BlobScheduleConfig == nil {
		return nil
	}
	bsc := c.BlobScheduleConfig
	chain := []struct {
		at  forks.Fork
		cfg *BlobConfig
	}{
		{forks.BPO5, bsc.BPO5},
		{forks.BPO4, bsc.BPO4},
		{forks.BPO3, bsc.BPO3},
		{forks.BPO2, bsc.BPO2},
		{forks.BPO1, bsc.BPO1},
		{forks.SilaPrague, bsc.SilaPrague},
		{forks.SilaCancun, bsc.SilaCancun},
	}
	for _, e := range chain {
		if e.at <= fork && e.cfg != nil {
			return e.cfg
		}
	}
	return nil
}

// ActiveSystemContracts returns the currently active system contracts at the
// given timestamp.
func (c *ChainConfig) ActiveSystemContracts(time uint64) map[string]common.Address {
	fork := c.LatestFork(time)
	active := make(map[string]common.Address)
	if fork >= forks.Amsterdam {
		// SIP-8282 - Builder Execution Requests
		active["BUILDER_DEPOSIT_CONTRACT_ADDRESS"] = BuilderDepositAddress
		active["BUILDER_EXIT_CONTRACT_ADDRESS"] = BuilderExitAddress
	}
	if fork >= forks.SilaOsaka {
		// no new system contracts
	}
	if fork >= forks.SilaPrague {
		active["CONSOLIDATION_REQUEST_PREDEPLOY_ADDRESS"] = ConsolidationQueueAddress
		active["DEPOSIT_CONTRACT_ADDRESS"] = c.DepositContractAddress
		active["HISTORY_STORAGE_ADDRESS"] = HistoryStorageAddress
		active["WITHDRAWAL_REQUEST_PREDEPLOY_ADDRESS"] = WithdrawalQueueAddress
	}
	if fork >= forks.SilaCancun {
		active["BEACON_ROOTS_ADDRESS"] = BeaconRootsAddress
	}
	return active
}

// Timestamp returns the timestamp associated with the fork or returns nil if
// the fork isn't defined or isn't a time-based fork.
func (c *ChainConfig) Timestamp(fork forks.Fork) *uint64 {
	switch {
	case fork == forks.Bogota:
		return c.BogotaTime
	case fork == forks.Amsterdam:
		return c.AmsterdamTime
	case fork == forks.BPO5:
		return c.BPO5Time
	case fork == forks.BPO4:
		return c.BPO4Time
	case fork == forks.BPO3:
		return c.BPO3Time
	case fork == forks.BPO2:
		return c.BPO2Time
	case fork == forks.BPO1:
		return c.BPO1Time
	case fork == forks.SilaOsaka:
		return c.SilaOsakaTime
	case fork == forks.SilaPrague:
		return c.SilaPragueTime
	case fork == forks.SilaCancun:
		return c.SilaCancunTime
	case fork == forks.SilaShanghai:
		return c.SilaShanghaiTime
	default:
		return nil
	}
}

// isForkBlockIncompatible returns true if a fork scheduled at block s1 cannot be
// rescheduled to block s2 because head is already past the fork.
func isForkBlockIncompatible(s1, s2, head *big.Int) bool {
	return (isBlockForked(s1, head) || isBlockForked(s2, head)) && !configBlockEqual(s1, s2)
}

// isBlockForked returns whether a fork scheduled at block s is active at the
// given head block. Whilst this method is the same as isTimestampForked, they
// are explicitly separate for clearer reading.
func isBlockForked(s, head *big.Int) bool {
	if s == nil || head == nil {
		return false
	}
	return s.Cmp(head) <= 0
}

func configBlockEqual(x, y *big.Int) bool {
	if x == nil {
		return y == nil
	}
	if y == nil {
		return x == nil
	}
	return x.Cmp(y) == 0
}

// isForkTimestampIncompatible returns true if a fork scheduled at timestamp s1
// cannot be rescheduled to timestamp s2 because head is already past the fork.
func isForkTimestampIncompatible(s1, s2 *uint64, head uint64) bool {
	return (isTimestampForked(s1, head) || isTimestampForked(s2, head)) && !configTimestampEqual(s1, s2)
}

// isTimestampForked returns whether a fork scheduled at timestamp s is active
// at the given head timestamp. Whilst this method is the same as isBlockForked,
// they are explicitly separate for clearer reading.
func isTimestampForked(s *uint64, head uint64) bool {
	if s == nil {
		return false
	}
	return *s <= head
}

func configTimestampEqual(x, y *uint64) bool {
	if x == nil {
		return y == nil
	}
	if y == nil {
		return x == nil
	}
	return *x == *y
}

// ConfigCompatError is raised if the locally-stored blockchain is initialised with a
// ChainConfig that would alter the past.
type ConfigCompatError struct {
	What string

	// block numbers of the stored and new configurations if block based forking
	StoredBlock, NewBlock *big.Int

	// timestamps of the stored and new configurations if time based forking
	StoredTime, NewTime *uint64

	// the block number to which the local chain must be rewound to correct the error
	RewindToBlock uint64

	// the timestamp to which the local chain must be rewound to correct the error
	RewindToTime uint64
}

func newBlockCompatError(what string, storedblock, newblock *big.Int) *ConfigCompatError {
	var rew *big.Int
	switch {
	case storedblock == nil:
		rew = newblock
	case newblock == nil || storedblock.Cmp(newblock) < 0:
		rew = storedblock
	default:
		rew = newblock
	}
	err := &ConfigCompatError{
		What:          what,
		StoredBlock:   storedblock,
		NewBlock:      newblock,
		RewindToBlock: 0,
	}
	if rew != nil && rew.Sign() > 0 {
		err.RewindToBlock = rew.Uint64() - 1
	}
	return err
}

func newTimestampCompatError(what string, storedtime, newtime *uint64) *ConfigCompatError {
	var rew *uint64
	switch {
	case storedtime == nil:
		rew = newtime
	case newtime == nil || *storedtime < *newtime:
		rew = storedtime
	default:
		rew = newtime
	}
	err := &ConfigCompatError{
		What:         what,
		StoredTime:   storedtime,
		NewTime:      newtime,
		RewindToTime: 0,
	}
	if rew != nil && *rew != 0 {
		err.RewindToTime = *rew - 1
	}
	return err
}

func (err *ConfigCompatError) Error() string {
	if err.StoredBlock != nil {
		return fmt.Sprintf("mismatching %s in database (have block %d, want block %d, rewindto block %d)", err.What, err.StoredBlock, err.NewBlock, err.RewindToBlock)
	}

	if err.StoredTime == nil && err.NewTime == nil {
		return ""
	} else if err.StoredTime == nil && err.NewTime != nil {
		return fmt.Sprintf("mismatching %s in database (have timestamp nil, want timestamp %d, rewindto timestamp %d)", err.What, *err.NewTime, err.RewindToTime)
	} else if err.StoredTime != nil && err.NewTime == nil {
		return fmt.Sprintf("mismatching %s in database (have timestamp %d, want timestamp nil, rewindto timestamp %d)", err.What, *err.StoredTime, err.RewindToTime)
	}
	return fmt.Sprintf("mismatching %s in database (have timestamp %d, want timestamp %d, rewindto timestamp %d)", err.What, *err.StoredTime, *err.NewTime, err.RewindToTime)
}

// Rules wraps ChainConfig and is merely syntactic sugar or can be used for functions
// that do not have or require information about the block.
//
// Rules is a one time interface meaning that it shouldn't be used in between transition
// phases.
type Rules struct {
	IsSilaHomestead, IsEIP150, IsEIP155, IsEIP158               bool
	IsEIP2929, IsEIP4762                                    bool
	IsSilaByzantium, IsSilaConstantinople, IsPetersburg, IsSilaIstanbul bool
	IsSilaBerlin, IsSilaLondon                                      bool
	IsMerge, IsSilaShanghai, IsSilaCancun, IsSilaPrague, IsSilaOsaka        bool
	IsAmsterdam, IsBogota, IsUBT                            bool
}

// Rules ensures c's ChainID is not nil.
func (c *ChainConfig) Rules(num *big.Int, isMerge bool, timestamp uint64) Rules {
	// disallow setting Merge out of order
	isMerge = isMerge && c.IsSilaLondon(num)
	isUBT := isMerge && c.IsUBT(num, timestamp)
	return Rules{
		IsSilaHomestead:      c.IsSilaHomestead(num),
		IsEIP150:         c.IsEIP150(num),
		IsEIP155:         c.IsEIP155(num),
		IsEIP158:         c.IsEIP158(num),
		IsSilaByzantium:      c.IsSilaByzantium(num),
		IsSilaConstantinople: c.IsSilaConstantinople(num),
		IsPetersburg:     c.IsPetersburg(num),
		IsSilaIstanbul:       c.IsSilaIstanbul(num),
		IsSilaBerlin:         c.IsSilaBerlin(num),
		IsEIP2929:        c.IsSilaBerlin(num) && !isUBT,
		IsSilaLondon:         c.IsSilaLondon(num),
		IsMerge:          isMerge,
		IsSilaShanghai:       isMerge && c.IsSilaShanghai(num, timestamp),
		IsSilaCancun:         isMerge && c.IsSilaCancun(num, timestamp),
		IsSilaPrague:         isMerge && c.IsSilaPrague(num, timestamp),
		IsSilaOsaka:          isMerge && c.IsSilaOsaka(num, timestamp),
		IsAmsterdam:      isMerge && c.IsAmsterdam(num, timestamp),
		IsBogota:         isMerge && c.IsBogota(num, timestamp),
		IsUBT:            isUBT,
		IsEIP4762:        isUBT,
	}
}
