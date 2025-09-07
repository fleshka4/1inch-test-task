package service

import (
	"context"
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/fleshka4/1inch-test-task/internal/service/dto"
	"github.com/fleshka4/1inch-test-task/internal/uniswapv2"
)

var (
	// ErrInvalidArgument is returned when the request parameters are invalid.
	ErrInvalidArgument = errors.New("invalid argument")

	// ErrInsufficientLiquidity is returned when the pool does not have enough
	// reserves to satisfy the requested swap.
	ErrInsufficientLiquidity = errors.New("insufficient liquidity")

	// ErrPairRead is returned when fetching pair data (tokens or reserves) fails,
	// typically due to an RPC or ABI decoding error.
	ErrPairRead = errors.New("pair read failed")
)

// Estimate performs the complete business logic for off-chain swap calculation.
//
// It validates the request parameters, reads the Uniswap V2 pair contract state
// (tokens and reserves) through the infra client, and calculates the output
// amount using the Uniswap V2 constant product formula with fee adjustment.
func (s *EstimatorService) Estimate(ctx context.Context, req dto.EstimateRequest) (*big.Int, error) {
	if req.Pool == (common.Address{}) || req.Src == (common.Address{}) || req.Dst == (common.Address{}) {
		return nil, ErrInvalidArgument
	}
	if strings.EqualFold(req.Src.Hex(), req.Dst.Hex()) {
		return nil, ErrInvalidArgument
	}
	if req.SrcAmount == nil || req.SrcAmount.Sign() <= 0 {
		return nil, ErrInvalidArgument
	}

	token0, token1, err := s.cli.GetPairTokens(ctx, req.Pool)
	if err != nil {
		return nil, ErrPairRead
	}

	// Check that src/dst belong to the pair.
	var reserveIn, reserveOut *big.Int
	if strings.EqualFold(req.Src.Hex(), token0.Hex()) && strings.EqualFold(req.Dst.Hex(), token1.Hex()) {
		r0, r1, err := s.cli.GetPairReserves(ctx, req.Pool)
		if err != nil {
			return nil, ErrPairRead
		}
		reserveIn, reserveOut = r0, r1
	} else if strings.EqualFold(req.Src.Hex(), token1.Hex()) && strings.EqualFold(req.Dst.Hex(), token0.Hex()) {
		r0, r1, err := s.cli.GetPairReserves(ctx, req.Pool)
		if err != nil {
			return nil, ErrPairRead
		}
		reserveIn, reserveOut = r1, r0
	} else {
		return nil, ErrInvalidArgument
	}

	amountOut, ok := uniswapv2.GetAmountOut(req.SrcAmount, reserveIn, reserveOut)
	if !ok {
		return nil, ErrInsufficientLiquidity
	}
	return amountOut, nil
}
