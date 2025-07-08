package intrinsics

import "github.com/miretskiy/simba/internal/ffi"

// SumU8 returns the sum of all bytes in data modulo 2^32 by delegating to the
// SIMD-accelerated Rust kernel via cgo.
func SumU8(data []byte) uint32 {
	return ffi.SumU8(data)
}
