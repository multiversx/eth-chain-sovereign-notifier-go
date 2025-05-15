package client

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type clientWrapper struct {
	client *ethclient.Client
}

// NewClient creates a new instance of clientWrapper with the specified URL.
// It establishes a connection to the Ethereum client using the provided URL.
func NewClient(url string) (*clientWrapper, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}

	return &clientWrapper{
		client: client,
	}, nil
}

func (cw *clientWrapper) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	return cw.client.SubscribeNewHead(ctx, ch)
}

func (cw *clientWrapper) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	return cw.client.FilterLogs(ctx, q)
}

func (cw *clientWrapper) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return cw.client.HeaderByNumber(ctx, number)
}

// Close closes the underlying eth client connection
func (cw *clientWrapper) Close() {
	cw.client.Close()
}

// IsInterfaceNil checks if the underlying pointer is nil
func (cw *clientWrapper) IsInterfaceNil() bool {
	return cw == nil
}
