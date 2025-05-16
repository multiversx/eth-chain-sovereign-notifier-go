package tracker

import (
	"context"
	"encoding/binary"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/eth-chain-sovereign-notifier-go/testscommon"
)

func createHash(id uint64) common.Hash {
	var hash [32]byte
	binary.BigEndian.PutUint64(hash[24:], id)
	return hash
}

func TestBlockCache_ExtractFinalizedBlocksWithChainReorganization(t *testing.T) {
	t.Parallel()

	logger.SetLogLevel("*:DEBUG")

	cache, _ := NewBlockCache(ArgsBlockCache{
		MaxSize:          10,
		MinConfirmations: 2,
		Client:           &testscommon.ETHClientHandlerMock{},
	})

	ctx := context.Background()

	header1 := &types.Header{Number: big.NewInt(99)}
	err := cache.Add(ctx, header1)
	require.Nil(t, err)

	header2 := &types.Header{Number: big.NewInt(100)}
	err = cache.Add(ctx, header2)
	require.Nil(t, err)

	header3 := &types.Header{Number: big.NewInt(101)}
	err = cache.Add(ctx, header3)
	require.Nil(t, err)

	header4 := &types.Header{Number: big.NewInt(102)}
	err = cache.Add(ctx, header4)
	require.Nil(t, err)

	// reorg chain
	header5 := &types.Header{Number: big.NewInt(100)}
	err = cache.Add(ctx, header5)
	require.Nil(t, err)

	finalizedHeaders := cache.ExtractFinalizedBlocks()
	require.Equal(t, []*types.Header{header1, header2}, finalizedHeaders)
	require.True(t, finalizedHeaders[0] == header1) // pointer check
	require.True(t, finalizedHeaders[1] == header5) // pointer check
	require.Equal(t, []uint64{101, 102}, cache.nonceOrder)

	header6 := &types.Header{Number: big.NewInt(101)}
	err = cache.Add(ctx, header6)
	require.Nil(t, err)

	finalizedHeaders = cache.ExtractFinalizedBlocks()
	require.Empty(t, finalizedHeaders)
	require.Equal(t, []uint64{101, 102}, cache.nonceOrder)

	header7 := &types.Header{Number: big.NewInt(103)}
	err = cache.Add(ctx, header7)
	require.Nil(t, err)

	finalizedHeaders = cache.ExtractFinalizedBlocks()
	require.Equal(t, []*types.Header{header6}, finalizedHeaders)
	require.True(t, finalizedHeaders[0] == header6) // pointer check

	header8 := &types.Header{Number: big.NewInt(104)}
	err = cache.Add(ctx, header8)
	require.Nil(t, err)

	finalizedHeaders = cache.ExtractFinalizedBlocks()
	require.Equal(t, []*types.Header{header4}, finalizedHeaders)
	require.True(t, finalizedHeaders[0] == header4) // pointer check

	require.Equal(t, []uint64{103, 104}, cache.nonceOrder)
	require.Equal(t, map[uint64]*types.Header{
		103: header7,
		104: header8,
	}, cache.headers)
}

func TestBlockCache_ExtractFinalizedBlocksWithDiscardedChainReorg(t *testing.T) {
	t.Parallel()

	logger.SetLogLevel("*:DEBUG")

	cache, _ := NewBlockCache(ArgsBlockCache{
		MaxSize:          10,
		MinConfirmations: 2,
		Client:           &testscommon.ETHClientHandlerMock{},
	})

	ctx := context.Background()

	header1 := &types.Header{Number: big.NewInt(99), Root: createHash(1)}
	err := cache.Add(ctx, header1)
	require.Nil(t, err)

	header2 := &types.Header{Number: big.NewInt(100), Root: createHash(2)}
	err = cache.Add(ctx, header2)
	require.Nil(t, err)

	header3 := &types.Header{Number: big.NewInt(101), Root: createHash(3)}
	err = cache.Add(ctx, header3)
	require.Nil(t, err)

	// Chain reorganization with discarded header
	header4 := &types.Header{Number: big.NewInt(100), Root: createHash(4)}
	err = cache.Add(ctx, header4)
	require.Nil(t, err)

	header5 := &types.Header{Number: big.NewInt(102), Root: createHash(5)}
	err = cache.Add(ctx, header5)
	require.Nil(t, err)

	header6 := &types.Header{Number: big.NewInt(103), Root: createHash(6)}
	err = cache.Add(ctx, header6)
	require.Nil(t, err)

	finalizedHeaders := cache.ExtractFinalizedBlocks()
	require.Equal(t, []*types.Header{header1, header2, header3}, finalizedHeaders)
	require.True(t, finalizedHeaders[0] == header1) // pointer check
	require.True(t, finalizedHeaders[1] == header2) // pointer check
	require.True(t, finalizedHeaders[2] == header3) // pointer check

	// Discarded header is not in map
	for _, headerInMap := range cache.headers {
		require.False(t, headerInMap == header4)
	}

	require.Equal(t, []uint64{102, 103}, cache.nonceOrder)
	require.Equal(t, map[uint64]*types.Header{
		102: header5,
		103: header6,
	}, cache.headers)
}

func TestBlockCache_CacheSizeFull(t *testing.T) {
	t.Parallel()

	logger.SetLogLevel("*:DEBUG")

	cache, _ := NewBlockCache(ArgsBlockCache{
		MaxSize:          5,
		MinConfirmations: 0,
		Client:           &testscommon.ETHClientHandlerMock{},
	})

	ctx := context.Background()

	initialHdr := make([]*types.Header, 0)
	for i := int64(1); i < 8; i++ {
		header := &types.Header{Number: big.NewInt(i), Root: createHash(uint64(i))}
		err := cache.Add(ctx, header)
		require.Nil(t, err)

		initialHdr = append(initialHdr, header)
	}

	require.Equal(t, []uint64{3, 4, 5, 6, 7}, cache.nonceOrder)
	initialHdr = initialHdr[2:]

	headerReorgNonce5 := &types.Header{Number: big.NewInt(5), Root: createHash(5)}
	err := cache.Add(ctx, headerReorgNonce5)
	require.Nil(t, err)

	headerReorgNonce6Discarded := &types.Header{Number: big.NewInt(6), Root: createHash(100)}
	err = cache.Add(ctx, headerReorgNonce6Discarded)
	require.Nil(t, err)
	require.Equal(t, []uint64{3, 4, 5, 6, 7}, cache.nonceOrder)

	extractedHeaders := cache.ExtractFinalizedBlocks()
	require.Len(t, extractedHeaders, 5)
	require.Empty(t, cache.nonceOrder)
	require.Empty(t, cache.headers)

	for i := 0; i < 5; i++ {
		if i == 2 {
			require.True(t, extractedHeaders[2] == headerReorgNonce5)
			continue
		}

		require.True(t, extractedHeaders[i] == initialHdr[i])
	}
}
