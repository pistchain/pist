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

// Package pist implements the Pistchain protocol.
package pist

import (
	"errors"
	"fmt"
	"git.taiyue.io/pist/go-pist/accounts"
	"git.taiyue.io/pist/go-pist/common/hexutil"
	"git.taiyue.io/pist/go-pist/consensus"
	elect "git.taiyue.io/pist/go-pist/consensus/election"
	ethash "git.taiyue.io/pist/go-pist/consensus/minerva"
	"git.taiyue.io/pist/go-pist/consensus/tbft"
	"git.taiyue.io/pist/go-pist/core"
	"git.taiyue.io/pist/go-pist/core/bloombits"
	"git.taiyue.io/pist/go-pist/core/rawdb"
	"git.taiyue.io/pist/go-pist/core/types"
	"git.taiyue.io/pist/go-pist/core/vm"
	"git.taiyue.io/pist/go-pist/crypto"
	"git.taiyue.io/pist/go-pist/event"
	"git.taiyue.io/pist/go-pist/internal/pistapi"
	"git.taiyue.io/pist/go-pist/log"
	"git.taiyue.io/pist/go-pist/node"
	"git.taiyue.io/pist/go-pist/p2p"
	"git.taiyue.io/pist/go-pist/params"
	config "git.taiyue.io/pist/go-pist/params"
	"git.taiyue.io/pist/go-pist/pist/downloader"
	"git.taiyue.io/pist/go-pist/pist/filters"
	"git.taiyue.io/pist/go-pist/pist/gasprice"
	"git.taiyue.io/pist/go-pist/pistdb"
	"git.taiyue.io/pist/go-pist/rlp"
	"git.taiyue.io/pist/go-pist/rpc"
	"math/big"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *core.ChainIndexer)
}

// Pistchain implements the Pistchain full node service.
type Pistchain struct {
	config      *Config
	chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan chan bool // Channel for shutting down the Pistchain
	// Handlers
	txPool          *core.TxPool
	agent           *PbftAgent
	election        *elect.Election
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer
	// DB interfaces
	chainDb pistdb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	APIBackend    *TrueAPIBackend
	gasPrice      *big.Int
	networkID     uint64
	netRPCService *pistapi.PublicNetAPI

	pbftServer *tbft.Node

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price)
}

func (s *Pistchain) AddLesServer(ls LesServer) {
	s.lesServer = ls
	ls.SetBloomBitsIndexer(s.bloomIndexer)
}

// New creates a new Pistchain object (including the
// initialisation of the common Pistchain object)
func New(ctx *node.ServiceContext, config *Config) (*Pistchain, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run pist.Pistchain in light sync mode, use les.LightTruechain")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDb, err := CreateDB(ctx, config, "chaindata")
	//chainDb, err := CreateDB(ctx, config, path)
	if err != nil {
		return nil, err
	}

	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	/*if config.Genesis != nil {
		config.MinerGasFloor = config.Genesis.GasLimit * 9 / 10
		config.MinerGasCeil = config.Genesis.GasLimit * 11 / 10
	}*/

	pist := &Pistchain{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, &ethash.Config{PowMode: ethash.ModeNormal}, chainDb),
		shutdownChan:   make(chan bool),
		networkID:      config.NetworkId,
		gasPrice:       config.GasPrice,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainDb, params.BloomBitsBlocks, params.BloomConfirms),
	}

	log.Info("Initialising Pistchain protocol", "versions", ProtocolVersions, "network", config.NetworkId, "syncmode", config.SyncMode)

	if !config.SkipBcVersionCheck {
		bcVersion := rawdb.ReadDatabaseVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run gpist upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		rawdb.WriteDatabaseVersion(chainDb, core.BlockChainVersion)
	}
	var (
		vmConfig    = vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
		cacheConfig = &core.CacheConfig{Deleted: config.DeletedState, Disabled: config.NoPruning, TrieNodeLimit: config.TrieCache, TrieTimeLimit: config.TrieTimeout}
	)

	pist.blockchain, err = core.NewBlockChain(chainDb, cacheConfig, pist.chainConfig, pist.engine, vmConfig)
	if err != nil {
		return nil, err
	}

	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		pist.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	pist.bloomIndexer.Start(pist.blockchain)
	consensus.InitDPos(chainConfig)
	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	pist.txPool = core.NewTxPool(config.TxPool, pist.chainConfig, pist.blockchain)
	pist.election = elect.NewElection(pist.chainConfig, pist.blockchain, pist.config)
	checkpoint := config.Checkpoint
	cacheLimit := cacheConfig.TrieCleanLimit

	pist.engine.SetElection(pist.election)
	pist.election.SetEngine(pist.engine)

	pist.agent = NewPbftAgent(pist, pist.chainConfig, pist.engine, pist.election, config.MinerGasFloor, config.MinerGasCeil)

	if pist.protocolManager, err = NewProtocolManager(
		pist.chainConfig, checkpoint, config.SyncMode, config.NetworkId,
		pist.eventMux, pist.txPool, pist.engine,
		pist.blockchain, chainDb, pist.agent, cacheLimit, config.Whitelist); err != nil {
		return nil, err
	}

	pist.APIBackend = &TrueAPIBackend{pist, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	pist.APIBackend.gpo = gasprice.NewOracle(pist.APIBackend, gpoParams)
	return pist, nil
}
func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"gpist",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (pistdb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*pistdb.LDBDatabase); ok {
		db.Meter("pist/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an Pistchain service
func CreateConsensusEngine(ctx *node.ServiceContext, config *ethash.Config, db pistdb.Database) consensus.Engine {
	// If proof-of-authority is requested, set it up
	/*
		if chainConfig.Clique != nil {
			return clique.New(chainConfig.Clique, db)
		}*/
	// Otherwise assume proof-of-work
	switch config.PowMode {
	case ethash.ModeFake:
		log.Info("-----Fake mode")
		log.Warn("Ethash used in fake mode")
		return ethash.NewFaker()
	case ethash.ModeTest:
		log.Warn("Ethash used in test mode")
		return ethash.NewTester()
	case ethash.ModeShared:
		log.Warn("Ethash used in shared mode")
		return ethash.NewShared()
	default:
		engine := ethash.New(ethash.Config{})
		engine.SetThreads(-1) // Disable CPU mining
		return engine
	}
}

// APIs return the collection of RPC services the pist package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Pistchain) APIs() []rpc.API {
	apis := pistapi.GetAPIs(s.APIBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append pist	APIs and  Eth APIs
	namespaces := []string{"pist", "eth"}
	for _, name := range namespaces {
		apis = append(apis, []rpc.API{
			{
				Namespace: name,
				Version:   "1.0",
				Service:   NewPublicTruechainAPI(s),
				Public:    true,
			}, {
				Namespace: name,
				Version:   "1.0",
				Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
				Public:    true,
			}, {
				Namespace: name,
				Version:   "1.0",
				Service:   filters.NewPublicFilterAPI(s.APIBackend, false),
				Public:    true,
			},
		}...)
	}
	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s.chainConfig, s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *Pistchain) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Pistchain) ResetWithFastGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}
func (s *Pistchain) PbftAgent() *PbftAgent             { return s.agent }
func (s *Pistchain) AccountManager() *accounts.Manager { return s.accountManager }
func (s *Pistchain) BlockChain() *core.BlockChain      { return s.blockchain }
func (s *Pistchain) Config() *Config                   { return s.config }
func (s *Pistchain) TxPool() *core.TxPool              { return s.txPool }

func (s *Pistchain) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Pistchain) Engine() consensus.Engine           { return s.engine }
func (s *Pistchain) ChainDb() pistdb.Database           { return s.chainDb }
func (s *Pistchain) IsListening() bool                  { return true } // Always listening
func (s *Pistchain) EthVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *Pistchain) NetVersion() uint64                 { return s.networkID }
func (s *Pistchain) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *Pistchain) Synced() bool                       { return atomic.LoadUint32(&s.protocolManager.acceptTxs) == 1 }
func (s *Pistchain) ArchiveMode() bool                  { return s.config.NoPruning }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Pistchain) Protocols() []p2p.Protocol {
	if s.lesServer == nil {
		return s.protocolManager.SubProtocols
	}
	return append(s.protocolManager.SubProtocols, s.lesServer.Protocols()...)
}

// Start implements node.Service, starting all internal goroutines needed by the
// Pistchain protocol implementation.
func (s *Pistchain) Start(srvr *p2p.Server) error {

	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = pistapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers

	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}
	s.startPbftServer()
	if s.pbftServer == nil {
		log.Error("start pbft server failed.")
		return errors.New("start pbft server failed.")
	}
	s.agent.server = s.pbftServer
	log.Info("", "server", s.agent.server)
	s.agent.Start()

	s.election.Start()
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Pistchain protocol.
func (s *Pistchain) Stop() error {
	s.stopPbftServer()
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}
	s.txPool.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}

func (s *Pistchain) startPbftServer() error {
	priv, err := crypto.ToECDSA(s.config.CommitteeKey)
	if err != nil {
		return err
	}

	cfg := config.DefaultConfig()
	cfg.P2P.ListenAddress1 = "tcp://0.0.0.0:" + strconv.Itoa(s.config.Port)
	cfg.P2P.ListenAddress2 = "tcp://0.0.0.0:" + strconv.Itoa(s.config.StandbyPort)

	n1, err := tbft.NewNode(cfg, "1", priv, s.agent)
	if err != nil {
		return err
	}
	s.pbftServer = n1
	return n1.Start()
}

func (s *Pistchain) stopPbftServer() error {
	if s.pbftServer != nil {
		s.pbftServer.Stop()
	}
	return nil
}
