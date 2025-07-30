package intrinsics

import "github.com/miretskiy/simba/internal/ffi"

// Crc32Update updates CRC32 checksum with additional data using SIMD kernels.
// The function never falls back to scalar – call the algo layer if you want
// automatic fallback for short buffers.
func Crc32Update(data []byte, init uint32) uint32 {
	switch n := len(data); {
	case n == 0:
		return init
	case n >= 64:
		return ffi.Crc32Update64(data, init)
	default:
		return ffi.Crc32Update32(data, init)
	}
}
