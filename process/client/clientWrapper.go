package client

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type clientWrapper struct {
	url    string
	client *ethclient.Client
}

// NewClient creates a new instance of clientWrapper with the specified URL.
// It  will establish a connection to the Ethereum client using the provided URL.
func NewClient(url string) (*clientWrapper, error) {
	return &clientWrapper{
		client: nil,
		url:    url,
	}, nil
}

// Dial connects the underlying client
func (cw *clientWrapper) Dial() error {
	if cw.client != nil {
		return nil
	}

	client, err := ethclient.Dial(cw.url)
	if err != nil {
		return err
	}

	cw.client = client
	return nil
}

// SubscribeNewHead subscribes to notifications about the current blockchain head on the given channel.
func (cw *clientWrapper) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	if cw.client == nil {
		return nil, errConnectionNotOpened
	}

	return cw.client.SubscribeNewHead(ctx, ch)
}

// FilterLogs executes a filter query.
func (cw *clientWrapper) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	if cw.client == nil {
		return nil, errConnectionNotOpened
	}

	return cw.client.FilterLogs(ctx, q)
}

// HeaderByNumber returns a block header from the current canonical chain. If number is nil, the latest known header is returned.
func (cw *clientWrapper) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	if cw.client == nil {
		return nil, errConnectionNotOpened
	}

	return cw.client.HeaderByNumber(ctx, number)
}

// Close closes the underlying eth client connection
func (cw *clientWrapper) Close() {
	if cw.client != nil {
		cw.client.Close()
	}
	cw.client = nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (cw *clientWrapper) IsInterfaceNil() bool {
	return cw == nil
}
