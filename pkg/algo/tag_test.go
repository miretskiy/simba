package algo

import (
	"strings"
	"testing"
)

func TestValidateTagASCII(t *testing.T) {
	good := []string{"a", "foo", "foo_bar", "a123", "abc:def", "abc/def-ghi", strings.Repeat("a", 100)}
	bad := []string{"", "_abc", "Abc", "abc__def", "abc_", "abc💩"}
	for _, s := range good {
		if !ValidateTagASCII(s) {
			t.Errorf("expected valid: %q", s)
		}
	}
	for _, s := range bad {
		if ValidateTagASCII(s) {
			t.Errorf("expected invalid: %q", s)
		}
	}
}
