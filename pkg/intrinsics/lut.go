package intrinsics

import "github.com/miretskiy/simba/internal/ffi"

// AllBytesInSet reports whether every byte in data exists in the provided LUT.
// intrinsics layer always uses SIMD; scalar path is in algo.
func AllBytesInSet(data []byte, lut *[256]byte) bool {
	n := len(data)
	if n == 0 {
		return true
	}
	if n >= 64 {
		return ffi.AllBytesInSet64(data, lut)
	}
	return ffi.AllBytesInSet32(data, lut)
}
