package service

import (
	"context"
	"math/big"

	"github.com/fleshka4/1inch-test-task/internal/infra/uniswap"
	"github.com/fleshka4/1inch-test-task/internal/service/dto"
)

// Service represents interface for business logic.
type Service interface {
	Estimate(ctx context.Context, req dto.EstimateRequest) (*big.Int, error)
}

// EstimatorService represents struct for business logic.
type EstimatorService struct {
	uniswapClient uniswap.Client
}

// NewEstimatorService creates EstimatorService.
func NewEstimatorService(cli uniswap.Client) *EstimatorService {
	return &EstimatorService{uniswapClient: cli}
}
