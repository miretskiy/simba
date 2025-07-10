package algo

import "github.com/miretskiy/simba/pkg/intrinsics"

// With the lightweight syso trampoline the SIMD path wins once the slice is
// roughly 16 bytes or larger (~0.3 ns fixed cost).  Tune per-CPU if needed.
const simdLUTThreshold = 16

// AllBytesInSet returns true if every byte in data exists in the provided
// lookup table. For tiny slices it uses an inlined scalar loop; for longer
// inputs the SIMD-accelerated FFI path is used.
func AllBytesInSet(data []byte, lut *ByteSet) bool {
	if len(data) < simdLUTThreshold {
		for _, b := range data {
			if (*lut)[b] == 0 {
				return false
			}
		}
		return true
	}
	return intrinsics.AllBytesInSet(data, (*[256]byte)(lut))
}
