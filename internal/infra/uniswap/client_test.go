package uniswap

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/fleshka4/1inch-test-task/internal/infra/uniswap/mock"
)

const timeout = 2 * time.Second

func TestNewClient(t *testing.T) {
	t.Parallel()

	t.Run("dial error", func(t *testing.T) {
		client, err := NewClient("invalid://url", timeout)
		require.Error(t, err)
		require.Nil(t, client)
	})
}

func TestCallMethod(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock.NewMockEthCaller(ctrl)
	client := &ethClientImpl{caller: mockCaller}

	pairABI, err := abi.JSON(strings.NewReader(pairABIJSON))
	require.NoError(t, err)
	client.pairABI = pairABI

	t.Run("pack error", func(t *testing.T) {
		t.Parallel()

		invalidClient := &ethClientImpl{caller: mockCaller}
		invalidABI, err := abi.JSON(strings.NewReader(`[]`))
		require.NoError(t, err)

		invalidClient.pairABI = invalidABI
		_, err = invalidClient.call(context.Background(), common.Address{}, "nonexistent")
		require.Error(t, err)
	})

	t.Run("call contract error", func(t *testing.T) {
		t.Parallel()

		mockCaller.EXPECT().
			CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
			Return(nil, errors.New("call error"))

		_, err := client.call(context.Background(), common.Address{}, "token0")
		require.Error(t, err)
	})

	t.Run("unpack error", func(t *testing.T) {
		t.Parallel()

		mockCaller.EXPECT().
			CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
			Return([]byte("invalid data"), nil)

		_, err := client.call(context.Background(), common.Address{}, "token0")
		require.Error(t, err)
	})
}

func TestGetPairTokens(t *testing.T) {
	t.Parallel()

	addr0 := common.HexToAddress("0x0000000000000000000000000000000000000001")
	addr1 := common.HexToAddress("0x0000000000000000000000000000000000000002")

	tests := []struct {
		name      string
		mockSetup func(*mock.MockEthCaller, *ethClientImpl)
		wantErr   assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			mockSetup: func(mc *mock.MockEthCaller, client *ethClientImpl) {
				mc.EXPECT().
					CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
					Return(mustPackAddr(t, "token0", addr0), nil)
				mc.EXPECT().
					CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
					Return(mustPackAddr(t, "token1", addr1), nil)
			},
			wantErr: assert.NoError,
		},
		{
			name: "token0 call error",
			mockSetup: func(mc *mock.MockEthCaller, client *ethClientImpl) {
				mc.EXPECT().
					CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
					Return(nil, errors.New("call error"))
				mc.EXPECT().
					CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
					Return(mustPackAddr(t, "token1", addr0), nil)
			},
			wantErr: assert.Error,
		},
		{
			name: "token1 call error",
			mockSetup: func(mc *mock.MockEthCaller, client *ethClientImpl) {
				mc.EXPECT().
					CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
					Return(mustPackAddr(t, "token0", addr0), nil)
				mc.EXPECT().
					CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
					Return(nil, errors.New("call error"))
			},
			wantErr: assert.Error,
		},
		{
			name: "token0 cast error",
			mockSetup: func(mc *mock.MockEthCaller, client *ethClientImpl) {
				invalidABI, err := abi.JSON(strings.NewReader(`[
					{"inputs":[],"name":"token0","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}
				]`))
				require.NoError(t, err)
				client.pairABI = invalidABI

				mc.EXPECT().
					CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
					Return(mustPackUint(t, "token0", big.NewInt(123)), nil)
			},
			wantErr: assert.Error,
		},
		{
			name: "token1 cast error",
			mockSetup: func(mc *mock.MockEthCaller, client *ethClientImpl) {
				invalidABI, err := abi.JSON(strings.NewReader(`[
					{"inputs":[],"name":"token0","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},
					{"inputs":[],"name":"token1","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}
				]`))
				require.NoError(t, err)
				client.pairABI = invalidABI

				mc.EXPECT().
					CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
					Return(mustPackAddr(t, "token0", addr0), nil)
				mc.EXPECT().
					CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
					Return(mustPackUint(t, "token1", big.NewInt(123)), nil)
			},
			wantErr: assert.Error,
		},
		{
			name:      "context timeout",
			mockSetup: func(mc *mock.MockEthCaller, client *ethClientImpl) {},
			wantErr:   assert.Error,
		},
		{
			name:      "context cancellation",
			mockSetup: func(mc *mock.MockEthCaller, client *ethClientImpl) {},
			wantErr:   assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCaller := mock.NewMockEthCaller(ctrl)
			client, err := newClientWithCaller(mockCaller, timeout)
			require.NoError(t, err)

			ethClient := client.(*ethClientImpl)

			originalABI := ethClient.pairABI
			defer func() { ethClient.pairABI = originalABI }()

			if tt.mockSetup != nil {
				tt.mockSetup(mockCaller, ethClient)
			}

			var (
				ctx    context.Context
				cancel context.CancelFunc
			)

			switch tt.name {
			case "context timeout":
				ctx, cancel = context.WithTimeout(context.Background(), 1*time.Millisecond)
				defer cancel()
				time.Sleep(2 * time.Millisecond)
			case "context cancellation":
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			default:
				ctx = context.Background()
			}

			got0, got1, err := client.GetPairTokens(ctx, common.Address{})
			tt.wantErr(t, err)

			if err == nil {
				addresses := []common.Address{got0, got1}
				require.Contains(t, addresses, addr0)
				require.Contains(t, addresses, addr1)
				require.NotEqual(t, got0, got1)
			}
		})
	}
}

func TestGetPairReserves(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock.NewMockEthCaller(ctrl)
	client, err := newClientWithCaller(mockCaller, timeout)
	require.NoError(t, err)

	r0 := big.NewInt(123)
	r1 := big.NewInt(456)
	timestamp := uint32(1234567890)

	t.Run("success", func(t *testing.T) {
		mockCaller.EXPECT().
			CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
			Return(mustPackReserves(t, pairABIJSON, "getReserves", r0, r1, timestamp), nil)

		got0, got1, err := client.GetPairReserves(context.Background(), common.Address{})
		require.NoError(t, err)
		require.Equal(t, r0, got0)
		require.Equal(t, r1, got1)
	})

	t.Run("call error", func(t *testing.T) {
		mockCaller.EXPECT().
			CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
			Return(nil, errors.New("call error"))

		_, _, err := client.GetPairReserves(context.Background(), common.Address{})
		require.Error(t, err)
	})
}

func mustPackAddr(t *testing.T, method string, addr common.Address) []byte {
	t.Helper()

	abiFromJSON, err := abi.JSON(strings.NewReader(`[
		{"inputs":[],"name":"` + method + `","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"}
	]`))
	require.NoError(t, err)

	_, err = abiFromJSON.Pack(method)
	require.NoError(t, err)

	result, err := abiFromJSON.Methods[method].Outputs.Pack(addr)
	require.NoError(t, err)

	return result
}

func mustPackUint(t *testing.T, method string, value *big.Int) []byte {
	t.Helper()

	abiFromJSON, err := abi.JSON(strings.NewReader(`[
		{"inputs":[],"name":"` + method + `","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}
	]`))
	require.NoError(t, err)

	result, err := abiFromJSON.Methods[method].Outputs.Pack(value)
	require.NoError(t, err)

	return result
}

func mustPackReserves(t *testing.T, abiJSON, method string, r0, r1 *big.Int, ts uint32) []byte {
	t.Helper()

	a, err := abi.JSON(strings.NewReader(abiJSON))
	require.NoError(t, err)

	b, err := a.Methods[method].Outputs.Pack(r0, r1, ts)
	require.NoError(t, err)

	return b
}
