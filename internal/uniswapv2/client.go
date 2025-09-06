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
	"github.com/pkg/errors"
)

// Минимальный ABI пары для чтения токенов и резервов.
var pairABIJSON = `[
	{"inputs":[],"name":"token0","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},
	{"inputs":[],"name":"token1","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},
	{"inputs":[],"name":"getReserves","outputs":[{"internalType":"uint112","name":"_reserve0","type":"uint112"},{"internalType":"uint112","name":"_reserve1","type":"uint112"},{"internalType":"uint32","name":"_blockTimestampLast","type":"uint32"}],"stateMutability":"view","type":"function"}
]`

// Client provides read-only access to Uniswap V2 pair contracts.
type Client struct {
	rpc     *ethclient.Client
	pairABI abi.ABI
}

// NewClient creates a new Uniswap V2 client.
func NewClient(rpcURL string) (*Client, error) {
	rpc, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, errors.Wrap(err, "ethclient.Dial")
	}
	a, err := abi.JSON(stringsNewReader(pairABIJSON))
	if err != nil {
		return nil, errors.Wrap(err, "abi.JSON")
	}
	return &Client{rpc: rpc, pairABI: a}, nil
}

// PairTokens returns token0 and token1 addresses of a pair.
func (c *Client) PairTokens(ctx context.Context, pair common.Address) (common.Address, common.Address, error) {
	// token0
	var out0 []interface{}
	if err := c.call(ctx, pair, "token0", &out0); err != nil {
		return common.Address{}, common.Address{}, errors.Wrap(err, "c.call")
	}
	// token1
	var out1 []interface{}
	if err := c.call(ctx, pair, "token1", &out1); err != nil {
		return common.Address{}, common.Address{}, errors.Wrap(err, "c.call")
	}
	return out0[0].(common.Address), out1[0].(common.Address), nil
}

// PairReserves returns reserve0 and reserve1 of the pair.
func (c *Client) PairReserves(ctx context.Context, pair common.Address) (*big.Int, *big.Int, error) {
	var outs []interface{}
	if err := c.call(ctx, pair, "getReserves", &outs); err != nil {
		return nil, nil, errors.Wrap(err, "c.call")
	}
	// getReserves returns uint112, но go-ethereum мапит в *big.Int
	r0 := new(big.Int).Set(outs[0].(*big.Int))
	r1 := new(big.Int).Set(outs[1].(*big.Int))
	return r0, r1, nil
}

func (c *Client) call(ctx context.Context, to common.Address, method string, out *[]interface{}) error {
	input, err := c.pairABI.Pack(method)
	if err != nil {
		return errors.Wrap(err, "c.pairABI.Pack")
	}

	// Формируем CallMsg вручную
	callMsg := ethereum.CallMsg{
		To:   &to,
		Data: input,
	}

	// Выполняем вызов
	res, err := c.rpc.CallContract(ctx, callMsg, nil)
	if err != nil {
		return errors.Wrap(err, "c.rpc.CallContract")
	}

	decoded, err := c.pairABI.Unpack(method, res)
	if err != nil {
		return errors.Wrap(err, "c.pairABI.Unpack")
	}
	*out = decoded
	return nil
}

func stringsNewReader(s string) *stringsReader {
	return &stringsReader{str: s, num: 0}
}

type stringsReader struct {
	str string
	num int64
}

// Read implements the io.Reader interface for stringsReader.
// It reads data from the underlying string into the provided byte slice.
func (r *stringsReader) Read(p []byte) (int, error) {
	if r.num >= int64(len(r.str)) {
		return 0, io.EOF
	}
	n := copy(p, r.str[r.num:])
	r.num += int64(n)
	return n, nil
}
