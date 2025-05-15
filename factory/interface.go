package factory

import (
	"context"
)

// ETHClient defines what a websocket client should do
type ETHClient interface {
	Start(ctx context.Context)
	Close()
}
