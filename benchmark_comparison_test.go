package simba

import (
	"testing"

	"github.com/miretskiy/simba/internal/ffi"
	"github.com/miretskiy/simba/internal/ffi_cgo"
)

// BenchmarkFFIComparison provides a comprehensive comparison between
// SIMBA's assembly trampoline approach and traditional CGO for
// measuring pure function call overhead.
func BenchmarkFFIComparison(b *testing.B) {
	b.Run("SIMBA", func(b *testing.B) {
		// SIMBA approach: Go → assembly trampoline → Rust
		for i := 0; i < b.N; i++ {
			ffi.Noop()
		}
	})

	b.Run("CGO", func(b *testing.B) {
		// Traditional CGO: Go → CGO runtime → C
		for i := 0; i < b.N; i++ {
			ffi_cgo.NoopCGO()
		}
	})
}
