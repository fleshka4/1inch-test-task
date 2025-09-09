package service

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/fleshka4/1inch-test-task/internal/apperrors"
	"github.com/fleshka4/1inch-test-task/internal/dexmath"
	"github.com/fleshka4/1inch-test-task/internal/service/dto"
	"github.com/fleshka4/1inch-test-task/internal/service/validate"
)

// Estimate performs the complete business logic for off-chain swap calculation.
//
// It validates the request parameters, reads the Uniswap V2 pair contract state
// (tokens and reserves) through the infra client, and calculates the output
// amount using the Uniswap V2 constant product formula with fee adjustment.
func (s *EstimatorService) Estimate(ctx context.Context, req dto.EstimateRequest) (*big.Int, error) {
	if err := validate.EstimateRequestValidate(req); err != nil {
		return nil, errors.Wrap(err, "validate.EstimateRequestValidate")
	}

	token0, token1, err := s.uniswapClient.GetPairTokens(ctx, req.Pool)
	if err != nil {
		return nil, errors.Wrap(err, "s.uniswapClient.GetPairTokens")
	}

	var reserves struct {
		in  *big.Int
		out *big.Int
	}

	switch {
	case isTokenMatch(req.Src, token0) && isTokenMatch(req.Dst, token1):
		r0, r1, err := s.uniswapClient.GetPairReserves(ctx, req.Pool)
		if err != nil {
			return nil, errors.Wrap(err, "s.uniswapClient.GetPairReserves")
		}
		reserves.in, reserves.out = r0, r1

	case isTokenMatch(req.Src, token1) && isTokenMatch(req.Dst, token0):
		r0, r1, err := s.uniswapClient.GetPairReserves(ctx, req.Pool)
		if err != nil {
			return nil, errors.Wrap(err, "s.uniswapClient.GetPairReserves")
		}
		reserves.in, reserves.out = r1, r0

	default:
		return nil, errors.Wrapf(
			apperrors.ErrInvalidArgument,
			"src/dst does not match pool tokens: pool has %s and %s",
			token0.Hex(), token1.Hex(),
		)
	}

	out := new(big.Int)
	if !dexmath.GetAmountOutInto(out, req.SrcAmount, reserves.in, reserves.out) || out.Sign() == 0 {
		return nil, errors.Wrap(apperrors.ErrInsufficientLiquidity, "bad estimate")
	}

	return out, nil
}

func isTokenMatch(addr1, addr2 common.Address) bool {
	return strings.EqualFold(addr1.Hex(), addr2.Hex())
}
