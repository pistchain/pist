package conn

import (
	cryptoAmino "git.taiyue.io/pist/go-pist/consensus/tbft/crypto/cryptoamino"
	amino "github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()

func init() {
	cryptoAmino.RegisterAmino(cdc)
	RegisterPacket(cdc)
}
