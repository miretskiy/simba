//go:build simba_purego || (!cgo && !simba_cgo)
// +build simba_purego !cgo,!simba_cgo

package ffi

import (
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

var (
	sumAddr   uintptr
	asciiAddr uintptr
)

func init() {
	libPath := selectLib()
	lib, err := purego.Dlopen(libPath, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		panic("ffi: dlopen failed: " + err.Error())
	}

	sumAddr, err = purego.Dlsym(lib, "sum_u8")
	if err != nil {
		panic(err)
	}
	asciiAddr, err = purego.Dlsym(lib, "is_ascii")
	if err != nil {
		panic(err)
	}
}

func selectLib() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(libDir(), "libsimba.dylib")
	case "linux":
		return filepath.Join(libDir(), "libsimba.so")
	default:
		panic("ffi: unsupported OS " + runtime.GOOS)
	}
}

// libDir returns directory of this package binary (placeholder: current working dir).
func libDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

func SumU8(data []byte) uint32 {
	if len(data) == 0 {
		return 0
	}
	r1, _, _ := purego.SyscallN(sumAddr,
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
	)
	return uint32(r1)
}

// IsASCII returns true if every byte in data is < 0x80.
func IsASCII(data []byte) bool {
	if len(data) == 0 {
		return true
	}

	r1, _, _ := purego.SyscallN(asciiAddr,
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
	)

	return r1 != 0
}
