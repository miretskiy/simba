//go:build simba_syso && !cgo
// +build simba_syso,!cgo

// This file exposes the same public API as ffi.go but is implemented via a
// static, position-independent object (`libsimba.syso`) linked directly by the
// Go tool-chain.  The foreign code is the Rust SIMD crate in ../../rust.
//
// Build/update the .syso with:
//
//	 go generate ./internal/ffi
//
//	ld -r -o libsimba.syso ../../target/x86_64-unknown-linux-gnu/release/deps/*.o"
//
// (The `ld -r` step merges all crate objects into a single relocatable blob
// that the Go linker will pick up automatically.)
//
//go:generate bash -c "cargo rustc --manifest-path ../../rust/Cargo.toml --lib --release --target x86_64-unknown-linux-gnu -- -C relocation-model=pic && \
package ffi

// SumU8 returns the sum of all bytes in the slice.
func SumU8(data []byte) uint32 {
	if len(data) == 0 {
		return 0
	}
	return sum_u8_raw(&data[0], uintptr(len(data)))
}

// IsASCII returns true if every byte in data < 0x80.
func IsASCII(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	return is_ascii_raw(&data[0], uintptr(len(data))) != 0
}

// AllBytesInSet validates using a 256-byte LUT.
func AllBytesInSet(data []byte, lut *[256]byte) bool {
	if len(data) == 0 {
		return true
	}
	return validate_u8_lut_raw(&data[0], uintptr(len(data)), &lut[0]) != 0
}

// MapBytes maps src through lut into dst.
func MapBytes(dst, src []byte, lut *[256]byte) {
	if len(src) == 0 {
		return
	}
	if len(dst) < len(src) {
		panic("ffi: MapBytes dst slice too short")
	}
	map_u8_lut_raw(&src[0], uintptr(len(src)), &dst[0], &lut[0])
}

// Noop measures bare FFI overhead.
func Noop() {
	noop_raw()
}

// --- raw syscall signatures implemented in assembly ---

func sum_u8_raw(ptr *byte, n uintptr) uint32
func is_ascii_raw(ptr *byte, n uintptr) uint8
func validate_u8_lut_raw(ptr *byte, n uintptr, lut *byte) uint8
func map_u8_lut_raw(src *byte, n uintptr, dst *byte, lut *byte)
func noop_raw()
