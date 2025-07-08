package intrinsics

import "github.com/miretskiy/simba/internal/ffi"

// IsASCII reports whether all bytes in data are standard 7-bit ASCII.
// An empty slice returns true.
func IsASCII(data []byte) bool {
	return ffi.IsASCII(data)
}
