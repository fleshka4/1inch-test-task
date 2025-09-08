package uniswapv2

import (
	"math/big"
	"testing"
)

func bi(s string) *big.Int {
	z, _ := new(big.Int).SetString(s, 10)
	return z
}

func TestGetAmountOutInto_Basic(t *testing.T) {
	t.Parallel()

	out := new(big.Int)
	ok := GetAmountOutInto(out, bi("100"), bi("1000"), bi("1000"))
	if !ok {
		t.Fatalf("ok=false")
	}
	if out.Cmp(bi("90")) != 0 { // 90.6... -> 90
		t.Fatalf("want 90 got %s", out.String())
	}
}

func TestGetAmountOutInto_Zeroes(t *testing.T) {
	t.Parallel()

	out := new(big.Int)
	if ok := GetAmountOutInto(out, bi("0"), bi("1"), bi("1")); ok {
		t.Fatal("zero amountIn should be false")
	}
	if ok := GetAmountOutInto(out, bi("1"), bi("0"), bi("1")); ok {
		t.Fatal("zero reserveIn should be false")
	}
	if ok := GetAmountOutInto(out, bi("1"), bi("1"), bi("0")); ok {
		t.Fatal("zero reserveOut should be false")
	}
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

func BenchmarkGetAmountOut_Allocating(b *testing.B) {
	ain := bi("1000000000000000000")
	rIn := bi("1234567890000000000000")
	rOut := bi("987654321000000000000000")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, ok := GetAmountOut(ain, rIn, rOut); !ok {
			b.Fatal("unexpected false")
		}
	}
}

func BenchmarkGetAmountOut_NoAllocs(b *testing.B) {
	ain := bi("1000000000000000000")
	rIn := bi("1234567890000000000000")
	rOut := bi("987654321000000000000000")
	out := new(big.Int) // allocate once and reuse
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if !GetAmountOutInto(out, ain, rIn, rOut) {
			b.Fatal("unexpected false")
		}
	}
}
