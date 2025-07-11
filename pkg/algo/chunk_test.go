package algo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEachHelpers(t *testing.T) {
	data := make([]byte, 100)
	var c64, c32, c16 int
	tail := Each64(data, func(chunk []byte) { c64++ })
	tail = Each32(tail, func(chunk []byte) { c32++ })
	tail = Each16(tail, func(chunk []byte) { c16++ })

	require.Equal(t, 1, c64, "64-chunk count")
	require.Equal(t, 1, c32, "32-chunk count")
	require.Equal(t, 0, c16, "16-chunk count")
	require.Equal(t, 4, len(tail), "tail length")
}
