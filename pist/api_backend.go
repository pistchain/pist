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
	"context"
	"math/big"

	"git.taiyue.io/pist/go-pist/accounts"
	"git.taiyue.io/pist/go-pist/common"
	"git.taiyue.io/pist/go-pist/common/math"
	"git.taiyue.io/pist/go-pist/core"
	"git.taiyue.io/pist/go-pist/core/bloombits"
	"git.taiyue.io/pist/go-pist/core/rawdb"
	"git.taiyue.io/pist/go-pist/core/state"
	"git.taiyue.io/pist/go-pist/core/types"
	"git.taiyue.io/pist/go-pist/core/vm"
	"git.taiyue.io/pist/go-pist/event"
	"git.taiyue.io/pist/go-pist/params"
	"git.taiyue.io/pist/go-pist/pist/downloader"
	"git.taiyue.io/pist/go-pist/pist/gasprice"
	"git.taiyue.io/pist/go-pist/pistdb"
	"git.taiyue.io/pist/go-pist/rpc"
)

// TRUEAPIBackend implements ethapi.Backend for full nodes
type TrueAPIBackend struct {
	pist *Pistchain
	gpo  *gasprice.Oracle
}

// ChainConfig returns the active chain configuration.
func (b *TrueAPIBackend) ChainConfig() *params.ChainConfig {
	return b.pist.chainConfig
}

// CurrentBlock return the fast chain current Block
func (b *TrueAPIBackend) CurrentBlock() *types.Block {
	return b.pist.blockchain.CurrentBlock()
}

// SetHead Set the newest position of Fast Chain, that will reset the fast blockchain comment
func (b *TrueAPIBackend) SetHead(number uint64) {
	b.pist.protocolManager.downloader.Cancel()
	b.pist.blockchain.SetHead(number)
}

// HeaderByNumber returns Header of fast chain by the number
// rpc.PendingBlockNumber == "pending"; rpc.LatestBlockNumber == "latest" ; rpc.LatestBlockNumber == "earliest"
func (b *TrueAPIBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.pist.blockchain.CurrentBlock().Header(), nil
	}
	return b.pist.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

// HeaderByHash returns header of fast chain by the hash
func (b *TrueAPIBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return b.pist.blockchain.GetHeaderByHash(hash), nil
}

// BlockByNumber returns block of fast chain by the number
func (b *TrueAPIBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	if blockNr == rpc.PendingBlockNumber {
		block := b.pist.blockchain.CurrentBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.pist.blockchain.CurrentBlock(), nil
	}
	return b.pist.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

// StateAndHeaderByNumber returns the state of block by the number
func (b *TrueAPIBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		state, _ := b.pist.blockchain.State()
		block := b.pist.blockchain.CurrentBlock()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.pist.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

// GetBlock returns the block by the block's hash
func (b *TrueAPIBackend) GetBlock(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return b.pist.blockchain.GetBlockByHash(hash), nil
}

// GetReceipts returns the Receipt details by txhash
func (b *TrueAPIBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	if number := rawdb.ReadHeaderNumber(b.pist.chainDb, hash); number != nil {
		return rawdb.ReadReceipts(b.pist.chainDb, hash, *number), nil
	}
	return nil, nil
}

// GetLogs returns the logs by txhash
func (b *TrueAPIBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	number := rawdb.ReadHeaderNumber(b.pist.chainDb, hash)
	if number == nil {
		return nil, nil
	}
	receipts := rawdb.ReadReceipts(b.pist.chainDb, hash, *number)
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

// GetEVM returns the EVM
func (b *TrueAPIBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.pist.BlockChain(), nil, nil)
	return vm.NewEVM(context, state, b.pist.chainConfig, vmCfg), vmError, nil
}

// SubscribeRemovedLogsEvent registers a subscription of RemovedLogsEvent in fast blockchain
func (b *TrueAPIBackend) SubscribeRemovedLogsEvent(ch chan<- types.RemovedLogsEvent) event.Subscription {
	return b.pist.BlockChain().SubscribeRemovedLogsEvent(ch)
}

// SubscribeChainEvent registers a subscription of chainEvnet in fast blockchain
func (b *TrueAPIBackend) SubscribeChainEvent(ch chan<- types.FastChainEvent) event.Subscription {
	return b.pist.BlockChain().SubscribeChainEvent(ch)
}

// SubscribeChainHeadEvent registers a subscription of chainHeadEvnet in fast blockchain
func (b *TrueAPIBackend) SubscribeChainHeadEvent(ch chan<- types.FastChainHeadEvent) event.Subscription {
	return b.pist.BlockChain().SubscribeChainHeadEvent(ch)
}

// SubscribeChainSideEvent registers a subscription of chainSideEvnet in fast blockchain,deprecated
func (b *TrueAPIBackend) SubscribeChainSideEvent(ch chan<- types.FastChainSideEvent) event.Subscription {
	return b.pist.BlockChain().SubscribeChainSideEvent(ch)
}

// SubscribeLogsEvent registers a subscription of log in fast blockchain
func (b *TrueAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.pist.BlockChain().SubscribeLogsEvent(ch)
}

func (b *TrueAPIBackend) GetChainRewardContent(blockNr rpc.BlockNumber) *types.ChainReward {
	sheight := uint64(blockNr)
	return b.pist.blockchain.GetRewardInfos(sheight)
}

// GetStateChangeByFastNumber returns the Committee info by committee number
func (b *TrueAPIBackend) GetStateChangeByFastNumber(fastNumber rpc.BlockNumber) *types.BlockBalance {
	return b.pist.blockchain.GetBalanceInfos(uint64(fastNumber))
}

func (b *TrueAPIBackend) GetCommittee(number rpc.BlockNumber) (map[string]interface{}, error) {
	if number == rpc.LatestBlockNumber {
		return b.pist.election.GetCommitteeById(new(big.Int).SetUint64(b.pist.agent.CommitteeNumber())), nil
	}
	return b.pist.election.GetCommitteeById(big.NewInt(number.Int64())), nil
}

func (b *TrueAPIBackend) GetCurrentCommitteeNumber() *big.Int {
	return b.pist.election.GetCurrentCommitteeNumber()
}

// SendTx returns nil by success to add local txpool
func (b *TrueAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.pist.txPool.AddLocal(signedTx)
}

// GetPoolTransactions returns Transactions by pending state in txpool
func (b *TrueAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.pist.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

// GetPoolTransaction returns Transaction by txHash in txpool
func (b *TrueAPIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.pist.txPool.Get(hash)
}

// GetPoolNonce returns user nonce by user address in txpool
func (b *TrueAPIBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.pist.txPool.State().GetNonce(addr), nil
}

// Stats returns the count tx in txpool
func (b *TrueAPIBackend) Stats() (pending int, queued int) {
	return b.pist.txPool.Stats()
}

func (b *TrueAPIBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.pist.TxPool().Content()
}

// SubscribeNewTxsEvent returns the subscript event of new tx
func (b *TrueAPIBackend) SubscribeNewTxsEvent(ch chan<- types.NewTxsEvent) event.Subscription {
	return b.pist.TxPool().SubscribeNewTxsEvent(ch)
}

// Downloader returns the fast downloader
func (b *TrueAPIBackend) Downloader() *downloader.Downloader {
	return b.pist.Downloader()
}

// ProtocolVersion returns the version of protocol
func (b *TrueAPIBackend) ProtocolVersion() int {
	return b.pist.EthVersion()
}

// SuggestPrice returns tht suggest gas price
func (b *TrueAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

// ChainDb returns tht database of fastchain
func (b *TrueAPIBackend) ChainDb() pistdb.Database {
	return b.pist.ChainDb()
}

// EventMux returns Event locker
func (b *TrueAPIBackend) EventMux() *event.TypeMux {
	return b.pist.EventMux()
}

// AccountManager returns Account Manager
func (b *TrueAPIBackend) AccountManager() *accounts.Manager {
	return b.pist.AccountManager()
}

// BloomStatus returns Bloom Status
func (b *TrueAPIBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.pist.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

// ServiceFilter make the Filter for the truechian
func (b *TrueAPIBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.pist.bloomRequests)
	}
}
