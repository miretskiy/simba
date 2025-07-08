package algo

import "github.com/miretskiy/simba/pkg/intrinsics"

// SumU8 adds all bytes in the slice modulo 2^32.
//
// This high-level helper mirrors intrinsics.SumU8 but first checks if the input
// is small enough that a scalar loop is faster than paying the ~30 ns cgo
// overhead (see str_test.go benchmark table).  On Apple M-series silicon an
// ~128-byte threshold is optimal; tailor as needed per platform.
func SumU8(data []byte) uint32 {
	if len(data) < simdThreshold {
		var acc uint32
		for _, b := range data {
			acc += uint32(b)
		}
		return acc
	}
	return intrinsics.SumU8(data)
}
