//go:build syso_experiment
// +build syso_experiment

package ffi

import "testing"

func TestAddU32(t *testing.T) {
	const a, b = 12345, 67890
	got := AddU32(a, b)
	want := uint32(a + b)
	if got != want {
		t.Fatalf("AddU32(%d,%d)=%d, want %d", a, b, got, want)
	}
}
