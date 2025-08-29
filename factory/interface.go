package factory

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/core/sovereign"
)

// ETHClient defines what a websocket client should do
type ETHClient interface {
	Start(ctx context.Context) error
	RegisterHandler(handler sovereign.IncomingHeaderSubscriber) error
	Close()
	IsInterfaceNil() bool
}
