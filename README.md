# 🦁 SIMBA - SIMD Binary Accelerator

**SIMBA** (SIMD Binary Accelerator) is a high-performance runtime and tooling layer that empowers **Go applications** to leverage **SIMD** (Single Instruction, Multiple Data) through **Rust intrinsics** and **CGo**.

Whether you're building data-intensive pipelines, number-crunching algorithms, or real-time systems, SIMBA lets your Go code *roar* with the speed of native vectorized instructions — without sacrificing code clarity or portability.

---

## 🚀 Why SIMBA?

- 🧠 **Simple interface** – Access powerful SIMD instructions from Go via intuitive wrappers.
- ⚙️ **Powered by Rust** – Leverages mature SIMD support in Rust for portability and safety.
- 🦾 **No assembly needed** – Write expressive Go code, while SIMBA handles the dirty bits.
- 🛠 **Tooling included** – Optional CLI tooling to build, inspect, and test SIMD-accelerated modules.
- 📦 **Modular** – Use SIMD intrinsics where you need them, and fall back to pure Go when you don’t.

---

## 🛠 How It Works

SIMBA compiles **Rust functions** that use platform-specific `std::arch` intrinsics and exposes them via a **C ABI**, which is then invoked from Go using **CGo**. This allows SIMD-capable hot paths to be written in Rust while keeping the rest of your app in idiomatic Go.

```
[ Go Code ] <-- CGo --> [ C ABI Shim ] <-- FFI --> [ Rust SIMD ]
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
/*
#cgo LDFLAGS: -L${SRCDIR}/target/release -lsimba
#include "simba.h"
*/
import "C"

func Sum(data []byte) uint32 {
    return uint32(C.sum_u8_avx2((*C.uchar)(unsafe.Pointer(&data[0])), C.size_t(len(data))))
}
```

To build without C tool-chain (pure Go, cross-compile friendly):

```bash
# Optionally control which FFI engine is used via Go build tags.

# By default the build picks the FFI backend automatically:
#   • If CGO is enabled -> cgo engine (lowest latency)
#   • If CGO is disabled -> purego engine (no external tool-chain)
#
# Override this behaviour with explicit tags:
#   -tags=simba_cgo      # force cgo bindings even when CGO_ENABLED=0
#   -tags=simba_purego   # force pure-Go (purego) bindings even when CGO is on

# examples
CGO_ENABLED=0 go test -tags=simba_cgo ./...      # cross-compile but still use cgo engine
go test -tags=simba_purego ./...                 # use purego even when CGO is enabled

```

---

## 🧩 Composing Intrinsics & Choosing Granularity

Calling a single SIMD kernel is cheap once the data size amortises the fixed FFI cost (≈ 30 ns via cgo, ≈ 90 ns via purego). The moment you **chain** two kernels back-to-back you pay that gateway latency twice, which can wipe out the win for small/medium slices.

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
