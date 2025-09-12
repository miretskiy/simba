package simba

import (
	"runtime"
	"testing"

	"github.com/miretskiy/simba/internal/ffi"
)

// TestMemorySafety validates that SIMBA properly handles memory ownership
// across FFI boundaries by ensuring slices remain alive during foreign calls.
func TestMemorySafety(t *testing.T) {
	// Test with a reasonably sized buffer to ensure the Rust code
	// has time to access the memory
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Force garbage collection before the FFI call to maximize the chance
	// of detecting use-after-free if memory safety is broken
	runtime.GC()
	runtime.GC() // Call twice to be thorough

	// This should work correctly with our runtime.KeepAlive calls
	sum1 := ffi.SumU8_32(data)
	sum2 := ffi.SumU8_64(data)

	// Verify the results are consistent and reasonable
	if sum1 != sum2 {
		t.Errorf("SumU8_32 and SumU8_64 produced different results: %d vs %d", sum1, sum2)
	}

	// Expected sum: 0+1+2+...+255+0+1+2+...+255+... (4 full cycles)
	expectedSum := uint32(4 * (255 * 256 / 2)) // 4 * 32640 = 130560
	if sum1 != expectedSum {
		t.Errorf("SumU8 result %d doesn't match expected %d", sum1, expectedSum)
	}

	// Test other functions to ensure they all have proper memory safety
	isASCII := ffi.IsASCII32([]byte("Hello, World!"))
	if !isASCII {
		t.Error("ASCII test failed on valid ASCII string")
	}

	// Test LUT functions
	lut := &[256]byte{}
	for i := range lut {
		lut[i] = 1 // Mark all bytes as valid
	}

	runtime.GC()
	runtime.GC()

	valid := ffi.AllBytesInSet32(data, lut)
	if !valid {
		t.Error("AllBytesInSet32 should return true when all bytes are marked valid in LUT")
	}

	// Test mapping functions
	dst := make([]byte, len(data))
	for i := range lut {
		lut[i] = byte(255 - i) // Invert mapping
	}

	runtime.GC()
	runtime.GC()

	ffi.MapBytes32(dst, data, lut)

	// Verify the mapping worked
	for i, b := range data {
		expected := byte(255 - int(b))
		if dst[i] != expected {
			t.Errorf("MapBytes32 failed at index %d: got %d, expected %d", i, dst[i], expected)
			break
		}
	}

	t.Logf("Memory safety test completed successfully - all FFI calls preserved slice data correctly")
}
