package algo

import "testing"

func TestAlgoIsASCII(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want bool
	}{
		{"empty", nil, true},
		{"ascii", []byte("Hello"), true},
		{"nonascii", []byte{0xFF, 'A'}, false},
	}
	for _, c := range cases {
		if got := IsASCII(c.data); got != c.want {
			t.Errorf("%s: want %v, got %v", c.name, c.want, got)
		}
	}
}
