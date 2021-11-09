// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"git.taiyue.io/pist/go-pist/core/vm"

	"git.taiyue.io/pist/go-pist/common"
	"git.taiyue.io/pist/go-pist/common/hexutil"
	"git.taiyue.io/pist/go-pist/common/math"
	"git.taiyue.io/pist/go-pist/consensus"
	"git.taiyue.io/pist/go-pist/core/rawdb"
	"git.taiyue.io/pist/go-pist/core/state"
	"git.taiyue.io/pist/go-pist/core/types"
	"git.taiyue.io/pist/go-pist/crypto"
	"git.taiyue.io/pist/go-pist/log"
	"git.taiyue.io/pist/go-pist/params"
	"git.taiyue.io/pist/go-pist/pistdb"
	"git.taiyue.io/pist/go-pist/rlp"
)

//go:generate gencodec -type Genesis -field-override genesisSpecMarshaling -out gen_genesis.go
//go:generate gencodec -type GenesisAccount -field-override genesisAccountMarshaling -out gen_genesis_account.go

var errGenesisNoConfig = errors.New("genesis has no chain configuration")

// Genesis specifies the header fields, state of a genesis block. It also defines hard
// fork switch-over blocks through the chain configuration.
type Genesis struct {
	Config     *params.ChainConfig      `json:"config"`
	Nonce      uint64                   `json:"nonce"`
	Timestamp  uint64                   `json:"timestamp"`
	ExtraData  []byte                   `json:"extraData"`
	GasLimit   uint64                   `json:"gasLimit"   gencodec:"required"`
	Difficulty *big.Int                 `json:"difficulty" gencodec:"required"`
	Mixhash    common.Hash              `json:"mixHash"`
	Coinbase   common.Address           `json:"coinbase"`
	Alloc      types.GenesisAlloc       `json:"alloc"      gencodec:"required"`
	Committee  []*types.CommitteeMember `json:"committee"      gencodec:"required"`

	// These fields are used for consensus tests. Please don't use them
	// in actual genesis blocks.
	Number     uint64      `json:"number"`
	GasUsed    uint64      `json:"gasUsed"`
	ParentHash common.Hash `json:"parentHash"`
}

// GenesisAccount is an account in the state of the genesis block.
type GenesisAccount struct {
	Code       []byte                      `json:"code,omitempty"`
	Storage    map[common.Hash]common.Hash `json:"storage,omitempty"`
	Balance    *big.Int                    `json:"balance" gencodec:"required"`
	Nonce      uint64                      `json:"nonce,omitempty"`
	PrivateKey []byte                      `json:"secretKey,omitempty"` // for tests
}

// field type overrides for gencodec
type genesisSpecMarshaling struct {
	Nonce      math.HexOrDecimal64
	Timestamp  math.HexOrDecimal64
	ExtraData  hexutil.Bytes
	GasLimit   math.HexOrDecimal64
	GasUsed    math.HexOrDecimal64
	Number     math.HexOrDecimal64
	Difficulty *math.HexOrDecimal256
	Alloc      map[common.UnprefixedAddress]GenesisAccount
}

type genesisAccountMarshaling struct {
	Code       hexutil.Bytes
	Balance    *math.HexOrDecimal256
	Nonce      math.HexOrDecimal64
	Storage    map[storageJSON]storageJSON
	PrivateKey hexutil.Bytes
}

// storageJSON represents a 256 bit byte array, but allows less than 256 bits when
// unmarshaling from hex.
type storageJSON common.Hash

func (h *storageJSON) UnmarshalText(text []byte) error {
	text = bytes.TrimPrefix(text, []byte("0x"))
	if len(text) > 64 {
		return fmt.Errorf("too many hex characters in storage key/value %q", text)
	}
	offset := len(h) - len(text)/2 // pad on the left
	if _, err := hex.Decode(h[offset:], text); err != nil {
		fmt.Println(err)
		return fmt.Errorf("invalid hex storage key/value %q", text)
	}
	return nil
}

func (h storageJSON) MarshalText() ([]byte, error) {
	return hexutil.Bytes(h[:]).MarshalText()
}

// GenesisMismatchError is raised when trying to overwrite an existing
// genesis block with an incompatible one.
type GenesisMismatchError struct {
	Stored, New common.Hash
}

func (e *GenesisMismatchError) Error() string {
	return fmt.Sprintf("database already contains an incompatible genesis block (have %x, new %x)", e.Stored[:8], e.New[:8])
}

// SetupGenesisBlock writes or updates the genesis block in db.
// The block that will be used is:
//
//                          genesis == nil       genesis != nil
//                       +------------------------------------------
//     db has no genesis |  main-net default  |  genesis
//     db has genesis    |  from DB           |  genesis (if compatible)
//
// The stored chain configuration will be updated if it is compatible (i.e. does not
// specify a fork block below the local head block). In case of a conflict, the
// error is a *params.ConfigCompatError and the new, unwritten config is returned.
//
// The returned chain configuration is never nil.
func SetupGenesisBlock(db pistdb.Database, genesis *Genesis) (*params.ChainConfig, common.Hash, error) {
	if genesis != nil && genesis.Config == nil {
		return params.AllMinervaProtocolChanges, common.Hash{}, errGenesisNoConfig
	}

	fastConfig, fastHash, fastErr := setupFastGenesisBlock(db, genesis)

	return fastConfig, fastHash, fastErr

}

// setupFastGenesisBlock writes or updates the fast genesis block in db.
// The block that will be used is:
//
//                          genesis == nil       genesis != nil
//                       +------------------------------------------
//     db has no genesis |  main-net default  |  genesis
//     db has genesis    |  from DB           |  genesis (if compatible)
//
// The stored chain configuration will be updated if it is compatible (i.e. does not
// specify a fork block below the local head block). In case of a conflict, the
// error is a *params.ConfigCompatError and the new, unwritten config is returned.
//
// The returned chain configuration is never nil.
func setupFastGenesisBlock(db pistdb.Database, genesis *Genesis) (*params.ChainConfig, common.Hash, error) {
	if genesis != nil && genesis.Config == nil {
		return params.AllMinervaProtocolChanges, common.Hash{}, errGenesisNoConfig
	}

	// Just commit the new block if there is no stored genesis block.
	stored := rawdb.ReadCanonicalHash(db, 0)
	if (stored == common.Hash{}) {
		if genesis == nil {
			log.Info("Writing default main-net genesis block")
			genesis = DefaultGenesisBlock()
		} else {
			log.Info("Writing custom genesis block")
		}
		block, err := genesis.CommitFast(db)
		return genesis.Config, block.Hash(), err
	}

	// Check whether the genesis block is already written.
	if genesis != nil {
		hash := genesis.ToFastBlock(nil).Hash()
		if hash != stored {
			return genesis.Config, hash, &GenesisMismatchError{stored, hash}
		}
	}

	// Get the existing chain configuration.
	newcfg := genesis.configOrDefault(stored)
	storedcfg := rawdb.ReadChainConfig(db, stored)
	if storedcfg == nil {
		log.Warn("Found genesis block without chain config")
		rawdb.WriteChainConfig(db, stored, newcfg)
		return newcfg, stored, nil
	}
	// Special case: don't change the existing config of a non-mainnet chain if no new
	// config is supplied. These chains would get AllProtocolChanges (and a compat error)
	// if we just continued here.
	if genesis == nil && stored != params.MainnetGenesisHash {
		return storedcfg, stored, nil
	}

	// Check config compatibility and write the config. Compatibility errors
	// are returned to the caller unless we're already at block zero.
	height := rawdb.ReadHeaderNumber(db, rawdb.ReadHeadHeaderHash(db))
	if height == nil {
		return newcfg, stored, fmt.Errorf("missing block number for head header hash")
	}
	compatErr := storedcfg.CheckCompatible(newcfg, *height)
	if compatErr != nil && *height != 0 && compatErr.RewindTo != 0 {
		return newcfg, stored, compatErr
	}
	rawdb.WriteChainConfig(db, stored, newcfg)
	return newcfg, stored, nil
}

// CommitFast writes the block and state of a genesis specification to the database.
// The block is committed as the canonical head block.
func (g *Genesis) CommitFast(db pistdb.Database) (*types.Block, error) {
	block := g.ToFastBlock(db)
	if block.Number().Sign() != 0 {
		return nil, fmt.Errorf("can't commit genesis block with number > 0")
	}
	rawdb.WriteBlock(db, block)
	rawdb.WriteReceipts(db, block.Hash(), block.NumberU64(), nil)
	rawdb.WriteCanonicalHash(db, block.Hash(), block.NumberU64())
	rawdb.WriteHeadBlockHash(db, block.Hash())
	rawdb.WriteHeadHeaderHash(db, block.Hash())
	rawdb.WriteStateGcBR(db, block.NumberU64())

	config := g.Config
	if config == nil {
		config = params.AllMinervaProtocolChanges
	}
	rawdb.WriteChainConfig(db, block.Hash(), config)
	return block, nil
}

// ToFastBlock creates the genesis block and writes state of a genesis specification
// to the given database (or discards it if nil).
func (g *Genesis) ToFastBlock(db pistdb.Database) *types.Block {
	if db == nil {
		db = pistdb.NewMemDatabase()
	}
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(db))
	for addr, account := range g.Alloc {
		statedb.AddBalance(addr, account.Balance)
		statedb.SetCode(addr, account.Code)
		statedb.SetNonce(addr, account.Nonce)
		for key, value := range account.Storage {
			statedb.SetState(addr, key, value)
		}
	}
	consensus.OnceInitImpawnState(statedb)
	impl := vm.NewImpawnImpl()
	hh := g.Number
	bigint := big.NewInt(2147483648)
	for _, member := range g.Committee {
		err := impl.InsertSAccount2(hh, member.Coinbase, member.Publickey, params.ElectionMinLimitForStaking, big.NewInt(100), true)
		if err != nil {
			log.Error("ToFastBlock InsertSAccount", "error", err)
		}
		// Judge the main chain to add pist to the membership of the committee
		if bigint.Cmp(g.Difficulty) == 0 {
			statedb.SetPOSLocked(member.Coinbase, new(big.Int).Set(params.ElectionMinLimitForStaking))
		}
		//statedb.AddBalance(types.StakingAddress, params.ElectionMinLimitForStaking)
	}
	_, err := impl.DoElections(1, 0)
	if err != nil {
		log.Error("ToFastBlock DoElections", "error", err)
	}
	err = impl.Shift(1)
	if err != nil {
		log.Error("ToFastBlock Shift", "error", err)
	}
	err = impl.Save(statedb, types.StakingAddress)
	if err != nil {
		log.Error("ToFastBlock IMPL Save", "error", err)
	}

	root := statedb.IntermediateRoot(false)

	head := &types.Header{
		Number:     new(big.Int).SetUint64(g.Number),
		Time:       new(big.Int).SetUint64(g.Timestamp),
		ParentHash: g.ParentHash,
		Extra:      g.ExtraData,
		GasLimit:   g.GasLimit,
		GasUsed:    g.GasUsed,
		Root:       root,
	}
	if g.GasLimit == 0 {
		head.GasLimit = params.GenesisGasLimit
	}
	statedb.Commit(false)
	statedb.Database().TrieDB().Commit(root, true)

	// All genesis committee members are included in switchinfo of block #0
	committee := &types.SwitchInfos{CID: common.Big0, Members: g.Committee, BackMembers: make([]*types.CommitteeMember, 0), Vals: make([]*types.SwitchEnter, 0)}
	for _, member := range committee.Members {
		pubkey, _ := crypto.UnmarshalPubkey(member.Publickey)
		member.Flag = types.StateUsedFlag
		member.MType = types.TypeFixed
		member.CommitteeBase = crypto.PubkeyToAddress(*pubkey)
	}
	return types.NewBlock(head, nil, nil, nil, committee.Members)
}

// MustFastCommit writes the genesis block and state to db, panicking on error.
// The block is committed as the canonical head block.
func (g *Genesis) MustFastCommit(db pistdb.Database) *types.Block {
	block, err := g.CommitFast(db)
	if err != nil {
		panic(err)
	}
	return block
}

// GenesisBlockForTesting creates and writes a block in which addr has the given wei balance.
func GenesisBlockForTesting(db pistdb.Database, addr common.Address, balance *big.Int) *types.Block {
	g := Genesis{Alloc: types.GenesisAlloc{addr: {Balance: balance}}, Config: params.AllMinervaProtocolChanges}
	return g.MustFastCommit(db)
}

// DefaultGenesisBlock returns the Pistchain main net snail block.
func DefaultGenesisBlock() *Genesis {
	allocAmount := new(big.Int).Mul(big.NewInt(1000000000-70000), big.NewInt(1e18))

	key1 := hexutil.MustDecode("0x046d840f61f399e28bce5f75d2c10ed3d69d327fa6c1d7e3b88a52da94d6ea2ba386ac4421dda8526e68d9e36bbd29b78a97083a76591405a8bad91d9e879c9fdf")
	key2 := hexutil.MustDecode("0x047623965238eccade5d3c5198f85236d61a3ccc8abb2e049b1d0a5aa0405d4691adf13566c60963c28fbf6138a05be6cc6480f12f865093cc4dba17f1a99fddc3")
	key3 := hexutil.MustDecode("0x04e0ae7e80bfa8101210af7facd60a78edce4763b609907eb6d90d7899c2899d4accb0f9bc2b3009e5cc14292be112c81621453f48d029b9731570782b88b53d8c")
	key4 := hexutil.MustDecode("0x0438dd2b3628761b757c1739341948440f8046a334e797c594a6504880742a74c8ac9bfe60d511a9afa354954b95bc0e7088e77d83d0f4d50c5315c6978a71d280")
	key5 := hexutil.MustDecode("0x0469c3915b634b680dd524ec65fd968ffea968022c4d9c52908dac0d12b2aca8cdb1d5ac86ded9cd414b0a75216f2bdc60c0ebb4cf8b06794d173da0c9b6b1fa68")
	key6 := hexutil.MustDecode("0x040fd309938e7a1e0f5ba92051122e1893955f1774aad080b9cd525d7cbeb0b78bb397c5748a6081f4dc07094859cc4de7de3ffc514b6b0d1e054c0fb2512a1edd")
	key7 := hexutil.MustDecode("0x043bb94b172e4dc9bd7460c6b9320080b56a558952beac159bce3e24102579345852fac212f00f8a20f43cf3e578b218d9673a715e3f2201e2f24f20da51d6be1a")
	i := new(big.Int).Mul(big.NewInt(10000), big.NewInt(1e18))
	return &Genesis{
		Config:     params.MainnetChainConfig,
		Nonce:      330,
		ExtraData:  hexutil.MustDecode("0x54727565436861696E204D61696E4E6574"),
		GasLimit:   16777216,
		Difficulty: big.NewInt(2147483648),
		//Timestamp:  1553918400,
		Coinbase:   common.HexToAddress("0x0000000000000000000000000000000000000000"),
		Mixhash:    common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
		ParentHash: common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
		//Alloc:      decodePrealloc(mainnetAllocData),
		Alloc: map[common.Address]types.GenesisAccount{
			common.HexToAddress("0x2c369246170c231004e4f9E179bAE0cCB4792B9b"): {Balance: allocAmount},
			common.HexToAddress("0x36a9711dCDAEddfE9A72dC8ca02053D7206cb61c"): {Balance: i},
			common.HexToAddress("0x9f2Dc90001FD9aE2Dfa1b2013655a974C289Df44"): {Balance: i},
			common.HexToAddress("0xA0225995FD956581bD58bdAac05e757DD6696080"): {Balance: i},
			common.HexToAddress("0x6f2297DB7c72685FFdA70761011E18aB37b9361E"): {Balance: i},
			common.HexToAddress("0x8932615E18f9549d738Ea3B1b014474a9287610E"): {Balance: i},
			common.HexToAddress("0x7a74B00b4fDA168Eee790Cea6cAd7ce0f845caeE"): {Balance: i},
			common.HexToAddress("0x0859C4b866Cf1A9e137704B60c2d369021Ec1584"): {Balance: i},
		},
		Committee: []*types.CommitteeMember{
			&types.CommitteeMember{Coinbase: common.HexToAddress("0x36a9711dCDAEddfE9A72dC8ca02053D7206cb61c"), Publickey: key1},
			&types.CommitteeMember{Coinbase: common.HexToAddress("0x9f2Dc90001FD9aE2Dfa1b2013655a974C289Df44"), Publickey: key2},
			&types.CommitteeMember{Coinbase: common.HexToAddress("0xA0225995FD956581bD58bdAac05e757DD6696080"), Publickey: key3},
			&types.CommitteeMember{Coinbase: common.HexToAddress("0x6f2297DB7c72685FFdA70761011E18aB37b9361E"), Publickey: key4},
			&types.CommitteeMember{Coinbase: common.HexToAddress("0x8932615E18f9549d738Ea3B1b014474a9287610E"), Publickey: key5},
			&types.CommitteeMember{Coinbase: common.HexToAddress("0x7a74B00b4fDA168Eee790Cea6cAd7ce0f845caeE"), Publickey: key6},
			&types.CommitteeMember{Coinbase: common.HexToAddress("0x0859C4b866Cf1A9e137704B60c2d369021Ec1584"), Publickey: key7},
		},
	}
}

func (g *Genesis) configOrDefault(ghash common.Hash) *params.ChainConfig {
	switch {
	case g != nil:
		return g.Config
	case ghash == params.MainnetGenesisHash:
		return params.MainnetChainConfig

	case ghash == params.TestnetGenesisHash:
		return params.TestnetChainConfig

	default:
		return params.AllMinervaProtocolChanges
	}
}

func decodePrealloc(data string) types.GenesisAlloc {
	var p []struct{ Addr, Balance *big.Int }
	if err := rlp.NewStream(strings.NewReader(data), 0).Decode(&p); err != nil {
		panic(err)
	}
	ga := make(types.GenesisAlloc, len(p))
	for _, account := range p {
		ga[common.BigToAddress(account.Addr)] = types.GenesisAccount{Balance: account.Balance}
	}
	return ga
}

// GenesisFastBlockForTesting creates and writes a block in which addr has the given wei balance.
func GenesisFastBlockForTesting(db pistdb.Database, addr common.Address, balance *big.Int) *types.Block {
	g := Genesis{Alloc: types.GenesisAlloc{addr: {Balance: balance}}, Config: params.AllMinervaProtocolChanges}
	return g.MustFastCommit(db)
}

// DefaultDevGenesisBlock returns the Rinkeby network genesis block.
func DefaultDevGenesisBlock() *Genesis {
	i, _ := new(big.Int).SetString("90000000000000000000000", 10)
	// priv1: 55dcdfd62f565a66e1886959e82a365e4987ed0b405adc43614a42c3481edd1a
	// addr1: 0x3e3429F72450A39CE227026E8DdeF331E9973E4d
	key1 := hexutil.MustDecode("0x04600254af4ce74276f54b4f9df193f2cb72ed76b7341cb144f4d6f1408402dc10719eebdcb947ced9ac6fe9a690e004692db6222de7867cbab712246eb23a50b7")
	// priv2: a0eb966cae593e0d85c7eda4ad4815d0c857bee9a7085a8b19e52e3227138ae4
	// addr2: 0xf353ab1417177F766497bF716D7aAd4ECd5f36C8
	key2 := hexutil.MustDecode("0x043ae657860b05d119351eac9d2f4531811ade3895ee2df00661368ca528ee36ceb850315f7bb566c6bbebf765e2c15f6af16b253a4d3d930cca7a191ae14af80d")
	// priv3: 5b743d4234c54710a644ff93a6f5284af065d2a42fff5b51de73a7c13d427b1c
	// addr3: 0x8fF345746C3d3435a105538E4c024Af5FE700598
	key3 := hexutil.MustDecode("0x049e0a67955d69e28faabe654b4a8f85e7d32b32fd2687a080e6357b53ec9413ad4f472d979bdccfe21cb135c7e144ca90f2beeb728b06e59f80918c7e52fbc6ff")
	// priv4: 229ca04fb83ec698296037c7d2b04a731905df53b96c260555cbeed9e4c64036
	// addr4: 0xf0C8898B2016Afa0Ec5912413ebe403930446779
	key4 := hexutil.MustDecode("0x04718502f879a949ca5fa29f78f1d3cef362ecdc36ee42a3023cca80371c2e1936d1f632a0ec5bf5edb2af228a5ba1669d31ea55df87548de172e5767b9201097d")

	return &Genesis{
		Config:     params.DevnetChainConfig,
		Nonce:      928,
		ExtraData:  nil,
		GasLimit:   88080384,
		Difficulty: big.NewInt(20000),
		//Alloc:      decodePrealloc(mainnetAllocData),
		Alloc: map[common.Address]types.GenesisAccount{
			common.HexToAddress("0x8dd832d7db11f4a7cae758641d84ad5af6e4833f"): {Balance: i},
			common.HexToAddress("0x3672631830709b0f7ebefd7c24d437867b638979"): {Balance: i},
			common.HexToAddress("0x26a187c6d2e77fb339fc6d721f67a9495b98a81a"): {Balance: i},
			common.HexToAddress("0x228f78fc398db973b96ed666c92e78753b9466eb"): {Balance: i},
		},
		Committee: []*types.CommitteeMember{
			{Coinbase: common.HexToAddress("0x3e3429F72450A39CE227026E8DdeF331E9973E4d"), Publickey: key1},
			{Coinbase: common.HexToAddress("0xf353ab1417177F766497bF716D7aAd4ECd5f36C8"), Publickey: key2},
			{Coinbase: common.HexToAddress("0x8fF345746C3d3435a105538E4c024Af5FE700598"), Publickey: key3},
			{Coinbase: common.HexToAddress("0xf0C8898B2016Afa0Ec5912413ebe403930446779"), Publickey: key4},
		},
	}
}

func DefaultSingleNodeGenesisBlock() *Genesis {
	i, _ := new(big.Int).SetString("90000000000000000000000", 10)
	key1 := hexutil.MustDecode("0x04044308742b61976de7344edb8662d6d10be1c477dd46e8e4c433c1288442a79183480894107299ff7b0706490f1fb9c9b7c9e62ae62d57bd84a1e469460d8ac1")

	return &Genesis{
		Config:     params.SingleNodeChainConfig,
		Nonce:      66,
		ExtraData:  nil,
		GasLimit:   22020096,
		Difficulty: big.NewInt(256),
		//Alloc:      decodePrealloc(mainnetAllocData),
		Alloc: map[common.Address]types.GenesisAccount{
			common.HexToAddress("0xbd54a6c8298a70e9636d0555a77ffa412abdd71a"): {Balance: i},
			common.HexToAddress("0x3c2e0a65a023465090aaedaa6ed2975aec9ef7f9"): {Balance: i},
			common.HexToAddress("0x7c357530174275dd30e46319b89f71186256e4f7"): {Balance: i},
			common.HexToAddress("0xeeb69c67751e9f4917b605840fa9a28be4517871"): {Balance: i},
			common.HexToAddress("0x9810a954bb88fdc251374d666ed7e06748ea672d"): {Balance: i},
		},
		Committee: []*types.CommitteeMember{
			{Coinbase: common.HexToAddress("0x76ea2f3a002431fede1141b660dbb75c26ba6d97"), Publickey: key1},
		},
	}
}

// DefaultTestnetGenesisBlock returns the Ropsten network genesis block.
func DefaultTestnetGenesisBlock() *Genesis {
	allocAmount := new(big.Int).Mul(big.NewInt(1000000000-40000), big.NewInt(1e18))
	// priv1: 55dcdfd62f565a66e1886959e82a365e4987ed0b405adc43614a42c3481edd1a
	// addr1: 0x3e3429F72450A39CE227026E8DdeF331E9973E4d
	key1 := hexutil.MustDecode("0x04600254af4ce74276f54b4f9df193f2cb72ed76b7341cb144f4d6f1408402dc10719eebdcb947ced9ac6fe9a690e004692db6222de7867cbab712246eb23a50b7")
	// priv2: a0eb966cae593e0d85c7eda4ad4815d0c857bee9a7085a8b19e52e3227138ae4
	// addr2: 0xf353ab1417177F766497bF716D7aAd4ECd5f36C8
	key2 := hexutil.MustDecode("0x043ae657860b05d119351eac9d2f4531811ade3895ee2df00661368ca528ee36ceb850315f7bb566c6bbebf765e2c15f6af16b253a4d3d930cca7a191ae14af80d")
	// priv3: 5b743d4234c54710a644ff93a6f5284af065d2a42fff5b51de73a7c13d427b1c
	// addr3: 0x8fF345746C3d3435a105538E4c024Af5FE700598
	key3 := hexutil.MustDecode("0x049e0a67955d69e28faabe654b4a8f85e7d32b32fd2687a080e6357b53ec9413ad4f472d979bdccfe21cb135c7e144ca90f2beeb728b06e59f80918c7e52fbc6ff")
	// priv4: 229ca04fb83ec698296037c7d2b04a731905df53b96c260555cbeed9e4c64036
	// addr4: 0xf0C8898B2016Afa0Ec5912413ebe403930446779
	key4 := hexutil.MustDecode("0x04718502f879a949ca5fa29f78f1d3cef362ecdc36ee42a3023cca80371c2e1936d1f632a0ec5bf5edb2af228a5ba1669d31ea55df87548de172e5767b9201097d")

	return &Genesis{
		Config:     params.TestnetChainConfig,
		Nonce:      928,
		ExtraData:  hexutil.MustDecode("0x54727565436861696E20546573744E6574203035"),
		GasLimit:   20971520,
		Difficulty: big.NewInt(100000),
		Timestamp:  1537891200,
		Coinbase:   common.HexToAddress("0x0000000000000000000000000000000000000000"),
		Mixhash:    common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
		ParentHash: common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
		Alloc: map[common.Address]types.GenesisAccount{
			common.HexToAddress("0x8dd832d7db11f4a7cae758641d84ad5af6e4833f"): {Balance: allocAmount},
		},
		Committee: []*types.CommitteeMember{
			{Coinbase: common.HexToAddress("0x3e3429F72450A39CE227026E8DdeF331E9973E4d"), Publickey: key1},
			{Coinbase: common.HexToAddress("0xf353ab1417177F766497bF716D7aAd4ECd5f36C8"), Publickey: key2},
			{Coinbase: common.HexToAddress("0x8fF345746C3d3435a105538E4c024Af5FE700598"), Publickey: key3},
			{Coinbase: common.HexToAddress("0xf0C8898B2016Afa0Ec5912413ebe403930446779"), Publickey: key4},
		},
	}
}
