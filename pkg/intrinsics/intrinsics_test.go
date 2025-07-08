package intrinsics

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/miretskiy/simba/internal/ffi"
	"github.com/stretchr/testify/require"
)

var sink uint32 // used to prevent compiler optimizations in benchmarks

func TestSumU8(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(r.Intn(256))
	}

	var expected uint32
	for _, b := range data {
		expected += uint32(b)
	}

	require.Equal(t, expected, SumU8(data))
}

func TestSumU8Wrap(t *testing.T) {
	const n = 16_843_010 // ceil(2^32 / 255)
	data := bytes.Repeat([]byte{0xFF}, n)
	want := uint32((255 * n) & 0xFFFFFFFF)
	require.Equal(t, want, SumU8(data))
}

// scalarSum is a simple Go loop used as the pure-Go baseline in benchmarks.
func scalarSum(data []byte) uint32 {
	var acc uint32
	for _, b := range data {
		acc += uint32(b)
	}
	return acc
}

func BenchmarkSumU8(b *testing.B) {
	sizes := []int{0, 64, 128, 256, 1024, 8192, 65536}
	for _, n := range sizes {
		// Generate input once per size.
		r := rand.New(rand.NewSource(42))
		data := make([]byte, n)
		for i := range data {
			data[i] = byte(r.Intn(256))
		}

		b.Run(fmt.Sprintf("Scalar_%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = scalarSum(data)
			}
		})

		b.Run(fmt.Sprintf("SIMD_%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = SumU8(data)
			}
		})

		b.Run(fmt.Sprintf("FFI_%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = ffi.SumU8(data)
			}
		})
	}
}
