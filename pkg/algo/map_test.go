package algo

import "testing"

// build lower-case LUT once
var lowerLUT = func() *[256]byte {
	var t [256]byte
	for i := 0; i < 256; i++ {
		t[i] = byte(i)
	}
	for c := byte('A'); c <= 'Z'; c++ {
		t[c] = c + ('a' - 'A')
	}
	return &t
}()

func TestMapBytesLowercase(t *testing.T) {
	src := []byte("HeLLo_World123")
	dst := make([]byte, len(src))
	n := MapBytes(dst, src, lowerLUT)
	if n != len(src) {
		t.Fatalf("expected %d bytes written, got %d", len(src), n)
	}
	want := "hello_world123"
	if string(dst) != want {
		t.Fatalf("want %q, got %q", want, string(dst))
	}
}

func TestMapBytesShortDst(t *testing.T) {
	src := []byte("ABCDE")
	dst := make([]byte, 3)
	n := MapBytes(dst, src, lowerLUT)
	if n != 3 {
		t.Fatalf("expected 3 bytes written, got %d", n)
	}
	if string(dst) != "abc" {
		t.Fatalf("unexpected dst %q", string(dst))
	}
}
