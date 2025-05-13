package client

import (
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

// Close closes the underlying eth client connection
func (cw *clientWrapper) Close() {
	cw.client.Close()
}

// IsInterfaceNil checks if the underlying pointer is nil
func (cw *clientWrapper) IsInterfaceNil() bool {
	return cw == nil
}
