package tracker

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/closing"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("eth-block-tracker")

type SubscribedETHEvent struct {
	Address common.Address
	Topic   common.Hash
}

type blockTrackerNotifier struct {
	minConfirmations    uint8
	subscribedETHEvents []SubscribedETHEvent

	closer                  core.SafeCloser
	client                  ETHClientHandler
	blockCache              BlockCache
	incomingHeadersNotifier IncomingHeadersNotifierHandler
	incomingHeaderCreator   IncomingHeaderCreator
}

// ArgsETHBlockTracker is a struct placeholder for args needed to create a block tracker
type ArgsETHBlockTracker struct {
	MinConfirmations    uint8
	SubscribedETHEvents []SubscribedETHEvent

	Client                  ETHClientHandler
	BlockCache              BlockCache
	IncomingHeaderCreator   IncomingHeaderCreator
	IncomingHeadersNotifier IncomingHeadersNotifierHandler
}

// NewBlockTrackerNotifier creates a new eth block tracker notifier
func NewBlockTrackerNotifier(args ArgsETHBlockTracker) (*blockTrackerNotifier, error) {
	return &blockTrackerNotifier{
		client:                  args.Client,
		closer:                  closing.NewSafeChanCloser(),
		minConfirmations:        args.MinConfirmations,
		subscribedETHEvents:     args.SubscribedETHEvents,
		incomingHeadersNotifier: args.IncomingHeadersNotifier,
		incomingHeaderCreator:   args.IncomingHeaderCreator,
		blockCache:              args.BlockCache,
	}, nil
}

// Start will start subscribing to new incoming headers. Upon receiving a new block, it will store it in its
// block cache tracker and check for confirmed(finalized) blocks. If any finalized blocks are found
// it will create its corresponding incoming header and notify subscribed components.
func (btn *blockTrackerNotifier) Start(ctx context.Context) error {
	headers := make(chan *types.Header)
	sub, err := btn.client.SubscribeNewHead(ctx, headers)
	if err != nil {
		return err
	}

	defer func() {
		sub.Unsubscribe()
		btn.Close()
	}()

	for {
		select {
		case err = <-sub.Err():
			log.Error("blockTrackerNotifier.subscribeToNewHeaders", "err", err)
			return err
		case header := <-headers:
			err = btn.processBlock(ctx, header)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			log.Debug("blockTrackerNotifier.btn.ctx.Done()")
			return nil
		case <-btn.closer.ChanClose():
			log.Debug("blockTrackerNotifier.btn.closer.ChanClose()")
			return nil

		}
	}
}

func (btn *blockTrackerNotifier) processBlock(ctx context.Context, header *types.Header) error {
	log.Info("received new ETH block in tracker", "nonce", header.Number.Uint64(), "hash", header.Hash().Hex())

	errCache := btn.blockCache.Add(ctx, header)
	if errCache != nil {
		return errCache
	}

	finalizedHeaders := btn.blockCache.ExtractFinalizedBlocks()
	if len(finalizedHeaders) == 0 {
		return nil
	}

	for _, finalizedHeader := range finalizedHeaders {
		logs, err := btn.getLogs(ctx, finalizedHeader)
		if err != nil {
			return err
		}

		incomingHeader, err := btn.incomingHeaderCreator.CreateIncomingHeader(finalizedHeader, logs)
		if err != nil {
			return err
		}

		err = btn.incomingHeadersNotifier.NotifyHeaderSubscribers(incomingHeader)
		if err != nil {
			return err
		}
	}

	return nil
}

func (btn *blockTrackerNotifier) getLogs(ctx context.Context, header *types.Header) ([]types.Log, error) {
	logs := make([]types.Log, 0)
	for _, subEvent := range btn.subscribedETHEvents {
		query := ethereum.FilterQuery{
			FromBlock: header.Number,
			ToBlock:   header.Number,
			Addresses: []common.Address{subEvent.Address},
			Topics:    [][]common.Hash{{subEvent.Topic}}, // query to match specific topic in first position
		}

		currLogs, err := btn.client.FilterLogs(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to filter logs: %w", err)

		}

		logs = append(logs, currLogs...)
	}

	return logs, nil
}

// Close will close the underlying client and closer chan
func (btn *blockTrackerNotifier) Close() {
	defer btn.closer.Close() // should always be last
	btn.client.Close()
}
