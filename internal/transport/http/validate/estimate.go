package validate

import (
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/fleshka4/1inch-test-task/internal/transport/http/dto"
)

// EstimateRequestValidate validates /estimate request and returns dto.
func EstimateRequestValidate(r *http.Request) (*dto.EstimateRequest, int, error) {
	q := r.URL.Query()
	p := q.Get("pool")
	src := q.Get("src")
	dst := q.Get("dst")
	amt := q.Get("src_amount")
	if p == "" || src == "" || dst == "" || amt == "" {
		return nil, http.StatusBadRequest, errors.New("missing params")
	}
	if !common.IsHexAddress(p) || !common.IsHexAddress(src) || !common.IsHexAddress(dst) {
		return nil, http.StatusBadRequest, errors.New("bad address format")
	}
	a, ok := new(big.Int).SetString(amt, 10)
	if !ok || a.Sign() <= 0 {
		return nil, http.StatusBadRequest, errors.New("bad src_amount")
	}
	return &dto.EstimateRequest{
		Pool:      common.HexToAddress(p),
		Src:       common.HexToAddress(src),
		Dst:       common.HexToAddress(dst),
		SrcAmount: a,
	}, 0, nil
}
