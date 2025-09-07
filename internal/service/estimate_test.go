package service

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/fleshka4/1inch-test-task/internal/infra/uniswap/mock"
	"github.com/fleshka4/1inch-test-task/internal/service/dto"
)

func TestEstimate(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockClient(ctrl)
	service := NewEstimatorService(mockClient)

	poolAddr := common.HexToAddress("0x1234")
	token0 := common.HexToAddress("0x5678")
	token1 := common.HexToAddress("0x12345678")
	srcAmount := big.NewInt(1000)

	t.Run("success token0 to token1", func(t *testing.T) {
		mockClient.EXPECT().
			GetPairTokens(gomock.Any(), poolAddr).
			Return(token0, token1, nil)
		mockClient.EXPECT().
			GetPairReserves(gomock.Any(), poolAddr).
			Return(big.NewInt(10000), big.NewInt(20000), nil)

		req := dto.EstimateRequest{
			Pool:      poolAddr,
			Src:       token0,
			Dst:       token1,
			SrcAmount: srcAmount,
		}

		result, err := service.Estimate(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, result.Cmp(big.NewInt(0)) > 0)
	})

	t.Run("invalid argument - empty pool", func(t *testing.T) {
		req := dto.EstimateRequest{
			Pool:      common.Address{},
			Src:       token0,
			Dst:       token1,
			SrcAmount: srcAmount,
		}

		result, err := service.Estimate(context.Background(), req)
		require.Error(t, err)
		require.Equal(t, ErrInvalidArgument, err)
		require.Nil(t, result)
	})

	t.Run("invalid argument - empty src", func(t *testing.T) {
		req := dto.EstimateRequest{
			Pool:      poolAddr,
			Src:       common.Address{},
			Dst:       token1,
			SrcAmount: srcAmount,
		}

		result, err := service.Estimate(context.Background(), req)
		require.Error(t, err)
		require.Equal(t, ErrInvalidArgument, err)
		require.Nil(t, result)
	})

	t.Run("invalid argument - empty dst", func(t *testing.T) {
		req := dto.EstimateRequest{
			Pool:      poolAddr,
			Src:       token0,
			Dst:       common.Address{},
			SrcAmount: srcAmount,
		}

		result, err := service.Estimate(context.Background(), req)
		require.Error(t, err)
		require.Equal(t, ErrInvalidArgument, err)
		require.Nil(t, result)
	})

	t.Run("invalid argument - same src and dst", func(t *testing.T) {
		req := dto.EstimateRequest{
			Pool:      poolAddr,
			Src:       token0,
			Dst:       token0,
			SrcAmount: srcAmount,
		}

		result, err := service.Estimate(context.Background(), req)
		require.Error(t, err)
		require.Equal(t, ErrInvalidArgument, err)
		require.Nil(t, result)
	})

	t.Run("invalid argument - zero src amount", func(t *testing.T) {
		req := dto.EstimateRequest{
			Pool:      poolAddr,
			Src:       token0,
			Dst:       token1,
			SrcAmount: big.NewInt(0),
		}

		result, err := service.Estimate(context.Background(), req)
		require.Error(t, err)
		require.Equal(t, ErrInvalidArgument, err)
		require.Nil(t, result)
	})

	t.Run("invalid argument - negative src amount", func(t *testing.T) {
		req := dto.EstimateRequest{
			Pool:      poolAddr,
			Src:       token0,
			Dst:       token1,
			SrcAmount: big.NewInt(-100),
		}

		result, err := service.Estimate(context.Background(), req)
		require.Error(t, err)
		require.Equal(t, ErrInvalidArgument, err)
		require.Nil(t, result)
	})

	t.Run("pair read error - GetPairTokens fails", func(t *testing.T) {
		mockClient.EXPECT().
			GetPairTokens(gomock.Any(), poolAddr).
			Return(common.Address{}, common.Address{}, errors.New("RPC error"))

		req := dto.EstimateRequest{
			Pool:      poolAddr,
			Src:       token0,
			Dst:       token1,
			SrcAmount: srcAmount,
		}

		result, err := service.Estimate(context.Background(), req)
		require.Error(t, err)
		require.Equal(t, ErrPairRead, err)
		require.Nil(t, result)
	})

	t.Run("pair read error - GetPairReserves fails", func(t *testing.T) {
		mockClient.EXPECT().
			GetPairTokens(gomock.Any(), poolAddr).
			Return(token0, token1, nil)
		mockClient.EXPECT().
			GetPairReserves(gomock.Any(), poolAddr).
			Return(nil, nil, errors.New("RPC error"))

		req := dto.EstimateRequest{
			Pool:      poolAddr,
			Src:       token0,
			Dst:       token1,
			SrcAmount: srcAmount,
		}

		result, err := service.Estimate(context.Background(), req)
		require.Error(t, err)
		require.Equal(t, ErrPairRead, err)
		require.Nil(t, result)
	})

	t.Run("invalid argument - src and dst not in pair", func(t *testing.T) {
		otherToken := common.HexToAddress("0x99999999999")

		mockClient.EXPECT().
			GetPairTokens(gomock.Any(), poolAddr).
			Return(token0, token1, nil)

		req := dto.EstimateRequest{
			Pool:      poolAddr,
			Src:       otherToken,
			Dst:       token1,
			SrcAmount: srcAmount,
		}

		result, err := service.Estimate(context.Background(), req)
		require.Error(t, err)
		require.Equal(t, ErrInvalidArgument, err)
		require.Nil(t, result)
	})

	t.Run("insufficient liquidity", func(t *testing.T) {
		smallAmount := big.NewInt(1)
		bigReserveIn := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(18), nil)
		bigReserveOut := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(18), nil)

		mockClient.EXPECT().
			GetPairTokens(gomock.Any(), poolAddr).
			Return(token0, token1, nil)
		mockClient.EXPECT().
			GetPairReserves(gomock.Any(), poolAddr).
			Return(bigReserveIn, bigReserveOut, nil)

		req := dto.EstimateRequest{
			Pool:      poolAddr,
			Src:       token0,
			Dst:       token1,
			SrcAmount: smallAmount,
		}

		result, err := service.Estimate(context.Background(), req)
		require.Error(t, err)
		require.Equal(t, ErrInsufficientLiquidity, err)
		require.Nil(t, result)
	})

	t.Run("case insensitive address comparison", func(t *testing.T) {
		req := dto.EstimateRequest{
			Pool:      poolAddr,
			Src:       common.HexToAddress("0x" + "abcdef1234567890abcdef1234567890abcdef12"),
			Dst:       common.HexToAddress("0x" + "ABCDEF1234567890ABCDEF1234567890ABCDEF12"),
			SrcAmount: srcAmount,
		}

		result, err := service.Estimate(context.Background(), req)
		require.Error(t, err)
		require.Nil(t, result)
	})
}
