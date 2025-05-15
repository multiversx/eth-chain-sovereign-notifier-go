package tracker

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/multiversx/mx-chain-core-go/core/sovereign"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
)

type blockTracker struct {
	client           ETHClientHandler
	blocks           map[common.Hash]*types.Header // todo: here cacher component + cacher for blocks
	minConfirmations uint8

	addresses []common.Address
	topics    [][]common.Hash

	finalizedBlockNonce uint64

	incomingHeadersNotifier IncomingHeadersNotifierHandler
	incomingHeaderCreator   IncomingHeaderCreator
}

// todo: here, pass directly expected eth data, not our config

type ArgsETHBlockTracker struct {
	Marshaller marshal.Marshalizer
	Hasher     hashing.Hasher
}

func NewBlockTracker(args ArgsETHBlockTracker) (*blockTracker, error) {

	hn, err := sovereign.NewHeadersNotifier(args.Marshaller, args.Hasher)
	if err != nil {
		return nil, err
	}

	return &blockTracker{
		blocks:                  map[common.Hash]*types.Header{},
		minConfirmations:        2,
		incomingHeadersNotifier: hn,
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
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(bt.finalizedBlockNonce)),
		ToBlock:   big.NewInt(int64(bt.finalizedBlockNonce)),
		Addresses: bt.addresses,
		Topics:    bt.topics,
	}

	logs, err := bt.client.FilterLogs(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to filter logs: %v", err)

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
