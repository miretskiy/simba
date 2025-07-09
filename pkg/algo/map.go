package algo

import "github.com/miretskiy/simba/pkg/intrinsics"

// ByteSet is a 256-entry lookup table used by AllBytesInSet/MapBytes. A value
// of 0 means the byte is disallowed (or maps to zero), any non-zero value is
// treated as allowed / mapped.  Defined as a type alias so *ByteSet can be
// passed to functions that expect *[256]byte without conversion.
type ByteSet = [256]byte

// MakeByteSet builds a ByteSet with the given bytes set to 1.  Typical usage
// at package init time:
//
//	var digits = algo.MakeByteSet('0','1',…)
func MakeByteSet(bytes ...byte) *ByteSet {
	var bs ByteSet
	for _, b := range bytes {
		bs[b] = 1
	}
	return &bs
}

// simdMapThreshold mirrors the crossover observed for AllBytesInSet.
const simdMapThreshold = 64

// MapBytes maps bytes from src into dst via lut, like
//
//	dst[i] = lut[src[i]]
//
// and returns the number of bytes written, matching the semantics of the
// built-in copy.  It processes up to min(len(src), len(dst)) bytes and never
// panics on length mismatch.
func MapBytes(dst, src []byte, lut *ByteSet) int {
	n := len(src)
	if len(dst) < n {
		n = len(dst)
	}
	if n == 0 {
		return 0
	}

	// Fast path for tiny slices.
	if n < simdMapThreshold {
		for i := 0; i < n; i++ {
			dst[i] = (*lut)[src[i]]
		}
		return n
	}

	intrinsics.MapBytes(dst[:n], src[:n], (*[256]byte)(lut))
	return n
}
