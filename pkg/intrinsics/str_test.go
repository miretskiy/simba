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

// Benchmark insight (Apple M2 Max, Go 1.24, **syso trampoline – no CGO**, 10× 1 s)
//
// | Size | SIMD (ns/op) | Scalar (ns/op) | Winner |
// |------|--------------|----------------|--------|
// |   0  |   ~0.30      |    ~0.60       | SIMD 2× |
// |   1  |   ~2.12      |    ~0.90       | Scalar  |
// |  15  |   ~7.30      |    ~5.64       | Scalar  |
// |  32  |   ~14.1      |   ~10.6        | Scalar  |
// |  64  |   ~2.43      |   ~20.4        | SIMD 8× |
// | 256  |   ~4.11      |   ~86          | SIMD 21×|
// | 1 KiB|   ~9.44      |  ~326          | SIMD 34×|
// | 4 KiB|   ~30.3      | ~1 260         | SIMD 42×|
// | 64 KiB|  ~472       | ~20 000        | SIMD 42×|
//
// The lightweight assembly shim adds just **~0.3 ns** per call, yet the SIMD
// implementation for IsASCII only starts winning at **64 B** because the
// scalar loop is highly efficient on tiny inputs. Other intrinsics such as
// SumU8 beat scalar much earlier, so higher-level packages may choose a lower
// crossover, or expose algorithm-specific thresholds.

// BenchmarkIsASCII measures the SIMD-backed implementation.
func BenchmarkIsASCII(b *testing.B) {
	sizes := []int{0, 1, 15, 32, 64, 256, 1024, 4096, 1 << 16} // 64 KiB upper bound

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
