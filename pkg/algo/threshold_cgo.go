//go:build cgo
// +build cgo

package algo

// simdThreshold is the slice length (bytes) above which the SIMD/Rust
// kernels outperform a scalar Go loop *when cgo is enabled*.
// Micro-benchmarks (pkg/intrinsics, size=0 cases) put the fixed cgo gateway
// at ~30 ns.  The scalar loop processes ≈0.6 ns per byte, so 30 ns/0.6 ns ≈ 50 B
// is the analytical tie-point; 128 B is chosen to stay comfortably past that
// on slower Xeon or laptop cores.
const simdThreshold = 128
