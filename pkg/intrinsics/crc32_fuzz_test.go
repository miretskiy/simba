package intrinsics

import (
	"hash/crc32"
	"testing"
)

// Property: for all byte slices A, B
//
//	CRC(A∥B) computed via streaming Update equals the value reconstructed
//	algebraically with combineGeneric (CRC(A) + CRC(B) + len(B)).
//
// combineGeneric is defined in crc32_bench_test.go and is available to all
// tests in this package.
func FuzzCRC32CombineVsUpdate(f *testing.F) {
	tbl := crc32.MakeTable(crc32.Castagnoli)

	// deterministic seeds
	f.Add([]byte(""), []byte(""))
	f.Add([]byte("foo"), []byte("bar"))
	f.Add(make([]byte, 2048), make([]byte, 7))

	f.Fuzz(func(t *testing.T, a, b []byte) {
		crcA := crc32.Update(0, tbl, a)
		crcB := crc32.Update(0, tbl, b)

		streaming := crc32.Update(crcA, tbl, b)
		combined := combineGeneric(tbl, crcA, crcB, len(b))

		if streaming != combined {
			t.Fatalf("mismatch update=%08x combine=%08x", streaming, combined)
		}
	})
}
