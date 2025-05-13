package factory

import (
	"github.com/multiversx/eth-chain-sovereign-notifier-go/config"
	"github.com/multiversx/eth-chain-sovereign-notifier-go/process/client"
)

// ETHClient defines what a websocket client should do
type ETHClient interface {
	Close()
}

func CreateWSETHNotifier(cfg config.Config) (ETHClient, error) {
	return client.NewClient(cfg.ClientConfig.Url)
}
