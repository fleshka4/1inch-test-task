package uniswap

import (
	"context"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/fleshka4/1inch-test-task/internal/infra/uniswap/mock"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	t.Run("dial error", func(t *testing.T) {
		// Используем невалидный URL для проверки ошибки
		client, err := NewClient("invalid://url")
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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock.NewMockEthCaller(ctrl)
	client, err := newClientWithCaller(mockCaller)
	require.NoError(t, err)

	addr0 := common.HexToAddress("0x0000000000000000000000000000000000000001")
	addr1 := common.HexToAddress("0x0000000000000000000000000000000000000002")

	t.Run("success", func(t *testing.T) {
		mockCaller.EXPECT().
			CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
			Return(mustPackAddr(t, "token0", addr0), nil)
		mockCaller.EXPECT().
			CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
			Return(mustPackAddr(t, "token1", addr1), nil)

		got0, got1, err := client.GetPairTokens(context.Background(), common.Address{})
		require.NoError(t, err)
		require.Equal(t, addr0, got0)
		require.Equal(t, addr1, got1)
	})

	t.Run("token0 call error", func(t *testing.T) {
		mockCaller.EXPECT().
			CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
			Return(nil, errors.New("call error"))

		_, _, err := client.GetPairTokens(context.Background(), common.Address{})
		require.Error(t, err)
	})

	t.Run("token1 call error", func(t *testing.T) {
		mockCaller.EXPECT().
			CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
			Return(mustPackAddr(t, "token0", addr0), nil)
		mockCaller.EXPECT().
			CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
			Return(nil, errors.New("call error"))

		_, _, err := client.GetPairTokens(context.Background(), common.Address{})
		require.Error(t, err)
	})

	t.Run("token0 cast error", func(t *testing.T) {
		c := client.(*ethClientImpl)
		invalidABI, err := abi.JSON(strings.NewReader(`[
			{"inputs":[],"name":"token0","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}
		]`))
		require.NoError(t, err)
		originalABI := c.pairABI
		c.pairABI = invalidABI
		defer func() { c.pairABI = originalABI }()

		mockCaller.EXPECT().
			CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
			Return(mustPackUint(t, "token0", big.NewInt(123)), nil)

		_, _, err = client.GetPairTokens(context.Background(), common.Address{})
		require.Error(t, err)
	})

	t.Run("token1 cast error", func(t *testing.T) {
		c := client.(*ethClientImpl)
		invalidABI, err := abi.JSON(strings.NewReader(`[
			{"inputs":[],"name":"token0","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},
			{"inputs":[],"name":"token1","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}
		]`))
		require.NoError(t, err)

		originalABI := c.pairABI
		c.pairABI = invalidABI
		defer func() { c.pairABI = originalABI }()

		mockCaller.EXPECT().
			CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
			Return(mustPackAddr(t, "token0", addr0), nil)
		mockCaller.EXPECT().
			CallContract(gomock.Any(), gomock.Any(), gomock.Nil()).
			Return(mustPackUint(t, "token1", big.NewInt(123)), nil)

		_, _, err = client.GetPairTokens(context.Background(), common.Address{})
		require.Error(t, err)
	})
}

func TestGetPairReserves(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock.NewMockEthCaller(ctrl)
	client, err := newClientWithCaller(mockCaller)
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

func mustPackUint(t *testing.T, method string, value *big.Int) []byte {
	a, err := abi.JSON(strings.NewReader(`[
		{"inputs":[],"name":"` + method + `","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}
	]`))
	require.NoError(t, err)

	b, err := a.Methods[method].Outputs.Pack(value)
	require.NoError(t, err)

	return b
}

func mustPackAddr(t *testing.T, method string, addr common.Address) []byte {
	a, err := abi.JSON(strings.NewReader(pairABIJSON))
	require.NoError(t, err)

	b, err := a.Methods[method].Outputs.Pack(addr)
	require.NoError(t, err)

	return b
}

func mustPackReserves(t *testing.T, abiJSON, method string, r0, r1 *big.Int, ts uint32) []byte {
	a, err := abi.JSON(strings.NewReader(abiJSON))
	require.NoError(t, err)

	b, err := a.Methods[method].Outputs.Pack(r0, r1, ts)
	require.NoError(t, err)

	return b
}
