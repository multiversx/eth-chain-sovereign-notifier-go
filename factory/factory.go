package factory

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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

	subscribedEvents, err := createETHSubscribedEvents(cfg.SubscribedEvents)
	if err != nil {
		return nil, err
	}

	argsBlockCache := tracker.ArgsBlockCache{
		MaxSize:          cfg.BlockCacheSize,
		MinConfirmations: uint64(cfg.MinBlocksConfirmation),
		Client:           ethClient,
	}
	blockCache, err := tracker.NewBlockCache(argsBlockCache)
	if err != nil {
		return nil, err
	}

	argsBlockTracker := tracker.ArgsETHBlockTracker{
		SubscribedETHEvents:     subscribedEvents,
		MinConfirmations:        cfg.MinBlocksConfirmation,
		Client:                  ethClient,
		IncomingHeaderCreator:   tracker.NewIncomingHeadersCreator(),
		IncomingHeadersNotifier: headersNotifier,
		BlockCache:              blockCache,
	}

	return tracker.NewBlockTrackerNotifier(argsBlockTracker)
}

func createETHSubscribedEvents(subscribedEvents []config.SubscribedEvent) ([]tracker.SubscribedETHEvent, error) {
	ret := make([]tracker.SubscribedETHEvent, len(subscribedEvents))
	for idx, subEvent := range subscribedEvents {
		ethAddr := common.HexToAddress(subEvent.Address)
		if len(ethAddr.String()) == 0 {
			return nil, fmt.Errorf("%w from config, value: %s", errInvalidETHAddress, subEvent.Address)
		}

		eventSignature := []byte(subEvent.Identifier)
		eventHash := crypto.Keccak256Hash(eventSignature)

		ret[idx] = tracker.SubscribedETHEvent{
			Address: ethAddr,
			Topic:   eventHash,
		}
	}

	return ret, nil
}
