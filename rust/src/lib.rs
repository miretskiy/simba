#![feature(portable_simd)]
use core::simd::prelude::SimdUint;
use core::simd::Simd;

// === Portable SIMD byte-sum ===================================================

/// Returns the sum of all bytes in the buffer modulo 2^32.
///
/// The function accumulates into a 32-bit unsigned integer; if the arithmetic
/// overflows, it wraps around just like normal `u32` addition.
///
/// Internally this wrapper loops over 64-byte blocks, calling the low-level
/// `sum_u8_block` SIMD kernel for each block, and finishes the tail with a
/// scalar loop.  A single call keeps the cgo overhead low (~80–120 ns on Apple
/// Silicon) while each 64-byte kernel retires in roughly 2–3 ns – a 40×
/// difference that makes it worthwhile to batch the work here rather than from
/// the Go side.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn sum_u8(ptr: *const u8, len: usize) -> u32 {
    if ptr.is_null() || len == 0 {
        return 0;
    }

    let data = core::slice::from_raw_parts(ptr, len);
    sum_u8_impl(data)
}

/// Number of bytes processed per SIMD block.
const LANES: usize = 64;

/// Low-level kernel: sums exactly 64 bytes pointed to by `ptr`.
///
/// # Safety
/// * `ptr` must be valid for **at least 64 readable bytes**.
/// * The 64-byte region must be properly aligned for `u8` (any pointer on most
///   ABIs) and may not overlap with mutable data for the duration of the call.
///
/// This function is marked `unsafe` so that higher-level safe wrappers must
/// uphold these pre-conditions.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn sum_u8_block(ptr: *const u8) -> u32 {
    debug_assert!(!ptr.is_null());
    let slice = core::slice::from_raw_parts(ptr, LANES);
    let v = Simd::<u8, LANES>::from_slice(slice);
    let v32: Simd<u32, LANES> = v.cast();
    v32.reduce_sum()
}

fn sum_u8_impl(data: &[u8]) -> u32 {
    let mut total: u64 = 0;

    let mut chunks = data.chunks_exact(LANES);
    for chunk in &mut chunks {
        // SAFETY: `chunk` is exactly 64 bytes long by construction.
        total += unsafe { sum_u8_block(chunk.as_ptr()) } as u64;
    }

    for &b in chunks.remainder() {
        total += b as u64;
    }

    (total & 0xFFFF_FFFF) as u32
}

// -----------------------------------------------------------------------------

#[unsafe(no_mangle)]
pub unsafe extern "C" fn is_ascii(ptr: *const u8, len: usize) -> u8 {
    if ptr.is_null() || len == 0 {
        return 1; // empty slice is ASCII
    }
    let data = core::slice::from_raw_parts(ptr, len);
    is_ascii_impl(data) as u8
}

/// Low-level kernel: returns 1 if the 64-byte block is all ASCII, 0 otherwise.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn is_ascii_block(ptr: *const u8) -> u8 {
    debug_assert!(!ptr.is_null());
    let slice = core::slice::from_raw_parts(ptr, LANES);
    let v = Simd::<u8, LANES>::from_slice(slice);
    let max_val: u8 = v.reduce_max();
    (max_val < 0x80) as u8
}

fn is_ascii_impl(data: &[u8]) -> bool {
    let mut chunks = data.chunks_exact(LANES);
    for chunk in &mut chunks {
        // SAFETY: chunk is exactly 64 bytes.
        if unsafe { is_ascii_block(chunk.as_ptr()) } == 0 {
            return false;
        }
    }
    for &b in chunks.remainder() {
        if b & 0x80 != 0 {
            return false;
        }
    }
    true
}

#[cfg(test)]
mod tests {
    #[test]
    fn test_sum_u8() {
        let data: Vec<u8> = (0u8..=255u8).collect();
        let expected: u32 = data.iter().map(|&b| b as u32).sum();
        let result = unsafe { super::sum_u8(data.as_ptr(), data.len()) };
        assert_eq!(result, expected);
    }

    #[test]
    fn test_sum_u8_wrap() {
        // Number of 0xFF bytes needed to overflow u32.
        const LEN: usize = 16_843_010; // ceil(2^32 / 255)
        let data = vec![0xFFu8; LEN];
        // Expected result is (255 * LEN) mod 2^32.
        let expected = ((255u64 * LEN as u64) & 0xFFFF_FFFF) as u32;
        let result = unsafe { super::sum_u8(data.as_ptr(), data.len()) };
        assert_eq!(result, expected);
    }

    #[test]
    fn test_sum_u8_various_lengths() {
        // Stress a variety of lengths to make sure chunk/remainder logic works.
        const LANES: usize = super::LANES;
        let mut lengths = vec![
            0usize,
            1,
            LANES - 1,
            LANES,
            LANES + 1,
            2 * LANES - 3,
            2 * LANES,
            4096,
            10_000,
        ];
        lengths.push(123_456); // an arbitrary large size

        for len in lengths {
            // Create data pattern 0,1,2,3,... wrapping every 256.
            let data: Vec<u8> = (0..len as u32).map(|i| (i % 256) as u8).collect();

            let expected: u32 = data.iter().map(|&b| b as u32).sum();
            let got = unsafe { super::sum_u8(data.as_ptr(), data.len()) };
            assert_eq!(got, expected, "failed at len={}", len);
        }
    }

    #[test]
    fn test_is_ascii() {
        let ascii = b"Hello, world!";
        let non_ascii = [0x48u8, 0x80u8, 0x49u8];

        unsafe {
            assert_eq!(super::is_ascii(ascii.as_ptr(), ascii.len()), 1);
            assert_eq!(super::is_ascii(non_ascii.as_ptr(), non_ascii.len()), 0);
            assert_eq!(super::is_ascii(core::ptr::null(), 0), 1);
        }
    }
}
