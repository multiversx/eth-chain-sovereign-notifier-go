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

func NewIncomingHeadersCreator() *incomingHeaderCreator {
	return &incomingHeaderCreator{}
}

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

	for idx, log := range logs {
		incomingEvents[idx] = &transaction.Event{
			Address:    log.Address.Bytes(),
			Identifier: nil, // todo
			Topics:     getTopics(log.Topics),
			Data:       log.Data,
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
