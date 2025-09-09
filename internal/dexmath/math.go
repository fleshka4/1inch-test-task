package dexmath

import (
	"math/big"
	"sync"
)

var (
	// Fee constants.
	feeMul = big.NewInt(997)
	feeDen = big.NewInt(1000)

	defaultMath = newMathService()
)

type mathTmp struct {
	a *big.Int
	b *big.Int
	c *big.Int
}

type mathService struct {
	pool *sync.Pool
}

func newMathService() *mathService {
	return &mathService{
		pool: &sync.Pool{
			New: func() any {
				return &mathTmp{
					a: new(big.Int),
					b: new(big.Int),
					c: new(big.Int),
				}
			},
		},
	}
}

func (m *mathService) getAmountOutInto(out, amountIn, reserveIn, reserveOut *big.Int) bool {
	if out == nil {
		return false
	}
	// basic validation.
	if amountIn.Sign() <= 0 || reserveIn.Sign() <= 0 || reserveOut.Sign() <= 0 {
		out.SetInt64(0)
		return false
	}

	t := m.pool.Get().(*mathTmp)

	// ainFee := amountIn * 997.
	t.a.Mul(amountIn, feeMul)

	// num := ainFee * reserveOut.
	t.b.Mul(t.a, reserveOut)

	// den := reserveIn * 1000 + ainFee.
	t.c.Mul(reserveIn, feeDen)
	t.c.Add(t.c, t.a)

	if t.c.Sign() == 0 {
		out.SetInt64(0)
		m.pool.Put(t)
		return false
	}

	// out = num / den.
	out.Quo(t.b, t.c)

	// return temps to pool.
	m.pool.Put(t)
	return true
}

// GetAmountOutInto computes the amount of output tokens received for a given input amount,
// using Uniswap V2 formula with 0.3% fee (997/1000).
//
// Returns (output, true) if calculation is successful, or (0, false) if any value is zero.
// It writes the result into out and returns ok.
// out must be non-nil; this function does not allocate for temporaries
// if the pool is warm. Caller should reuse `out` when possible.
func GetAmountOutInto(out, amountIn, reserveIn, reserveOut *big.Int) bool {
	return defaultMath.getAmountOutInto(out, amountIn, reserveIn, reserveOut)
}

// GetAmountOut computes the amount of output tokens received for a given input amount,
// using Uniswap V2 formula with 0.3% fee (997/1000).
//
// Returns (output, true) if calculation is successful, or (0, false) if any value is zero.
// It represents backwards-compatible allocator: returns a newly allocated *big.Int (uses pool for temps).
func GetAmountOut(amountIn, reserveIn, reserveOut *big.Int) (*big.Int, bool) {
	out := new(big.Int)
	ok := defaultMath.getAmountOutInto(out, amountIn, reserveIn, reserveOut)
	return out, ok
}
