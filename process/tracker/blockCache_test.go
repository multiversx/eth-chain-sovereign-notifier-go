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

func TestBlockCache_ExtractFinalizedBlocks(t *testing.T) {
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

	header3 := &types.Header{Number: big.NewInt(101)}
	err = cache.Add(ctx, header3)

	header4 := &types.Header{Number: big.NewInt(102)}
	err = cache.Add(ctx, header4)

	header5 := &types.Header{Number: big.NewInt(100)}
	err = cache.Add(ctx, header5)

	finalizedHeaders := cache.ExtractFinalizedBlocks()
	require.Empty(t, finalizedHeaders)

	header6 := &types.Header{Number: big.NewInt(101)}
	err = cache.Add(ctx, header6)

	finalizedHeaders = cache.ExtractFinalizedBlocks()
	require.Equal(t, []*types.Header{header1}, finalizedHeaders)
	require.True(t, finalizedHeaders[0] == header1) // pointer check

	header7 := &types.Header{Number: big.NewInt(103)}
	err = cache.Add(ctx, header7)

	finalizedHeaders = cache.ExtractFinalizedBlocks()
	require.Equal(t, []*types.Header{header5, header6}, finalizedHeaders)
	require.True(t, finalizedHeaders[0] == header5) // pointer check
	require.True(t, finalizedHeaders[1] == header6) // pointer check

	header8 := &types.Header{Number: big.NewInt(104)}
	err = cache.Add(ctx, header8)

	finalizedHeaders = cache.ExtractFinalizedBlocks()
	require.Equal(t, []*types.Header{header4}, finalizedHeaders)
	require.True(t, finalizedHeaders[0] == header4) // pointer check

	require.Equal(t, []uint64{103, 104}, cache.nonceOrder)
	require.Equal(t, map[uint64]*types.Header{
		103: header7,
		104: header8,
	}, cache.headers)
}
