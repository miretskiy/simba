// Package intrinsics exposes thin Go wrappers around SIMD-accelerated Rust
// kernels linked via cgo. These functions operate on raw slice data and are
// intended as low-level building blocks; higher-level algorithms should reside
// in sibling packages (e.g., algo) that provide safer fallbacks and richer APIs.
//
// Design note: why we don’t expose 32-lane or 512-lane variants
// ------------------------------------------------------------
//
//  1. Portable SIMD already widens to the best registers available.
//     With `std::simd::Simd<T, N>` you specify only the *logical* lane count N.
//     The compiler chooses whatever hardware registers (128-, 256-, or 512-bit)
//     yield the highest throughput. As the Rust docs put it, operations
//     “compile to the best available SIMD instructions.”
//
//  2. Separate 32- or 64-lane bindings would call the same Rust kernel with a
//     different N and gain nothing; they would simply bloat the FFI surface and
//     test matrix.
//
//  3. Hand-rolled AVX-512 intrinsics are not yet worth the cost. AVX-512 is
//     rare, can down-clock the core, and requires runtime dispatch plus a
//     fallback path. Until benchmarks show clear wins on real hardware, the
//     complexity outweighs the upside.
//
//  4. Existing SSE/AVX intrinsics already handle sub-word elements (u8/u16) just
//     fine, so “smaller” lane bindings bring no benefit.
//
// In short, sticking to a single, portable implementation keeps the API small
// and yields robust performance across CPUs while leaving room to add wider
// variants later if evidence demands it.
package intrinsics
