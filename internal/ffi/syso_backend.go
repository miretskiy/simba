// This file exposes the same public API as ffi.go but is implemented via a
// static, position-independent object (`libsimba.syso`) linked directly by the
// Go tool-chain.  The foreign code is the Rust SIMD crate in ../../rust.
//
// Build/update the .syso archives with:
//
//	go generate ./internal/ffi
//
// The command below cross-compiles the Rust crate for macOS/amd64 (Mach-O)
// as a PIC static library, then renames the archive to `libsimba.syso` so
// the Go linker auto-links it.  Building for darwin/amd64 works reliably on
// both Apple Silicon and Intel hosts and is sufficient for the Go linker to
// consume regardless of the *build* GOOS/GOARCH—you can cross-link an ELF or
// Mach-O syso into any target binary.
//
//	target=x86_64-apple-darwin; \
//	toolchain=nightly; \
//	rustup target add $target --toolchain $toolchain >/dev/null 2>&1 || true; \
//	cargo +$toolchain rustc --manifest-path ../../rust/Cargo.toml --release --lib --target $target -- -C relocation-model=pic; \
//	cp ../../rust/target/$target/release/libsimba.a libsimba.syso"
//
//go:generate ./scripts/build_syso.sh
package ffi

// Width-specific thin wrappers around the raw assembly syscalls.  Higher-level
// packages decide which lane width to use based on slice length.

// SumU8_32 adds all bytes using a 32-lane SIMD kernel.
func SumU8_32(data []byte) uint32 {
	if len(data) == 0 {
		return 0
	}
	return sum_u8_32_raw(&data[0], uintptr(len(data)))
}

// SumU8_64 adds all bytes using a 64-lane SIMD kernel.
func SumU8_64(data []byte) uint32 {
	if len(data) == 0 {
		return 0
	}
	return sum_u8_64_raw(&data[0], uintptr(len(data)))
}

// IsASCII32 returns 1 if every byte in data < 0x80 using the 32-lane kernel.
func IsASCII32(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	return is_ascii32_raw(&data[0], uintptr(len(data))) != 0
}

// IsASCII64 returns 1 if every byte in data < 0x80 using the 64-lane kernel.
func IsASCII64(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	return is_ascii64_raw(&data[0], uintptr(len(data))) != 0
}

// AllBytesInSet32 validates data via LUT using 32 lanes.
func AllBytesInSet32(data []byte, lut *[256]byte) bool {
	if len(data) == 0 {
		return true
	}
	return validate_u8_lut32_raw(&data[0], uintptr(len(data)), &lut[0]) != 0
}

// AllBytesInSet64 validates data via LUT using 64 lanes.
func AllBytesInSet64(data []byte, lut *[256]byte) bool {
	if len(data) == 0 {
		return true
	}
	return validate_u8_lut64_raw(&data[0], uintptr(len(data)), &lut[0]) != 0
}

// MapBytes32 maps src through lut into dst using 32-lane kernel.
func MapBytes32(dst, src []byte, lut *[256]byte) {
	if len(src) == 0 {
		return
	}
	if len(dst) < len(src) {
		panic("ffi: MapBytes dst slice too short")
	}
	map_u8_lut32_raw(&src[0], uintptr(len(src)), &dst[0], &lut[0])
}

// MapBytes64 maps src through lut into dst using 64-lane kernel.
func MapBytes64(dst, src []byte, lut *[256]byte) {
	if len(src) == 0 {
		return
	}
	if len(dst) < len(src) {
		panic("ffi: MapBytes dst slice too short")
	}
	map_u8_lut64_raw(&src[0], uintptr(len(src)), &dst[0], &lut[0])
}

//go:noinline
func Noop() {
	noop_raw()
}

// --- raw syscall signatures implemented in assembly ---

//go:noescape
func sum_u8_32_raw(ptr *byte, n uintptr) uint32

//go:noescape
func sum_u8_64_raw(ptr *byte, n uintptr) uint32

//go:noescape
func is_ascii32_raw(ptr *byte, n uintptr) uint8

//go:noescape
func is_ascii64_raw(ptr *byte, n uintptr) uint8

//go:noescape
func validate_u8_lut32_raw(ptr *byte, n uintptr, lut *byte) uint8

//go:noescape
func validate_u8_lut64_raw(ptr *byte, n uintptr, lut *byte) uint8

//go:noescape
func map_u8_lut32_raw(src *byte, n uintptr, dst *byte, lut *byte)

//go:noescape
func map_u8_lut64_raw(src *byte, n uintptr, dst *byte, lut *byte)

//go:noescape
func noop_raw()
