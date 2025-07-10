package intrinsics

import "github.com/miretskiy/simba/internal/ffi"

// MapBytes applies the LUT to src and writes into dst via SIMD. intrinsics do
// not implement a scalar path.
func MapBytes(dst, src []byte, lut *[256]byte) {
	n := len(src)
	if n == 0 {
		return
	}
	if len(dst) < n {
		panic("intrinsics: MapBytes dst slice too short")
	}
	if n >= 64 {
		ffi.MapBytes64(dst, src, lut)
	} else {
		ffi.MapBytes32(dst, src, lut)
	}
}
