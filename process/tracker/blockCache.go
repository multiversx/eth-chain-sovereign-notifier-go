package tracker

import (
	"context"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
)

type blockCache struct {
	mut              sync.Mutex
	headers          map[uint64]*types.Header
	nonceOrder       []uint64 // Pentru pruning eficient
	maxSize          uint64
	minConfirmations uint64
	client           ETHClientHandler
}

type ArgsBlockCache struct {
	MaxSize          uint64
	MinConfirmations uint64
	Client           ETHClientHandler
}

func NewBlockCache(args ArgsBlockCache) (*blockCache, error) {
	if args.Client == nil || args.MaxSize == 0 {
		return nil, fmt.Errorf("invalid args: client=%v, maxSize=%d", args.Client, args.MaxSize)
	}
	return &blockCache{
		headers:          make(map[uint64]*types.Header),
		nonceOrder:       make([]uint64, 0, args.MaxSize),
		maxSize:          args.MaxSize,
		client:           args.Client,
		minConfirmations: args.MinConfirmations,
	}, nil
}

func (bc *blockCache) Add(ctx context.Context, header *types.Header) error {
	bc.mut.Lock()
	defer bc.mut.Unlock()

	hdrNonce := header.Number.Uint64()
	hash := header.Hash()

	if existingHdr, contains := bc.headers[hdrNonce]; contains && existingHdr.Hash() != hash {
		log.Debug("eth chain reorg detected", "nonce", hdrNonce, "old hash", existingHdr.Hash().Hex(), "new hash", hash.Hex())
		canonicalHdr, err := bc.client.HeaderByNumber(ctx, header.Number)
		if err != nil {
			return fmt.Errorf("blockCache.Add.client.HeaderByNumber error: %w, nonce: %d", err, hdrNonce)
		}
		if canonicalHdr.Hash() != hash {
			log.Debug("new header is not in canonical chain, discard it", "nonce", hdrNonce, "hash", hash.Hex())
			return nil
		}
	}

	bc.headers[hdrNonce] = header
	bc.nonceOrder = append(bc.nonceOrder, hdrNonce)

	log.Debug("Added header", "nonce", hdrNonce, "hash", hash.Hex())

	bc.resizeCacheIfNeeded()
	return nil
}

func (bc *blockCache) resizeCacheIfNeeded() {
	if len(bc.nonceOrder) > int(bc.maxSize) {
		for i := 0; i < len(bc.nonceOrder) && len(bc.nonceOrder) > int(bc.maxSize); i++ {
			log.Debug("Pruning block", "nonce", bc.nonceOrder[i])
			delete(bc.headers, bc.nonceOrder[i])
		}

		bc.nonceOrder = bc.nonceOrder[len(bc.nonceOrder)-int(bc.maxSize):]
	}
}

func (bc *blockCache) ExtractFinalizedBlocks() []*types.Header {
	bc.mut.Lock()
	defer bc.mut.Unlock()

	finalizedHeaders := make([]*types.Header, 0)
	if len(bc.nonceOrder) == 0 {
		return finalizedHeaders
	}

	latestNonce := bc.nonceOrder[len(bc.nonceOrder)-1]
	for i, nonce := range bc.nonceOrder {
		if nonce <= latestNonce-bc.minConfirmations {
			if header, found := bc.headers[nonce]; found {
				finalizedHeaders = append(finalizedHeaders, header)
				delete(bc.headers, nonce)
			}
		} else {
			bc.nonceOrder = bc.nonceOrder[i:]
			break
		}
	}

	log.Debug("Extracted finalized blocks", "count", len(finalizedHeaders))
	return finalizedHeaders
}
