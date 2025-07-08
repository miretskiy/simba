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
	"github.com/miretskiy/simba/pkg/algo"
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

// ValidateTagASCII is the accelerated version that optionally uses SIMD to
// reject non-ASCII bytes in bulk when the tag length is 128 bytes or more.
func ValidateTagASCII(tag string) bool {
	n := len(tag)
	// Fast ASCII rejection for long tags; scalar loop is cheaper for short ones.
	if n >= 128 && !algo.IsASCII([]byte(tag)) {
		return false
	}
	return validateTagASCIIScalar(tag)
}
