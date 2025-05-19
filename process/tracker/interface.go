package tracker

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	sovCore "github.com/multiversx/mx-chain-core-go/core/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

// ETHClientHandler defines an eth client behavior
type ETHClientHandler interface {
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	IsInterfaceNil() bool
	Close()
}

// IncomingHeaderCreator defines an incoming header creator behavior
type IncomingHeaderCreator interface {
	CreateIncomingHeader(header *types.Header, logs []types.Log) (sovereign.IncomingHeaderHandler, error)
	IsInterfaceNil() bool
}

// IncomingHeadersNotifierHandler defines an incoming header notifier behavior
type IncomingHeadersNotifierHandler interface {
	NotifyHeaderSubscribers(header sovereign.IncomingHeaderHandler) error
	RegisterSubscriber(handler sovCore.IncomingHeaderSubscriber) error
}

// BlockCache defines a block cache behavior
type BlockCache interface {
	Add(ctx context.Context, header *types.Header) error
	ExtractFinalizedBlocks() []*types.Header
}
