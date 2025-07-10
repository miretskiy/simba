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

// Benchmark insight (Apple M2 Max, Go 1.24, **syso trampoline – no CGO**, `go test -bench=IsASCII -count=10`)
//
// Representative results (Apple M2 Max, Go 1.24; median of 10 runs):
//
// | Size (B) | SIMD (ns/op) | Scalar (ns/op) | Winner |
// |----------|--------------|---------------|---------|
// | 0        | ~1.95        | ~0.60         | Scalar  |
// | 1        | ~3.30        | ~0.90         | Scalar  |
// | 15       | ~8.30        | ~5.80         | Scalar  |
// | 32       | ~3.34        | ~10.9         | SIMD 3× |
// | 63       | ~15.0        | ~20.4         | SIMD 1.4×|
// | 64       | ~3.32        | ~20.6         | SIMD 6× |
// | 127      | ~27.7        | ~47.6         | SIMD 1.7×|
// | 256      | ~5.46        | ~88.3         | SIMD 16×|
// | 319      | ~29.0        | ~111          | SIMD 3.8×|
// | 1 KiB    | ~10.2        | ~329          | SIMD 32×|
// | 4 KiB    | ~31.0        | ~1 260        | SIMD 41×|
// | 64 KiB   | ~469         | ~19 800       | SIMD 42×|
//
// The lightweight assembly shim adds just **~0.3 ns** per call. With 16/32/64-lane
// dispatch SIMD wins from 32 B upward and dominates as input grows.

// BenchmarkIsASCII measures the SIMD-backed implementation.
func BenchmarkIsASCII(b *testing.B) {
	sizes := []int{0, 1, 15, 32, 63, 64, 127, 256, 319, 1024, 1023, 4095, 4096, 1 << 16} // add half-lane & off-alignment cases

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
