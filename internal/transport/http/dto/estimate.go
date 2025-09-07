package dto

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// EstimateRequest represents a parsed HTTP request for the /estimate endpoint.
type EstimateRequest struct {
	Pool      common.Address
	Src       common.Address
	Dst       common.Address
	SrcAmount *big.Int
}
