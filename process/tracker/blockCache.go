package tracker

import (
	"context"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
)

type blockCache struct {
	mut     sync.Mutex
	headers map[uint64]*types.Header

	maxSize      uint64
	highestNonce uint64
	oldestNonce  uint64

	minConfirmations uint64

	client ETHClientHandler
}

type ArgsBlockCache struct {
	MaxSize          uint64
	MinConfirmations uint64
	Client           ETHClientHandler
}

func NewBlockCache(args ArgsBlockCache) (*blockCache, error) {
	return &blockCache{
		mut:              sync.Mutex{},
		headers:          make(map[uint64]*types.Header),
		maxSize:          0,
		highestNonce:     0,
		oldestNonce:      0,
		minConfirmations: 0,
		client:           nil,
	}, nil
}

func (bc *blockCache) Add(ctx context.Context, header *types.Header) error {
	bc.mut.Lock()
	defer bc.mut.Unlock()

	hdrNonce := header.Number.Uint64()
	hash := header.Hash()

	if existingHdr, contains := bc.headers[hdrNonce]; contains && existingHdr.Hash() != hash {
		log.Debug("eth chain reorg detected",
			"nonce", hdrNonce,
			"old hash", existingHdr.Hash().Hex(),
			"new hash", hash.Hex(),
		)

		canonicalHdr, err := bc.client.HeaderByNumber(ctx, header.Number)
		if err != nil {
			return fmt.Errorf("blockCache.Add.client.HeaderByNumber error: %w, nonce: %d", err, hdrNonce)
		}

		if canonicalHdr.Hash() != hash {
			log.Debug("new header is not in canonical chain, discard it", "nonce", hdrNonce, "hash", hash.Hex())
			return nil
		}
	}

	bc.updateInternalData(hdrNonce)
	return nil
}

func (bc *blockCache) updateInternalData(hdrNonce uint64) {
	if hdrNonce > bc.highestNonce {
		bc.highestNonce = hdrNonce
	}
	if hdrNonce < bc.oldestNonce {
		bc.oldestNonce = hdrNonce
	}

	bc.resizeCacheIfNeeded()
}

func (bc *blockCache) resizeCacheIfNeeded() {
	if len(bc.headers) > int(bc.maxSize) {
		delete(bc.headers, bc.oldestNonce)
		bc.oldestNonce++
	}
}

func (bc *blockCache) ExtractFinalizedBlocks() []*types.Header {
	bc.mut.Lock()
	defer bc.mut.Unlock()

	finalizedHeaders := make([]*types.Header, 0)
	for nonce := bc.oldestNonce; nonce < bc.highestNonce-bc.minConfirmations; nonce++ {
		if header, found := bc.headers[nonce]; found {
			finalizedHeaders = append(finalizedHeaders, header)
			delete(bc.headers, nonce)
		}

		bc.oldestNonce++
	}

	return finalizedHeaders
}
