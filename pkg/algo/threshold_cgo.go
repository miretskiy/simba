//go:build cgo
// +build cgo

package algo

// simdThreshold is the slice length (in bytes) above which the Rust/SIMD
// kernel beats a hand-written scalar loop *in cgo builds*.
//
// 2025-07-08 Apple M2 Max measurements:
//   - Pure cgo gateway latency (noop call):           ~13 ns
//   - Scalar byte-add loop throughput:                ~0.6 ns/B
//   - Break-even size  = 13 ns / 0.6 ns ≈ 22 B
//
// We round the crossover up to one cache-line (32 B) and add more head-room
// for older Intel/Zen cores where scalar ~0.8–1 ns/B and the gateway is
// ≈20–25 ns.  The result is a conservative yet tighter threshold of 64 B,
// down from the previous 128 B.
const simdThreshold = 64
