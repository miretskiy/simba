package algo

import "github.com/miretskiy/simba/pkg/intrinsics"

// IsASCII is a high-level helper that chooses between a fast scalar loop for
// tiny slices and the SIMD-accelerated intrinsic for larger inputs.
//
// The actual threshold is held in the build-tag–specific constant
// `simdThreshold` (see threshold_cgo.go and threshold_purego.go).
//   - cgo build:   simdThreshold = 128  // cgo gateway ≈30 ns per call¹
//   - purego build: simdThreshold = 256 // purego SyscallN path ≈90 ns per call
//
// ¹ The ~30 ns figure comes from the intrinsics micro-benchmarks in
//
//	pkg/intrinsics/str_test.go (size=0 cases).  With purego the same
//	benchmark reports ~0.3 ns, so the crossover can be pushed down to one
//	cache-line (32 B).
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
