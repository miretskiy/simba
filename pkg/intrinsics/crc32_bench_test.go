package intrinsics

import (
	"crypto/rand"
	"fmt"
	"hash/crc32"
	"testing"

	"github.com/miretskiy/simba/internal/ffi"
)

// --- generic Go Combine implementation (gf(2) matrix) ---

type gf2Mat [32]uint32

func (m *gf2Mat) mul(vec uint32) (sum uint32) {
	for i := 0; vec != 0; i, vec = i+4, vec>>4 {
		sum ^= (m[i] * (vec & 1)) ^
			(m[i+1] * ((vec >> 1) & 1)) ^
			(m[i+2] * ((vec >> 2) & 1)) ^
			(m[i+3] * ((vec >> 3) & 1))
	}
	return sum
}

func (m *gf2Mat) pow2(lhs *gf2Mat) {
	for n, val := range lhs {
		m[n] = lhs.mul(val)
	}
}

var (
	initOdd = gf2Mat{
		0, 0, 0, 0, 1 << 0, 1 << 1, 1 << 2, 1 << 3,
		1 << 4, 1 << 5, 1 << 6, 1 << 7, 1 << 8, 1 << 9, 1 << 10, 1 << 11,
		1 << 12, 1 << 13, 1 << 14, 1 << 15, 1 << 16, 1 << 17, 1 << 18, 1 << 19,
		1 << 20, 1 << 21, 1 << 22, 1 << 23, 1 << 24, 1 << 25, 1 << 26, 1 << 27,
	}
	initEven = gf2Mat{
		0, 0, 1 << 0, 1 << 1, 1 << 2, 1 << 3, 1 << 4, 1 << 5,
		1 << 6, 1 << 7, 1 << 8, 1 << 9, 1 << 10, 1 << 11, 1 << 12, 1 << 13,
		1 << 14, 1 << 15, 1 << 16, 1 << 17, 1 << 18, 1 << 19, 1 << 20, 1 << 21,
		1 << 22, 1 << 23, 1 << 24, 1 << 25, 1 << 26, 1 << 27, 1 << 28, 1 << 29,
	}
)

func combineGeneric(poly *crc32.Table, crc1, crc2 uint32, length int) uint32 {
	if length <= 0 {
		return crc1
	}

	var (
		odd  = initOdd
		even = initEven
	)

	// set polynomial seeds from table (Castagnoli only in our benchmarks)
	odd[0], odd[1], odd[2], odd[3] = poly[1<<4], poly[1<<5], poly[1<<6], poly[1<<7]
	even[0], even[1] = poly[1<<6], poly[1<<7]

	for pOdd, pEven := &odd, &even; length > 0; pOdd, pEven = pEven, pOdd {
		pEven.pow2(pOdd)
		if length&1 != 0 {
			crc1 = pEven.mul(crc1)
		}
		length >>= 1
	}

	return crc1 ^ crc2
}

// Prevent compiler from optimizing away benchmarked results.
var crc32Sink uint32

func benchSizes() []int { return []int{16, 64, 128, 256, 512, 1024, 4096, 16384, 65536} }

// BenchmarkCRC32 runs both the native Go hash/crc32 implementation and the SIMD
// intrinsics back-end across a range of input sizes. Sub-benchmarks are named
// using the pattern "impl=(native|simd)/<size>B" so that benchstat can compare
// columns with `-col /impl`.
func BenchmarkCRC32(b *testing.B) {
	sizes := benchSizes()
	// Allocate a buffer large enough for the biggest size and fill it once.
	data := make([]byte, sizes[len(sizes)-1])
	_, _ = rand.Read(data)

	nativeTbl := crc32.MakeTable(crc32.Castagnoli)

	for _, sz := range sizes {
		buf := data[:sz]

		// Native (scalar) implementation.
		b.Run(fmt.Sprintf("impl=native/%dB", sz), func(sb *testing.B) {
			for i := 0; i < sb.N; i++ {
				crc32Sink = crc32.Checksum(buf, nativeTbl)
			}
		})

		// SIMD intrinsics implementation.
		b.Run(fmt.Sprintf("impl=simd/%dB", sz), func(sb *testing.B) {
			for i := 0; i < sb.N; i++ {
				crc32Sink = Crc32Update(buf, 0)
			}
		})
	}
}

// BenchmarkSecureCRC32 emulates Datadog's SecureWrite pattern: compute the
// CRC32C of a chunk and immediately combine it into a running digest. We
// measure the native Go path (hash/crc32.Checksum + ffi.Crc32Combine) against
// the SIMD intrinsics plus the same FFI combine helper.
//
// Sub-benchmarks are named "impl=<native|simd>/<size>B" so that benchstat can
// compare columns with `-col /impl`.
func BenchmarkSecureCRC32(b *testing.B) {
	sizes := benchSizes()

	data := make([]byte, sizes[len(sizes)-1])
	_, _ = rand.Read(data)

	nativeTbl := crc32.MakeTable(crc32.Castagnoli)

	for _, sz := range sizes {
		buf := data[:sz]

		b.Run(fmt.Sprintf("impl=runtime/%dB", sz), func(sb *testing.B) {
			running := uint32(0)
			dummy := uint32(0)
			for i := 0; i < sb.N; i++ {
				sum := crc32.Update(0, nativeTbl, buf) // same as Checksum
				running = crc32.Update(running, nativeTbl, buf)
				dummy ^= sum
			}
			crc32Sink = running ^ dummy
		})

		b.Run(fmt.Sprintf("impl=native/%dB", sz), func(sb *testing.B) {
			running := uint32(0)
			for i := 0; i < sb.N; i++ {
				sum := crc32.Checksum(buf, nativeTbl)
				running = combineGeneric(nativeTbl, running, sum, len(buf))
			}
			crc32Sink = running
		})

		b.Run(fmt.Sprintf("impl=simd/%dB", sz), func(sb *testing.B) {
			running := uint32(0)
			for i := 0; i < sb.N; i++ {
				sum := Crc32Update(buf, 0)
				running = ffi.Crc32Combine(running, sum, len(buf))
			}
			crc32Sink = running
		})
	}
}
