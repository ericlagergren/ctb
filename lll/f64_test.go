package lll

import (
	"testing"

	"gonum.org/v1/gonum/floats"
)

func TestReduction64(t *testing.T) {
	for i, tc := range []struct {
		basis [][]float64
		want  [][]float64
		delta float64
	}{
		{
			basis: [][]float64{
				{1, 1, 1},
				{-1, 0, 2},
				{3, 5, 6},
			},
			want: [][]float64{
				{0, 1, 0},
				{1, 0, 1},
				{-1, 0, 2},
			},
			delta: 0.75,
		},
		{
			basis: [][]float64{
				{105, 821, 404, 328},
				{881, 667, 644, 927},
				{181, 483, 87, 500},
				{893, 834, 732, 441},
			},
			want: [][]float64{
				{76, -338, -317, 172},
				{88, -171, -229, -314},
				{269, 312, -142, 186},
				{519, -299, 470, -73},
			},
			delta: 0.75,
		},
	} {
		got := Reduction64(tc.delta, tc.basis)
		if !equal64(got, tc.want, 0) {
			t.Fatalf("#%d: wanted %v, got %v", i, tc.want, got)
		}
	}
}

func TestGramSchmidt64(t *testing.T) {
	for i, tc := range []struct {
		v    [][]float64
		want [][]float64
	}{
		{
			v:    [][]float64{{3, 1}, {2, 2}},
			want: [][]float64{{3, 1}, {-2. / 5, 6. / 5}},
		},
		{
			v:    [][]float64{{4, 1, 2}, {4, 7, 2}, {3, 1, 7}},
			want: [][]float64{{4, 1, 2}, {-8. / 7, 40. / 7, -4. / 7}, {-11. / 5, 0, 22. / 5}},
		},
	} {
		got := gramSchmidt64(nil, tc.v)
		if !equal64(got, tc.want, 0.00000001) {
			t.Fatalf("#%d: wanted %v, got %v", i, tc.want, got)
		}
	}
}

func BenchmarkReduction64(b *testing.B) {
	data := []struct {
		basis [][]float64
		want  [][]float64
		delta float64
	}{
		{
			basis: [][]float64{
				{1, 1, 1},
				{-1, 0, 2},
				{3, 5, 6},
			},
			want:  [][]float64{{0, 1, 0}, {1, 0, 1}, {-1, 0, 2}},
			delta: 0.75,
		},
		{
			basis: [][]float64{
				{105, 821, 404, 328},
				{881, 667, 644, 927},
				{181, 483, 87, 500},
				{893, 834, 732, 441},
			},
			want: [][]float64{
				{76, -338, -317, 172},
				{88, -171, -229, -314},
				{269, 312, -142, 186},
				{519, -299, 470, -73},
			},
			delta: 0.75,
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := data[i%len(data)]
		Sink64 = Reduction64(d.delta, d.basis)
	}
}

var Sink64 [][]float64

func equal64(a, b [][]float64, tol float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i, ai := range a {
		if !floats.EqualApprox(ai, b[i], tol) {
			return false
		}
	}
	return true
}
