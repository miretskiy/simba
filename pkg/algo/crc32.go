package algo

import (
	"hash/crc32"

	"github.com/miretskiy/simba/internal/ffi"
	"github.com/miretskiy/simba/pkg/intrinsics"
)

// Threshold at which the SIMD FFI path overtakes Go's built-in crc32.  Recent
// July-2025 benchmarks on an Apple M2 Max show that SIMD starts winning
// around 1 KiB (break-even ≈ 800 B, still ~9 % slower at 512 B).  Keep the
// threshold at a clean 1 KiB to stay conservative and to avoid a perf cliff
// on other CPUs.
const crc32Threshold = 1024

// Castagnoli table (CRC-32C / iSCSI polynomial) — this is the only
// polynomial supported by the Rust SIMD backend, so everything in the Simba
// CRC layer is hard-wired to it.
var castagnoliTable = crc32.MakeTable(crc32.Castagnoli)

// CRC32 returns the Castagnoli CRC32C of data by default.
// For short buffers we keep the tiny, table-driven Go implementation because
// the call overhead of the SIMD FFI path outweighs its benefit.  For ≥256 B
// we jump directly to the 32/64-lane kernels exposed by the intrinsics
// package.
func CRC32(data []byte) uint32 {
	if len(data) < crc32Threshold {
		return crc32.Checksum(data, castagnoliTable)
	}
	return intrinsics.Crc32Update(data, 0)
}

// Update extends an existing CRC32C (Castagnoli) value with additional data.
// For long buffers (>256 B) it routes through SIMD kernels; otherwise it
// falls back to Go's scalar routine.
func CRC32Update(data []byte, init uint32) uint32 {
	if len(data) < crc32Threshold {
		return crc32.Update(init, castagnoliTable, data)
	}
	return intrinsics.Crc32Update(data, init)
}

// Combine concatenates two CRC32C digests. For other tables use
// github.com/DataDog/dd-go/pkg/crc32combine or similar reference code.
func CRC32Combine(crc1, crc2 uint32, len2 int) uint32 {
	return ffi.Crc32Combine(crc1, crc2, len2)
}
