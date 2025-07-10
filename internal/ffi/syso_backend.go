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
//go:generate go run ../../scripts/gen_trampolines
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
//
// Each Go prototype below is paired with a **trampoline** implemented in
// architecture-specific assembly (`syso_*.s`).  To keep the two in sync we
// tag every prototype with:
//
//   //simba:trampoline <arches>
//
// The `gen_trampolines` generator (invoked via `go:generate` above) scans the
// package, finds these tags, and auto-writes the minimal `MOV / CALL / RET`
// stubs for the listed architectures.  Adding a new FFI symbol now requires
// only the Go prototype plus this comment—no hand-edited assembly.
//

//simba:trampoline amd64 arm64
//go:noescape
func sum_u8_32_raw(ptr *byte, n uintptr) uint32

//simba:trampoline amd64 arm64
//go:noescape
func sum_u8_64_raw(ptr *byte, n uintptr) uint32

//simba:trampoline amd64 arm64
//go:noescape
func is_ascii32_raw(ptr *byte, n uintptr) uint8

//simba:trampoline amd64 arm64
//go:noescape
func is_ascii64_raw(ptr *byte, n uintptr) uint8

//simba:trampoline amd64 arm64
//go:noescape
func validate_u8_lut32_raw(ptr *byte, n uintptr, lut *byte) uint8

//simba:trampoline amd64 arm64
//go:noescape
func validate_u8_lut64_raw(ptr *byte, n uintptr, lut *byte) uint8

//simba:trampoline amd64 arm64
//go:noescape
func map_u8_lut32_raw(src *byte, n uintptr, dst *byte, lut *byte)

//simba:trampoline amd64 arm64
//go:noescape
func map_u8_lut64_raw(src *byte, n uintptr, dst *byte, lut *byte)

//simba:trampoline amd64 arm64
//go:noescape
func noop_raw()
