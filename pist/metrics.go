// Copyright 2015 The go-ethereum Authors
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

package pist

import (
	"git.taiyue.io/pist/go-pist/metrics"
	"git.taiyue.io/pist/go-pist/p2p"
)

var (
	propTxnInPacketsMeter     = metrics.NewRegisteredMeter("pist/prop/txns/in/packets", nil)
	propTxnInTxsMeter         = metrics.NewRegisteredMeter("pist/prop/txns/in/txs", nil)
	propTxnInTrafficMeter     = metrics.NewRegisteredMeter("pist/prop/txns/in/traffic", nil)
	propTxnOutPacketsMeter    = metrics.NewRegisteredMeter("pist/prop/txns/out/packets", nil)
	propTxnOutTrafficMeter    = metrics.NewRegisteredMeter("pist/prop/txns/out/traffic", nil)
	propFtnInPacketsMeter     = metrics.NewRegisteredMeter("pist/prop/ftns/in/packets", nil)
	propFtnInTrafficMeter     = metrics.NewRegisteredMeter("pist/prop/ftns/in/traffic", nil)
	propFtnOutPacketsMeter    = metrics.NewRegisteredMeter("pist/prop/ftns/out/packets", nil)
	propFtnOutTrafficMeter    = metrics.NewRegisteredMeter("pist/prop/ftns/out/traffic", nil)
	propFHashInPacketsMeter   = metrics.NewRegisteredMeter("pist/prop/fhashes/in/packets", nil)
	propFHashInTrafficMeter   = metrics.NewRegisteredMeter("pist/prop/fhashes/in/traffic", nil)
	propFHashOutPacketsMeter  = metrics.NewRegisteredMeter("pist/prop/fhashes/out/packets", nil)
	propFHashOutTrafficMeter  = metrics.NewRegisteredMeter("pist/prop/fhashes/out/traffic", nil)
	propSHashInPacketsMeter   = metrics.NewRegisteredMeter("pist/prop/shashes/in/packets", nil)
	propSHashInTrafficMeter   = metrics.NewRegisteredMeter("pist/prop/shashes/in/traffic", nil)
	propSHashOutPacketsMeter  = metrics.NewRegisteredMeter("pist/prop/shashes/out/packets", nil)
	propSHashOutTrafficMeter  = metrics.NewRegisteredMeter("pist/prop/shashes/out/traffic", nil)
	propFBlockInPacketsMeter  = metrics.NewRegisteredMeter("pist/prop/fblocks/in/packets", nil)
	propFBlockInTrafficMeter  = metrics.NewRegisteredMeter("pist/prop/fblocks/in/traffic", nil)
	propFBlockOutPacketsMeter = metrics.NewRegisteredMeter("pist/prop/fblocks/out/packets", nil)
	propFBlockOutTrafficMeter = metrics.NewRegisteredMeter("pist/prop/fblocks/out/traffic", nil)
	propSBlockInPacketsMeter  = metrics.NewRegisteredMeter("pist/prop/sblocks/in/packets", nil)
	propSBlockInTrafficMeter  = metrics.NewRegisteredMeter("pist/prop/sblocks/in/traffic", nil)
	propSBlockOutPacketsMeter = metrics.NewRegisteredMeter("pist/prop/sblocks/out/packets", nil)
	propSBlockOutTrafficMeter = metrics.NewRegisteredMeter("pist/prop/sblocks/out/traffic", nil)

	propNodeInfoInPacketsMeter  = metrics.NewRegisteredMeter("pist/prop/nodeinfo/in/packets", nil)
	propNodeInfoInTrafficMeter  = metrics.NewRegisteredMeter("pist/prop/nodeinfo/in/traffic", nil)
	propNodeInfoOutPacketsMeter = metrics.NewRegisteredMeter("pist/prop/nodeinfo/out/packets", nil)
	propNodeInfoOutTrafficMeter = metrics.NewRegisteredMeter("pist/prop/nodeinfo/out/traffic", nil)

	propNodeInfoHashInPacketsMeter  = metrics.NewRegisteredMeter("pist/prop/nodeinfohash/in/packets", nil)
	propNodeInfoHashInTrafficMeter  = metrics.NewRegisteredMeter("pist/prop/nodeinfohash/in/traffic", nil)
	propNodeInfoHashOutPacketsMeter = metrics.NewRegisteredMeter("pist/prop/nodeinfohash/out/packets", nil)
	propNodeInfoHashOutTrafficMeter = metrics.NewRegisteredMeter("pist/prop/nodeinfohash/out/traffic", nil)

	reqFHeaderInPacketsMeter  = metrics.NewRegisteredMeter("pist/req/headers/in/packets", nil)
	reqFHeaderInTrafficMeter  = metrics.NewRegisteredMeter("pist/req/headers/in/traffic", nil)
	reqFHeaderOutPacketsMeter = metrics.NewRegisteredMeter("pist/req/headers/out/packets", nil)
	reqFHeaderOutTrafficMeter = metrics.NewRegisteredMeter("pist/req/headers/out/traffic", nil)

	reqFBodyInPacketsMeter  = metrics.NewRegisteredMeter("pist/req/fbodies/in/packets", nil)
	reqFBodyInTrafficMeter  = metrics.NewRegisteredMeter("pist/req/fbodies/in/traffic", nil)
	reqFBodyOutPacketsMeter = metrics.NewRegisteredMeter("pist/req/fbodies/out/packets", nil)
	reqFBodyOutTrafficMeter = metrics.NewRegisteredMeter("pist/req/fbodies/out/traffic", nil)

	reqStateInPacketsMeter    = metrics.NewRegisteredMeter("pist/req/states/in/packets", nil)
	reqStateInTrafficMeter    = metrics.NewRegisteredMeter("pist/req/states/in/traffic", nil)
	reqStateOutPacketsMeter   = metrics.NewRegisteredMeter("pist/req/states/out/packets", nil)
	reqStateOutTrafficMeter   = metrics.NewRegisteredMeter("pist/req/states/out/traffic", nil)
	reqReceiptInPacketsMeter  = metrics.NewRegisteredMeter("pist/req/receipts/in/packets", nil)
	reqReceiptInTrafficMeter  = metrics.NewRegisteredMeter("pist/req/receipts/in/traffic", nil)
	reqReceiptOutPacketsMeter = metrics.NewRegisteredMeter("pist/req/receipts/out/packets", nil)
	reqReceiptOutTrafficMeter = metrics.NewRegisteredMeter("pist/req/receipts/out/traffic", nil)

	getHeadInPacketsMeter  = metrics.NewRegisteredMeter("pist/get/head/in/packets", nil)
	getHeadInTrafficMeter  = metrics.NewRegisteredMeter("pist/get/head/in/traffic", nil)
	getHeadOutPacketsMeter = metrics.NewRegisteredMeter("pist/get/head/out/packets", nil)
	getHeadOutTrafficMeter = metrics.NewRegisteredMeter("pist/get/head/out/traffic", nil)

	getNodeInfoInPacketsMeter  = metrics.NewRegisteredMeter("pist/get/nodeinfo/in/packets", nil)
	getNodeInfoInTrafficMeter  = metrics.NewRegisteredMeter("pist/get/nodeinfo/in/traffic", nil)
	getNodeInfoOutPacketsMeter = metrics.NewRegisteredMeter("pist/get/nodeinfo/out/packets", nil)
	getNodeInfoOutTrafficMeter = metrics.NewRegisteredMeter("pist/get/nodeinfo/out/traffic", nil)

	miscInPacketsMeter  = metrics.NewRegisteredMeter("pist/misc/in/packets", nil)
	miscInTrafficMeter  = metrics.NewRegisteredMeter("pist/misc/in/traffic", nil)
	miscOutPacketsMeter = metrics.NewRegisteredMeter("pist/misc/out/packets", nil)
	miscOutTrafficMeter = metrics.NewRegisteredMeter("pist/misc/out/traffic", nil)
)

// meteredMsgReadWriter is a wrapper around a p2p.MsgReadWriter, capable of
// accumulating the above defined metrics based on the data stream contents.
type meteredMsgReadWriter struct {
	p2p.MsgReadWriter     // Wrapped message stream to meter
	version           int // Protocol version to select correct meters
}

// newMeteredMsgWriter wraps a p2p MsgReadWriter with metering support. If the
// metrics system is disabled, this function returns the original object.
func newMeteredMsgWriter(rw p2p.MsgReadWriter) p2p.MsgReadWriter {
	if !metrics.Enabled {
		return rw
	}
	return &meteredMsgReadWriter{MsgReadWriter: rw}
}

// Init sets the protocol version used by the stream to know which meters to
// increment in case of overlapping message ids between protocol versions.
func (rw *meteredMsgReadWriter) Init(version int) {
	rw.version = version
}

func (rw *meteredMsgReadWriter) ReadMsg() (p2p.Msg, error) {
	// Read the message and short circuit in case of an error
	msg, err := rw.MsgReadWriter.ReadMsg()
	if err != nil {
		return msg, err
	}
	// Account for the data traffic
	packets, traffic := miscInPacketsMeter, miscInTrafficMeter
	switch {
	case msg.Code == BlockHeadersMsg:
		packets, traffic = reqFHeaderInPacketsMeter, reqFHeaderInTrafficMeter
	case msg.Code == BlockBodiesMsg:
		packets, traffic = reqFBodyInPacketsMeter, reqFBodyInTrafficMeter
	case msg.Code == NodeDataMsg:
		packets, traffic = reqStateInPacketsMeter, reqStateInTrafficMeter
	case msg.Code == ReceiptsMsg:
		packets, traffic = reqReceiptInPacketsMeter, reqReceiptInTrafficMeter

	case msg.Code == NewBlockHashesMsg:
		packets, traffic = propFHashInPacketsMeter, propFHashInTrafficMeter
	case msg.Code == NewBlockMsg:
		packets, traffic = propFBlockInPacketsMeter, propFBlockInTrafficMeter
	case msg.Code == TransactionMsg:
		packets, traffic = propTxnInPacketsMeter, propTxnInTrafficMeter
	case msg.Code == TbftNodeInfoMsg:
		packets, traffic = propNodeInfoInPacketsMeter, propNodeInfoInTrafficMeter
	case msg.Code == TbftNodeInfoHashMsg:
		packets, traffic = propNodeInfoHashInPacketsMeter, propNodeInfoHashInTrafficMeter
	case msg.Code == GetTbftNodeInfoMsg:
		packets, traffic = getNodeInfoInPacketsMeter, getNodeInfoInTrafficMeter
	case msg.Code == GetBlockHeadersMsg:
		packets, traffic = getHeadInPacketsMeter, getHeadInTrafficMeter
	case msg.Code == GetBlockBodiesMsg:
		packets, traffic = getHeadInPacketsMeter, getHeadInTrafficMeter
	}
	packets.Mark(1)
	traffic.Mark(int64(msg.Size))
	return msg, err
}

func (rw *meteredMsgReadWriter) WriteMsg(msg p2p.Msg) error {
	// Account for the data traffic
	packets, traffic := miscOutPacketsMeter, miscOutTrafficMeter
	switch {
	case msg.Code == BlockHeadersMsg:
		packets, traffic = reqFHeaderOutPacketsMeter, reqFHeaderOutTrafficMeter
	case msg.Code == BlockBodiesMsg:
		packets, traffic = reqFBodyOutPacketsMeter, reqFBodyOutTrafficMeter
	case msg.Code == NodeDataMsg:
		packets, traffic = reqStateOutPacketsMeter, reqStateOutTrafficMeter
	case msg.Code == ReceiptsMsg:
		packets, traffic = reqReceiptOutPacketsMeter, reqReceiptOutTrafficMeter

	case msg.Code == NewBlockHashesMsg:
		packets, traffic = propFHashOutPacketsMeter, propFHashOutTrafficMeter
	case msg.Code == NewBlockMsg:
		packets, traffic = propFBlockOutPacketsMeter, propFBlockOutTrafficMeter
	case msg.Code == TransactionMsg:
		packets, traffic = propTxnOutPacketsMeter, propTxnOutTrafficMeter
	case msg.Code == TbftNodeInfoMsg:
		packets, traffic = propNodeInfoOutPacketsMeter, propNodeInfoOutTrafficMeter
	case msg.Code == TbftNodeInfoHashMsg:
		packets, traffic = propNodeInfoHashOutPacketsMeter, propNodeInfoHashOutTrafficMeter
	case msg.Code == GetTbftNodeInfoMsg:
		packets, traffic = getNodeInfoOutPacketsMeter, getNodeInfoOutTrafficMeter
	case msg.Code == GetBlockHeadersMsg:
		packets, traffic = getHeadOutPacketsMeter, getHeadOutTrafficMeter
	case msg.Code == GetBlockBodiesMsg:
		packets, traffic = getHeadInPacketsMeter, getHeadOutTrafficMeter
	}
	packets.Mark(1)
	traffic.Mark(int64(msg.Size))

	// Send the packet to the p2p layer
	return rw.MsgReadWriter.WriteMsg(msg)
}
