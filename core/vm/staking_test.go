package vm

import (
	"math/big"
	"testing"

	"git.taiyue.io/pist/go-pist/common"
	"git.taiyue.io/pist/go-pist/core/state"
	"git.taiyue.io/pist/go-pist/core/types"
	"git.taiyue.io/pist/go-pist/crypto"
	"git.taiyue.io/pist/go-pist/log"
	"git.taiyue.io/pist/go-pist/params"
	"git.taiyue.io/pist/go-pist/pistdb"
)

func TestDeposit(t *testing.T) {

	priKey, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(priKey.PublicKey)
	pub := crypto.FromECDSAPub(&priKey.PublicKey)
	value := big.NewInt(1000)

	statedb, _ := state.New(common.Hash{}, state.NewDatabase(pistdb.NewMemDatabase()))
	statedb.GetOrNewStateObject(types.StakingAddress)
	evm := NewEVM(Context{}, statedb, params.TestChainConfig, Config{})

	log.Info("Staking deposit", "address", from, "value", value)
	impawn := NewImpawnImpl()
	impawn.Load(evm.StateDB, types.StakingAddress)

	impawn.InsertSAccount2(1000, from, pub, value, big.NewInt(0), true)
	impawn.Save(evm.StateDB, types.StakingAddress)

	impawn1 := NewImpawnImpl()
	impawn1.Load(evm.StateDB, types.StakingAddress)
}
