package tracker

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	sovCore "github.com/multiversx/mx-chain-core-go/core/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

type ETHClientHandler interface {
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	Close()
}

type IncomingHeaderCreator interface {
	CreateIncomingHeader(header *types.Header, logs []types.Log) (sovereign.IncomingHeaderHandler, error)
}

type IncomingHeadersNotifierHandler interface {
	NotifyHeaderSubscribers(header sovereign.IncomingHeaderHandler) error
	RegisterSubscriber(handler sovCore.IncomingHeaderSubscriber) error
}
