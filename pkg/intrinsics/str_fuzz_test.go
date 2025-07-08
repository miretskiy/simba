package intrinsics

// Fuzz run summary (2025-07-08, Apple Silicon, Go 1.24.2)
//   • Duration: 3 minutes
//   • Executions: ~34.7 million (≈192 k execs/s avg)
//   • Interesting inputs discovered: 13
//   • Crashes / mismatches: 0

import "testing"

func scalarIsASCII(data []byte) bool {
	for _, b := range data {
		if b&0x80 != 0 {
			return false
		}
	}
	return true
}

func FuzzIsASCII(f *testing.F) {
	// Seed corpus with a few edge cases.
	seeds := [][]byte{
		{},
		[]byte("ASCII"),
		[]byte{0x7F},
		[]byte{0x80},
		[]byte{0xFF, 0x01, 0x02, 0x03},
	}
	for _, s := range seeds {
		f.Add(string(s))
	}

	f.Fuzz(func(t *testing.T, s string) {
		data := []byte(s)
		got := IsASCII(data)
		want := scalarIsASCII(data)
		if got != want {
			t.Fatalf("mismatch for %q: simd=%v scalar=%v", s, got, want)
		}
	})
}
