package algo

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/miretskiy/simba/pkg/intrinsics"
	"github.com/stretchr/testify/require"
)

func TestSumU8Algo(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	data := make([]byte, 1000)
	for i := range data {
		data[i] = byte(r.Intn(256))
	}
	require.Equal(t, intrinsics.SumU8(data), SumU8(data))
}

var sinkAlgo uint32

func BenchmarkSumU8Algo(b *testing.B) {
	sizes := []int{64, 128, 256, 1024, 8192, 65536}
	for _, n := range sizes {
		r := rand.New(rand.NewSource(42))
		data := make([]byte, n)
		for i := range data {
			data[i] = byte(r.Intn(256))
		}

		b.Run(fmt.Sprintf("Algo_%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sinkAlgo = SumU8(data)
			}
		})
	}
}
