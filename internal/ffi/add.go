//go:build syso_experiment
// +build syso_experiment

package ffi

// NOTE: This file demonstrates the “.syso + ABI shim” technique, which
// lets Go call into a foreign object file without using cgo or purego.  The
// steps are:
//   1. `go generate` compiles add.c -> add_amd64.syso.
//   2. The Go assembler stub in add_amd64.s provides a Go‐ABI friendly
//      entry point `addU32` that simply forwards the arguments (already in
//      RDI/RSI) to the external symbol `add_u32`.
//   3. Go code can import this package and call AddU32 with ~2 ns overhead.
//
// Run `go generate` in this directory whenever you change add.c.
//go:generate bash -c "rustc -O --emit=obj -C relocation-model=pic -o add_amd64.o add.rs && mv add_amd64.o add_amd64.syso"

// AddU32 adds two uint32 numbers together by calling a symbol that lives in
// add_amd64.syso (compiled from C).  Because the call obeys the Go internal
// ABI, no cgo trampoline is required.
func AddU32(a, b uint32) uint32 // implemented in assembly
