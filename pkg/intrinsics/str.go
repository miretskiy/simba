package intrinsics

import "github.com/miretskiy/simba/internal/ffi"

// IsASCII reports whether all bytes in data are 7-bit ASCII. intrinsics always
// use SIMD; scalar fallback lives in the algo layer.
func IsASCII(data []byte) bool {
	n := len(data)
	if n == 0 {
		return true
	}
	if n >= 64 {
		return ffi.IsASCII64(data)
	}
	return ffi.IsASCII32(data)
}
