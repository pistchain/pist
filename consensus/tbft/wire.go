package tbft

import (
	"git.taiyue.io/pist/go-pist/consensus/tbft/types"
	"github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()

func init() {
	RegisterConsensusMessages(cdc)
	// RegisterWALMessages(cdc)
	types.RegisterBlockAmino(cdc)
}
