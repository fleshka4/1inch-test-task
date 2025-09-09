package validate

import (
	"github.com/fleshka4/1inch-test-task/internal/apperrors"
	"github.com/fleshka4/1inch-test-task/internal/service/dto"
	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"
)

// EstimateRequestValidate validates business logic request and returns dto.
func EstimateRequestValidate(req dto.EstimateRequest) error {
	var zeroAddress = common.Address{}

	if req.Pool == zeroAddress || req.Src == zeroAddress || req.Dst == zeroAddress {
		return errors.Wrap(apperrors.ErrInvalidArgument, "address cannot be empty")
	}

	if req.Src == req.Dst {
		return errors.Wrap(apperrors.ErrInvalidArgument, "destination address cannot be the same as source address")
	}

	if req.SrcAmount == nil || req.SrcAmount.Sign() <= 0 {
		return errors.Wrap(apperrors.ErrInvalidArgument, "source amount cannot be zero or negative")
	}

	return nil
}
