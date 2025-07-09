package algo

import (
	"unsafe"

	"github.com/miretskiy/simba/internal/ffi"
)

const maxTagLength = 200

// Scalar lookup tables (bool → no pointer indirection) -----------------------
var scalarStart = [256]bool{
	'a': true, 'b': true, 'c': true, 'd': true, 'e': true, 'f': true, 'g': true, 'h': true,
	'i': true, 'j': true, 'k': true, 'l': true, 'm': true, 'n': true, 'o': true, 'p': true,
	'q': true, 'r': true, 's': true, 't': true, 'u': true, 'v': true, 'w': true, 'x': true,
	'y': true, 'z': true, ':': true,
}

var scalarMid = [256]bool{
	'a': true, 'b': true, 'c': true, 'd': true, 'e': true, 'f': true, 'g': true, 'h': true,
	'i': true, 'j': true, 'k': true, 'l': true, 'm': true, 'n': true, 'o': true, 'p': true,
	'q': true, 'r': true, 's': true, 't': true, 'u': true, 'v': true, 'w': true, 'x': true,
	'y': true, 'z': true, '0': true, '1': true, '2': true, '3': true, '4': true, '5': true,
	'6': true, '7': true, '8': true, '9': true, ':': true, '.': true, '/': true, '-': true,
	'_': true,
}

const simdTagThreshold = 64 // reuse same crossover logic

// ValidateTagASCII validates Datadog-style tag (ASCII subset). Returns true if
// the tag meets all constraints.
func ValidateTagASCII(tag string) bool {
	l := len(tag)
	if l == 0 || l > maxTagLength {
		return false
	}
	// first byte
	b0 := tag[0]
	if !scalarStart[b0] {
		return false
	}
	if l == 1 {
		return true
	}
	// last byte cannot be '_'
	if tag[l-1] == '_' {
		return false
	}

	mid := tag[1 : l-1]
	if len(mid) >= simdTagThreshold {
		// Zero-allocation: convert string slice to []byte without copying.
		midBytes := unsafe.Slice(unsafe.StringData(mid), len(mid))
		if !ffi.ValidateTagInner(midBytes) {
			return false
		}
		// also need to check last mid byte for tagCharSet? validateTagInner already covers
		return true
	}
	// scalar path – borrow tight loop from original validator.
	bytes := unsafe.Slice(unsafe.StringData(tag), l)
	for i := 1; i < l-1; i++ {
		c := bytes[i]
		if !scalarMid[c] || (c == '_' && bytes[i-1] == '_') {
			return false
		}
	}
	// We already checked last byte != '_' earlier; ensure it is in allowed set.
	if !scalarMid[bytes[l-1]] {
		return false
	}
	return true
}

// validateTagScalar runs the scalar validation path unconditionally (no SIMD).
// It exists for benchmarking purposes.
func validateTagScalar(tag string) bool {
	l := len(tag)
	if l == 0 || l > maxTagLength {
		return false
	}
	if !scalarStart[tag[0]] {
		return false
	}
	if l == 1 {
		return true
	}
	if tag[l-1] == '_' {
		return false
	}
	bytes := unsafe.Slice(unsafe.StringData(tag), l)
	for i := 1; i < l-1; i++ {
		c := bytes[i]
		if !scalarMid[c] || (c == '_' && bytes[i-1] == '_') {
			return false
		}
	}
	if !scalarMid[bytes[l-1]] {
		return false
	}
	return true
}

// validateTagSIMD runs the SIMD kernel unconditionally for the inner bytes,
// performing scalar checks only for the first/last byte. Used for benchmarks.
func validateTagSIMD(tag string) bool {
	l := len(tag)
	if l == 0 || l > maxTagLength {
		return false
	}
	if !scalarStart[tag[0]] {
		return false
	}
	if l == 1 {
		return true
	}
	if tag[l-1] == '_' {
		return false
	}
	mid := tag[1 : l-1]
	midBytes := unsafe.Slice(unsafe.StringData(mid), len(mid))
	if !ffi.ValidateTagInner(midBytes) {
		return false
	}
	if !scalarMid[tag[l-1]] {
		return false
	}
	return true
}
