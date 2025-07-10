package intrinsics

import "github.com/miretskiy/simba/internal/ffi"

// SumU8 adds all bytes modulo 2^32.  intrinsics always delegate to SIMD; they
// never fall back to scalar—that choice is made at the algo layer.
func SumU8(data []byte) uint32 {
	n := len(data)
	if n == 0 {
		return 0
	}
	if n >= 64 {
		return ffi.SumU8_64(data)
	}
	return ffi.SumU8_32(data)
}
