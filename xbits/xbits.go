// Package xbits augments math/bits with larger integers.
package xbits

import (
	"crypto/subtle"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"math/bits"

	"github.com/elagergren/ctb/xbits/ct"
)

// Uint256 is an unsigned 256-bit integer.
type Uint256 struct {
	u0, u1, u2, u3 uint64
}

var (
	_ fmt.Stringer  = Uint256{}
	_ fmt.Formatter = Uint256{}
)

// U256 creates a Uint256 from a uint64.
func U256(x uint64) Uint256 {
	return Uint256{x, 0, 0, 0}
}

// Add returns x + y.
//
// This function's execution time does not depend on its inputs.
//go:noinline
func (x Uint256) Add(y Uint256) Uint256 {
	var z Uint256
	var c uint64
	z.u0, c = bits.Add64(x.u0, y.u0, c)
	z.u1, c = bits.Add64(x.u1, y.u1, c)
	z.u2, c = bits.Add64(x.u2, y.u2, c)
	z.u3, _ = bits.Add64(x.u3, y.u3, c)
	return z
}

// And returns x & y.
//
// This function's execution time does not depend on its inputs.
func (x Uint256) And(y Uint256) Uint256 {
	var z Uint256
	z.u0 = x.u0 & y.u0
	z.u1 = x.u1 & y.u1
	z.u2 = x.u2 & y.u2
	z.u3 = x.u3 & y.u3
	return z
}

// Bit returns the value of bit at index i.
//
// That is, Bit returns (x>>i)&1.
// The index must be >= 0.
func (x Uint256) Bit(i int) uint {
	if i < 0 {
		panic("negative bit index")
	}
	if i >= 256 {
		return 0
	}
	n := make([]uint64, 4)
	n[0] = x.u0
	n[1] = x.u1
	n[2] = x.u2
	n[3] = x.u3
	return uint(n[i/64] >> (i % 64) & 1)
}

// BitLen returns the absolute value of x in bits.
//
// In other words, the number of bits needed to
// represent x.
func (x Uint256) BitLen() int {
	switch {
	case x.u3 != 0:
		return 192 + bits.Len64(x.u3)
	case x.u2 != 0:
		return 128 + bits.Len64(x.u2)
	case x.u1 != 0:
		return 64 + bits.Len64(x.u1)
	default:
		return bits.Len64(x.u0)
	}
}

// Cmp compares u and x and returns
//
//    +1 if x > y
//     0 if x == y
//    -1 if x < y
//
func (x Uint256) Cmp(y Uint256) int {
	var z Uint256
	var b uint64
	z.u0, b = bits.Sub64(x.u0, y.u0, b)
	z.u1, b = bits.Sub64(x.u1, y.u1, b)
	z.u2, b = bits.Sub64(x.u2, y.u2, b)
	z.u3, b = bits.Sub64(x.u3, y.u3, b)

	r := z.u0 ^ z.u1 ^ z.u2 ^ z.u3
	// If r == 0 then x == y
	// If r != 0 then x != y
	// If b == 0 then x > y
	// If b == 1 then x <= y
	if b == 0 {
		return +1
	}
	if r == 0 {
		return +0
	}
	return -1
}

// Exp returns x**y mod m.
func (x Uint256) Exp(y, m Uint256) Uint256 {
	x1 := x
	x2 := x.MulMod(x, m)
	for i := 256 - 2; i >= 0; i-- {
		if y.Bit(i) == 0 {
			// x2 = x1*x2 mod m
			x2 = x1.MulMod(x2, m)
			// x1 = x1^2 mod m
			x1 = x1.MulMod(x1, m)
		} else {
			// x1 = x1*x2 mod m
			x1 = x1.MulMod(x2, m)
			// x2 = x2^2 mod m
			x2 = x2.MulMod(x2, m)
		}
	}
	return x1
}

// FillBytes sets buf to x, storing it as a zero-extended
// big-endian byte slice, and returns buf.
//
// If x does not fit into buf, FillBytes will panic.
func (x Uint256) FillBytes(buf []byte) []byte {
	binary.BigEndian.PutUint64(buf[24:], x.u0)
	binary.BigEndian.PutUint64(buf[16:24], x.u1)
	binary.BigEndian.PutUint64(buf[8:16], x.u2)
	binary.BigEndian.PutUint64(buf[:8], x.u3)
	return buf
}

func (x Uint256) Format(s fmt.State, ch rune) {
	// Implementation borrowed from math/big.

	var base int
	switch ch {
	case 'b':
		base = 2
	case 'o', 'O':
		base = 8
	case 'd', 's', 'v':
		base = 10
	case 'x', 'X':
		base = 16
	default:
		fmt.Fprintf(s, "%%!%c(xbits.Uint256=%s)", ch, x.String())
		return
	}

	sign := ""
	switch {
	case s.Flag('+'):
		sign = "+"
	case s.Flag(' '):
		sign = " "
	}

	prefix := ""
	if s.Flag('#') {
		switch ch {
		case 'b':
			prefix = "0b"
		case 'o':
			prefix = "0"
		case 'x':
			prefix = "0x"
		case 'X':
			prefix = "0X"
		}
	}
	if ch == 'O' {
		prefix = "0o"
	}

	digits := []byte(x.Text(base))
	if ch == 'X' {
		for i, d := range digits {
			if 'a' <= d && d <= 'z' {
				digits[i] = 'A' + (d - 'a')
			}
		}
	}

	var left int
	var zeros int
	var right int

	precision, precisionSet := s.Precision()
	if precisionSet {
		switch {
		case len(digits) < precision:
			zeros = precision - len(digits)
		case len(digits) == 1 && digits[0] == '0' && precision == 0:
			return
		}
	}

	length := len(sign) + len(prefix) + zeros + len(digits)
	if width, widthSet := s.Width(); widthSet && length < width {
		switch d := width - length; {
		case s.Flag('-'):
			right = d
		case s.Flag('0') && !precisionSet:
			zeros = d
		default:
			left = d
		}
	}

	writeMultiple(s, " ", left)
	writeMultiple(s, sign, 1)
	writeMultiple(s, prefix, 1)
	writeMultiple(s, "0", zeros)
	s.Write(digits)
	writeMultiple(s, " ", right)
}

// write count copies of text to s
func writeMultiple(s fmt.State, text string, count int) {
	if len(text) > 0 {
		b := []byte(text)
		for ; count > 0; count-- {
			s.Write(b)
		}
	}
}

// Lsh returns x<<n.
//
// This function's execution time does not depend on its inputs.
func (x Uint256) Lsh(n uint) Uint256 {
	s := n % 64
	ŝ := 64 - s

	// If n is in [0, 256) set i = n/64.
	// Otherwise, set i = 4.
	i := subtle.ConstantTimeSelect(int(ct.LessOrEq(n, 255)), int(n/64), 4)

	res := make([]uint64, 8)
	res[i+3] = x.u3<<s | x.u2>>ŝ
	res[i+2] = x.u2<<s | x.u1>>ŝ
	res[i+1] = x.u1<<s | x.u0>>ŝ
	res[i+0] = x.u0 << s

	var z Uint256
	z.u0 = res[0]
	z.u1 = res[1]
	z.u2 = res[2]
	z.u3 = res[3]
	return z
}

// shr sets z = x<<n for n in [0, 64].
func shl(z, x []uint64, n uint) uint64 {
	s := n % 64
	ŝ := 64 - s

	c := x[len(z)-1] >> ŝ
	for i := len(z) - 1; i > 0; i-- {
		z[i] = x[i]<<s | x[i-1]>>ŝ
	}
	z[0] = x[0] << s
	return c
}

// ModInverse returns the multiplicative inverse
// of x in the ring ℤ/nℤ.
//
// If x and n are not relatively prime, x has no
// multiplicative inverse in the ring ℤ/nℤ and
// ModInverse will panic.
func (x Uint256) ModInverse(n Uint256) Uint256 {
	panic("TODO")
}

// Mul returns x * y.
//
// This function's execution time does not depend on its inputs.
func (x Uint256) Mul(y Uint256) Uint256 {
	z := make([]uint64, 8)
	mul512(z, x, y)
	return Uint256{z[0], z[1], z[2], z[3]}
}

// MulMod returns x*y mod m.
func (x Uint256) MulMod(y, m Uint256) Uint256 {
	z := make([]uint64, 8)
	mul512(z, x, y)

	if m.BitLen() <= 64 {
		return U256(mod64(z, m.u0))
	}

	v := make([]uint64, 4)
	v[0] = m.u0
	v[1] = m.u1
	v[2] = m.u2
	v[3] = m.u3

	q := make([]uint64, 8)
	r := div512(q, z, v)
	return Uint256{r[0], r[1], r[2], r[3]}
}

// mul512 returns the 512-bit product of x*y.
//
// mul512 has the following conditions:
//
//    len(z) == 8
//
// This function's execution time does not depend on its inputs.
func mul512(z []uint64, x, y Uint256) {
	var (
		c      uint64
		z1, z0 uint64
	)

	// y_0 * x
	//
	// Store in z[0:4]

	c, z[0] = mul128(x.u0, y.u0, z[0])

	z1, z0 = mul128(x.u1, y.u0, z[1])
	lo, cc := bits.Add64(z0, c, 0)
	c, z[1] = cc, lo
	c += z1

	z1, z0 = mul128(x.u2, y.u0, z[2])
	lo, cc = bits.Add64(z0, c, 0)
	c, z[2] = cc, lo
	c += z1

	z1, z0 = mul128(x.u3, y.u0, z[3])
	lo, cc = bits.Add64(z0, c, 0)
	c, z[3] = cc, lo
	c += z1

	z[4] = c

	// y_1 * x
	//
	// Store in z[1:5]

	c, z[1] = mul128(x.u0, y.u1, z[1])

	z1, z0 = mul128(x.u1, y.u1, z[2])
	lo, cc = bits.Add64(z0, c, 0)
	c, z[2] = cc, lo
	c += z1

	z1, z0 = mul128(x.u2, y.u1, z[3])
	lo, cc = bits.Add64(z0, c, 0)
	c, z[3] = cc, lo
	c += z1

	z1, z0 = mul128(x.u3, y.u1, z[4])
	lo, cc = bits.Add64(z0, c, 0)
	c, z[4] = cc, lo
	c += z1

	z[5] = c

	// y_2 * x
	//
	// Store in z[2:6]

	c, z[2] = mul128(x.u0, y.u2, z[2])

	z1, z0 = mul128(x.u1, y.u2, z[3])
	lo, cc = bits.Add64(z0, c, 0)
	c, z[3] = cc, lo
	c += z1

	z1, z0 = mul128(x.u2, y.u2, z[4])
	lo, cc = bits.Add64(z0, c, 0)
	c, z[4] = cc, lo
	c += z1

	z1, z0 = mul128(x.u3, y.u2, z[5])
	lo, cc = bits.Add64(z0, c, 0)
	c, z[5] = cc, lo
	c += z1

	z[6] = c

	// y_3 * x
	//
	// Store in z[3:7]

	c, z[3] = mul128(x.u0, y.u3, z[3])

	z1, z0 = mul128(x.u1, y.u3, z[4])
	lo, cc = bits.Add64(z0, c, 0)
	c, z[4] = cc, lo
	c += z1

	z1, z0 = mul128(x.u2, y.u3, z[5])
	lo, cc = bits.Add64(z0, c, 0)
	c, z[5] = cc, lo
	c += z1

	z1, z0 = mul128(x.u3, y.u3, z[6])
	lo, cc = bits.Add64(z0, c, 0)
	c, z[6] = cc, lo
	c += z1

	z[7] = c
}

func mul128(x, y, c uint64) (z1, z0 uint64) {
	hi, lo := bits.Mul64(x, y)
	lo, c = bits.Add64(lo, c, 0)
	return hi + c, lo
}

// LeadingZeros returns the number of leading
// zero bits in x.
//
// The result is 256 if x == 0.
func (x Uint256) LeadingZeros() int {
	return 256 - x.BitLen()
}

// OnesCount returns the number of one bits
// in x.
//
// Also known as the "population count."
func (x Uint256) OnesCount() int {
	var n int
	n += bits.OnesCount64(x.u0)
	n += bits.OnesCount64(x.u1)
	n += bits.OnesCount64(x.u2)
	n += bits.OnesCount64(x.u3)
	return n
}

// Or returns x | y.
//
// This function's execution time does not depend on its inputs.
func (x Uint256) Or(y Uint256) Uint256 {
	var z Uint256
	z.u0 = x.u0 | y.u0
	z.u1 = x.u1 | y.u1
	z.u2 = x.u2 | y.u2
	z.u3 = x.u3 | y.u3
	return z
}

// Quo returns x / y.
//
// Quo implements truncated division, like Go.
func (x Uint256) Quo(y Uint256) Uint256 {
	q, _ := x.QuoRem(y)
	return q
}

// QuoRem returns x / y and x % y.
//
// QuoRem implements truncated division and
// modulus, like Go.
func (x Uint256) QuoRem(y Uint256) (Uint256, Uint256) {
	// QuoRem is largely borrowed from math/big.

	if l := y.BitLen(); l <= 64 {
		if l == 0 {
			panic("division by zero")
		}
		y := y.u0
		rec := reciprocal(y)
		var q Uint256
		var r uint64
		q.u3, r = divWW(r, x.u3, y, rec)
		q.u2, r = divWW(r, x.u2, y, rec)
		q.u1, r = divWW(r, x.u1, y, rec)
		q.u0, r = divWW(r, x.u0, y, rec)
		return q, U256(r)
	}

	u := make([]uint64, 5)
	u[0] = x.u0
	u[1] = x.u1
	u[2] = x.u2
	u[3] = x.u3

	v := make([]uint64, 4)
	v[0] = y.u0
	v[1] = y.u1
	v[2] = y.u2
	v[3] = y.u3

	q := make([]uint64, 5)
	r := div512(q, u, v)

	quo := Uint256{q[0], q[1], q[2], q[3]}
	rem := Uint256{r[0], r[1], r[2], r[3]}
	return quo, rem
}

// mod64 return u%v.
func mod64(uIn []uint64, v uint64) (r uint64) {
	rec := reciprocal(v)
	for i := len(uIn) - 1; i >= 0; i-- {
		_, r = divWW(r, uIn[i], v, rec)
	}
	return r
}

// div512 sets q = u/v and returns r = u%v.
//
// div512 has the following conditions:
//
//    len(v) >= 2
//    len(u) >= len(v) + 1
//    u[len(u)-1] must be zero (reserved for r)
//    len(q) >= len(u)
//    u must not alias z
//
// div512 reuses u as storage for r.
// div512 modifies v.
func div512(q, uIn, vIn []uint64) (r []uint64) {
	_ = vIn[2-1]
	_ = uIn[len(vIn)+1-1]
	_ = q[len(uIn)-len(vIn)+1-1]

	// Normalize v.
	n := len(vIn)
	for n > 0 && vIn[n-1] == 0 {
		n--
	}
	vIn = vIn[:n]

	// Normalize u.
	uIn[len(uIn)-1] = 0
	m := len(uIn) - 1
	for m > 0 && uIn[m-1] == 0 {
		m--
	}
	uIn = uIn[:m]
	m -= n

	// D1.
	shift := uint(bits.LeadingZeros64(vIn[n-1]))

	v := vIn
	shl(v, vIn, shift)

	u := uIn[:len(uIn)+1]
	u[m] = shl(u[:len(uIn)], uIn, shift)

	q = q[:m+1]

	qhatv := make([]uint64, 8)
	qhatvLen := n + 1

	// D2.
	vn1 := v[n-1]
	rec := reciprocal(vn1)
	for j := m; j >= 0; j-- {
		// D3.
		const mask = 1<<64 - 1
		qhat := uint64(mask)
		var ujn uint64
		if j+n < len(u) {
			ujn = u[j+n]
		}
		if ujn != vn1 {
			var rhat uint64
			qhat, rhat = divWW(ujn, u[j+n-1], vn1, rec)

			// x1 | x2 = q̂v_{n-2}
			vn2 := v[n-2]
			x1, x2 := bits.Mul64(qhat, vn2)
			// test if q̂v_{n-2} > br̂ + u_{j+n-2}
			ujn2 := u[j+n-2]
			for greaterThan(x1, x2, rhat, ujn2) {
				qhat--
				prevRhat := rhat
				rhat += vn1
				// v[n-1] >= 0, so this tests for overflow.
				if rhat < prevRhat {
					break
				}
				x1, x2 = bits.Mul64(qhat, vn2)
			}
		}

		// D4.
		// Compute the remainder u - (q̂*v) << (_W*j).
		// The subtraction may overflow if q̂ estimate was off by one.
		qhatv[n] = mulAddVWW(qhatv[0:n], v, qhat, 0)
		qhl := qhatvLen
		if j+qhl > len(u) && qhatv[n] == 0 {
			qhl--
		}
		c := subVV(u[j:j+qhl], u[j:], qhatv)
		if c != 0 {
			c := addVV(u[j:j+n], u[j:], v)
			// If n == qhl, the carry from subVV and the carry from addVV
			// cancel out and don't affect u[j+n].
			if n < qhl {
				u[j+n] += c
			}
			qhat--
		}

		if j == m && m == len(q) && qhat == 0 {
			continue
		}
		q[j] = qhat
	}
	shr(u, u, shift)
	r = u
	return r
}

// Rem returns x % y.
//
// Rem implements truncated modulus, like Go.
func (x Uint256) Rem(y Uint256) Uint256 {
	_, r := x.QuoRem(y)
	return r
}

// RotateLeft returns the value of x rotated left
// by (k mod 256) bits.
//
// This function's execution time does not depend on its inputs.
func (x Uint256) RotateLeft(k int) Uint256 {
	const n = 256
	s := uint(k) & (n - 1)
	return x.Lsh(s).Or(x.Rsh(n - s))
}

// RoateRight returns the value of x rotated right
// by (k mod 256) bits.
//
// This function's execution time does not depend on its inputs.
func (x Uint256) RotateRight(k int) Uint256 {
	return x.RotateLeft(-k)
}

// Reverse returns the value of x with its bits in
// reversed order.
func (x Uint256) Reverse() Uint256 {
	var z Uint256
	z.u0 = bits.Reverse64(x.u0)
	z.u1 = bits.Reverse64(x.u1)
	z.u2 = bits.Reverse64(x.u2)
	z.u3 = bits.Reverse64(x.u3)
	return z
}

// ReverseBytes returns the value of x with its bytes
// in reversed order.
//
// This function's execution time does not depend on its inputs.
func (x Uint256) ReverseBytes() Uint256 {
	var z Uint256
	z.u0 = bits.ReverseBytes64(x.u0)
	z.u1 = bits.ReverseBytes64(x.u1)
	z.u2 = bits.ReverseBytes64(x.u2)
	z.u3 = bits.ReverseBytes64(x.u3)
	return z
}

// Rsh returns x>>n.
//
// This function's execution time does not depend on its inputs.
func (x Uint256) Rsh(n uint) Uint256 {
	s := n % 64
	ŝ := 64 - s

	res := make([]uint64, 8)
	res[0] = x.u0>>s | x.u1<<ŝ
	res[1] = x.u1>>s | x.u2<<ŝ
	res[2] = x.u2>>s | x.u3<<ŝ
	res[3] = x.u3 >> s

	// If n is in [0, 256) set i = n/64.
	// Otherwise, set i = 4.
	i := subtle.ConstantTimeSelect(int(ct.LessOrEq(n, 255)), int(n/64), 4)

	var z Uint256
	z.u0 = res[i+0]
	z.u1 = res[i+1]
	z.u2 = res[i+2]
	z.u3 = res[i+3]
	return z
}

// shr sets z = x>>n for n in [0, 64].
func shr(z, x []uint64, n uint) uint64 {
	s := n % 64
	ŝ := 64 - s

	c := x[0] << ŝ
	z[0] = x[0]>>s | x[1]<<ŝ
	z[1] = x[1]>>s | x[2]<<ŝ
	z[2] = x[2]>>s | x[3]<<ŝ
	z[3] = x[3] >> s
	return c
}

// SetBytes sets z to the big-endian unsigned integer buf.
//
// SetBytes panics if buf overflows z (buf > 1<<256-1).
func (z *Uint256) SetBytes(buf []byte) {
	if len(buf) > 32 {
		panic("SetBytes: integer too large")
	}

	*z = Uint256{}
	if len(buf) > 24 {
		z.u3 = be64(buf[:len(buf)-24])
		buf = buf[len(buf)-24:]
	}
	if len(buf) > 16 {
		z.u2 = be64(buf[:len(buf)-16])
		buf = buf[len(buf)-16:]
	}
	if len(buf) > 8 {
		z.u1 = be64(buf[:len(buf)-8])
		buf = buf[len(buf)-8:]
	}
	if len(buf) > 0 {
		z.u0 = be64(buf)
	}
}

func be64(b []byte) uint64 {
	switch len(b) {
	case 8:
		return binary.BigEndian.Uint64(b)
	case 7:
		_ = b[6] // bounds check hint to compiler; see golang.org/issue/14808
		return uint64(b[6]) | uint64(b[5])<<8 | uint64(b[4])<<16 |
			uint64(b[3])<<24 | uint64(b[2])<<32 | uint64(b[1])<<40 |
			uint64(b[0])<<48
	case 6:
		_ = b[5] // bounds check hint to compiler; see golang.org/issue/14808
		return uint64(b[5]) | uint64(b[4])<<8 | uint64(b[3])<<16 |
			uint64(b[2])<<24 | uint64(b[1])<<32 | uint64(b[0])<<40
	case 5:
		_ = b[4] // bounds check hint to compiler; see golang.org/issue/14808
		return uint64(b[4]) | uint64(b[3])<<8 | uint64(b[2])<<16 |
			uint64(b[1])<<24 | uint64(b[0])<<32
	case 4:
		return uint64(binary.BigEndian.Uint32(b))
	case 3:
		_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
		return uint64(b[2]) | uint64(b[1])<<8 | uint64(b[0])<<16
	case 2:
		return uint64(binary.BigEndian.Uint16(b))
	case 1:
		return uint64(b[0])
	case 0:
		return 0
	default:
		panic("unreachable")
	}
}

// Sub returns x - y.
//
// This function's execution time does not depend on its inputs.
func (x Uint256) Sub(y Uint256) Uint256 {
	var z Uint256
	var b uint64
	z.u0, b = bits.Sub64(x.u0, y.u0, b)
	z.u1, b = bits.Sub64(x.u1, y.u1, b)
	z.u2, b = bits.Sub64(x.u2, y.u2, b)
	z.u3, _ = bits.Sub64(x.u3, y.u3, b)
	return z
}

func (x Uint256) String() string {
	return x.Text(10)
}

// Text returns the textual representation of
// x in the provided base.
//
// The base must be in [2, 62].
func (x Uint256) Text(base int) string {
	var z big.Int
	setInt(&z, x)
	return z.Text(base)
}

// TrailingZeros returns the number of trailing
// zero bits in x.
//
// The result is 256 if x == 0.
func (x Uint256) TrailingZeros() int {
	switch {
	case x.u0 != 0:
		return bits.TrailingZeros64(x.u0)
	case x.u1 != 0:
		return 64 + bits.TrailingZeros64(x.u1)
	case x.u2 != 0:
		return 128 + bits.TrailingZeros64(x.u2)
	default:
		return 192 + bits.TrailingZeros64(x.u3)
	}
}

// Uint64 returns the uint64 representation of x.
//
// The result is undefined if x cannot be
// represented as a uint64.
func (x Uint256) Uint64() uint64 {
	return x.u0
}

// Xor returns x ^ y.
//
// This function's execution time does not depend on its inputs.
func (x Uint256) Xor(y Uint256) Uint256 {
	var z Uint256
	z.u0 = x.u0 ^ y.u0
	z.u1 = x.u1 ^ y.u1
	z.u2 = x.u2 ^ y.u2
	z.u3 = x.u3 ^ y.u3
	return z
}

// Rand256 returns a Uint256 in [0, max).
//
// Rand256 panics if max = 0.
func Rand256(rand io.Reader, max Uint256) (Uint256, error) {
	// Implementation borrowed from crypto/rand.
	if max.BitLen() == 0 {
		panic("xbits: argument to Rand256 is 0")
	}
	n := max.Sub(U256(1))
	// bitLen is the maximum bit length needed to encode a value < max.
	bitLen := n.BitLen()
	if bitLen == 0 {
		// the only valid result is 0
		return U256(0), nil
	}
	// k is the maximum byte length needed to encode a value < max.
	k := (bitLen + 7) / 8
	// b is the number of bits in the most significant byte of max-1.
	b := uint(bitLen % 8)
	if b == 0 {
		b = 8
	}

	bytes := make([]byte, k)
	for {
		_, err := io.ReadFull(rand, bytes)
		if err != nil {
			return U256(0), err
		}

		// Clear bits in the first byte to increase the probability
		// that the candidate is < max.
		bytes[0] &= uint8(int(1<<b) - 1)

		n.SetBytes(bytes)
		if n.Cmp(max) < 0 {
			return n, nil
		}
	}
}

func setInt(z *big.Int, x Uint256) {
	const _W = bits.UintSize
	if _W == 64 {
		z.SetBits([]big.Word{
			big.Word(x.u0),
			big.Word(x.u1),
			big.Word(x.u2),
			big.Word(x.u3),
		})
	} else {
		// _W == 32
		z.SetBits([]big.Word{
			big.Word(x.u0),
			big.Word(x.u0 >> 32),
			big.Word(x.u1),
			big.Word(x.u1 >> 32),
			big.Word(x.u2),
			big.Word(x.u2 >> 32),
			big.Word(x.u3),
			big.Word(x.u3 >> 32),
		})
	}
}
