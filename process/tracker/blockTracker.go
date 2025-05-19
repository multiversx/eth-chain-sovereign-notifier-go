package tracker

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("eth-block-tracker")

type SubscribedETHEvent struct {
	Address common.Address
	Topic   common.Hash
}

type blockTracker struct {
	minConfirmations    uint8
	subscribedETHEvents []SubscribedETHEvent

	client                  ETHClientHandler
	incomingHeadersNotifier IncomingHeadersNotifierHandler
	incomingHeaderCreator   IncomingHeaderCreator
	blockCache              BlockCache
}

// todo: here, pass directly expected eth data, not our config

type ArgsETHBlockTracker struct {
	SubscribedETHEvents []SubscribedETHEvent

	MinConfirmations        uint8
	Client                  ETHClientHandler
	IncomingHeaderCreator   IncomingHeaderCreator
	IncomingHeadersNotifier IncomingHeadersNotifierHandler
	BlockCache              BlockCache
}

func NewBlockTracker(args ArgsETHBlockTracker) (*blockTracker, error) {
	return &blockTracker{
		client:                  args.Client,
		minConfirmations:        args.MinConfirmations,
		subscribedETHEvents:     args.SubscribedETHEvents,
		incomingHeadersNotifier: args.IncomingHeadersNotifier,
		incomingHeaderCreator:   args.IncomingHeaderCreator,
		blockCache:              args.BlockCache,
	}, nil
}

func (bt *blockTracker) Start(ctx context.Context) {
	bt.subscribeToNewHeaders(ctx)
}

func (bt *blockTracker) subscribeToNewHeaders(ctx context.Context) {
	headers := make(chan *types.Header)
	sub, err := bt.client.SubscribeNewHead(ctx, headers)
	if err != nil {
		log.LogIfError(fmt.Errorf("failed to subscribe to new headers: %v", err))
		return
	}

	defer func() {
		sub.Unsubscribe()
		bt.client.Close()
	}()

	for {
		select {
		case err = <-sub.Err():
			log.Error("blockTracker.subscribeToNewHeaders", "err", err)
			return
		case header := <-headers:
			err = bt.processBlock(ctx, header)
			log.LogIfError(err)
		case <-ctx.Done():
			return
		}
	}
}

func (bt *blockTracker) processBlock(ctx context.Context, header *types.Header) error {
	log.Info("received new eth block in tracker, will process latest finalized block", "nonce", header.Number.Uint64())

	errCache := bt.blockCache.Add(ctx, header)
	if errCache != nil {
		return errCache
	}

	finalizedHeaders := bt.blockCache.ExtractFinalizedBlocks()
	if len(finalizedHeaders) == 0 {
		return nil
	}

	for _, finalizedHeader := range finalizedHeaders {
		logs, err := bt.getLogs(ctx, finalizedHeader)
		if err != nil {
			return err
		}

		incomingHeader, err := bt.incomingHeaderCreator.CreateIncomingHeader(finalizedHeader, logs)
		if err != nil {
			return err
		}

		err = bt.incomingHeadersNotifier.NotifyHeaderSubscribers(incomingHeader)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bt *blockTracker) getLogs(ctx context.Context, header *types.Header) ([]types.Log, error) {
	logs := make([]types.Log, 0)
	for _, subEvent := range bt.subscribedETHEvents {
		query := ethereum.FilterQuery{
			FromBlock: header.Number,
			ToBlock:   header.Number,
			Addresses: []common.Address{subEvent.Address},
			Topics:    [][]common.Hash{{subEvent.Topic}}, // matches topic in first position
		}

		currLogs, err := bt.client.FilterLogs(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to filter logs: %v", err)

		}

		logs = append(logs, currLogs...)
	}

	return logs, nil
}

func (bt *blockTracker) Close() {
	// todo: here chan closer and add in Start for {} loop
	bt.client.Close()
}
