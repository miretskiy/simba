package intrinsics

import "github.com/miretskiy/simba/internal/ffi"

// EqU8Masks64 compares each byte in `data` to `needle` using a 64-lane SIMD
// kernel and writes the resulting 64-bit masks into `out` – one mask word per
// **full** 64-byte chunk (no remainder handling).  `out` must have room for
// `len(data)/64` elements.  The function returns the number of bytes processed.
func EqU8Masks64(data []byte, needle byte, out []uint64) int {
	return ffi.EqU8Masks64(data, needle, out)
}

// EqU8Masks32 is the 32-lane variant (uint32 masks, one per 32-byte chunk; tail
// bytes `len(data)%32` are ignored). Returns bytes processed.
func EqU8Masks32(data []byte, needle byte, out []uint32) int {
	return ffi.EqU8Masks32(data, needle, out)
}

// EqU8Masks16 is the 16-lane variant (uint16 masks, one per 16-byte chunk). It
// can also be used to mop up a tail left by a wider-lane call. Returns bytes processed.
func EqU8Masks16(data []byte, needle byte, out []uint16) int {
	return ffi.EqU8Masks16(data, needle, out)
}
