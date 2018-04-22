package dac

import (
	"context"
	"math/big"

	"github.com/dacchain/dacapp/accounts"
	"github.com/dacchain/dacapp/dac/gasprice"
	"github.com/dacchain/dacapp/common"
	"github.com/dacchain/dacapp/common/math"
	"github.com/dacchain/dacapp/core"
	"github.com/dacchain/dacapp/core/state"
	"github.com/dacchain/dacapp/core/types"
	"github.com/dacchain/dacapp/core/vm"
	"github.com/dacchain/dacapp/db"
	"github.com/dacchain/dacapp/event"
	"github.com/dacchain/dacapp/params"
	"github.com/dacchain/dacapp/rpc"
)

// dacApiBackend implements dacapi.Backend for full nodes
type dacApiBackend struct {
	dac *dac
	gpo *gasprice.Oracle
}

func (b *dacApiBackend) ChainConfig() *params.ChainConfig {
	return b.dac.chainConfig
}

func (b *dacApiBackend) CurrentBlock() *types.Block {
	return b.dac.blockchain.CurrentBlock()
}

func (b *dacApiBackend) SetHead(number uint64) {
	b.dac.protocolManager.downloader.Cancel()
	b.dac.blockchain.SetHead(number)
}

func (b *dacApiBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.dac.txPool.Get(hash)
}

func (b *dacApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.dac.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.dac.blockchain.CurrentBlock().Header(), nil
	}
	return b.dac.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *dacApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.dac.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.dac.blockchain.CurrentBlock(), nil
	}
	return b.dac.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *dacApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.dac.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.dac.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *dacApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.dac.blockchain.GetBlockByHash(blockHash), nil
}

func (b *dacApiBackend) GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
	return core.GetBlockReceipts(b.dac.chainDb, blockHash, core.GetBlockNumber(b.dac.chainDb, blockHash)), nil
}

func (b *dacApiBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.dac.blockchain.GetTdByHash(blockHash)
}

func (b *dacApiBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.dac.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *dacApiBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.dac.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *dacApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.dac.txPool.AddLocal(signedTx)
}

func (b *dacApiBackend) AccountManager() *accounts.Manager {
	return b.dac.AccountManager()
}

func (b *dacApiBackend) ChainDb() db.Database {
	return b.dac.ChainDb()
}

func (b *dacApiBackend) ProtocolVersion() int {
	return b.dac.dacVersion()
}

func (b *dacApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *dacApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.dac.txPool.State().GetNonce(addr), nil
}

func (b *dacApiBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.dac.BlockChain(), nil)
	return vm.NewEVM(context, state, b.dac.chainConfig, vmCfg), vmError, nil
}
