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
