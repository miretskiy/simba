package tagvalidate

// Example of building a high-level ASCII-tag validator atop the SIMD helpers
// in the simba library. The tag syntax and length limit follow Datadog’s
// guidelines – see https://docs.datadoghq.com/getting_started/tagging/ .
//
// Rules enforced (mirrors the scalar reference implementation):
//   1. Tag length must be 1..200 bytes.
//   2. First byte: lowercase a-z or ':'
//   3. Subsequent bytes: a-z, 0-9, ':', '.', '/', '-', '_' (but no double
//      underscore "__").
//   4. Last byte cannot be an underscore.
//
// The function accelerates the expensive "all bytes < 0x80?" check by calling
// algo.IsASCII once the tag length crosses the 128-byte SIMD threshold; for
// shorter tags the pure scalar path remains faster.

import (
	"github.com/miretskiy/simba/pkg/intrinsics"
)

const maxTagLength = 200

// lookup tables – one load per byte, zero branches --------------------------------
var validASCIIStartChar = [256]bool{
	'a': true, 'b': true, 'c': true, 'd': true, 'e': true, 'f': true, 'g': true, 'h': true,
	'i': true, 'j': true, 'k': true, 'l': true, 'm': true, 'n': true, 'o': true, 'p': true,
	'q': true, 'r': true, 's': true, 't': true, 'u': true, 'v': true, 'w': true, 'x': true,
	'y': true, 'z': true, ':': true,
}

var validASCIITagChar = [256]bool{
	'a': true, 'b': true, 'c': true, 'd': true, 'e': true, 'f': true, 'g': true, 'h': true,
	'i': true, 'j': true, 'k': true, 'l': true, 'm': true, 'n': true, 'o': true, 'p': true,
	'q': true, 'r': true, 's': true, 't': true, 'u': true, 'v': true, 'w': true, 'x': true,
	'y': true, 'z': true, '0': true, '1': true, '2': true, '3': true, '4': true, '5': true,
	'6': true, '7': true, '8': true, '9': true, ':': true, '.': true, '/': true, '-': true,
	'_': true,
}

// validateTagASCIIScalar mirrors the production-optimized loop used in
// Datadog’s tag validator. It is fully bounds-check–free and avoids per-iteration
// boolean state.
func validateTagASCIIScalar(tag string) bool {
	if len(tag) == 0 || len(tag) > maxTagLength {
		return false
	}

	if !validASCIIStartChar[tag[0]] {
		return false
	}
	if len(tag) == 1 {
		return true
	}

	// Loop from byte 1 to penultimate byte (l-2). The compiler eliminates
	// bounds checks because the slice length is constant inside the loop and
	// tag[i-1] is guaranteed valid (i starts at 1).
	l := len(tag)
	for i := 1; i < l-1; i++ {
		c := tag[i]
		if !validASCIITagChar[c] || (c == '_' && tag[i-1] == '_') {
			return false
		}
	}

	// Trailing underscore not allowed.
	last := tag[l-1]
	if last == '_' {
		return false
	}
	return validASCIITagChar[last]
}

// tagFlags is a 256-byte table where:
//
//	bit0 = byte allowed (rule #3)
//	bit1 = byte is '_'
//
// This lets one SIMD gather serve both the allowed-byte test and
// double-underscore detection.
var tagFlags = func() *[256]byte {
	var t [256]byte
	for i, ok := range validASCIITagChar {
		if ok {
			t[i] = 1 // allowed bit
		}
	}
	t['_'] = 3 // allowed + underscore
	return &t
}()

// fastMiddleValid returns true if all bytes in data[1:] are valid tag chars
// *and* there is no double underscore. It uses a single SIMD LUT gather per
// 32-byte chunk and a scalar tail.
func fastMiddleValid(data []byte) bool {
	if len(data) <= 1 {
		return true
	}

	// Skip first byte (already checked by start-char rule).
	body := data[1:]
	n := len(body)

	if n < 32 {
		// Scalar fall-back: validate allowed chars & no "__".
		prev := false
		for _, c := range body {
			if tagFlags[c]&1 == 0 { // disallowed
				return false
			}
			if c == '_' {
				if prev {
					return false
				}
				prev = true
			} else {
				prev = false
			}
		}
		return true
	}

	blocks := n / 32
	var flags [32]byte
	prevLastBit := false // underscore flag from previous block

	for i := 0; i < blocks; i++ {
		start := i * 32
		intrinsics.MapBytes(flags[:], body[start:start+32], tagFlags)

		var allowedOr byte
		var underscoreMask uint32
		for idx, f := range flags {
			allowedOr |= f & 1
			underscoreMask |= uint32(f>>1) << idx
		}
		if allowedOr == 0 { // found disallowed char
			return false
		}
		if prevLastBit && (underscoreMask&1) != 0 {
			return false // boundary "__"
		}
		if (underscoreMask & (underscoreMask << 1)) != 0 {
			return false // "__" within block
		}
		prevLastBit = (underscoreMask>>31)&1 == 1
	}

	// Scalar tail
	prev := prevLastBit
	for _, c := range body[blocks*32:] {
		if tagFlags[c]&1 == 0 {
			return false
		}
		if c == '_' {
			if prev {
				return false
			}
			prev = true
		} else {
			prev = false
		}
	}
	return true
}

// ValidateTagASCII accelerates three heavy checks for long tags:
//  1. All bytes are ASCII (algo.IsASCII → SIMD).
//  2. No double underscores "__" (equality-mask SIMD scan).
//  3. All bytes after the first are in the allowed set (algo.AllBytesInSet → SIMD).
//
// For shorter inputs the original scalar validator is fastest.
func ValidateTagASCII(tag string) bool {
	n := len(tag)
	if n == 0 || n > maxTagLength {
		return false
	}

	// Fast ASCII rejection for long tags; scalar loop is cheaper for very short ones.
	if n >= 64 && !intrinsics.IsASCII([]byte(tag)) {
		return false
	}

	// First-character rule is still scalar (single byte).
	if !validASCIIStartChar[tag[0]] {
		return false
	}

	if n == 1 {
		return true
	}

	// Trailing underscore rule (cheap scalar check).
	if tag[n-1] == '_' {
		return false
	}

	data := []byte(tag)

	// Combined SIMD validation for body when beneficial.
	if n >= 32 && !fastMiddleValid(data) {
		return false
	}

	// Fall back to scalar validator for remaining nuanced checks on short tags
	// or for edge cases the SIMD path didn’t cover (small length).
	return validateTagASCIIScalar(tag)
}
