//go:build cgo || simba_cgo
// +build cgo simba_cgo

// Package ffi provides Go bindings to the Rust simba library.
//
//go:generate bash ../../scripts/build.sh
package ffi

/*
#cgo LDFLAGS: -L${SRCDIR} -lsimba
#include "simba.h"
*/
import "C"
import "unsafe"

// SumU8 returns the sum of all bytes in the slice.
func SumU8(data []byte) uint32 {
	if len(data) == 0 {
		return 0
	}
	return uint32(C.sum_u8((*C.uchar)(unsafe.Pointer(&data[0])), C.size_t(len(data))))
}

// IsASCII returns true if every byte in data is < 0x80.
func IsASCII(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	res := C.is_ascii((*C.uchar)(unsafe.Pointer(&data[0])), C.size_t(len(data)))
	return res != 0
}

// Noop calls an empty Rust function to measure pure FFI overhead.
func Noop() {
	C.noop()
}

// AllBytesInSet returns true if every byte in data indexes a non-zero entry
// in the provided 256-byte lookup table.
func AllBytesInSet(data []byte, lut *[256]byte) bool {
	if len(data) == 0 {
		return true
	}
	res := C.validate_u8_lut(
		(*C.uchar)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
		(*C.uchar)(unsafe.Pointer(&(*lut)[0])),
	)
	return res != 0
}

// MapBytes maps each byte of src through lut and stores the result in dst.
// It copies exactly len(src) bytes. Panic if dst is shorter than src, mirroring
// the safety contract of the built-in copy function.
func MapBytes(dst, src []byte, lut *[256]byte) {
	if len(src) == 0 {
		return
	}
	if len(dst) < len(src) {
		panic("ffi: MapBytes dst slice too short")
	}
	C.map_u8_lut(
		(*C.uchar)(unsafe.Pointer(&src[0])),
		C.size_t(len(src)),
		(*C.uchar)(unsafe.Pointer(&dst[0])),
		(*C.uchar)(unsafe.Pointer(&(*lut)[0])),
	)
}

// ValidateTagInner checks that all bytes are allowed tag chars and no double
// underscores appear. It assumes the first and last byte have already been
// validated by the caller.
func ValidateTagInner(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	res := C.validate_tag_inner(
		(*C.uchar)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
	)
	return res != 0
}
