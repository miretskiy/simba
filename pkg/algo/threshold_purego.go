//go:build !cgo
// +build !cgo

package algo

// simdThreshold is the slice length (bytes) above which the purego FFI version
// (using SyscallN) outperforms the scalar loop.  Current fixed cost is ≈90 ns
// due to two runtime escape-tracking allocations, which break even around
// 150 B (90 ns / 0.6 ns per byte).  We round up to the next cache-aligned
// power-of-two: 256 B.
const simdThreshold = 256
