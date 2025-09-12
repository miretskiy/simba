package ffi_cgo

import "testing"

// BenchmarkNoopCGO measures the per-call overhead of crossing CGO
// (Go → CGO runtime → C no-op).  This provides a baseline comparison
// against the SIMBA approach.
func BenchmarkNoopCGO(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NoopCGO()
	}
}
