package dto

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// EstimateRequest represents a request to calculate an off-chain Uniswap V2 swap.
type EstimateRequest struct {
	Pool      common.Address
	Src       common.Address
	Dst       common.Address
	SrcAmount *big.Int
}
