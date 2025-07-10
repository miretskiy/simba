package ffi

import "testing"

// pbench runs a parallel benchmark helper to measure function latency.
func pbench(b *testing.B, f func()) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f()
		}
	})
}

// BenchmarkNoop measures the overhead of calling the empty Rust function
// through the active FFI backend (cgo).
func BenchmarkNoop(b *testing.B) {
	pbench(b, Noop)
}
