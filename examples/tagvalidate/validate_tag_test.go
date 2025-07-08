package tagvalidate

import (
	"math/rand"
	"testing"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

func TestValidateTagASCII(t *testing.T) {
	cases := []struct {
		tag    string
		expect bool
	}{
		{"env:prod", true},
		{"service-api", true},
		{"bad__double", false},
		{"bad_trailing_", false},
		{"BadUpper", false},
		{string(make([]byte, 201)), false}, // too long
	}
	for _, c := range cases {
		got := ValidateTagASCII(c.tag)
		if got != c.expect {
			t.Errorf("ValidateTagASCII(%q)=%v, want %v", c.tag, got, c.expect)
		}
	}
}

var benchResult bool

func BenchmarkValidateTagASCII(b *testing.B) {
	rep := representativeCorpus

	run := func(b *testing.B, data []string, fn func(string) bool) {
		var r bool
		n := len(data)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			r = fn(data[i%n])
		}
		benchResult = r
	}

	b.Run("SIMD", func(b *testing.B) { run(b, rep, ValidateTagASCII) })
	b.Run("Scalar", func(b *testing.B) { run(b, rep, validateTagASCIIScalar) })
}

// representativeCorpus simulates production tag distribution: 32-byte tags,
// 11 % chance of uppercase letters and 1 % chance of a unicode rune.
var representativeCorpus = func() []string {
	const (
		corpusSize      = 1000
		tagLen          = 32
		upperProb       = 0.11
		unicodeProb     = 0.01
		asciiLowerStart = 'a'
		asciiLowerEnd   = 'z'
	)

	rnd := rand.New(rand.NewSource(42))
	makeTag := func() string {
		buf := make([]byte, 0, tagLen)
		for len(buf) < tagLen {
			// decide character category
			p := rnd.Float64()
			switch {
			case p < unicodeProb:
				// insert random 1–3 byte UTF-8 rune (avoid surrogate range)
				r := rune(rnd.Intn(utf8.MaxRune))
				if utf16.IsSurrogate(r) {
					r = '☃'
				}
				tmp := make([]byte, 4)
				n := utf8.EncodeRune(tmp, r)
				if len(buf)+n > tagLen {
					continue // skip, keep length exact
				}
				buf = append(buf, tmp[:n]...)
			case p < unicodeProb+upperProb:
				c := byte(rnd.Intn(asciiLowerEnd-asciiLowerStart+1)) + asciiLowerStart
				buf = append(buf, byte(unicode.ToUpper(rune(c))))
			default:
				c := byte(rnd.Intn(asciiLowerEnd-asciiLowerStart+1)) + asciiLowerStart
				buf = append(buf, c)
			}
		}
		return string(buf)
	}

	corpus := make([]string, corpusSize)
	for i := 0; i < corpusSize; i++ {
		corpus[i] = makeTag()
	}
	return corpus
}()
