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

4. **FFI back-end selection and thresholds**

   * Implemented two interchangeable gateways: cgo (default when CGO is enabled) and purego (uses `purego.Dlopen` + `SyscallN`).
   * Added build tags `simba_cgo` and `simba_purego` to force a back-end; default logic picks cgo if available, otherwise purego.
   * Threshold constants extracted to build-tag-specific files: `threshold_cgo.go` (64 B) and `threshold_purego.go` (256 B), based on benchmarked crossover points.

5. **Benchmarks and gateway latency**

   * Introduced a Rust `noop` kernel and Go wrappers; micro-benchmarks give raw gateway cost.
   * Results on Apple M2 Max: cgo `Noop` ≈ 13.5 ns, purego ≈ 90 ns.
   * Re-ran `sum_u8` benchmarks (5 × 1 s each); scalar vs SIMD crossover occurs around 64 B for cgo and 256 B for purego.

6. **Build scripts and documentation updates**

   * `scripts/build.sh` now builds both static (`.a`) and shared (`.dylib/.so`) Rust libraries with correct `crate-type` flags.
   * `internal/ffi/simba.h` expanded with prototypes for `noop`, `is_ascii`, and `sum_u8`.
   * README instructions updated with build-tag usage examples and benchmark rationale.

7. **Tag validation SIMD kernel – findings and decision**

   * Implemented `validate_tag_inner` kernel that checks allowed characters and "no double underscore" rule using SIMD.
   * Benchmarks on Apple M2 Max (cgo):

     | Tag length | Scalar ns/op | SIMD ns/op |
     |-----------:|-------------:|-----------:|
     | 8          |    ~4.6      | ~35        |
     | 64         |    ~31       | ~68        |
     | 199        |    ~99       | ~156       |

     – Fixed FFI cost ≈ 30 ns (13–15 ns CGO gateway + 12–15 ns Rust prologue/epilogue) dominates for short tags; SIMD wins only beyond ~250 B – unrealistic for real-world tags.
   * Purego adds ~60 ns extra (allocations + SyscallN), pushing crossover even further.
   * Conclusion: the single-tag SIMD kernel is not useful. Future speed-ups require batching multiple tags per call or fusing more logic into one kernel.
   * The experimental tag validator code (`pkg/algo/tag.go`, tests, benchmarks) will be deleted after checkpoint commit; history retains the prototype.

## 2025-07-10 – Dual-lane SIMD refactor

* **Rust crate** now exposes lane-specific symbols (`*_32`, `*_64`) implemented with a generic `LaneCount<N>` helper.
* **Assembly stubs** duplicated for both `arm64` and `amd64`; legacy non-width stubs deleted.
* **Go FFI backend** (`syso_backend.go`) converted to *thin* wrappers – one Go function per Rust symbol (e.g. `SumU8_32`). No length heuristics.
* **Intrinsics layer** takes over width selection (≥64 B → 64-lane, else 32-lane). No scalar fallback.
* **Algo layer** provides scalar fallback; thresholds updated: ASCII = 32 B, other helpers = 16 B.
* Benchmarks (Apple M2 Max, syso FFI): SIMD overtakes scalar at 32 B for ASCII, at 16 B for `SumU8` / LUT ops.
* README updated with the new dual-lane table and threshold policy.


_This file is for informational purposes only and is not required for building the project._ 