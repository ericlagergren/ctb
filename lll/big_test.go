package lll

import (
	"testing"
)

func TestRound(t *testing.T) {
	for i, tc := range []struct {
		in, out T
	}{
		{F64(21, 2), I64(11)},                              // 10.5
		{F64(-21, 2), I64(-11)},                            // -10.5
		{F64(6283, 2000), I64(3)},                          // 3.1415
		{F64(-6283, 2000), I64(-3)},                        // -3.1415
		{F64(9285714285714286, 10000000000000000), I64(1)}, // 0.9285714285714286
	} {
		got := round(tc.in)
		if tc.out.Cmp(got) != 0 {
			t.Fatalf("#%d: expected %s, got %s", i, tc.out, got)
		}
	}
}

func TestReduction(t *testing.T) {
	for i, tc := range []struct {
		basis [][]T
		want  [][]T
		delta T
	}{
		{
			basis: [][]T{
				{I64(1), I64(1), I64(1)},
				{I64(-1), I64(0), I64(2)},
				{I64(3), I64(5), I64(6)},
			},
			want: [][]T{
				{I64(0), I64(1), I64(0)},
				{I64(1), I64(0), I64(1)},
				{I64(-1), I64(0), I64(2)},
			},
			delta: F64(3, 4),
		},
		{
			basis: [][]T{
				{I64(105), I64(821), I64(404), I64(328)},
				{I64(881), I64(667), I64(644), I64(927)},
				{I64(181), I64(483), I64(87), I64(500)},
				{I64(893), I64(834), I64(732), I64(441)},
			},
			want: [][]T{
				{I64(76), I64(-338), I64(-317), I64(172)},
				{I64(88), I64(-171), I64(-229), I64(-314)},
				{I64(269), I64(312), I64(-142), I64(186)},
				{I64(519), I64(-299), I64(470), I64(-73)},
			},
			delta: F64(3, 4),
		},
	} {
		got := Reduction(tc.delta, tc.basis)
		if !equal(got, tc.want) {
			t.Fatalf("#%d: wanted %v, got %v", i, tc.want, got)
		}
	}
}

func TestGramSchmidt(t *testing.T) {
	for i, tc := range []struct {
		v    [][]T
		want [][]T
	}{
		{
			v: [][]T{
				{I64(3), I64(1)},
				{I64(2), I64(2)},
			},
			want: [][]T{
				{I64(3), I64(1)},
				{F64(-2, 5), F64(6, 5)},
			},
		},
		{
			v: [][]T{
				{I64(4), I64(1), I64(2)},
				{I64(4), I64(7), I64(2)},
				{I64(3), I64(1), I64(7)},
			},
			want: [][]T{
				{I64(4), I64(1), I64(2)},
				{F64(-8, 7), F64(40, 7), F64(-4, 7)},
				{F64(-11, 5), I64(0), F64(22, 5)},
			},
		},
	} {
		got := gramSchmidt(nil, tc.v)
		if !equal(got, tc.want) {
			t.Fatalf("#%d: wanted %v, got %v", i, tc.want, got)
		}
	}
}

func BenchmarkReduction(b *testing.B) {
	data := []struct {
		basis [][]T
		want  [][]T
		delta T
	}{
		{
			basis: [][]T{
				{I64(1), I64(1), I64(1)},
				{I64(-1), I64(0), I64(2)},
				{I64(3), I64(5), I64(6)},
			},
			want: [][]T{
				{I64(0), I64(1), I64(0)},
				{I64(1), I64(0), I64(1)},
				{I64(-1), I64(0), I64(2)},
			},
			delta: F64(3, 4),
		},
		{
			basis: [][]T{
				{I64(105), I64(821), I64(404), I64(328)},
				{I64(881), I64(667), I64(644), I64(927)},
				{I64(181), I64(483), I64(87), I64(500)},
				{I64(893), I64(834), I64(732), I64(441)},
			},
			want: [][]T{
				{I64(76), I64(-338), I64(-317), I64(172)},
				{I64(88), I64(-171), I64(-229), I64(-314)},
				{I64(269), I64(312), I64(-142), I64(186)},
				{I64(519), I64(-299), I64(470), I64(-73)},
			},
			delta: F64(3, 4),
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := data[i%len(data)]
		Sink = Reduction(d.delta, d.basis)
	}
}

var Sink [][]T

func equal(a, b [][]T) bool {
	if len(a) != len(b) {
		return false
	}
	for i, ai := range a {
		for j, aj := range ai {
			if aj.Cmp(b[i][j]) != 0 {
				return false
			}
		}
	}
	return true
}
