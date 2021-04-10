// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xbits

// Many of the loops in this file are of the form
//   for i := 0; i < len(z) && i < len(x) && i < len(y); i++
// i < len(z) is the real condition.
// However, checking i < len(x) && i < len(y) as well is faster than
// having the compiler do a bounds check in the body of the loop;
// remarkably it is even faster than hoisting the bounds check
// out of the loop, by doing something like
//   _, _ = x[len(z)-1], y[len(z)-1]
// There are other ways to hoist the bounds check out of the loop,
// but the compiler's BCE isn't powerful enough for them (yet?).
// See the discussion in CL 164966.

import "math/bits"

func mulAddVWW(z, x []uint64, y, r uint64) (c uint64) {
	c = r
	// The comment near the top of this file discusses this for loop condition.
	for i := 0; i < len(z) && i < len(x); i++ {
		c, z[i] = mul128(x[i], y, c)
	}
	return
}

// The resulting carry c is either 0 or 1.
func addVV(z, x, y []uint64) (c uint64) {
	// The comment near the top of this file discusses this for loop condition.
	for i := 0; i < len(z) && i < len(x) && i < len(y); i++ {
		zi, cc := bits.Add64(x[i], y[i], c)
		z[i] = zi
		c = cc
	}
	return
}

// The resulting carry c is either 0 or 1.
func subVV(z, x, y []uint64) (c uint64) {
	// The comment near the top of this file discusses this for loop condition.
	for i := 0; i < len(z) && i < len(x) && i < len(y); i++ {
		zi, cc := bits.Sub(uint(x[i]), uint(y[i]), uint(c))
		z[i] = uint64(zi)
		c = uint64(cc)
	}
	return
}

// q = ( x1 << _W + x0 - r)/y. m = floor(( _B^2 - 1 ) / d - _B). Requiring x1<y.
// An approximate reciprocal with a reference to "Improved Division by Invariant Integers
// (IEEE Transactions on Computers, 11 Jun. 2010)"
func divWW(x1, x0, y, m uint64) (q, r uint64) {
	s := bits.LeadingZeros64(y)
	if s != 0 {
		x1 = x1<<s | x0>>(64-s)
		x0 <<= s
		y <<= s
	}
	d := uint(y)
	// We know that
	//   m = ⎣(B^2-1)/d⎦-B
	//   ⎣(B^2-1)/d⎦ = m+B
	//   (B^2-1)/d = m+B+delta1    0 <= delta1 <= (d-1)/d
	//   B^2/d = m+B+delta2        0 <= delta2 <= 1
	// The quotient we're trying to compute is
	//   quotient = ⎣(x1*B+x0)/d⎦
	//            = ⎣(x1*B*(B^2/d)+x0*(B^2/d))/B^2⎦
	//            = ⎣(x1*B*(m+B+delta2)+x0*(m+B+delta2))/B^2⎦
	//            = ⎣(x1*m+x1*B+x0)/B + x0*m/B^2 + delta2*(x1*B+x0)/B^2⎦
	// The latter two terms of this three-term sum are between 0 and 1.
	// So we can compute just the first term, and we will be low by at most 2.
	t1, t0 := bits.Mul(uint(m), uint(x1))
	_, c := bits.Add(t0, uint(x0), 0)
	t1, _ = bits.Add(t1, uint(x1), c)
	// The quotient is either t1, t1+1, or t1+2.
	// We'll try t1 and adjust if needed.
	qq := t1
	// compute remainder r=x-d*q.
	dq1, dq0 := bits.Mul(d, qq)
	r0, b := bits.Sub(uint(x0), dq0, 0)
	r1, _ := bits.Sub(uint(x1), dq1, b)
	// The remainder we just computed is bounded above by B+d:
	// r = x1*B + x0 - d*q.
	//   = x1*B + x0 - d*⎣(x1*m+x1*B+x0)/B⎦
	//   = x1*B + x0 - d*((x1*m+x1*B+x0)/B-alpha)                                   0 <= alpha < 1
	//   = x1*B + x0 - x1*d/B*m                         - x1*d - x0*d/B + d*alpha
	//   = x1*B + x0 - x1*d/B*⎣(B^2-1)/d-B⎦             - x1*d - x0*d/B + d*alpha
	//   = x1*B + x0 - x1*d/B*⎣(B^2-1)/d-B⎦             - x1*d - x0*d/B + d*alpha
	//   = x1*B + x0 - x1*d/B*((B^2-1)/d-B-beta)        - x1*d - x0*d/B + d*alpha   0 <= beta < 1
	//   = x1*B + x0 - x1*B + x1/B + x1*d + x1*d/B*beta - x1*d - x0*d/B + d*alpha
	//   =        x0        + x1/B        + x1*d/B*beta        - x0*d/B + d*alpha
	//   = x0*(1-d/B) + x1*(1+d*beta)/B + d*alpha
	//   <  B*(1-d/B) +  d*B/B          + d          because x0<B (and 1-d/B>0), x1<d, 1+d*beta<=B, alpha<1
	//   =  B - d     +  d              + d
	//   = B+d
	// So r1 can only be 0 or 1. If r1 is 1, then we know q was too small.
	// Add 1 to q and subtract d from r. That guarantees that r is <B, so
	// we no longer need to keep track of r1.
	if r1 != 0 {
		qq++
		r0 -= d
	}
	// If the remainder is still too large, increment q one more time.
	if r0 >= d {
		qq++
		r0 -= d
	}
	return uint64(qq), uint64(r0 >> s)
}

// greaterThan reports whether (x1<<_W + x2) > (y1<<_W + y2)
func greaterThan(x1, x2, y1, y2 uint64) bool {
	return x1 > y1 || x1 == y1 && x2 > y2
}

func reciprocal(d1 uint64) uint64 {
	const mask = 1<<64 - 1
	u := d1 << bits.LeadingZeros64(d1)
	// (_B^2-1)/U-_B = (_B*(_M-C)+_M)/U
	rec, _ := bits.Div64(^u, mask, u)
	return rec
}
