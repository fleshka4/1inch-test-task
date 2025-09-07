package uniswap

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

// Client defines an abstraction for reading Uniswap V2 pair data from the Ethereum blockchain.
type Client interface {
	// GetPairTokens returns the addresses of token0 and token1 for a given pair contract.
	GetPairTokens(ctx context.Context, pair common.Address) (common.Address, common.Address, error)
	// GetPairReserves returns the current reserves of token0 and token1 for a given pair contract.
	GetPairReserves(ctx context.Context, pair common.Address) (*big.Int, *big.Int, error)
}

type ethClientImpl struct {
	rpc     *ethclient.Client
	pairABI abi.ABI
}

// NewClient creates a new Uniswap Client backed by an Ethereum RPC connection.
func NewClient(rpcURL string) (Client, error) {
	r, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, errors.Wrap(err, "ethclient.Dial")
	}

	a, err := abi.JSON(strings.NewReader(pairABIJSON))
	if err != nil {
		return nil, errors.Wrap(err, "abi.JSON")
	}

	return &ethClientImpl{rpc: r, pairABI: a}, nil
}

const pairABIJSON = `[
	{"inputs":[],"name":"token0","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},
	{"inputs":[],"name":"token1","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},
	{"inputs":[],"name":"getReserves","outputs":[{"internalType":"uint112","name":"_reserve0","type":"uint112"},{"internalType":"uint112","name":"_reserve1","type":"uint112"},{"internalType":"uint32","name":"_blockTimestampLast","type":"uint32"}],"stateMutability":"view","type":"function"}
]`

func (c *ethClientImpl) call(ctx context.Context, to common.Address, method string) ([]interface{}, error) {
	data, err := c.pairABI.Pack(method)
	if err != nil {
		return nil, errors.Wrap(err, "c.pairABI.Pack")
	}

	res, err := c.rpc.CallContract(
		ctx,
		ethereum.CallMsg{
			To:   &to,
			Data: data,
		},
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "c.rpc.CallContract")
	}

	out, err := c.pairABI.Unpack(method, res)
	if err != nil {
		return nil, errors.Wrap(err, "c.pairABI.Unpack")
	}

	return out, nil
}

// GetPairTokens returns the addresses of token0 and token1 for a given pair contract.
func (c *ethClientImpl) GetPairTokens(ctx context.Context, pair common.Address) (common.Address, common.Address, error) {
	out0, err := c.call(ctx, pair, "token0")
	if err != nil {
		return common.Address{}, common.Address{}, errors.Wrap(err, "c.call")
	}

	out1, err := c.call(ctx, pair, "token1")
	if err != nil {
		return common.Address{}, common.Address{}, errors.Wrap(err, "c.call")
	}

	a0, ok := out0[0].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, errors.New("token0 cast")
	}

	a1, ok := out1[0].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, errors.New("token1 cast")
	}

	return a0, a1, nil
}

// GetPairReserves returns the current reserves of token0 and token1 for a given pair contract.
func (c *ethClientImpl) GetPairReserves(ctx context.Context, pair common.Address) (*big.Int, *big.Int, error) {
	out, err := c.call(ctx, pair, "getReserves")
	if err != nil {
		return nil, nil, errors.Wrap(err, "c.call")
	}

	r0 := new(big.Int)
	r1 := new(big.Int)
	switch v := out[0].(type) {
	case *big.Int:
		r0.Set(v)
	default:
		return nil, nil, errors.New("reserve0 cast")
	}

	switch v := out[1].(type) {
	case *big.Int:
		r1.Set(v)
	default:
		return nil, nil, errors.New("reserve1 cast")
	}

	return r0, r1, nil
}
