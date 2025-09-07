package uniswapv2

import "math/big"

// Fee constants.
var (
	feeMul = big.NewInt(997)
	feeDen = big.NewInt(1000)
)

// GetAmountOut computes the amount of output tokens received for a given input amount,
// using Uniswap V2 formula with 0.3% fee (997/1000).
//
// Returns (output, true) if calculation is successful, or (0, false) if any value is zero.
func GetAmountOut(amountIn, reserveIn, reserveOut *big.Int) (*big.Int, bool) {
	if amountIn.Sign() <= 0 || reserveIn.Sign() <= 0 || reserveOut.Sign() <= 0 {
		return big.NewInt(0), false
	}

	ainFee := new(big.Int).Mul(amountIn, feeMul) // amountIn*997
	num := new(big.Int).Mul(ainFee, reserveOut)  // * reserveOut
	den := new(big.Int).Mul(reserveIn, feeDen)   // reserveIn*1000
	den.Add(den, ainFee)                         // + amountIn*997
	if den.Sign() == 0 {
		return big.NewInt(0), false
	}
	out := new(big.Int).Quo(num, den)

	if out.Sign() <= 0 {
		return big.NewInt(0), false
	}

	return out, true
}
