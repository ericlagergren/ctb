package lll

import (
	"fmt"
	"math/big"
)

// T is an integer.
//
// Unlike math/big, T is a value type.
type T interface {
	Sign() int
	Cmp(T) int
	CmpAbs(T) int
	Add(T) T
	Mul(T) T
	Sub(T) T
	Quo(T) T
	String() string
}

func SetInt(z *big.Int, x T) {
	switch x := x.(type) {
	case *Int:
		z.Set(&x.x)
	case *Frac:
		if !x.x.IsInt() {
			SetInt(z, round(x))
		} else {
			z.Set(x.x.Num())
		}
	}
}

// Int is an integer.
//
// Int implements T.
type Int struct {
	x big.Int
}

var _ T = (*Int)(nil)

// I64 creates an Int from x.
func I64(x int64) T {
	var z Int
	z.x.SetInt64(x)
	return &z
}

// I copies x into a new Int.
func I(x *big.Int) T {
	var z Int
	z.x.Set(x)
	return &z
}

func (x *Int) String() string {
	return x.x.String()
}

func (x *Int) Sign() int {
	return x.x.Sign()
}

func (x *Int) CmpAbs(y T) int {
	switch {
	case x.Sign() == 0 && y.Sign() == 0:
		return 0
	}
	switch y := y.(type) {
	case *Int:
		return x.x.CmpAbs(&y.x)
	case *Frac:
		if y.x.IsInt() {
			return x.x.CmpAbs(y.x.Num())
		}
		var tmp big.Rat
		tmp.SetInt(&x.x)
		// Set sign(x) = sign(y)
		if y.Sign() < 0 {
			tmp.Neg(&tmp)
		} else {
			tmp.Abs(&tmp)
		}
		return tmp.Cmp(&y.x)
	default:
		panic(fmt.Sprintf("unknown type: %T", y))
	}
}

func (x *Int) Cmp(y T) int {
	switch {
	case x.Sign() < y.Sign():
		return -1
	case x.Sign() > y.Sign():
		return +1
	case x.Sign() == 0 && y.Sign() == 0:
		return 0
	}
	switch y := y.(type) {
	case *Int:
		return x.x.Cmp(&y.x)
	case *Frac:
		var tmp big.Rat
		tmp.SetInt(&x.x)
		return tmp.Cmp(&y.x)
	default:
		panic(fmt.Sprintf("unknown type: %T", y))
	}
}

func (x *Int) Add(y T) T {
	switch y := y.(type) {
	case *Int:
		var z Int
		z.x.Add(&x.x, &y.x)
		return &z
	case *Frac:
		if y.x.IsInt() {
			var z Int
			z.x.Add(&x.x, y.x.Num())
			return &z
		}
		var tmp big.Rat
		tmp.SetInt(&x.x)
		var z Frac
		z.x.Add(&tmp, &y.x)
		return &z
	default:
		panic(fmt.Sprintf("unknown type: %T", y))
	}
}

func (x *Int) Mul(y T) T {
	switch y := y.(type) {
	case *Int:
		var z Int
		z.x.Mul(&x.x, &y.x)
		return &z
	case *Frac:
		if y.x.IsInt() {
			var z Int
			z.x.Mul(&x.x, y.x.Num())
			return &z
		}
		var z Frac
		var tmp big.Rat
		tmp.SetInt(&x.x)
		z.x.Mul(&tmp, &y.x)
		return &z
	default:
		panic(fmt.Sprintf("unknown type: %T", y))
	}
}

func (x *Int) Sub(y T) T {
	switch y := y.(type) {
	case *Int:
		var z Int
		z.x.Sub(&x.x, &y.x)
		return &z
	case *Frac:
		if y.x.IsInt() {
			var z Int
			z.x.Sub(&x.x, y.x.Num())
			return &z
		}
		var tmp big.Rat
		tmp.SetInt(&x.x)
		var z Frac
		z.x.Sub(&tmp, &y.x)
		return &z
	default:
		panic(fmt.Sprintf("unknown type: %T", y))
	}
}

func (x *Int) Quo(y T) T {
	switch y := y.(type) {
	case *Int:
		var z Frac
		z.x.SetFrac(&x.x, &y.x)
		return &z
	case *Frac:
		var tmp big.Rat
		tmp.SetInt(&x.x)
		var z Frac
		z.x.Quo(&tmp, &y.x)
		return &z
	default:
		panic(fmt.Sprintf("unknown type: %T", y))
	}
}

// Frac is a fraction (rational) number.
//
// Frac implements T.
type Frac struct {
	x big.Rat
}

var _ T = (*Frac)(nil)

// F64 creates a Frac from a numerateor and denominator.
func F64(n, d int64) T {
	if d == 1 {
		return I64(n)
	}
	var z Frac
	z.x.SetFrac64(n, d)
	return &z
}

// F copies the numerator and denominator into a Frac.
func F(n, d *big.Int) T {
	var z Frac
	z.x.SetFrac(n, d)
	return &z
}

func (x *Frac) Sign() int {
	return x.x.Sign()
}

func (x *Frac) Cmp(y T) int {
	switch {
	case x.Sign() < y.Sign():
		return -1
	case x.Sign() > y.Sign():
		return +1
	case x.Sign() == 0 && y.Sign() == 0:
		return 0
	}
	switch y := y.(type) {
	case *Int:
		var tmp big.Rat
		tmp.SetInt(&y.x)
		return x.x.Cmp(&tmp)
	case *Frac:
		return x.x.Cmp(&y.x)
	default:
		panic(fmt.Sprintf("unknown type: %T", y))
	}
}

func (x *Frac) CmpAbs(y T) int {
	switch {
	case x.Sign() == 0 && y.Sign() == 0:
		return 0
	}
	switch y := y.(type) {
	case *Int:
		if x.x.IsInt() {
			return y.x.CmpAbs(x.x.Num())
		}
		r := +1
		var tmp big.Rat
		tmp.SetInt(&y.x)
		// Set sign(y) = sign(x)
		if x.Sign() < 0 {
			r = -1
			tmp.Neg(&tmp)
		} else {
			tmp.Abs(&tmp)
		}
		return x.x.Cmp(&tmp) * r
	case *Frac:
		if x.Sign() == y.Sign() {
			return x.x.Cmp(&y.x)
		}
		r := +1
		var tmp big.Rat
		// Set sign(y) = sign(x)
		if x.Sign() < 0 {
			r = -1
			tmp.Neg(&y.x)
		} else {
			tmp.Abs(&y.x)
		}
		return x.x.Cmp(&tmp) * r
	default:
		panic(fmt.Sprintf("unknown type: %T", y))
	}
}

func (x *Frac) Add(y T) T {
	switch y := y.(type) {
	case *Frac:
		var z Frac
		z.x.Add(&x.x, &y.x)
		return &z
	case *Int:
		var tmp big.Rat
		tmp.SetInt(&y.x)
		var z Frac
		z.x.Add(&x.x, &tmp)
		return &z
	default:
		panic(fmt.Sprintf("unknown type: %T", y))
	}
}

func (x *Frac) Mul(y T) T {
	switch y := y.(type) {
	case *Frac:
		var z Frac
		z.x.Mul(&x.x, &y.x)
		return &z
	case *Int:
		var tmp big.Rat
		tmp.SetInt(&y.x)
		var z Frac
		z.x.Mul(&x.x, &tmp)
		return &z
	default:
		panic(fmt.Sprintf("unknown type: %T", y))
	}
}

func (x *Frac) Sub(y T) T {
	switch y := y.(type) {
	case *Frac:
		var z Frac
		z.x.Sub(&x.x, &y.x)
		return &z
	case *Int:
		var tmp big.Rat
		tmp.SetInt(&y.x)
		var z Frac
		z.x.Sub(&x.x, &tmp)
		return &z
	default:
		panic(fmt.Sprintf("unknown type: %T", y))
	}
}

func (x *Frac) Quo(y T) T {
	switch y := y.(type) {
	case *Frac:
		var z Frac
		z.x.Quo(&x.x, &y.x)
		return &z
	case *Int:
		var tmp big.Rat
		tmp.SetInt(&y.x)
		var z Frac
		z.x.Quo(&x.x, &tmp)
		return &z
	default:
		panic(fmt.Sprintf("unknown type: %T", y))
	}
}

func (f *Frac) String() string {
	return f.x.String()
}

var bigOne = big.NewInt(1)

func round(x T) T {
	switch x := x.(type) {
	case *Int:
		return x
	case *Frac:
		if x.x.IsInt() {
			return x
		}

		var z Int     // result
		var r big.Int // scratch

		n := x.x.Num()
		d := x.x.Denom()

		// Rats are always normalized, meaning the following
		// holds:
		//    if x.IsInt then n.Cmp(d) != 0
		if n.CmpAbs(d) < 0 {
			// Proper fraction.
			if r.Add(n, n).CmpAbs(d) >= 0 {
				z.x.Add(&z.x, bigOne)
			}
			// Round down to zero.
			return &z
		}

		// Improper fraction.
		z.x.QuoRem(n, d, &r)
		// Is r >= 0.5? If so, round up away from zero.
		if r.Add(&r, &r).CmpAbs(d) >= 0 {
			if x.Sign() < 0 {
				z.x.Sub(&z.x, bigOne)
			} else {
				z.x.Add(&z.x, bigOne)
			}
		}
		return &z
	default:
		panic(fmt.Sprintf("unknown type: %T", x))
	}
}

func sq(x T) T {
	return x.Mul(x)
}

var (
	one   = I64(1)
	half  = F64(1, 2)
	quart = F64(1, 4)
)

// Reduction computes the Lenstra–Lenstra–Lovász
// lattice basis reduction algorithm.
//
// B is a lattice basis
//    b0, b1, ... bn in Z^m
// delta must be in (1/4, 1), typically 3/4.
func Reduction(delta T, B [][]T) [][]T {
	if delta.Cmp(quart) < 0 || delta.Cmp(one) >= 0 {
		panic("delta out of range")
	}
	Bstar := gramSchmidt(nil, B)
	mu := func(i, j int) T {
		return projCoff(Bstar[j], B[i])
	}
	n := len(B)
	k := 1
	for k < n {
		for j := k - 1; j >= 0; j-- {
			mukj := mu(k, j)
			if mukj.CmpAbs(half) > 0 {
				bj := scale(nil, B[j], round(mukj))
				B[k] = sub(B[k], B[k], bj)
				Bstar = gramSchmidt(Bstar, B)
			}
		}
		dmksq := delta.Sub(sq(mu(k, k-1)))
		pbsk1 := sdot(Bstar[k-1])
		if sdot(Bstar[k]).Cmp(dmksq.Mul(pbsk1)) >= 0 {
			k++
		} else {
			B[k], B[k-1] = B[k-1], B[k]
			Bstar = gramSchmidt(Bstar, B)
			k--
			if k < 1 {
				k = 1
			}
		}
	}
	return B
}

func gramSchmidt(u, v [][]T) [][]T {
	u = u[:0]
	for _, vi := range v {
		ui := vi
		for _, uj := range u {
			// ui -= uj*vi
			uj = proj(nil, uj, vi)
			ui = sub(nil, ui, uj)
		}
		if len(ui) > 0 {
			u = append(u, ui)
		}
	}
	return u
}

// scale is
//    for i := range x {
//        z[i] = x[i] * c
//    }
func scale(z, x []T, c T) []T {
	z = zmake(z, len(x))
	for i := range x {
		z[i] = x[i].Mul(c)
	}
	return z
}

// mul is
//    for i := range x {
//        z[i] = x[i] * y[i]
//    }
func mul(z, x, y []T) []T {
	z = zmake(z, len(x))
	for i := range x {
		z[i] = x[i].Mul(y[i])
	}
	return z
}

// sub is
//    for i := range x {
//        z[i] = x[i] - y[i]
//    }
func sub(z, x, y []T) []T {
	z = zmake(z, len(x))
	for i := range x {
		z[i] = x[i].Sub(y[i])
	}
	return z
}

// proj is
//    c := projCoff(x, y)
//    scale(z, x, c)
func proj(z, x, y []T) []T {
	z = zmake(z, len(x))
	return scale(z, x, projCoff(x, y))
}

// projCoff is
//    dot(x, y) / sdot(x)
func projCoff(x, y []T) T {
	return dot(x, y).Quo(sdot(x))
}

// dot is
//    for i := range x {
//        sum += x[i] * y[i]
//    }
func dot(x, y []T) T {
	sum := I64(0)
	for i := range x {
		sum = sum.Add(x[i].Mul(y[i]))
	}
	return sum
}

// sdot is
//    dot(x, x)
func sdot(x []T) T {
	return dot(x, x)
}

func zmake(z []T, n int) []T {
	if n <= cap(z) {
		return z[:n]
	}
	return make([]T, n)
}
