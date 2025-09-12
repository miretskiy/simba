// Package ffi_cgo provides CGO-based implementations for performance comparison
package ffi_cgo

/*
void noop_cgo();
*/
import "C"

//go:noinline
func NoopCGO() {
	C.noop_cgo()
}
