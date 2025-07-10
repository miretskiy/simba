package algo

import "github.com/miretskiy/simba/pkg/intrinsics"

// asciiThreshold is tuned specifically for IsASCII. Benchmarks show that the
// SIMD kernel overtakes the scalar loop once the slice length reaches 32
// bytes on Apple M-series CPUs and is expected to behave similarly on AWS
// Graviton.  Using 32 B strikes a good balance between minimum latency for
// tiny strings and maximum throughput on typical inputs.
const asciiThreshold = 32

// IsASCII is a high-level helper that chooses between a fast scalar loop for
// tiny slices and the SIMD-accelerated intrinsic for larger inputs.
//
// Benchmarks on Apple M2 Max with the syso trampoline show SIMD overtakes the
// scalar loop at around 64 bytes; for smaller inputs the scalar path is
// cheaper despite the ~0.3 ns FFI cost.
func IsASCII(data []byte) bool {
	if len(data) < asciiThreshold {
		for _, b := range data {
			if b&0x80 != 0 {
				return false
			}
		}
		return true
	}
	return intrinsics.IsASCII(data)
}
