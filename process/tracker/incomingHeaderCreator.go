package tracker

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

type incomingHeaderCreator struct {
}

// NewIncomingHeadersCreator creates a new incoming header creator
func NewIncomingHeadersCreator() *incomingHeaderCreator {
	return &incomingHeaderCreator{}
}

// CreateIncomingHeader will create an incoming header for MVX chain, based on the provided ETH header with its incoming logs
// For now, the proof represents the json bytes of the ETH header.
func (ihc *incomingHeaderCreator) CreateIncomingHeader(header *types.Header, logs []types.Log) (sovereign.IncomingHeaderHandler, error) {
	bytes, err := header.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return &sovereign.IncomingHeader{
		Proof:          bytes,
		SourceChainID:  dto.ETH,
		Nonce:          header.Nonce.Uint64(),
		IncomingEvents: createIncomingEvents(logs),
	}, nil
}

func createIncomingEvents(logs []types.Log) []*transaction.Event {
	incomingEvents := make([]*transaction.Event, len(logs))

	for idx, ethLog := range logs {
		incomingEvents[idx] = &transaction.Event{
			Address:    ethLog.Address.Bytes(),
			Identifier: nil, // todo
			Topics:     getTopics(ethLog.Topics),
			Data:       ethLog.Data,
		}
	}

	return incomingEvents
}

func getTopics(topics []common.Hash) [][]byte {
	res := make([][]byte, len(topics))
	for idx, topic := range topics {
		res[idx] = topic.Bytes()
	}

	return res
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ihc *incomingHeaderCreator) IsInterfaceNil() bool {
	return ihc == nil
}
