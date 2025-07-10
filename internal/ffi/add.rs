//! Minimal Rust equivalent of add.c demonstrating that the `.syso + ABI shim`
//! trick works with Rust too.
//!
//! Build notes (also encoded in `add.go`’s go:generate):
//!   rustc -O --emit=obj -C relocation-model=pic -o add_amd64.o add.rs
//!   mv add_amd64.o add_amd64.syso
//!
//! The `#[no_mangle]` keeps the symbol name stable so Go & the asm stub can
//! reference `add_u32` directly. `extern "C"` selects the System-V ABI, which
//! matches Go’s internal ABI on amd64.

#[no_mangle]
pub extern "C" fn add_u32(a: u32, b: u32) -> u32 {
    a + b
}
