package intrinsics

import "github.com/miretskiy/simba/internal/ffi"

// AllBytesInSet reports whether every byte in data exists in the provided LUT.
// intrinsics layer always uses SIMD; scalar path is in algo.
func AllBytesInSet(data []byte, lut *[256]byte) bool {
	switch n := len(data); {
	case n == 0:
		return true
	case n >= 64:
		return ffi.AllBytesInSet64(data, lut)
	case n >= 32:
		return ffi.AllBytesInSet32(data, lut)
	default:
		return ffi.AllBytesInSet16(data, lut)
	}
}
