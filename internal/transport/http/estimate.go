package http

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/fleshka4/1inch-test-task/internal/service"
	"github.com/fleshka4/1inch-test-task/internal/service/dto"
	"github.com/fleshka4/1inch-test-task/internal/transport/http/validate"
)

func (s *Server) handleEstimate(w http.ResponseWriter, r *http.Request) {
	req, code, err := validate.EstimateRequestValidate(r)
	if err != nil {
		if code == 0 {
			code = http.StatusBadRequest
		}
		http.Error(w, err.Error(), code)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	out, err := s.est.Estimate(ctx, dto.EstimateRequest{
		Pool:      req.Pool,
		Src:       req.Src,
		Dst:       req.Dst,
		SrcAmount: req.SrcAmount,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInsufficientLiquidity), errors.Is(err, service.ErrInvalidArgument):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, service.ErrPairRead):
			http.Error(w, err.Error(), http.StatusBadGateway)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if _, err := w.Write([]byte(out.String())); err != nil {
		log.Printf("estimate write error: %v", err)
	}
}
