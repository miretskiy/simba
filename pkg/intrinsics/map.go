package intrinsics

import "github.com/miretskiy/simba/internal/ffi"

// MapBytes applies the LUT to src and writes into dst via SIMD. intrinsics do
// not implement a scalar path.
func MapBytes(dst, src []byte, lut *[256]byte) {
	switch n := len(src); {
	case n == 0:
		return
	case len(dst) < n:
		panic("intrinsics: MapBytes dst slice too short")
	case n >= 64:
		ffi.MapBytes64(dst, src, lut)
	case n >= 32:
		ffi.MapBytes32(dst, src, lut)
	default:
		ffi.MapBytes16(dst, src, lut)
	}
}
