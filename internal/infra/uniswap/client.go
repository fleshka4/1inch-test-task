package uniswap

import (
	"context"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

const pairABIJSON = `[
	{"inputs":[],"name":"token0","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},
	{"inputs":[],"name":"token1","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},
	{"inputs":[],"name":"getReserves","outputs":[{"internalType":"uint112","name":"_reserve0","type":"uint112"},{"internalType":"uint112","name":"_reserve1","type":"uint112"},{"internalType":"uint32","name":"_blockTimestampLast","type":"uint32"}],"stateMutability":"view","type":"function"}
]`

// Client defines an abstraction for reading Uniswap V2 pair data from the Ethereum blockchain.
type Client interface {
	// GetPairTokens returns the addresses of token0 and token1 for a given pair contract.
	GetPairTokens(ctx context.Context, pair common.Address) (common.Address, common.Address, error)
	// GetPairReserves returns the current reserves of token0 and token1 for a given pair contract.
	GetPairReserves(ctx context.Context, pair common.Address) (*big.Int, *big.Int, error)
}

// EthCaller represents interface for calling contracts.
type EthCaller interface {
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
}

type ethClientImpl struct {
	caller  EthCaller
	pairABI abi.ABI

	callTimeout time.Duration
}

// NewClient creates a new Uniswap Client backed by an Ethereum RPC connection.
func NewClient(rpcURL string, callTimeout time.Duration) (Client, error) {
	caller, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, errors.Wrap(err, "ethclient.Dial")
	}

	return newClientWithCaller(caller, callTimeout)
}

func newClientWithCaller(caller EthCaller, callTimeout time.Duration) (Client, error) {
	pairABI, err := abi.JSON(strings.NewReader(pairABIJSON))
	if err != nil {
		return nil, errors.Wrap(err, "abi.JSON")
	}

	return &ethClientImpl{
		caller:  caller,
		pairABI: pairABI,

		callTimeout: callTimeout,
	}, nil
}

func (c *ethClientImpl) call(ctx context.Context, to common.Address, method string) ([]interface{}, error) {
	data, err := c.pairABI.Pack(method)
	if err != nil {
		return nil, errors.Wrap(err, "c.pairABI.Pack")
	}

	res, err := c.caller.CallContract(
		ctx,
		ethereum.CallMsg{
			To:   &to,
			Data: data,
		},
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "c.caller.CallContract")
	}

	out, err := c.pairABI.Unpack(method, res)
	if err != nil {
		return nil, errors.Wrap(err, "c.pairABI.Unpack")
	}

	return out, nil
}

// GetPairTokens returns the addresses of token0 and token1 for a given pair contract.
func (c *ethClientImpl) GetPairTokens(ctx context.Context, pair common.Address) (common.Address, common.Address, error) {
	const (
		numTokens    = 2
		token0Method = "token0"
		token1Method = "token1"
	)

	type tokenResult struct {
		token common.Address
		err   error
		name  string
	}

	var wg sync.WaitGroup
	ch := make(chan tokenResult, numTokens)

	getToken := func(method string) {
		defer wg.Done()

		ctxCall, cancel := context.WithTimeout(ctx, c.callTimeout)
		defer cancel()

		select {
		case <-ctxCall.Done():
			ch <- tokenResult{err: errors.Wrap(ctxCall.Err(), "context cancelled before call")}
			return
		default:
		}

		out, err := c.call(ctxCall, pair, method)
		if err != nil {
			ch <- tokenResult{err: errors.Wrapf(err, "failed to call %s", method)}
			return
		}

		addr, ok := out[0].(common.Address)
		if !ok {
			ch <- tokenResult{err: errors.Errorf("failed to cast %s result to address", method)}
			return
		}

		ch <- tokenResult{token: addr, name: method}
	}

	wg.Add(numTokens)
	go getToken(token0Method)
	go getToken(token1Method)

	go func() {
		wg.Wait()
		close(ch)
	}()

	var (
		token0, token1 common.Address
		combinedErr    error
	)

	for result := range ch {
		if result.err != nil {
			combinedErr = multierr.Append(combinedErr, result.err)
			continue
		}

		switch result.name {
		case token0Method:
			token0 = result.token
		case token1Method:
			token1 = result.token
		}
	}

	if combinedErr != nil {
		return common.Address{}, common.Address{}, errors.Wrap(combinedErr, "failed to get pair tokens")
	}

	return token0, token1, nil
}

// GetPairReserves returns the current reserves of token0 and token1 for a given pair contract.
func (c *ethClientImpl) GetPairReserves(ctx context.Context, pair common.Address) (*big.Int, *big.Int, error) {
	out, err := c.call(ctx, pair, "getReserves")
	if err != nil {
		return nil, nil, errors.Wrap(err, "c.call")
	}

	const requiredSize = 2
	if len(out) < requiredSize {
		return nil, nil, errors.Errorf("insufficient outputs from getReserves call: expected %d, got %d", requiredSize, len(out))
	}

	reserves := make([]*big.Int, requiredSize)
	reserveNames := []string{"reserve0", "reserve1"}

	for i := 0; i < requiredSize; i++ {
		reserve, ok := out[i].(*big.Int)
		if !ok {
			return nil, nil, errors.Errorf("failed to cast %s to *big.Int", reserveNames[i])
		}
		reserves[i] = reserve
	}

	return reserves[0], reserves[1], nil
}
