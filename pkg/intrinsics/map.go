package intrinsics

import "github.com/miretskiy/simba/internal/ffi"

// MapBytes applies the lookup table to src and writes into dst. For slices
// shorter than the SIMD thresholds callers should prefer a scalar loop.
func MapBytes(dst, src []byte, lut *[256]byte) {
	ffi.MapBytes(dst, src, lut)
}
