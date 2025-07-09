package algo

import (
	"fmt"
	"strings"
	"testing"
)

func benchmarkTag(b *testing.B, size int, fn func(string) bool) {
	tag := makeTag(size)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fn(tag)
	}
}

func makeTag(length int) string {
	if length < 1 {
		return ""
	}
	if length == 1 {
		return "a"
	}
	if length > maxTagLength {
		length = maxTagLength
	}
	midLen := length - 2
	return "a" + strings.Repeat("b", midLen) + "c"
}

func BenchmarkValidateTag(b *testing.B) {
	sizes := []int{1, 8, 15, 31, 63, 64, 100, 150, 199}
	for _, sz := range sizes {
		b.Run(fmt.Sprintf("Scalar/len=%d", sz), func(b *testing.B) {
			benchmarkTag(b, sz, validateTagScalar)
		})
		b.Run(fmt.Sprintf("SIMD/len=%d", sz), func(b *testing.B) {
			benchmarkTag(b, sz, validateTagSIMD)
		})
	}
}
