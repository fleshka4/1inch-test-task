package uniswapv2

import "math/big"

// Константы комиссии Uniswap V2: 0.3% = 997/1000.
var (
	feeMul = big.NewInt(997)
	feeDen = big.NewInt(1000)
)

// GetAmountOut считает оффчейн результат свопа по V2:
// amountOut = (amountIn*997 * reserveOut) / (reserveIn*1000 + amountIn*997)
// Возвращает ok=false если ликвидности недостаточно или деление на 0.
func GetAmountOut(amountIn, reserveIn, reserveOut *big.Int) (*big.Int, bool) {
	if amountIn.Sign() <= 0 || reserveIn.Sign() <= 0 || reserveOut.Sign() <= 0 {
		return big.NewInt(0), false
	}

	// Локальные big.Int чтобы снизить аллокации.
	ainFee := new(big.Int).Mul(amountIn, feeMul) // amountIn*997
	num := new(big.Int).Mul(ainFee, reserveOut)  // * reserveOut
	den := new(big.Int).Mul(reserveIn, feeDen)   // reserveIn*1000
	den.Add(den, ainFee)                         // + amountIn*997
	if den.Sign() == 0 {
		return big.NewInt(0), false
	}
	out := new(big.Int).Quo(num, den)
	return out, true
}
