package tracker

import (
	"context"
	"fmt"
	"math/big"

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
	client           ETHClientHandler
	minConfirmations uint8

	subscribedETHEvents []SubscribedETHEvent

	finalizedBlockNonce uint64

	incomingHeadersNotifier IncomingHeadersNotifierHandler
	incomingHeaderCreator   IncomingHeaderCreator
}

// todo: here, pass directly expected eth data, not our config

type ArgsETHBlockTracker struct {
	SubscribedETHEvents []SubscribedETHEvent

	MinConfirmations        uint8
	Client                  ETHClientHandler
	IncomingHeaderCreator   IncomingHeaderCreator
	IncomingHeadersNotifier IncomingHeadersNotifierHandler
}

func NewBlockTracker(args ArgsETHBlockTracker) (*blockTracker, error) {
	return &blockTracker{
		client:                  args.Client,
		minConfirmations:        args.MinConfirmations,
		subscribedETHEvents:     args.SubscribedETHEvents,
		incomingHeadersNotifier: args.IncomingHeadersNotifier,
		incomingHeaderCreator:   args.IncomingHeaderCreator,
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
	defer sub.Unsubscribe()

	for {
		select {
		case err = <-sub.Err():
			log.Error("DASDADAAA", "err", err)
			return
		case header := <-headers:
			bt.finalizedBlockNonce = header.Number.Uint64() - uint64(bt.minConfirmations)
			err = bt.processBlock(ctx)
			log.LogIfError(err)
		case <-ctx.Done():
			sub.Unsubscribe()
			return
		}
	}
}

func (bt *blockTracker) processBlock(ctx context.Context) error {
	logs := make([]types.Log, 0)

	for _, subEvent := range bt.subscribedETHEvents {
		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(bt.finalizedBlockNonce)),
			ToBlock:   big.NewInt(int64(bt.finalizedBlockNonce)),
			Addresses: []common.Address{subEvent.Address},
			Topics:    [][]common.Hash{{subEvent.Topic}}, // matches topic in first position
		}

		currLogs, err := bt.client.FilterLogs(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to filter logs: %v", err)

		}

		logs = append(logs, currLogs...)
	}

	log.Info("DSADAS", "num", bt.finalizedBlockNonce)

	finalizedHeader, err := bt.client.HeaderByNumber(ctx, big.NewInt(int64(bt.finalizedBlockNonce)))
	if err != nil {
		return fmt.Errorf("failed to get header by number: %v", err)
	}

	incomingHeader, err := bt.incomingHeaderCreator.CreateIncomingHeader(finalizedHeader, logs)
	if err != nil {
		return err
	}

	err = bt.incomingHeadersNotifier.NotifyHeaderSubscribers(incomingHeader)
	if err != nil {
		return err
	}

	return nil
}

func (bt *blockTracker) Close() {
	//todo: here, close on chan start
}
