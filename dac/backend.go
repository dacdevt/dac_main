package dac

import (
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/dacchain/dacapp/accounts"
	"github.com/dacchain/dacapp/dac/gasprice"
	"github.com/dacchain/dacapp/common"
	"github.com/dacchain/dacapp/consensus"
	"github.com/dacchain/dacapp/consensus/clique"
	"github.com/dacchain/dacapp/core"
	"github.com/dacchain/dacapp/core/bloombits"
	"github.com/dacchain/dacapp/core/vm"
	"github.com/dacchain/dacapp/db"
	"github.com/dacchain/dacapp/event"
	"github.com/dacchain/dacapp/log"
	"github.com/dacchain/dacapp/miner"
	"github.com/dacchain/dacapp/node"
	"github.com/dacchain/dacapp/p2p"
	"github.com/dacchain/dacapp/params"
	"github.com/dacchain/dacapp/rlp"
	"github.com/dacchain/dacapp/rpc"
	"github.com/dacchain/dacapp/utils/dacapi"
)

type dac struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan  chan bool    // Channel for shutting down the dacChain
	stopDbUpgrade func() error // stop chain db sequential key upgrade

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager

	// DB interfaces
	chainDb db.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	ApiBackend *dacApiBackend

	miner    *miner.Miner
	gasPrice *big.Int
	dacbase  common.Address

	networkId     uint64
	netRPCService *dacapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and dacbase)
}

// New creates a new dac object (including the
// initialisation of the common dacChain object)
func New(ctx *node.ServiceContext, config *Config) (*dac, error) {
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	//stopDbUpgrade := upgradeDeduplicateData(chainDb)
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	dac := &dac{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, chainConfig, chainDb),
		shutdownChan:   make(chan bool),
		//stopDbUpgrade: stopDbUpgrade,
		networkId:     config.NetworkId,
		gasPrice:      config.GasPrice,
		bloomRequests: make(chan chan *bloombits.Retrieval),
		//bloomIndexer:  NewBloomIndexer(chainDb, params.BloomBitsBlocks),
	}

	log.Info("Initialising dacChain protocol", "versions", ProtocolVersions, "network", config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := core.GetBlockChainVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run dacapp upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		core.WriteBlockChainVersion(chainDb, core.BlockChainVersion)
	}

	vmConfig := vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
	dac.blockchain, err = core.NewBlockChain(chainDb, dac.chainConfig, dac.engine, vmConfig)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		dac.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	//eth.bloomIndexer.Start(eth.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	dac.txPool = core.NewTxPool(config.TxPool, dac.chainConfig, dac.blockchain)

	if dac.protocolManager, err = NewProtocolManager(dac.chainConfig, config.NetworkId, dac.eventMux, dac.txPool, dac.engine, dac.blockchain, chainDb); err != nil {
		return nil, err
	}

	dac.miner = miner.New(dac, dac.chainConfig, dac.EventMux(), dac.engine)
	dac.miner.SetExtra(makeExtraData(config.ExtraData))

	dac.ApiBackend = &dacApiBackend{dac, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	dac.ApiBackend.gpo = gasprice.NewOracle(dac.ApiBackend, gpoParams)

	return dac, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"dac",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", common.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (db.Database, error) {
	dacdb, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if dacdb, ok := dacdb.(*db.LDBDatabase); ok {
		dacdb.Meter("dac/db/chaindata/")
	}
	return dacdb, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an dacChain service
func CreateConsensusEngine(ctx *node.ServiceContext, chainConfig *params.ChainConfig, db db.Database) consensus.Engine {
	// If proof-of-authority is requested, set it up
	if chainConfig.Clique != nil {
		return clique.New(chainConfig.Clique, db)
	}
	return nil
}

// APIs returns the collection of RPC services the dac package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *dac) APIs() []rpc.API {
	apis := dacapi.GetAPIs(s.ApiBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "dac",
			Version:   "1.0",
			Service:   NewPublicdacAPI(s),
			Public:    true,
		},
		{
			Namespace: "dac",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, /*{
			Namespace: "dac",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "dac",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "dac",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, false),
			Public:    true,
		}, */
		{
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		},
		/*, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s.chainConfig, s),
		}, */{
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *dac) BlockChain() *core.BlockChain { return s.blockchain }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *dac) Protocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
}

func (s *dac) dacbase() (eb common.Address, err error) {
	s.lock.RLock()
	dacbase := s.dacbase
	s.lock.RUnlock()

	if dacbase != (common.Address{}) {
		return dacbase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			dacbase := accounts[0].Address

			s.lock.Lock()
			s.dacbase = dacbase
			s.lock.Unlock()

			log.Info("dacChain automatically configured", "address", dacbase)
			return dacbase, nil
		}
	}
	return common.Address{}, fmt.Errorf("dacbase must be explicitly specified")
}

func (s *dac) StartMining(local bool) error {
	eb, err := s.dacbase()
	if err != nil {
		log.Error("Cannot start mining without dacbase", "err", err)
		return fmt.Errorf("dacrbase missing: %v", err)
	}
	if clique, ok := s.engine.(*clique.Clique); ok {
		wallet, err := s.accountManager.Find(accounts.Account{Address: eb})
		if wallet == nil || err != nil {
			log.Error("dacbase account unavailable locally", "err", err)
			return fmt.Errorf("signer missing: %v", err)
		}
		clique.Authorize(eb, wallet.SignHash)
	}
	if local {
		// If local (CPU) mining is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU mining on mainnet is ludicrous
		// so noone will ever hit this path, whereas marking sync done on CPU mining
		// will ensure that private networks work in single miner mode too.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)
	}
	go s.miner.Start(eb)
	return nil
}

// Start implements node.Service, starting all internal goroutines needed by the
// protocol implementation.
func (s *dac) Start(srvr *p2p.Server) error {
	// Start the bloom bits servicing goroutines
	//s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = dacapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// protocol.
func (s *dac) Stop() error {
	/*if s.stopDbUpgrade != nil {
		s.stopDbUpgrade()
	}
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}
	s.txPool.Stop()
	s.miner.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
	close(s.shutdownChan)*/

	return nil
}

func (s *dac) Engine() consensus.Engine          { return s.engine }
func (s *dac) EventMux() *event.TypeMux          { return s.eventMux }
func (s *dac) ChainDb() db.Database              { return s.chainDb }
func (s *dac) TxPool() *core.TxPool              { return s.txPool }
func (s *dac) AccountManager() *accounts.Manager { return s.accountManager }
func (s *dac) NetVersion() uint64                { return s.networkId }
func (s *dac) dacVersion() int                   { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *dac) IsMining() bool                    { return s.miner.Mining() }
func (s *dac) Miner() *miner.Miner               { return s.miner }
