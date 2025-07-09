package algo

import "github.com/miretskiy/simba/pkg/intrinsics"

// On Apple M2 Max the FFI/SIMD path overtakes the scalar loop around 64 B
// (cgo gateway ~30–35 ns). Adjust via benchmarks on other CPUs.
const simdLUTThreshold = 64

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
