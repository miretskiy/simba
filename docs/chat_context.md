# Chat Context (2025-07-08)

This document captures the main decisions and explanations exchanged during the Cursor session so the context is available even if the chat is closed.

## Key Actions

1. **Package comments added**
   * Added a proper package-level comment to `internal/ffi/ffi.go`.
   * Created `doc.go` at repo root with overview for the `simba` module.
   * Added `pkg/intrinsics/doc.go` with detailed design notes.

2. **Design rationale for lane widths**

   * Portable SIMD (`std::simd`) already compiles operations to the widest registers available on the host CPU, so separate 32-lane (256-bit) or 64-lane (512-bit) bindings would be redundant.
   * Hand-rolled AVX-512 variants add maintenance overhead, require runtime dispatch, and can reduce turbo frequency on unsupported hardware, so they are deferred until clear performance benefits are demonstrated.
   * Existing SSE/AVX intrinsics handle sub-word elements (e.g., `u8`, `u16`) adequately; “smaller lane” bindings offer no advantage.

   Reference: Rust portable-SIMD documentation — <https://doc.rust-lang.org/std/simd/index.html>

3. **Future reference**

   * Higher-level algorithms should reside in `pkg/algo`, which can provide scalar fallbacks for tiny slices.
   * The `go mod tidy` workflow is preferred for dependency updates.
    * Benchmarks belong in the regular `*_test.go` files alongside unit tests; avoid dedicated `*_bench_test.go` files. Fuzz tests, however, should live in separate `*_fuzz_test.go` files.

---

_This file is for informational purposes only and is not required for building the project._ 