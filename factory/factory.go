package factory

import (
	"github.com/multiversx/mx-chain-core-go/core/sovereign"
	hashingFactory "github.com/multiversx/mx-chain-core-go/hashing/factory"
	"github.com/multiversx/mx-chain-core-go/marshal/factory"

	"github.com/multiversx/eth-chain-sovereign-notifier-go/config"
	"github.com/multiversx/eth-chain-sovereign-notifier-go/process/client"
	"github.com/multiversx/eth-chain-sovereign-notifier-go/process/tracker"
)

// CreateWSETHClientNotifier creates a ws eth client notifier
func CreateWSETHClientNotifier(cfg config.Config) (ETHClient, error) {
	marshaller, err := factory.NewMarshalizer(cfg.MarshallerType)
	if err != nil {
		return nil, err
	}

	hasher, err := hashingFactory.NewHasher(cfg.HasherType)
	if err != nil {
		return nil, err
	}

	headersNotifier, err := sovereign.NewHeadersNotifier(marshaller, hasher)
	if err != nil {
		return nil, err
	}

	ethClient, err := client.NewClient(cfg.ClientConfig.Url)
	if err != nil {
		return nil, err
	}

	//common.HexToAddress(cfg.SubscribedEvents[0].)

	argsBlockTracker := tracker.ArgsETHBlockTracker{
		SubscribedETHEvents:     nil,
		MinConfirmations:        cfg.MinBlocksConfirmation,
		Client:                  ethClient,
		IncomingHeaderCreator:   tracker.NewIncomingHeadersCreator(),
		IncomingHeadersNotifier: headersNotifier,
	}

	return tracker.NewBlockTracker(argsBlockTracker)
}
