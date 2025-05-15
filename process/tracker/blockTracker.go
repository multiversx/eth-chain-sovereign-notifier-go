package tracker

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

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

func (bt *blockTracker) Start(ctx context.Context, errChan chan error) {
	latestHeader, err := bt.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return
	}

	bt.finalizedBlockNonce = latestHeader.Nonce.Uint64() - uint64(bt.minConfirmations)

	bt.subscribeToNewHeaders(ctx, errChan)
}

func (bt *blockTracker) subscribeToNewHeaders(ctx context.Context, errChan chan error) {
	headers := make(chan *types.Header)
	sub, err := bt.client.SubscribeNewHead(ctx, headers)
	if err != nil {
		errChan <- fmt.Errorf("failed to subscribe to new headers: %v", err)
		return
	}
	defer sub.Unsubscribe()

	for {
		select {
		case err := <-sub.Err():
			errChan <- fmt.Errorf("header subscription error: %v", err)
			return
		case <-headers:
			errChan <- bt.processBlock(ctx)
		case <-ctx.Done():
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

	bt.finalizedBlockNonce++
	return nil
}

func (bt *blockTracker) Close() {
	//todo: here, close on chan start
}
