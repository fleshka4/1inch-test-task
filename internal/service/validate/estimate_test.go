package validate

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fleshka4/1inch-test-task/internal/service/dto"
)

func TestEstimateRequestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     dto.EstimateRequest
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "valid request",
			req:     createValidRequest(),
			wantErr: assert.NoError,
		},
		{
			name: "zero pool address",
			req: dto.EstimateRequest{
				Pool:      common.Address{},
				Src:       common.HexToAddress("0x123"),
				Dst:       common.HexToAddress("0x456"),
				SrcAmount: big.NewInt(100),
			},
			wantErr: assert.Error,
		},
		{
			name: "zero src address",
			req: dto.EstimateRequest{
				Pool:      common.HexToAddress("0x789"),
				Src:       common.Address{},
				Dst:       common.HexToAddress("0x456"),
				SrcAmount: big.NewInt(100),
			},
			wantErr: assert.Error,
		},
		{
			name: "zero dst address",
			req: dto.EstimateRequest{
				Pool:      common.HexToAddress("0x789"),
				Src:       common.HexToAddress("0x123"),
				Dst:       common.Address{},
				SrcAmount: big.NewInt(100),
			},
			wantErr: assert.Error,
		},
		{
			name: "nil src amount",
			req: dto.EstimateRequest{
				Pool:      common.HexToAddress("0x789"),
				Src:       common.HexToAddress("0x123"),
				Dst:       common.HexToAddress("0x456"),
				SrcAmount: nil,
			},
			wantErr: assert.Error,
		},
		{
			name: "zero src amount",
			req: dto.EstimateRequest{
				Pool:      common.HexToAddress("0x789"),
				Src:       common.HexToAddress("0x123"),
				Dst:       common.HexToAddress("0x456"),
				SrcAmount: big.NewInt(0),
			},
			wantErr: assert.Error,
		},
		{
			name: "negative src amount",
			req: dto.EstimateRequest{
				Pool:      common.HexToAddress("0x789"),
				Src:       common.HexToAddress("0x123"),
				Dst:       common.HexToAddress("0x456"),
				SrcAmount: big.NewInt(-100),
			},
			wantErr: assert.Error,
		},
		{
			name: "same src and dst addresses",
			req: dto.EstimateRequest{
				Pool:      common.HexToAddress("0x789"),
				Src:       common.HexToAddress("0x123"),
				Dst:       common.HexToAddress("0x123"),
				SrcAmount: big.NewInt(100),
			},
			wantErr: assert.Error,
		},
		{
			name: "same addresses different case - should fail",
			req: dto.EstimateRequest{
				Pool:      common.HexToAddress("0x789"),
				Src:       common.HexToAddress("0xabc123"),
				Dst:       common.HexToAddress("0xABC123"),
				SrcAmount: big.NewInt(100),
			},
			wantErr: assert.Error,
		},
		{
			name: "different addresses - should pass",
			req: dto.EstimateRequest{
				Pool:      common.HexToAddress("0x789"),
				Src:       common.HexToAddress("0x123456"),
				Dst:       common.HexToAddress("0x789ABC"),
				SrcAmount: big.NewInt(100),
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := EstimateRequestValidate(tt.req)
			tt.wantErr(t, err)
		})
	}
}

func TestEstimateRequestValidate_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("very large amount", func(t *testing.T) {
		t.Parallel()

		req := createValidRequest()
		req.SrcAmount = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)

		err := EstimateRequestValidate(req)
		require.NoError(t, err)
	})

	t.Run("minimal positive amount", func(t *testing.T) {
		t.Parallel()

		req := createValidRequest()
		req.SrcAmount = big.NewInt(1)

		err := EstimateRequestValidate(req)
		require.NoError(t, err)
	})
}

// Вспомогательная функция для создания валидного запроса
func createValidRequest() dto.EstimateRequest {
	return dto.EstimateRequest{
		Pool:      common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc454e4438f44e"),
		Src:       common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"),
		Dst:       common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
		SrcAmount: big.NewInt(1000000000000000000), // 1 ETH
	}
}
