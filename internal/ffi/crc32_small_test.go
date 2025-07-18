package ffi
import (
    "bytes"
    "hash/crc32"
    "testing"
)
func TestCRC32Small(t *testing.T) {
    buf := bytes.Repeat([]byte{0xAB}, 50)
    tbl := crc32.MakeTable(crc32.Castagnoli)
    want := crc32.Checksum(buf, tbl)
    got := Crc32Update32(buf, 0)
    if got != want {
        t.Fatalf("small mismatch got %08x want %08x", got, want)
    }
}
