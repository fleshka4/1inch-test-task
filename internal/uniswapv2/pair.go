package uniswapv2

import (
	"context"
	_ "embed"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Минимальный ABI пары для чтения токенов и резервов.
var pairABIJSON = `[
	{"inputs":[],"name":"token0","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},
	{"inputs":[],"name":"token1","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},
	{"inputs":[],"name":"getReserves","outputs":[{"internalType":"uint112","name":"_reserve0","type":"uint112"},{"internalType":"uint112","name":"_reserve1","type":"uint112"},{"internalType":"uint32","name":"_blockTimestampLast","type":"uint32"}],"stateMutability":"view","type":"function"}
]`

type Client struct {
	rpc     *ethclient.Client
	pairABI abi.ABI
}

func NewClient(rpcURL string) (*Client, error) {
	rpc, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}
	a, err := abi.JSON(stringsNewReader(pairABIJSON))
	if err != nil {
		return nil, err
	}
	return &Client{rpc: rpc, pairABI: a}, nil
}

func stringsNewReader(s string) *stringsReader { // маленький хак без аллокаций интерфейса
	return &stringsReader{s: s, i: 0}
}

type stringsReader struct {
	s string
	i int64
}

func (r *stringsReader) Read(p []byte) (int, error) {
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}
	n := copy(p, r.s[r.i:])
	r.i += int64(n)
	return n, nil
}

func (c *Client) PairTokens(ctx context.Context, pair common.Address) (common.Address, common.Address, error) {
	// token0
	var out0 []interface{}
	if err := c.call(ctx, pair, "token0", &out0); err != nil {
		return common.Address{}, common.Address{}, err
	}
	// token1
	var out1 []interface{}
	if err := c.call(ctx, pair, "token1", &out1); err != nil {
		return common.Address{}, common.Address{}, err
	}
	return out0[0].(common.Address), out1[0].(common.Address), nil
}

func (c *Client) PairReserves(ctx context.Context, pair common.Address) (*big.Int, *big.Int, error) {
	var outs []interface{}
	if err := c.call(ctx, pair, "getReserves", &outs); err != nil {
		return nil, nil, err
	}
	// getReserves returns uint112, но go-ethereum мапит в *big.Int
	r0 := new(big.Int).Set(outs[0].(*big.Int))
	r1 := new(big.Int).Set(outs[1].(*big.Int))
	return r0, r1, nil
}

func (c *Client) call(ctx context.Context, to common.Address, method string, out *[]interface{}) error {
	input, err := c.pairABI.Pack(method)
	if err != nil {
		return err
	}

	// Формируем CallMsg вручную
	callMsg := ethereum.CallMsg{
		To:   &to,
		Data: input,
	}

	// Выполняем вызов
	res, err := c.rpc.CallContract(ctx, callMsg, nil)
	if err != nil {
		return err
	}

	decoded, err := c.pairABI.Unpack(method, res)
	if err != nil {
		return err
	}
	*out = decoded
	return nil
}
