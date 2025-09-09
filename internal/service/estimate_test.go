package service

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/fleshka4/1inch-test-task/internal/infra/uniswap/mock"
	"github.com/fleshka4/1inch-test-task/internal/service/dto"
)

func TestEstimate(t *testing.T) {
	t.Parallel()

	poolAddr := common.HexToAddress("0x1234")
	token0 := common.HexToAddress("0x5678")
	token1 := common.HexToAddress("0x12345678")
	srcAmount := big.NewInt(1000)

	tests := []struct {
		name      string
		mockSetup func(*mock.MockClient)
		req       dto.EstimateRequest
		wantErr   assert.ErrorAssertionFunc
	}{
		{
			name: "success token0 to token1",
			mockSetup: func(mc *mock.MockClient) {
				mc.EXPECT().
					GetPairTokens(gomock.Any(), poolAddr).
					Return(token0, token1, nil)
				mc.EXPECT().
					GetPairReserves(gomock.Any(), poolAddr).
					Return(big.NewInt(10000), big.NewInt(20000), nil)
			},
			req: dto.EstimateRequest{
				Pool:      poolAddr,
				Src:       token0,
				Dst:       token1,
				SrcAmount: srcAmount,
			},
			wantErr: assert.NoError,
		},
		{
			name:      "invalid argument - empty pool",
			mockSetup: nil,
			req: dto.EstimateRequest{
				Pool:      common.Address{},
				Src:       token0,
				Dst:       token1,
				SrcAmount: srcAmount,
			},
			wantErr: assert.Error,
		},
		{
			name:      "invalid argument - empty src",
			mockSetup: nil,
			req: dto.EstimateRequest{
				Pool:      poolAddr,
				Src:       common.Address{},
				Dst:       token1,
				SrcAmount: srcAmount,
			},
			wantErr: assert.Error,
		},
		{
			name:      "invalid argument - empty dst",
			mockSetup: nil,
			req: dto.EstimateRequest{
				Pool:      poolAddr,
				Src:       token0,
				Dst:       common.Address{},
				SrcAmount: srcAmount,
			},
			wantErr: assert.Error,
		},
		{
			name:      "invalid argument - same src and dst",
			mockSetup: nil,
			req: dto.EstimateRequest{
				Pool:      poolAddr,
				Src:       token0,
				Dst:       token0,
				SrcAmount: srcAmount,
			},
			wantErr: assert.Error,
		},
		{
			name:      "invalid argument - zero src amount",
			mockSetup: nil,
			req: dto.EstimateRequest{
				Pool:      poolAddr,
				Src:       token0,
				Dst:       token1,
				SrcAmount: big.NewInt(0),
			},
			wantErr: assert.Error,
		},
		{
			name:      "invalid argument - negative src amount",
			mockSetup: nil,
			req: dto.EstimateRequest{
				Pool:      poolAddr,
				Src:       token0,
				Dst:       token1,
				SrcAmount: big.NewInt(-100),
			},
			wantErr: assert.Error,
		},
		{
			name: "pair read error - GetPairTokens fails",
			mockSetup: func(mc *mock.MockClient) {
				mc.EXPECT().
					GetPairTokens(gomock.Any(), poolAddr).
					Return(common.Address{}, common.Address{}, errors.New("RPC error"))
			},
			req: dto.EstimateRequest{
				Pool:      poolAddr,
				Src:       token0,
				Dst:       token1,
				SrcAmount: srcAmount,
			},
			wantErr: assert.Error,
		},
		{
			name: "pair read error - GetPairReserves fails",
			mockSetup: func(mc *mock.MockClient) {
				mc.EXPECT().
					GetPairTokens(gomock.Any(), poolAddr).
					Return(token0, token1, nil)
				mc.EXPECT().
					GetPairReserves(gomock.Any(), poolAddr).
					Return(nil, nil, errors.New("RPC error"))
			},
			req: dto.EstimateRequest{
				Pool:      poolAddr,
				Src:       token0,
				Dst:       token1,
				SrcAmount: srcAmount,
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid argument - src and dst not in pair",
			mockSetup: func(mc *mock.MockClient) {
				mc.EXPECT().
					GetPairTokens(gomock.Any(), poolAddr).
					Return(token0, token1, nil)
			},
			req: dto.EstimateRequest{
				Pool:      poolAddr,
				Src:       common.HexToAddress("0x99999999999"),
				Dst:       token1,
				SrcAmount: srcAmount,
			},
			wantErr: assert.Error,
		},
		{
			name: "insufficient liquidity",
			mockSetup: func(mc *mock.MockClient) {
				bigReserveIn := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(18), nil)
				bigReserveOut := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(18), nil)
				mc.EXPECT().
					GetPairTokens(gomock.Any(), poolAddr).
					Return(token0, token1, nil)
				mc.EXPECT().
					GetPairReserves(gomock.Any(), poolAddr).
					Return(bigReserveIn, bigReserveOut, nil)
			},
			req: dto.EstimateRequest{
				Pool:      poolAddr,
				Src:       token0,
				Dst:       token1,
				SrcAmount: big.NewInt(1),
			},
			wantErr: assert.Error,
		},
		{
			name: "case insensitive address comparison - different addresses",
			req: dto.EstimateRequest{
				Pool:      poolAddr,
				Src:       common.HexToAddress("0x" + "abcdef1234567890abcdef1234567890abcdef12"),
				Dst:       common.HexToAddress("0x" + "ABCDEF1234567890ABCDEF1234567890ABCDEF12"),
				SrcAmount: srcAmount,
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mock.NewMockClient(ctrl)
			service := NewEstimatorService(mockClient)

			if tt.mockSetup != nil {
				tt.mockSetup(mockClient)
			}

			result, err := service.Estimate(context.Background(), tt.req)
			tt.wantErr(t, err)

			if err == nil {
				require.NotNil(t, result)
				require.True(t, result.Cmp(big.NewInt(0)) > 0)
			}
		})
	}
}
