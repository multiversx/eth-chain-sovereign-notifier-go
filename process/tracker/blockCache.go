package tracker

import (
	"context"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

const minCacheSize = 1
const maxCacheSize = 10000

type blockCache struct {
	mut              sync.Mutex
	headers          map[uint64]*types.Header
	nonceOrder       []uint64
	maxSize          uint64
	minConfirmations uint64
	client           ETHClientHandler
}

// ArgsBlockCache is a struct placeholder containing needed args for a block tracker cache
type ArgsBlockCache struct {
	MaxSize          uint64
	MinConfirmations uint64
	Client           ETHClientHandler
}

// NewBlockCache creates a new eth block tracker cacher
func NewBlockCache(args ArgsBlockCache) (*blockCache, error) {
	if check.IfNil(args.Client) {
		return nil, errNilClient
	}

	if args.MaxSize < minCacheSize || args.MaxSize > maxCacheSize {
		return nil, fmt.Errorf("%w: %d, min value: %d, max value: %d",
			errInvalidMaxCacheSize, args.MaxSize, minCacheSize, maxCacheSize)
	}

	if args.MinConfirmations > args.MaxSize {
		return nil, fmt.Errorf("%w : %d, should be less than max cache size: %d",
			errInvalidMinConfirmations, args.MinConfirmations, args.MaxSize)
	}

	return &blockCache{
		headers:          make(map[uint64]*types.Header),
		nonceOrder:       make([]uint64, 0, args.MaxSize),
		maxSize:          args.MaxSize,
		client:           args.Client,
		minConfirmations: args.MinConfirmations,
	}, nil
}

// Add inserts a block header into the cache, handling chain reorgs. It checks for reorgs by comparing hashes, verifies
// canonicity with HeaderByNumber, and discards non-canonical headers. For new blocks, it appends the nonce to nonceOrder.
// Updates the cache, logs the action, and resizes if needed.
func (bc *blockCache) Add(ctx context.Context, header *types.Header) error {
	bc.mut.Lock()
	defer bc.mut.Unlock()

	hdrNonce := header.Number.Uint64()
	hash := header.Hash()

	existingHdr, contains := bc.headers[hdrNonce]
	if contains && existingHdr.Hash() != hash {
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

	if !contains {
		bc.nonceOrder = append(bc.nonceOrder, hdrNonce)
	}

	bc.headers[hdrNonce] = header
	log.Debug("blockCache.Add", "nonce", hdrNonce, "hash", hash.Hex())

	bc.resizeCacheIfNeeded()
	return nil
}

func (bc *blockCache) resizeCacheIfNeeded() {
	numToRemove := len(bc.nonceOrder) - int(bc.maxSize)
	if numToRemove > 0 {
		for i := 0; i < numToRemove; i++ {
			log.Debug("blockCache.resizeCacheIfNeeded pruning block", "nonce", bc.nonceOrder[i])
			delete(bc.headers, bc.nonceOrder[i])
		}
		bc.nonceOrder = bc.nonceOrder[numToRemove:]
	}
}

// ExtractFinalizedBlocks returns headers with sufficient confirmations (nonce <= latestNonce - minConfirmations) and
// removes them from the cache. If minConfirmations==0, clears nonceOrder entirely.
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
			return finalizedHeaders
		}
	}

	bc.nonceOrder = bc.nonceOrder[:0]
	return finalizedHeaders
}

// IsInterfaceNil checks if the underlying pointer is nil
func (bc *blockCache) IsInterfaceNil() bool {
	return bc == nil
}
