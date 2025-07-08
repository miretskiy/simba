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
