package httpserver

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/fleshka4/inch-test-task/internal/uniswapv2"
)

type Server struct {
	client *uniswapv2.Client
	mux    *http.ServeMux
}

func New(rpcURL string) (*Server, error) {
	c, err := uniswapv2.NewClient(rpcURL)
	if err != nil {
		return nil, err
	}
	s := &Server{
		client: c,
		mux:    http.NewServeMux(),
	}
	s.mux.HandleFunc("/estimate", s.handleEstimate)
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return s, nil
}

func (s *Server) ListenAndServe(addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.logMiddleware(s.mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(ln)
}

func (s *Server) handleEstimate(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	poolStr := q.Get("pool")
	srcStr := q.Get("src")
	dstStr := q.Get("dst")
	amtStr := q.Get("src_amount")

	// Валидация
	if !isHexAddr(poolStr) || !isHexAddr(srcStr) || !isHexAddr(dstStr) || amtStr == "" {
		http.Error(w, "bad params", http.StatusBadRequest)
		return
	}
	if strings.EqualFold(srcStr, dstStr) {
		http.Error(w, "src == dst", http.StatusBadRequest)
		return
	}
	amountIn, ok := new(big.Int).SetString(amtStr, 10)
	if !ok || amountIn.Sign() <= 0 {
		http.Error(w, "bad src_amount", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	pair := common.HexToAddress(poolStr)
	src := common.HexToAddress(srcStr)
	dst := common.HexToAddress(dstStr)

	// Тянем токены пула и резервы на latest состоянии
	token0, token1, err := s.client.PairTokens(ctx, pair)
	if err != nil {
		http.Error(w, fmt.Sprintf("pair tokens: %v", err), http.StatusBadGateway)
		return
	}
	if !addrEq(src, token0) && !addrEq(src, token1) {
		http.Error(w, "src не принадлежит пулу", http.StatusBadRequest)
		return
	}
	if !addrEq(dst, token0) && !addrEq(dst, token1) {
		http.Error(w, "dst не принадлежит пулу", http.StatusBadRequest)
		return
	}
	if addrEq(src, dst) {
		http.Error(w, "src == dst in pool", http.StatusBadRequest)
		return
	}

	r0, r1, err := s.client.PairReserves(ctx, pair)
	if err != nil {
		http.Error(w, fmt.Sprintf("reserves: %v", err), http.StatusBadGateway)
		return
	}

	var reserveIn, reserveOut *big.Int
	if addrEq(src, token0) && addrEq(dst, token1) {
		reserveIn, reserveOut = r0, r1
	} else if addrEq(src, token1) && addrEq(dst, token0) {
		reserveIn, reserveOut = r1, r0
	} else {
		http.Error(w, "src/dst не соответствуют паре", http.StatusBadRequest)
		return
	}

	// Расчет оффчейн по формуле Uniswap V2
	amountOut, ok := uniswapv2.GetAmountOut(amountIn, reserveIn, reserveOut)
	if !ok {
		http.Error(w, "insufficient liquidity", http.StatusBadRequest)
		return
	}

	// Ответ строго числом в текст/plain как в примере
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(amountOut.String()))
}

func (s *Server) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.String(), time.Since(start))
	})
}

func isHexAddr(s string) bool {
	if !common.IsHexAddress(s) {
		return false
	}
	// быстрый отказ для мусора вроде 0x123 без длины
	b, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	return err == nil && len(b) == 20
}

func addrEq(a, b common.Address) bool {
	return strings.EqualFold(a.Hex(), b.Hex())
}
