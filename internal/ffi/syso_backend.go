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
//go:generate bash -c "set -euo pipefail; toolchain=nightly; for target in x86_64-apple-darwin aarch64-apple-darwin; do rustup target add $target --toolchain $toolchain >/dev/null 2>&1 || true; cargo +$toolchain rustc --manifest-path ../../rust/Cargo.toml --release --lib --target $target -- -C relocation-model=pic; if [ \"$target\" = \"x86_64-apple-darwin\" ]; then goarch=amd64; else goarch=arm64; fi; cp ../../rust/target/$target/release/libsimba.a libsimba_darwin_${goarch}.syso; done"
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

//go:noescape
func sum_u8_raw(ptr *byte, n uintptr) uint32

//go:noescape
func is_ascii_raw(ptr *byte, n uintptr) uint8

//go:noescape
func validate_u8_lut_raw(ptr *byte, n uintptr, lut *byte) uint8

//go:noescape
func map_u8_lut_raw(src *byte, n uintptr, dst *byte, lut *byte)

//go:noescape
func noop_raw()
