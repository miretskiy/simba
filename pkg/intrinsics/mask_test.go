package intrinsics

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEqU8MasksIntrinsics(t *testing.T) {
	// 16-lane
	src16 := make([]byte, 16)
	for i := range src16 {
		src16[i] = 'a'
	}
	src16[2] = '_'
	out16 := make([]uint16, 1)
	n := EqU8Masks16(src16, '_', out16)
	require.Equal(t, 16, n, "bytes16")
	want16 := uint16(1 << 2)
	require.Equal(t, want16, out16[0], "mask16")

	// 32-lane
	src32 := make([]byte, 32)
	copy(src32, append(src16, src16...))
	out32 := make([]uint32, 1)
	EqU8Masks32(src32, '_', out32)
	want32 := uint32(want16) | uint32(want16)<<16
	require.Equal(t, want32, out32[0], "mask32")

	// 64-lane
	src64 := make([]byte, 64)
	copy(src64, append(src32, src32...))
	out64 := make([]uint64, 1)
	EqU8Masks64(src64, '_', out64)
	want64 := uint64(out32[0]) | uint64(out32[0])<<32
	require.Equal(t, want64, out64[0], "mask64")
}

// Ensure that bytes beyond the last full lane are not included in the mask and
// that the function returns the correct number of words.
func TestEqU8MasksRemainderIgnored(t *testing.T) {
	// 32-lane case: 40-byte buffer → 1 mask word, 8-byte tail ignored.
	buf40 := make([]byte, 40)
	// put needle '_' only in the tail so mask should be zero.
	for i := 32; i < 40; i++ {
		buf40[i] = '_'
	}
	out32 := make([]uint32, 2) // length > needed to verify we don't overrun
	n32 := EqU8Masks32(buf40, '_', out32)
	require.Equal(t, 32, n32, "bytes32 remainder")
	require.Equal(t, uint32(0), out32[0], "mask32 remainder")

	// 16-lane case: 18-byte buffer → 1 word, last 2 bytes ignored.
	buf18 := make([]byte, 18)
	buf18[16] = '_'
	buf18[17] = '_'
	out16 := make([]uint16, 2)
	n16 := EqU8Masks16(buf18, '_', out16)
	require.Equal(t, 16, n16, "bytes16 remainder")
	require.Equal(t, uint16(0), out16[0], "mask16 remainder")
}
