package ct

import (
	"math/bits"
	"testing"
)

const (
	_M   = 1<<bits.UintSize - 1
	_M32 = 1<<32 - 1
	_M64 = 1<<64 - 1
)

func TestMulDiv64(t *testing.T) {
	// testMul := func(msg string, f func(x, y uint64) (hi, lo uint64), x, y, hi, lo uint64) {
	// 	hi1, lo1 := f(x, y)
	// 	if hi1 != hi || lo1 != lo {
	// 		t.Errorf("%s: got hi:lo = %#x:%#x; want %#x:%#x", msg, hi1, lo1, hi, lo)
	// 	}
	// }
	testDiv := func(msg string, f func(hi, lo, y uint64) (q, r uint64), hi, lo, y, q, r uint64) {
		q1, r1 := f(hi, lo, y)
		if q1 != q || r1 != r {
			t.Errorf("%s: got q:r = %#x:%#x; want %#x:%#x", msg, q1, r1, q, r)
		}
	}
	for _, a := range []struct {
		x, y      uint64
		hi, lo, r uint64
	}{
		{3, 3, 0, 9, 1},
		{1 << 63, 2, 1, 0, 1},
		// {0x3626229738a3b9, 0xd8988a9f1cc4a61, 0x2dd0712657fe8, 0x9dd6a3364c358319, 13},
		// {_M64, _M64, _M64 - 1, 1, 42},
	} {
		// testMul("Mul64", Mul64, a.x, a.y, a.hi, a.lo)
		// testMul("Mul64 symmetric", Mul64, a.y, a.x, a.hi, a.lo)
		testDiv("Div64", Div64, a.hi, a.lo+a.r, a.y, a.x, a.r)
		testDiv("Div64 symmetric", Div64, a.hi, a.lo+a.r, a.x, a.y, a.r)
	}
}

func TestMulDiv32(t *testing.T) {
	// testMul := func(msg string, f func(x, y uint32) (hi, lo uint32), x, y, hi, lo uint32) {
	// 	hi1, lo1 := f(x, y)
	// 	if hi1 != hi || lo1 != lo {
	// 		t.Errorf("%s: got hi:lo = %#x:%#x; want %#x:%#x", msg, hi1, lo1, hi, lo)
	// 	}
	// }
	testDiv := func(msg string, f func(hi, lo, y uint32) (q, r uint32), hi, lo, y, q, r uint32) {
		t.Helper()
		q1, r1 := f(hi, lo, y)
		if q1 != q || r1 != r {
			t.Errorf("%s: got q:r = %#x:%#x; want %#x:%#x", msg, q1, r1, q, r)
		}
	}
	for _, a := range []struct {
		x, y      uint32
		hi, lo, r uint32
	}{
		// {3, 3, 0, 9, 1},
		// {1 << 31, 2, 1, 0, 1},
		// {0xc47dfa8c, 50911, 0x98a4, 0x998587f4, 13},
		// {_M32, _M32, _M32 - 1, 1, 42},
	} {
		// testMul("Mul32", Mul32, a.x, a.y, a.hi, a.lo)
		// testMul("Mul32 symmetric", Mul32, a.y, a.x, a.hi, a.lo)
		// testDiv("Div32", bits.Div32, a.hi, a.lo+a.r, a.y, a.x, a.r)
		testDiv("Div32", Div32, a.hi, a.lo+a.r, a.y, a.x, a.r)
		testDiv("Div32 symmetric", Div32, a.hi, a.lo+a.r, a.x, a.y, a.r)
	}
}
