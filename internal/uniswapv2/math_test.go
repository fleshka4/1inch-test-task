package uniswapv2

import (
	"math/big"
	"testing"
)

func bi(s string) *big.Int {
	z, _ := new(big.Int).SetString(s, 10)
	return z
}

func TestGetAmountOut_Basic(t *testing.T) {
	t.Parallel()

	out, ok := GetAmountOut(bi("100"), bi("1000"), bi("1000"))
	if !ok {
		t.Fatalf("ok=false")
	}
	if out.Cmp(bi("90")) != 0 { // 90.6... -> 90
		t.Fatalf("want 90 got %s", out.String())
	}
}

func TestGetAmountOut_Zeroes(t *testing.T) {
	t.Parallel()

	if _, ok := GetAmountOut(bi("0"), bi("1"), bi("1")); ok {
		t.Fatal("zero amountIn should be false")
	}
	if _, ok := GetAmountOut(bi("1"), bi("0"), bi("1")); ok {
		t.Fatal("zero reserveIn should be false")
	}
	if _, ok := GetAmountOut(bi("1"), bi("1"), bi("0")); ok {
		t.Fatal("zero reserveOut should be false")
	}
}

func BenchmarkGetAmountOut(b *testing.B) {
	ain := bi("1000000000000000000")       // 1e18
	rIn := bi("1234567890000000000000")    // 1.234e21
	rOut := bi("987654321000000000000000") // 9.876e23
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, ok := GetAmountOut(ain, rIn, rOut); !ok {
			b.Fatal("unexpected false")
		}
	}
}
