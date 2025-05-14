package tracker

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/multiversx/eth-chain-sovereign-notifier-go/config"
)

type blockTracker struct {
	client           ETHClientHandler
	blocks           map[common.Hash]*types.Header // todo: here cacher component + cacher for blocks
	minConfirmations uint8

	addresses []common.Address
	topics    [][]common.Hash

	finalizedBlockNonce uint64

	incomingHeaderCreator IncomingHeaderCreator
}

// todo: here, pass directly expected eth data, not our config

func NewBlockTracker(subscribedEvents []config.SubscribedEvent, client ETHClientHandler) *blockTracker {

	return &blockTracker{
		blocks:           map[common.Hash]*types.Header{},
		minConfirmations: 2,
	}
}

func (bt *blockTracker) Start(ctx context.Context, errChan chan error) {
	// also somewhere get latest block on a separate go routine to check block finality perhaps

	latestHeader, err := bt.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return
	}

	bt.finalizedBlockNonce = latestHeader.Nonce.Uint64() - uint64(bt.minConfirmations)

	go bt.subscribeToNewHeaders(ctx, errChan)
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
			bt.processBlock(ctx, header, errChan)
		case <-ctx.Done():
			return
		}
	}
}

func (bt *blockTracker) processBlock(ctx context.Context, header *types.Header, errChan chan error) {
	query := ethereum.FilterQuery{
		FromBlock: header.Number,
		ToBlock:   header.Number,
		Addresses: bt.addresses,
		Topics:    bt.topics,
	}

	logs, err := bt.client.FilterLogs(ctx, query)
	if err != nil {
		errChan <- fmt.Errorf("failed to subscribe to events: %v", err)
		return
	}

	finalizedHeader, err := bt.client.HeaderByNumber(ctx, big.NewInt(int64(bt.finalizedBlockNonce)))
	if err != nil {
		errChan <- fmt.Errorf("failed to subscribe to events: %v", err)
		return
	}

	bt.incomingHeaderCreator.CreateIncomingHeader(finalizedHeader, logs)
}
