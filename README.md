# 🦁 SIMBA - SIMD Binary Accelerator

**SIMBA** (SIMD Binary Accelerator) is a high-performance runtime and tooling layer that lets **Go binaries** call **Rust SIMD intrinsics** _without_ CGO.

Whether you're building data-intensive pipelines, number-crunching algorithms, or real-time systems, SIMBA lets your Go code *roar* with the speed of native vectorized instructions — without sacrificing code clarity or portability.

---

## 🚀 Why SIMBA?

- 🧠 **Simple interface** – Access powerful SIMD instructions from Go via intuitive wrappers.
- ⚙️ **Powered by Rust** – Leverages mature SIMD support in Rust for portability and safety.
- 🦾 **No CGO needed** – one tiny assembly shim per function, no external linker or `cc` tool-chain.
- 🛠 **Tooling included** – Optional CLI tooling to build, inspect, and test SIMD-accelerated modules.
- 📦 **Modular** – Use SIMD intrinsics where you need them, and fall back to pure Go when you don’t.

---

## 🛠 How It Works

SIMBA compiles **Rust functions** into a position-independent static
library, renames it to `*.syso`, and the Go linker treats it like a
native object.  A 3-instruction assembly **trampoline** (one per
function) bridges Go’s internal ABI to the System-V / AAPCS64 calling
convention – no cgo, no dynamic loader, ~2 ns overhead.

```
[ Go Code ] --asm shim--> [ .syso object ] --> [ Rust SIMD ]
```

---

## 📦 Getting Started

### 1. Add SIMBA to Your Project

```bash
git submodule add https://github.com/yourname/simba
```

### 2. Define a SIMD-accelerated Rust function

```rust
#[no_mangle]
pub extern "C" fn sum_u8_avx2(ptr: *const u8, len: usize) -> u32 {
    // Rust SIMD code using AVX2 intrinsics
}
```

### 3. Call it from Go

```go
package algo

//go:generate go run ./internal/ffi   // rebuilds *.syso archive

// SumU8 adds a slice of bytes via SIMD.
func SumU8(b []byte) uint32 {
    if len(b) == 0 {
        return 0
    }
    return ffi.SumU8(b) // ~2 ns call-return
}
```

No build tags needed – **SIMBA always builds with CGO disabled**.  The
`go generate ./internal/ffi` step produces two files:

* `libsimba_darwin_amd64.syso`
* `libsimba_darwin_arm64.syso`

They are auto-linked by the Go tool-chain on any platform.

---

## 🆕 Dual-Lane SIMD Kernels (32- vs 64-byte)

SIMBA now ships **two lane widths** for every byte-wise primitive:

| Operation            | 32-lane symbol | 64-lane symbol |
|----------------------|----------------|----------------|
| Sum of bytes         | `sum_u8_32`    | `sum_u8_64`    |
| ASCII check          | `is_ascii32`   | `is_ascii64`   |
| Validate via LUT     | `validate_u8_lut32` | `validate_u8_lut64` |
| Map via LUT          | `map_u8_lut32` | `map_u8_lut64` |

The **intrinsics layer** (`pkg/intrinsics`) automatically picks the _widest_
kernel that amortises its 0.3 ns FFI cost:

```go
// ≥64 B → 64-lane kernel, else 32-lane
if len(b) >= 64 {
    return ffi.SumU8_64(b)
}
return ffi.SumU8_32(b)
```

The **algo layer** adds a **scalar fallback** for tiny slices where pure Go
still beats SIMD.  Current thresholds (Apple M-series):

* Generic helpers (`SumU8`, LUT ops): **16 B**
* ASCII check: **32 B**

These cut-offs are recorded in `pkg/algo/threshold_*.go` and can be tuned per
platform – early experiments on AWS Graviton look similar.

---

## 🧩 Composing Intrinsics & Choosing Granularity

Calling a single SIMD kernel is cheap once the data size amortises the fixed FFI cost (~2 ns via the syso trampoline). The moment you **chain** two kernels back-to-back you pay that gateway latency twice, which can wipe out the win for small/medium slices.

Design options:

1. **Custom merged kernels (recommended)**
   Write the exact combination you need (e.g. *lower-case + validate*). Rust’s generics/macros make adding a new symbol trivial and the call still costs one hop.

2. **Batch API**
   Pass a tiny *op-code list* to one exported function so multiple operations run inside one call. Keeps Go in charge but needs a small “mini-VM” on the Rust side.

3. **Handle / pipeline builder**
   Build the op list once, get back an opaque handle (`u64`), then execute it many times. Saves parameter marshaling but adds lifetime management.

We currently expose **low-level primitives** (`validate_u8_lut`, `map_u8_lut`) that you can stitch together in Go for rapid prototyping. For production paths, prefer **option 1**—generate a bespoke kernel and export it. It scales linearly with the number of unique pipelines and keeps the public API intuitive.

*(Waiting for “native Go SIMD” isn’t part of the near-term plan; the proposal has been open for years and still lacks a stable design.)*

## 🔬 Use Cases

- High-performance parsing (e.g., JSON, CSV, binary protocols)
- Fast image or video preprocessing
- Bitwise vector math
- Custom hashing or compression
- Filtering, mapping, scanning of large datasets

---

## 📚 Resources

- [SIMD in Rust (std::arch)](https://doc.rust-lang.org/core/arch/)
- [CGO Documentation](https://golang.org/cmd/cgo/)
- [Why Rust for SIMD](https://blog.rust-lang.org/inside-rust/2021/09/08/simd-in-rust.html)

---

## 📣 Roadmap

- [ ] Platform-independent vector dispatch
- [ ] Optional fallback to Go implementation
- [ ] Generator for wrappers from Rust → Go
- [ ] CLI: `simba build`, `simba inspect`, `simba bench`
- [ ] Docs site with examples

---

## 🧑‍💻 Contributing

Contributions are welcome! If you have ideas for performance improvements, target architecture support, or want to help with the CLI, open an issue or pull request.

---

## 🦁 Philosophy

SIMBA’s goal is to **democratize low-level performance** for Go developers, without forcing them to write unsafe, unreadable code. You should be able to think in Go — and roar with SIMD.

---

## 📜 License

MIT or Apache 2.0, whichever you prefer.

## ✅ Testing & CI

SIMBA ships a **trampoline-sanity** unit-test that exercises the FFI layer with
seven mixed-width arguments (pointer, usize, 8-/32-/64-bit ints, raw
`float32`/`float64` bit-patterns).  On amd64 the last argument spills to the
stack; on arm64 all fit in registers.  The Rust side recomputes a simple FNV
hash and Go asserts equality, so any future stub width/offset error fails
instantly in CI.

Run just this guard:

```bash
go test ./internal/ffi -run TestTrampolineSanity
```

`go generate ./internal/ffi` regenerates the assembly stubs; the test must stay
green on both amd64 and arm64.
