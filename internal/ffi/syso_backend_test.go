package ffi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// build ASCII lookup table once
var asciiLUT = func() *[256]byte {
	var t [256]byte
	for i := 0; i < 128; i++ {
		t[i] = 1
	}
	return &t
}()

func TestSumU8(t *testing.T) {
	cases := []struct {
		name string
		fn   func([]byte) uint32
	}{
		{"16", SumU8_16},
		{"32", SumU8_32},
		{"64", SumU8_64},
	}

	for _, c := range cases {
		require.Equal(t, uint32(0), c.fn(nil), "SumU8_%s(nil)", c.name)
		require.Equal(t, uint32(6), c.fn([]byte{1, 2, 3}), "SumU8_%s([1 2 3])", c.name)
	}
}

func TestIsASCII(t *testing.T) {
	cases := []struct {
		name string
		fn   func([]byte) bool
	}{
		{"16", IsASCII16},
		{"32", IsASCII32},
		{"64", IsASCII64},
	}

	for _, c := range cases {
		require.True(t, c.fn([]byte("abc")), "IsASCII%s positive", c.name)
		require.False(t, c.fn([]byte{0xFF}), "IsASCII%s negative", c.name)
	}
}

func TestAllBytesInSet(t *testing.T) {
	cases := []struct {
		name string
		fn   func([]byte, *[256]byte) bool
	}{
		{"16", AllBytesInSet16},
		{"32", AllBytesInSet32},
		{"64", AllBytesInSet64},
	}

	for _, c := range cases {
		require.True(t, c.fn([]byte("ABC"), asciiLUT), "AllBytesInSet%s positive", c.name)
		require.False(t, c.fn([]byte{0xC0}, asciiLUT), "AllBytesInSet%s negative", c.name)
	}
}

func TestMapBytes(t *testing.T) {
	// Build a simple +1 translation table.
	var mapTbl [256]byte
	for i := 0; i < 256; i++ {
		mapTbl[i] = byte((i + 1) & 0xFF)
	}

	cases := []struct {
		name string
		fn   func(dst, src []byte, lut *[256]byte)
	}{
		{"16", MapBytes16},
		{"32", MapBytes32},
		{"64", MapBytes64},
	}

	src := []byte{0x00, 0x7F, 0xFF}
	want := []byte{0x01, 0x80, 0x00}

	for _, c := range cases {
		dst := make([]byte, len(src))
		c.fn(dst, src, &mapTbl)
		require.Equal(t, want, dst, "MapBytes%s result", c.name)
	}
}

func TestNoop(t *testing.T) {
	// Simply make sure it links and can be called.
	Noop()
}

func TestEqU8Masks(t *testing.T) {
	// Build source buffers with a few '_' bytes so masks are non-zero.
	src16 := make([]byte, 16)
	for i := range src16 {
		src16[i] = 'a'
	}
	src16[0] = '_'
	src16[3] = '_'

	src32 := make([]byte, 32)
	copy(src32, append(src16, src16...))

	src64 := make([]byte, 64)
	copy(src64, append(src32, src32...))

	// Expected bitmasks
	want16 := uint16((1 << 0) | (1 << 3))
	want32 := uint32(want16) | uint32(want16)<<16
	want64 := uint64(want32) | uint64(want32)<<32

	out16 := make([]uint16, 1)
	out32 := make([]uint32, 1) // 32/32 =1 mask word
	out64 := make([]uint64, 1)

	require.Equal(t, 16, EqU8Masks16(src16, '_', out16))
	require.Equal(t, want16, out16[0])

	require.Equal(t, 32, EqU8Masks32(src32, '_', out32))
	require.Equal(t, want32, out32[0])

	require.Equal(t, 64, EqU8Masks64(src64, '_', out64))
	require.Equal(t, want64, out64[0])
}
