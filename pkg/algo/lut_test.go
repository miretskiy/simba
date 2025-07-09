package algo

import (
	"fmt"
	"testing"

	"github.com/miretskiy/simba/pkg/intrinsics"
)

// scalarAllBytesInSet is the reference implementation with no SIMD/FFI.
func scalarAllBytesInSet(data []byte, lut *[256]byte) bool {
	for _, b := range data {
		if (*lut)[b] == 0 {
			return false
		}
	}
	return true
}

// build a simple ASCII (0..127) table once.
var asciiLUT = func() *[256]byte {
	var t [256]byte
	for i := 0; i < 128; i++ {
		t[i] = 1
	}
	return &t
}()

func TestAllBytesInSet(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want bool
	}{
		{"empty", nil, true},
		{"ascii", []byte("Hello"), true},
		{"nonascii", []byte{0xC0, 'A'}, false},
	}
	for _, c := range cases {
		if got := AllBytesInSet(c.data, asciiLUT); got != c.want {
			t.Errorf("%s: want %v, got %v", c.name, c.want, got)
		}
	}
}

// Benchmark insight similar to IsASCII benchmarks.
func BenchmarkAllBytesInSet(b *testing.B) {
	sizes := []int{0, 1, 15, 32, 64, 128, 256, 1024, 4096, 1 << 16} // 64 KiB

	for _, sz := range sizes {
		buf := make([]byte, sz)
		for i := range buf {
			buf[i] = byte(i % 128) // always allowed
		}

		b.Run(fmt.Sprintf("size=%d/Algo", sz), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				AllBytesInSet(buf, asciiLUT)
			}
		})

		b.Run(fmt.Sprintf("size=%d/Intrinsics", sz), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				intrinsics.AllBytesInSet(buf, asciiLUT)
			}
		})

		b.Run(fmt.Sprintf("size=%d/Scalar", sz), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				scalarAllBytesInSet(buf, asciiLUT)
			}
		})
	}
}
