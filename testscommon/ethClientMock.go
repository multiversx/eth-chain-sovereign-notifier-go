package testscommon

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

// ETHClientHandlerMock is a mock implementation of the ETHClientHandler interface
type ETHClientHandlerMock struct {
	DialCalled             func() error
	SubscribeNewHeadCalled func(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)
	FilterLogsCalled       func(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
	HeaderByNumberCalled   func(ctx context.Context, number *big.Int) (*types.Header, error)
	CloseCalled            func()
}

// Dial mocks the Dial method
func (mock *ETHClientHandlerMock) Dial() error {
	if mock.DialCalled != nil {
		return mock.DialCalled()
	}
	return nil
}

// SubscribeNewHead mocks the SubscribeNewHead method
func (mock *ETHClientHandlerMock) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	if mock.SubscribeNewHeadCalled != nil {
		return mock.SubscribeNewHeadCalled(ctx, ch)
	}
	return nil, nil
}

// FilterLogs mocks the FilterLogs method
func (mock *ETHClientHandlerMock) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	if mock.FilterLogsCalled != nil {
		return mock.FilterLogsCalled(ctx, q)
	}
	return nil, nil
}

// HeaderByNumber mocks the HeaderByNumber method
func (mock *ETHClientHandlerMock) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	if mock.HeaderByNumberCalled != nil {
		return mock.HeaderByNumberCalled(ctx, number)
	}
	return nil, nil
}

// Close mocks the Close method
func (mock *ETHClientHandlerMock) Close() {
	if mock.CloseCalled != nil {
		mock.CloseCalled()
	}
}

// IsInterfaceNil checks if the mock is nil
func (mock *ETHClientHandlerMock) IsInterfaceNil() bool {
	return mock == nil
}
