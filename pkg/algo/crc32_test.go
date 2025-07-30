package algo

import (
	"bytes"
	"crypto/rand"
	"hash/crc32"
	"testing"
)

func randomBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return b
}

func TestCRC32Castagnoli(t *testing.T) {
	data := randomBytes(1024)
	want := crc32.Checksum(data, crc32.MakeTable(crc32.Castagnoli))
	got := CRC32(data)
	if got != want {
		t.Fatalf("CRC32 mismatch: want %x got %x", want, got)
	}
}

func TestCRC32UpdateAndCombine(t *testing.T) {
	buf1 := bytes.Repeat([]byte{0xAB}, 1500)
	buf2 := bytes.Repeat([]byte{0xCD}, 4096)

	// Reference scalar implementation
	tbl := crc32.MakeTable(crc32.Castagnoli)
	crc1 := crc32.Checksum(buf1, tbl)
	crc2 := crc32.Checksum(buf2, tbl)
	expectedConcat := crc32.Checksum(append(buf1, buf2...), tbl)

	// Simba update path
	got1 := CRC32Update(buf1, 0)
	if got1 != crc1 {
		t.Fatalf("CRC32Update mismatch first part: %x vs %x", got1, crc1)
	}

	got2 := CRC32Update(buf2, got1)
	if got2 != expectedConcat {
		t.Fatalf("CRC32Update sequential mismatch: %x vs %x", got2, expectedConcat)
	}

	// Combine
	combined := CRC32Combine(crc1, crc2, len(buf2))
	if combined != expectedConcat {
		t.Fatalf("CRC32Combine mismatch: %x vs %x", combined, expectedConcat)
	}
}

func TestCRC32GoldenVectors(t *testing.T) {
	vectors := []struct {
		in   string
		want uint32
	}{
		{"hello", 0x9a71bb4c},       // CRC32C("hello")
		{"hello world", 0xc99465aa}, // CRC32C("hello world")
	}

	for _, v := range vectors {
		got := CRC32([]byte(v.in))
		if got != v.want {
			t.Errorf("CRC32(%q) = %08x, want %08x", v.in, got, v.want)
		}
	}
}

func TestCRC32UpdateAndCombineGolden(t *testing.T) {
	part1 := []byte("hello")
	part2 := []byte(" world") // note leading space

	crc1 := CRC32(part1)
	crc2 := CRC32(part2)

	// One-shot hash of the concatenation is our ground truth.
	concat := append(part1, part2...)
	want := CRC32(concat)

	// Sequential update path.
	gotSeq := CRC32Update(part2, crc1)
	if gotSeq != want {
		t.Fatalf("CRC32Update result %08x want %08x", gotSeq, want)
	}

	// Combine path.
	gotComb := CRC32Combine(crc1, crc2, len(part2))
	if gotComb != want {
		t.Fatalf("CRC32Combine result %08x want %08x", gotComb, want)
	}
}
