package intrinsics

import "github.com/miretskiy/simba/internal/ffi"

// SumU8 adds all bytes modulo 2^32.  intrinsics always delegate to SIMD; they
// never fall back to scalar—that choice is made at the algo layer.
func SumU8(data []byte) uint32 {
	switch n := len(data); {
	case n == 0:
		return 0
	case n >= 64:
		return ffi.SumU8_64(data)
	case n >= 32:
		return ffi.SumU8_32(data)
	default:
		return ffi.SumU8_16(data)
	}
}
