package lll

import (
	"math"

	"gonum.org/v1/gonum/floats"
)

// Reduction64 computes the Lenstra–Lenstra–Lovász
// lattice basis reduction algorithm.
//
// B is a lattice basis
//    b0, b1, ... bn in Z^m
// delta must be in (1/4, 1), typically 3/4.
func Reduction64(delta float64, B [][]float64) [][]float64 {
	if delta < 1/4 || delta > 1 {
		panic("delta out of range")
	}
	Bstar := gramSchmidt64(nil, B)
	mu := func(i, j int) float64 {
		return projCoff64(Bstar[j], B[i])
	}
	n := len(B)
	k := 1
	for k < n {
		for j := k - 1; j >= 0; j-- {
			mukj := mu(k, j)
			if math.Abs(mukj) > 0.5 {
				bj := scale64(nil, B[j], math.Round(mukj))
				B[k] = sub64(B[k], B[k], bj)
				Bstar = gramSchmidt64(Bstar, B)
			}
		}
		if sdot64(Bstar[k]) >= (delta-math.Pow(mu(k, k-1), 2))*sdot64(Bstar[k-1]) {
			k++
		} else {
			B[k], B[k-1] = B[k-1], B[k]
			Bstar = gramSchmidt64(Bstar, B)
			k--
			if k < 1 {
				k = 1
			}
		}
	}
	return B
}

func gramSchmidt64(u, v [][]float64) [][]float64 {
	u = u[:0]
	for _, vi := range v {
		ui := vi
		for _, uj := range u {
			// ui -= uj*vi
			uj = proj64(nil, uj, vi)
			ui = sub64(nil, ui, uj)
		}
		if len(ui) > 0 {
			u = append(u, ui)
		}
	}
	return u
}

func scale64(z, x []float64, c float64) []float64 {
	z = zmake64(z, len(x))
	return floats.ScaleTo(z, c, x)
}

func mul64(z, x, y []float64) []float64 {
	z = zmake64(z, len(x))
	return floats.MulTo(z, x, y)
}

func sub64(z, x, y []float64) []float64 {
	z = zmake64(z, len(x))
	return floats.SubTo(z, x, y)
}

func proj64(z, x, y []float64) []float64 {
	z = zmake64(z, len(x))
	return scale64(z, x, projCoff64(x, y))
}

func projCoff64(x, y []float64) float64 {
	return dot64(x, y) / sdot64(x)
}

func dot64(x, y []float64) float64 {
	return floats.Dot(x, y)
}

func sdot64(x []float64) float64 {
	return dot64(x, x)
}

func zmake64(z []float64, n int) []float64 {
	if n <= cap(z) {
		return z[:n]
	}
	return make([]float64, n)
}
