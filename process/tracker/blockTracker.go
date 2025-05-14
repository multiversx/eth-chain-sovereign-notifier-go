package tracker

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/multiversx/eth-chain-sovereign-notifier-go/config"
)

type blockTracker struct {
	client           ETHClientHandler
	eventsByBlock    map[uint64][]types.Log // todo: here cacher component + cacher for blocks
	minConfirmations uint8

	addresses []common.Address
	topics    [][]common.Hash
}

// todo: here, pass directly expected eth data, not our config

func NewBlockTracker(subscribedEvents []config.SubscribedEvent) *blockTracker {
	return &blockTracker{
		eventsByBlock:    make(map[uint64][]types.Log),
		minConfirmations: 2,
	}
}

func (bt *blockTracker) Start(ctx context.Context, errChan chan error) {
	// also somewhere get latest block on a separate go routine to check block finality perhaps

	go bt.subscribeToNewHeaders(ctx, errChan)
	go bt.subscribeToEvents(ctx, errChan)
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
		case header := <-headers:
			bt.processBlock(header)
		case <-ctx.Done():
			return
		}
	}
}

func (bt *blockTracker) processBlock(header *types.Header) {
	_ = header
}

func (bt *blockTracker) subscribeToEvents(ctx context.Context, errChan chan error) {
	logs := make(chan types.Log)
	query := ethereum.FilterQuery{
		Addresses: bt.addresses,
		Topics:    bt.topics,
	}
	sub, err := bt.client.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		errChan <- fmt.Errorf("failed to subscribe to events: %v", err)
		return
	}

	defer sub.Unsubscribe()

	for {
		select {
		case err := <-sub.Err():
			errChan <- fmt.Errorf("event subscription error: %v", err)
			return
		case log := <-logs:
			bt.processEventLog(log)
		case <-ctx.Done():
			return
		}
	}
}

func (bt *blockTracker) processEventLog(log types.Log) {
	_ = log
}
