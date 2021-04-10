package xbits

import (
	"fmt"
	"io"
	"math"
	"math/big"
	"math/rand"
	"testing"
)

var (
	rng    = readFunc(rand.Read)
	max256 = Uint256{math.MaxUint64, math.MaxUint64, math.MaxUint64, math.MaxUint64}
	max128 = Uint256{math.MaxUint64, math.MaxUint64, 0, 0}
)

type readFunc func([]byte) (int, error)

var _ io.Reader = (readFunc)(nil)

func (fn readFunc) Read(p []byte) (int, error) {
	return fn(p)
}

func cmpInt(x *big.Int, y Uint256) int {
	var yy big.Int
	setInt(&yy, y)
	return x.Cmp(&yy)
}

// big256Mask is 1<<256-1.
var big256Mask = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

func TestFoo(t *testing.T) {
	u2 := Uint256{17562291160714782030, 13611842547513532036, 18446744073709551615, 9223372032559808512}
	N := Uint256{17562291160714782033, 13611842547513532036, 18446744073709551615, 18446744069414584320}

	sq := u2.Add(u2)
	fmt.Println(u2)
	fmt.Println(sq)
	fmt.Println(N)
}

func TestAdd256(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x, err := Rand256(rng, max256)
		if err != nil {
			t.Fatal(err)
		}
		y, err := Rand256(rng, max256)
		if err != nil {
			t.Fatal(err)
		}
		z := x.Add(y)

		var bz, bx, by big.Int
		setInt(&bx, x)
		setInt(&by, y)
		bz.Add(&bx, &by)
		bz.And(&bz, big256Mask)

		if cmpInt(&bz, z) != 0 {
			t.Fatalf("#%d: expected %s, got %d", i, bz.String(), z)
		}
	}
}

func TestSub256(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x, err := Rand256(rng, max256)
		if err != nil {
			t.Fatal(err)
		}
		y, err := Rand256(rng, max256)
		if err != nil {
			t.Fatal(err)
		}
		z := x.Sub(y)

		var bz, bx, by big.Int
		setInt(&bx, x)
		setInt(&by, y)
		bz.Sub(&bx, &by)
		bz.And(&bz, big256Mask)

		if cmpInt(&bz, z) != 0 {
			t.Fatalf("#%d: expected %s, got %d", i, bz.String(), z)
		}
	}
}

func TestMul256(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x, err := Rand256(rng, max256)
		if err != nil {
			t.Fatal(err)
		}
		y, err := Rand256(rng, max256)
		if err != nil {
			t.Fatal(err)
		}
		z := x.Mul(y)

		var bz, bx, by big.Int
		setInt(&bx, x)
		setInt(&by, y)
		bz.Mul(&bx, &by)
		bz.And(&bz, big256Mask)

		if cmpInt(&bz, z) != 0 {
			t.Fatalf("#%d: expected %s, got %d", i, bz.String(), z)
		}
	}
}

func TestLsh256(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x, err := Rand256(rng, max256)
		if err != nil {
			t.Fatal(err)
		}
		n := uint(rand.Intn(512))
		z := x.Lsh(n)

		var bz, bx big.Int
		setInt(&bx, x)
		bz.Lsh(&bx, n)
		bz.And(&bz, big256Mask)

		if cmpInt(&bz, z) != 0 {
			t.Fatalf("#%d: expected %s, got %d", i, bz.String(), z)
		}
	}
}

func TestRsh256(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x, err := Rand256(rng, max256)
		if err != nil {
			t.Fatal(err)
		}
		n := uint(rand.Intn(512))
		z := x.Rsh(n)

		var bz, bx big.Int
		setInt(&bx, x)
		bz.Rsh(&bx, n)
		bz.And(&bz, big256Mask)

		if cmpInt(&bz, z) != 0 {
			t.Fatalf("#%d: expected %s, got %d", i, bz.String(), z)
		}
	}
}

func TestQuoRem256(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x, err := Rand256(rng, max256)
		if err != nil {
			t.Fatal(err)
		}
		y, err := Rand256(rng, U256(4))
		if err != nil {
			t.Fatal(err)
		}
		if y.BitLen() == 0 {
			y = U256(1)
		}

		var bq, br, bx, by big.Int
		setInt(&bx, x)
		setInt(&by, y)
		bq.QuoRem(&bx, &by, &br)
		bq.And(&bq, big256Mask)
		q, r := x.QuoRem(y)

		if cmpInt(&bq, q) != 0 {
			t.Fatalf("#%d: expected %s, got %d", i, bq.String(), q)
		}
		if cmpInt(&br, r) != 0 {
			t.Fatalf("#%d: expected %s, got %d", i, br.String(), r)
		}
	}
}

func TestAnd256(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x, err := Rand256(rng, max256)
		if err != nil {
			t.Fatal(err)
		}
		y, err := Rand256(rng, max256)
		if err != nil {
			t.Fatal(err)
		}
		z := x.And(y)

		var bz, bx, by big.Int
		setInt(&bx, x)
		setInt(&by, y)
		bz.And(&bx, &by)
		bz.And(&bz, big256Mask)

		if cmpInt(&bz, z) != 0 {
			t.Fatalf("#%d: expected %s, got %d", i, bz.String(), z)
		}
	}
}

func TestXor256(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x, err := Rand256(rng, max256)
		if err != nil {
			t.Fatal(err)
		}
		y, err := Rand256(rng, max256)
		if err != nil {
			t.Fatal(err)
		}
		z := x.Xor(y)

		var bz, bx, by big.Int
		setInt(&bx, x)
		setInt(&by, y)
		bz.Xor(&bx, &by)
		bz.And(&bz, big256Mask)

		if cmpInt(&bz, z) != 0 {
			t.Fatalf("#%d: expected %s, got %d", i, bz.String(), z)
		}
	}
}

func TestOr256(t *testing.T) {
	for i := 0; i < 100_000; i++ {
		x, err := Rand256(rng, max256)
		if err != nil {
			t.Fatal(err)
		}
		y, err := Rand256(rng, max256)
		if err != nil {
			t.Fatal(err)
		}
		z := x.Or(y)

		var bz, bx, by big.Int
		setInt(&bx, x)
		setInt(&by, y)
		bz.Or(&bx, &by)
		bz.And(&bz, big256Mask)

		if cmpInt(&bz, z) != 0 {
			t.Fatalf("#%d: expected %s, got %d", i, bz.String(), z)
		}
	}
}

func TestTrailingZeros256(t *testing.T) {
	for i, tc := range []struct {
		x Uint256
		n int
	}{
		{Uint256{0, 0, 0, 0}, 256},
		{Uint256{0, 0, 0, math.MaxUint64}, 192},
		{Uint256{0, 0, math.MaxUint64, 0}, 128},
		{Uint256{0, math.MaxUint64, 0, 0}, 64},
		{Uint256{math.MaxUint64, 0, 0, 0}, 0},
		{Uint256{1, 0, 0, 0}, 0},
		{Uint256{0, 0, 0, 1}, 192},
		{Uint256{0, 0, 1 << 8, 0}, 136},
	} {
		got := tc.x.TrailingZeros()
		if got != tc.n {
			t.Fatalf("#%d: expected %d, got %d", i, tc.n, got)
		}
	}
}

func TestFillBytes256(t *testing.T) {
	for i := 0; i < 10_000; i++ {
		x, err := Rand256(rng, max256)
		if err != nil {
			t.Fatal(err)
		}
		buf := x.FillBytes(make([]byte, 32))

		var y Uint256
		y.SetBytes(buf)
		if x != y {
			t.Fatalf("#%d: expected %d, got %d", i, x, y)
		}
	}

	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic")
		}
	}()

	var x Uint256
	x.FillBytes(make([]byte, 31))
}

func TestExp256(t *testing.T) {
	for i := 0; i < 10_000; i++ {
		x, err := Rand256(rng, max256)
		if err != nil {
			t.Fatal(err)
		}
		y, err := Rand256(rng, max256)
		if err != nil {
			t.Fatal(err)
		}
		m, err := Rand256(rng, U256(512))
		if err != nil {
			t.Fatal(err)
		}
		if m.BitLen() == 0 {
			m = U256(1)
		}
		z := x.Exp(y, m)

		var bz, bx, by, bm big.Int
		setInt(&bx, x)
		setInt(&by, y)
		setInt(&bm, m)
		bz.Exp(&bx, &by, &bm)
		// exp(&bz, &bx, &by, &bm)
		bz.And(&bz, big256Mask)

		if cmpInt(&bz, z) != 0 {
			t.Fatalf("#%d: expected %s, got %d", i, bz.String(), z)
		}
	}
}

func exp(z, g, n, m *big.Int) *big.Int {
	x1 := new(big.Int).Set(g)
	x2 := new(big.Int).Mul(g, g)
	x2.Mod(x2, m)
	for i := 256 - 2; i >= 0; i-- {
		if n.Bit(i) == 0 {
			x2.Mul(x1, x2)
			x2.Mod(x2, m)
			x1.Mul(x1, x1)
			x1.Mod(x1, m)
		} else {
			x1.Mul(x1, x2)
			x1.Mod(x1, m)
			x2.Mul(x2, x2)
			x2.Mod(x2, m)
		}
	}
	return z.Set(x1)
}

func TestBits256(t *testing.T) {
	for i := 0; i < 10_000; i++ {
		x, err := Rand256(rng, max256)
		if err != nil {
			t.Fatal(err)
		}

		var bx big.Int
		setInt(&bx, x)

		for j := 0; j < 256*2; j++ {
			want := bx.Bit(j)
			got := x.Bit(j)
			if want != got {
				t.Fatalf("#%d: expected %d, got %d", i, want, got)
			}
		}
	}
}

var Sink256 Uint256

func BenchmarkAdd256(b *testing.B) {
	x, err := Rand256(rng, max256)
	if err != nil {
		b.Fatal(err)
	}
	y, err := Rand256(rng, max256)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sink256 = x.Add(y)
	}
}

func BenchmarkSub256(b *testing.B) {
	x, err := Rand256(rng, max256)
	if err != nil {
		b.Fatal(err)
	}
	y, err := Rand256(rng, max256)
	if err != nil {
		b.Fatal(err)
	}
	if x.Cmp(y) < 0 {
		x, y = y, x
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sink256 = x.Sub(y)
	}
}

func BenchmarkMul256(b *testing.B) {
	x, err := Rand256(rng, max256)
	if err != nil {
		b.Fatal(err)
	}
	y, err := Rand256(rng, max256)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sink256 = x.Mul(y)
	}
}

func BenchmarkLsh256(b *testing.B) {
	x, err := Rand256(rng, max256)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sink256 = x.Lsh(uint(i))
	}
}

func BenchmarkRsh256(b *testing.B) {
	x, err := Rand256(rng, max256)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sink256 = x.Rsh(uint(i))
	}
}

func BenchmarkQuoRem256(b *testing.B) {
	x, err := Rand256(rng, max256)
	if err != nil {
		b.Fatal(err)
	}
	y, err := Rand256(rng, max128)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sink256, Sink256 = x.QuoRem(y)
	}
}

func BenchmarkQuoRem256Small(b *testing.B) {
	x, err := Rand256(rng, max256)
	if err != nil {
		b.Fatal(err)
	}
	y, err := Rand256(rng, U256(math.MaxUint64))
	if err != nil {
		b.Fatal(err)
	}
	if y.BitLen() == 0 {
		y = U256(1)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sink256, Sink256 = x.QuoRem(y)
	}
}

func BenchmarkAnd256(b *testing.B) {
	x, err := Rand256(rng, max256)
	if err != nil {
		b.Fatal(err)
	}
	y, err := Rand256(rng, max256)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sink256 = x.And(y)
	}
}

func BenchmarkXor256(b *testing.B) {
	x, err := Rand256(rng, max256)
	if err != nil {
		b.Fatal(err)
	}
	y, err := Rand256(rng, max256)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sink256 = x.Xor(y)
	}
}

func BenchmarkOr256(b *testing.B) {
	x, err := Rand256(rng, max256)
	if err != nil {
		b.Fatal(err)
	}
	y, err := Rand256(rng, max256)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sink256 = x.Or(y)
	}
}
