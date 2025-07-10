package algo

import "github.com/miretskiy/simba/pkg/intrinsics"

// IsASCII is a high-level helper that chooses between a fast scalar loop for
// tiny slices and the SIMD-accelerated intrinsic for larger inputs.
//
// The actual threshold is held in the build-tag–specific constant
// `simdThreshold` (see threshold_cgo.go).
//
//	– In cgo builds the fixed FFI cost is ≈13 ns on Apple M2 Max, making
//	  64 B a safe crossover.  On x86 the cost is a bit higher (~20 ns) but
//	  64 B is still conservative.
func IsASCII(data []byte) bool {
	if len(data) < simdThreshold {
		for _, b := range data {
			if b&0x80 != 0 {
				return false
			}
		}
		return true
	}
	return intrinsics.IsASCII(data)
}
