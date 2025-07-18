package ffi
import (
    "bytes"
    "hash/crc32"
    "testing"
)
func TestCRC32UpdateMismatch(t *testing.T) {
    buf := bytes.Repeat([]byte{0xAB}, 1500)
    tbl := crc32.MakeTable(crc32.Castagnoli)
    want := crc32.Checksum(buf, tbl)
    got32 := Crc32Update32(buf, 0)
    got64 := Crc32Update64(buf, 0)
    if got32 != want {
        t.Fatalf("32 mismatch: want %08x got %08x", want, got32)
    }
    if got64 != want {
        t.Fatalf("64 mismatch: want %08x got %08x", want, got64)
    }
}
