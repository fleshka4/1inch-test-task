package apperrors

import "github.com/pkg/errors"

var (
	// ErrInvalidArgument is returned when the request parameters are invalid.
	ErrInvalidArgument = errors.New("invalid argument")

	// ErrInsufficientLiquidity is returned when the pool does not have enough
	// reserves to satisfy the requested swap.
	ErrInsufficientLiquidity = errors.New("insufficient liquidity")
)
