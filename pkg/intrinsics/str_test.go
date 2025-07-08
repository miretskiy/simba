package intrinsics

import (
	"fmt"
	"testing"
)

func TestIsASCII(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"empty", []byte{}, true},
		{"ascii", []byte("Hello, World!"), true},
		{"non-ascii", []byte{0xC3, 0x28}, false}, // invalid UTF-8 0xC3 0x28
		{"extended", []byte{0x7F, 0x20}, true},
		{"highbit", []byte{0x80}, false},
	}

	for _, tt := range tests {
		if got := IsASCII(tt.data); got != tt.want {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.want, got)
		}
	}
}

// Benchmark insight (Apple M2 Max, Go 1.24, 5× 1 s runs)
//
// | Size | SIMD (ns/op) | Scalar (ns/op) | Speed-up |
// |------|--------------|----------------|----------|
// |   0  |   ~1.94      |    ~0.59       | 0.3×      |
// |   1  |   ~31.6      |    ~0.89       | 0.03×     |
// |  15  |   ~35.3      |    ~5.66       | 0.16×     |
// |  64  |   ~32.2      |    ~20.2       | 0.63×     |
// | 256  |   ~33.6      |    ~86.2       | 2.6×      |
// | 1 KiB|   ~36.7      |    ~323        | 8.8×      |
// | 4 KiB|   ~58.9      |   ~1 245       | 21×       |
// | 64 KiB|  ~505       |  ~19 800       | 39×       |
//
// The SIMD implementation has an ~30 ns fixed overhead (dominated by the
// cgo boundary) that amortizes after roughly 100 B.  We therefore use 128 B as
// the switch-over threshold in higher-level APIs.

// BenchmarkIsASCII measures the SIMD-backed implementation.
func BenchmarkIsASCII(b *testing.B) {
	sizes := []int{0, 1, 15, 64, 256, 1024, 4096, 1 << 16} // 64 KiB upper bound

	for _, sz := range sizes {
		// Prepare ASCII-only buffer. Using all-ASCII data forces the
		// implementation to scan the *entire* slice instead of short-
		// circuiting on the first high bit, providing a stable baseline.
		buf := make([]byte, sz)
		for i := range buf {
			buf[i] = byte(i % 128)
		}

		b.Run(fmt.Sprintf("size=%d/SIMD", sz), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				IsASCII(buf)
			}
		})

		b.Run(fmt.Sprintf("size=%d/Scalar", sz), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				scalarIsASCII(buf)
			}
		})
	}
}
