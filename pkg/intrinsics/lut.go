package intrinsics

import "github.com/miretskiy/simba/internal/ffi"

// AllBytesInSet reports whether every byte in data exists in the provided
// lookup table (256 entries, non-zero ⇒ allowed). An empty slice returns true.
func AllBytesInSet(data []byte, lut *[256]byte) bool {
	return ffi.AllBytesInSet(data, lut)
}
