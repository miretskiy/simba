package ffi

import "testing"

// BenchmarkNoop measures the per-call overhead of crossing the pure-Go FFI
// trampoline (Go → tiny asm stub → Rust no-op).  It runs single-threaded on
// the calling goroutine so the number reported is unaffected by scheduler
// context switches.
func BenchmarkNoop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Noop()
	}
}
