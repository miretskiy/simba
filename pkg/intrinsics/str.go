package intrinsics

import "github.com/miretskiy/simba/internal/ffi"

// IsASCII reports whether all bytes in data are 7-bit ASCII. intrinsics always
// use SIMD; scalar fallback lives in the algo layer.
func IsASCII(data []byte) bool {
	switch n := len(data); {
	case n == 0:
		return true
	case n >= 64:
		return ffi.IsASCII64(data)
	case n >= 32:
		return ffi.IsASCII32(data)
	default:
		return ffi.IsASCII16(data)
	}
}
