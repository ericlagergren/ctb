package ct

import "fmt"

// LessOrEq returns 1 if x <= y and 0 otherwise.
//
// See golang.org/issues/42685.
func LessOrEq(x, y uint) uint {
	x64 := uint64(x)
	y64 := uint64(y)

	gtb := x64 & ^y64
	ltb := ^x64 & y64

	ltb |= (ltb >> 1)
	ltb |= (ltb >> 2)
	ltb |= (ltb >> 4)
	ltb |= (ltb >> 8)
	ltb |= (ltb >> 16)
	ltb |= (ltb >> 32)

	bit := gtb &^ ltb

	bit |= (bit >> 1)
	bit |= (bit >> 2)
	bit |= (bit >> 4)
	bit |= (bit >> 8)
	bit |= (bit >> 16)
	bit |= (bit >> 32)

	return uint(^bit & 1)
}

// Greater returns 1 if x > y and 0 otherwise.
//
// See golang.org/issues/42685.
func Greater(x, y uint64) uint64 {
	gtb := x & ^y
	ltb := ^x & y

	ltb |= (ltb >> 1)
	ltb |= (ltb >> 2)
	ltb |= (ltb >> 4)
	ltb |= (ltb >> 8)
	ltb |= (ltb >> 16)
	ltb |= (ltb >> 32)

	bit := gtb &^ ltb

	bit |= (bit >> 1)
	bit |= (bit >> 2)
	bit |= (bit >> 4)
	bit |= (bit >> 8)
	bit |= (bit >> 16)
	bit |= (bit >> 32)

	return bit & 1
}

// greater32 returns 1 if x > y and 0 otherwise.
//
// See golang.org/issues/42685.
func greater32(x, y uint32) uint32 {
	gtb := x & ^y
	ltb := ^x & y

	ltb |= (ltb >> 1)
	ltb |= (ltb >> 2)
	ltb |= (ltb >> 4)
	ltb |= (ltb >> 8)
	ltb |= (ltb >> 16)

	bit := gtb &^ ltb

	bit |= (bit >> 1)
	bit |= (bit >> 2)
	bit |= (bit >> 4)
	bit |= (bit >> 8)
	bit |= (bit >> 16)

	return bit & 1
}

func GreaterEq32(x, y uint32) uint32 {
	return greater32(y, x) ^ 1
}

func GreaterEq64(x, y uint64) uint64 {
	return Greater(y, x) ^ 1
}

// Equal64 returns 1 if x == y and 0 otherwise.
func Equal32(x, y uint32) uint32 {
	return uint32((uint64(x^y) - 1) >> 63)
}

// Equal64 returns 1 if x == y and 0 otherwise.
func Equal64(x, y uint64) uint64 {
	return (uint64(x^y) - 1) >> 63
}

// Select returns x if v == 1 and y if v == 0.
//
// The result is undefined if v is anything
// other than 1 or 0.
func Select32(v, x, y uint32) uint32 {
	return ^(v-1)&x | (v-1)&y
}

// Select returns x if v == 1 and y if v == 0.
//
// The result is undefined if v is anything
// other than 1 or 0.
func Select64(v, x, y uint64) uint64 {
	return ^(v-1)&x | (v-1)&y
}

// Div64 returns q = (hi, lo) / d and r = (hi, lo) % d.
func Div32(hi, lo, d uint32) (q, r uint32) {
	ch := Equal32(hi, d)
	hi = Select32(ch, 0, hi)
	for k := 31; k > 0; k-- {
		j := 32 - k
		w := (hi << j) | (lo >> k)
		ctl := GreaterEq32(w, d) | (hi >> k)
		hi2 := (w - d) >> j
		lo2 := lo - (d << k)
		hi = Select32(ctl, hi2, hi)
		lo = Select32(ctl, lo2, lo)
		q |= ctl << k
	}
	cf := GreaterEq32(lo, d) | hi
	q |= cf
	r = Select32(cf, lo-d, lo)
	return q, r
}

// Div64 returns q = (hi, lo) / d and r = (hi, lo) % d.
func Div64(hi, lo, d uint64) (q, r uint64) {
	ch := Equal64(hi, d)
	hi = Select64(ch, 0, hi)
	for k := 63; k > 0; k-- {
		j := 64 - k
		w := (hi << j) | (lo >> k)
		ctl := GreaterEq64(w, d) | (hi >> k)
		hi2 := (w - d) >> j
		lo2 := lo - (d << k)
		fmt.Printf("lo=%d lo2=%d d=%d k=%d\n", lo, lo2, d, k)
		hi = Select64(ctl, hi2, hi)
		lo = Select64(ctl, lo2, lo)
		q |= ctl << k
	}
	cf := GreaterEq64(lo, d) | hi
	q |= cf
	r = Select64(cf, lo-d, lo)
	return q, r
}
