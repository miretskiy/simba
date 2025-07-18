// Package ffi exposes the same public API as ffi.go but is implemented via a
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
//go:generate ../../scripts/build_syso.sh
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

// SumU8_16 adds all bytes using a 16-lane SIMD kernel.
func SumU8_16(data []byte) uint32 {
	if len(data) == 0 {
		return 0
	}
	return sum_u8_16_raw(&data[0], uintptr(len(data)))
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

// IsASCII16 returns true if every byte < 0x80 using the 16-lane kernel.
func IsASCII16(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	return is_ascii16_raw(&data[0], uintptr(len(data))) != 0
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

// AllBytesInSet16 validates data via LUT using 16 lanes.
func AllBytesInSet16(data []byte, lut *[256]byte) bool {
	if len(data) == 0 {
		return true
	}
	return validate_u8_lut16_raw(&data[0], uintptr(len(data)), &lut[0]) != 0
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

// MapBytes16 maps src through lut into dst using 16-lane kernel.
func MapBytes16(dst, src []byte, lut *[256]byte) {
	if len(src) == 0 {
		return
	}
	if len(dst) < len(src) {
		panic("ffi: MapBytes dst slice too short")
	}
	map_u8_lut16_raw(&src[0], uintptr(len(src)), &dst[0], &lut[0])
}

// EqU8Masks32 compares each byte in `data` to `needle` using a 32-lane SIMD
// kernel and stores one bitmask word per 32-byte chunk into `out`.  Each word
// has bit *i* set when byte *i* in the chunk equals `needle`.
//
// Only **whole** 32-byte chunks are processed; any tail `len(data)%32` bytes are
// ignored.  Callers must either pad the input, issue a second call with the
// 16-lane variant, or handle the remainder scalarly.
//
// `out` must hold at least `len(data)/32` elements.  The function returns the
// number of **bytes processed** (i.e. `len(data)/32 * 32`).
func EqU8Masks32(data []byte, needle byte, out []uint32) int {
	if len(data) == 0 {
		return 0
	}
	if len(out) < len(data)/32 {
		panic("ffi: EqU8Masks32 out slice too short")
	}
	chunks := len(data) / 32
	eq_u8_masks32_raw(&data[0], uintptr(len(data)), needle, &out[0])
	return chunks * 32
}

// EqU8Masks64 is identical to EqU8Masks32 but operates on 64-byte chunks and
// produces 64-bit masks.  Tail bytes `len(data)%64` are skipped. The return
// value is the number of bytes processed.
func EqU8Masks64(data []byte, needle byte, out []uint64) int {
	if len(data) == 0 {
		return 0
	}
	if len(out) < len(data)/64 {
		panic("ffi: EqU8Masks64 out slice too short")
	}
	chunks := len(data) / 64
	eq_u8_masks64_raw(&data[0], uintptr(len(data)), needle, &out[0])
	return chunks * 64
}

// EqU8Masks16 operates on 16-byte chunks producing 16-bit masks.  This helper
// is convenient for processing short buffers or cleaning up a remainder after
// a wider-lane call. Returns bytes processed.
func EqU8Masks16(data []byte, needle byte, out []uint16) int {
	if len(data) == 0 {
		return 0
	}
	if len(out) < len(data)/16 {
		panic("ffi: EqU8Masks16 out slice too short")
	}
	chunks := len(data) / 16
	eq_u8_masks16_raw(&data[0], uintptr(len(data)), needle, &out[0])
	return chunks * 16
}

//go:noinline
func Noop() {
	noop_raw()
}

// Crc32Update32 updates CRC32 using the 32-lane SIMD kernel.
func Crc32Update32(data []byte, init uint32) uint32 {
	if len(data) == 0 {
		return init
	}
	return crc32_update_32_raw(&data[0], uintptr(len(data)), init)
}

// Crc32Update64 updates CRC32 using the 64-lane SIMD kernel.
func Crc32Update64(data []byte, init uint32) uint32 {
	if len(data) == 0 {
		return init
	}
	return crc32_update_64_raw(&data[0], uintptr(len(data)), init)
}

// Crc32Combine returns CRC32 of concatenation of two buffers given their CRCs
// and the length of the second buffer.
func Crc32Combine(crc1, crc2 uint32, len2 int) uint32 {
	return crc32_combine_raw(crc1, crc2, uintptr(len2))
}

// Echo mirrors the rust Echo struct; used only in trampoline tests.
type Echo struct {
	Ptr     uintptr
	Len     uintptr
	V32     uint32
	V8      uint8
	_pad1   [3]byte // align
	V64     uint64
	F64Bits uint64
	F32Bits uint32
	_pad2   [4]byte
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
func sum_u8_16_raw(ptr *byte, n uintptr) uint32

//simba:trampoline amd64 arm64
//go:noescape
func is_ascii32_raw(ptr *byte, n uintptr) uint8

//simba:trampoline amd64 arm64
//go:noescape
func is_ascii64_raw(ptr *byte, n uintptr) uint8

//simba:trampoline amd64 arm64
//go:noescape
func is_ascii16_raw(ptr *byte, n uintptr) uint8

//simba:trampoline amd64 arm64
//go:noescape
func validate_u8_lut32_raw(ptr *byte, n uintptr, lut *byte) uint8

//simba:trampoline amd64 arm64
//go:noescape
func validate_u8_lut64_raw(ptr *byte, n uintptr, lut *byte) uint8

//simba:trampoline amd64 arm64
//go:noescape
func validate_u8_lut16_raw(ptr *byte, n uintptr, lut *byte) uint8

//simba:trampoline amd64 arm64
//go:noescape
func map_u8_lut32_raw(src *byte, n uintptr, dst *byte, lut *byte)

//simba:trampoline amd64 arm64
//go:noescape
func map_u8_lut64_raw(src *byte, n uintptr, dst *byte, lut *byte)

//simba:trampoline amd64 arm64
//go:noescape
func map_u8_lut16_raw(src *byte, n uintptr, dst *byte, lut *byte)

//simba:trampoline amd64 arm64
//go:noescape
func eq_u8_masks32_raw(src *byte, n uintptr, needle uint8, out *uint32) uintptr

//simba:trampoline amd64 arm64
//go:noescape
func eq_u8_masks64_raw(src *byte, n uintptr, needle uint8, out *uint64) uintptr

//simba:trampoline amd64 arm64
//go:noescape
func eq_u8_masks16_raw(src *byte, n uintptr, needle uint8, out *uint16) uintptr

//simba:trampoline amd64 arm64
//go:noescape
func noop_raw()

//simba:trampoline amd64 arm64
//go:noescape
func crc32_update_32_raw(ptr *byte, n uintptr, init uint32) uint32

//simba:trampoline amd64 arm64
//go:noescape
func crc32_update_64_raw(ptr *byte, n uintptr, init uint32) uint32

//simba:trampoline amd64 arm64
//go:noescape
func crc32_combine_raw(crc1 uint32, crc2 uint32, len2 uintptr) uint32

//simba:trampoline amd64 arm64
//go:noescape
func trampoline_sanity_raw(ptr *byte, n uintptr, val32 uint32, val8 uint8, val64 uint64, f64bits uint64, f32bits uint32) uintptr

// TrampolineSanityHash returns a 64-bit mix of the four arguments. Used by
// tests to verify that trampolines marshal arguments verbatim.
func TrampolineSanityHash(ptr *byte, length uintptr, v32 uint32, v8 uint8, v64 uint64, f64bits uint64, f32bits uint32) uintptr {
	return trampoline_sanity_raw(ptr, length, v32, v8, v64, f64bits, f32bits)
}

//simba:trampoline amd64 arm64
//go:noescape
func trampoline_echo_raw(ptr *byte, n uintptr, v32 uint32, v8 uint8, v64 uint64, f64bits uint64, f32bits uint32, out *Echo)

// TrampolineEcho calls the echo helper for debugging.
func TrampolineEcho(ptr *byte, length uintptr, v32 uint32, v8 uint8, v64 uint64, f64bits uint64, f32bits uint32) Echo {
	var e Echo
	trampoline_echo_raw(ptr, length, v32, v8, v64, f64bits, f32bits, &e)
	return e
}
