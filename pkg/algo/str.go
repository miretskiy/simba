package algo

import "github.com/miretskiy/simba/pkg/intrinsics"

// IsASCII is a high-level helper that chooses between a fast scalar loop for
// tiny slices and the SIMD-accelerated intrinsic for larger inputs. The
// 128-byte threshold is derived from microbenchmarks (see str_test.go) showing
// the ~30 ns cgo overhead amortizes after ≈100 B.
func IsASCII(data []byte) bool {
	const simdThreshold = 128 // bytes – tweak per platform
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
