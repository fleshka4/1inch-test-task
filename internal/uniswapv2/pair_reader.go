package uniswapv2

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// PairReader defines the interface required to read Uniswap-like pair data.
type PairReader interface {
	// PairTokens returns token0 and token1 addresses of a pair.
	PairTokens(ctx context.Context, pair common.Address) (common.Address, common.Address, error)

	// PairReserves returns reserve0 and reserve1 of the pair.
	PairReserves(ctx context.Context, pair common.Address) (*big.Int, *big.Int, error)
}
