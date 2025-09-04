package handlers

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/fleshka4/inch-test-task/internal/uniswapv2"
)

// NewEstimateHandler returns an HTTP handler that estimates the amountOut
// from a Uniswap V2-style liquidity pool, using off-chain math.
func NewEstimateHandler(client uniswapv2.PairReader, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		poolStr := q.Get("pool")
		srcStr := q.Get("src")
		dstStr := q.Get("dst")
		amtStr := q.Get("src_amount")

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

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		pair := common.HexToAddress(poolStr)
		src := common.HexToAddress(srcStr)
		dst := common.HexToAddress(dstStr)

		token0, token1, err := client.PairTokens(ctx, pair)
		if err != nil {
			http.Error(w, fmt.Sprintf("pair tokens: %v", err), http.StatusBadGateway)
			return
		}
		if !addrEq(src, token0) && !addrEq(src, token1) {
			http.Error(w, "src is not from pool", http.StatusBadRequest)
			return
		}
		if !addrEq(dst, token0) && !addrEq(dst, token1) {
			http.Error(w, "dst is not from pool", http.StatusBadRequest)
			return
		}
		if addrEq(src, dst) {
			http.Error(w, "src == dst in pool", http.StatusBadRequest)
			return
		}

		r0, r1, err := client.PairReserves(ctx, pair)
		if err != nil {
			http.Error(w, fmt.Sprintf("client.PairReserves: %v", err), http.StatusBadGateway)
			return
		}

		var reserveIn, reserveOut *big.Int
		if addrEq(src, token0) && addrEq(dst, token1) {
			reserveIn, reserveOut = r0, r1
		} else if addrEq(src, token1) && addrEq(dst, token0) {
			reserveIn, reserveOut = r1, r0
		} else {
			http.Error(w, "src/dst does not match to pair", http.StatusBadRequest)
			return
		}

		amountOut, ok := uniswapv2.GetAmountOut(amountIn, reserveIn, reserveOut)
		if !ok {
			http.Error(w, "insufficient liquidity", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, err := w.Write([]byte(amountOut.String())); err != nil {
			log.Printf("error writing response: %v", err)
		}
	}
}

func isHexAddr(s string) bool {
	if !common.IsHexAddress(s) {
		return false
	}
	b, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	return err == nil && len(b) == 20
}

func addrEq(a, b common.Address) bool {
	return strings.EqualFold(a.Hex(), b.Hex())
}
