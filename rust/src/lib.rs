//! Rust SIMD kernels for Simba FFI layer
#![feature(portable_simd)]
#![allow(unsafe_op_in_unsafe_fn)] // calls to unsafe APIs are audited and wrapped inside unsafe fns
use core::simd::prelude::SimdUint;
use core::simd::{LaneCount, Simd, SupportedLaneCount};

// === Portable SIMD byte-sum ===================================================

// no module-level LANES constant; tests use explicit 64.

// ---- Generic helpers --------------------------------------------------------

#[inline(always)]
unsafe fn sum_u8_impl<const LANES_N: usize>(data: &[u8]) -> u32
where
    LaneCount<LANES_N>: SupportedLaneCount,
{
    let mut total: u64 = 0;

    let mut chunks = data.chunks_exact(LANES_N);
    for chunk in &mut chunks {
        let v = Simd::<u8, LANES_N>::from_slice(chunk);
        let v32: Simd<u32, LANES_N> = v.cast();
        total += v32.reduce_sum() as u64;
    }
    for &b in chunks.remainder() {
        total += b as u64;
    }
    (total & 0xFFFF_FFFF) as u32
}

/* ─── 32-lane and 64-lane public exports ─────────────────────────────────── */

#[unsafe(no_mangle)]
pub unsafe extern "C" fn sum_u8_32(ptr: *const u8, len: usize) -> u32 {
    if ptr.is_null() || len == 0 {
        return 0;
    }
    let data = core::slice::from_raw_parts(ptr, len);
    sum_u8_impl::<32>(data)
}

#[unsafe(no_mangle)]
pub unsafe extern "C" fn sum_u8_64(ptr: *const u8, len: usize) -> u32 {
    if ptr.is_null() || len == 0 {
        return 0;
    }
    let data = core::slice::from_raw_parts(ptr, len);
    sum_u8_impl::<64>(data)
}

// -----------------------------------------------------------------------------

#[inline(always)]
unsafe fn is_ascii_impl<const N: usize>(data: &[u8]) -> bool
where
    LaneCount<N>: SupportedLaneCount,
{
    let mut chunks = data.chunks_exact(N);
    for chunk in &mut chunks {
        let v = Simd::<u8, N>::from_slice(chunk);
        if v.reduce_max() >= 0x80 {
            return false;
        }
    }
    chunks.remainder().iter().all(|&b| b < 0x80)
}

/* ─── 32-lane symbol ───────────────────────────────────── */
#[unsafe(no_mangle)]
pub unsafe extern "C" fn is_ascii32(ptr: *const u8, len: usize) -> u8 {
    if ptr.is_null() || len == 0 {
        return 1;
    }
    let data = core::slice::from_raw_parts(ptr, len);
    is_ascii_impl::<32>(data) as u8
}

/* ─── 64-lane symbol ───────────────────────────────────── */
#[unsafe(no_mangle)]
pub unsafe extern "C" fn is_ascii64(ptr: *const u8, len: usize) -> u8 {
    if ptr.is_null() || len == 0 {
        return 1;
    }
    let data = core::slice::from_raw_parts(ptr, len);
    is_ascii_impl::<64>(data) as u8
}

// legacy alias removed

// === Generic byte-set validator ============================================

#[inline(always)]
unsafe fn validate_u8_lut_impl<const L: usize>(data: &[u8], table: &[u8]) -> bool
where
    LaneCount<L>: SupportedLaneCount,
{
    let mut chunks = data.chunks_exact(L);
    for chunk in &mut chunks {
        let v = Simd::<u8, L>::from_slice(chunk);
        let idx: Simd<usize, L> = v.cast();
        let flags = Simd::<u8, L>::gather_or_default(table, idx);
        if flags.reduce_min() == 0 {
            return false;
        }
    }
    for &b in chunks.remainder() {
        if table[b as usize] == 0 {
            return false;
        }
    }
    true
}

#[unsafe(no_mangle)]
pub unsafe extern "C" fn validate_u8_lut32(ptr: *const u8, len: usize, lut: *const u8) -> u8 {
    if ptr.is_null() || len == 0 {
        return 1;
    }
    let data = core::slice::from_raw_parts(ptr, len);
    let table = core::slice::from_raw_parts(lut, 256);
    validate_u8_lut_impl::<32>(data, table) as u8
}

#[unsafe(no_mangle)]
pub unsafe extern "C" fn validate_u8_lut64(ptr: *const u8, len: usize, lut: *const u8) -> u8 {
    if ptr.is_null() || len == 0 {
        return 1;
    }
    let data = core::slice::from_raw_parts(ptr, len);
    let table = core::slice::from_raw_parts(lut, 256);
    validate_u8_lut_impl::<64>(data, table) as u8
}

// legacy alias removed

// === Byte mapping via LUT ====================================================

#[inline(always)]
unsafe fn map_u8_lut_impl<const L: usize>(
    src: *const u8,
    len: usize,
    dst: *mut u8,
    table: *const u8,
) where
    LaneCount<L>: SupportedLaneCount,
{
    let src_slice = core::slice::from_raw_parts(src, len);
    let dst_slice = core::slice::from_raw_parts_mut(dst, len);
    let map = core::slice::from_raw_parts(table, 256);

    let mut chunks = src_slice.chunks_exact(L);
    let mut out_chunks = dst_slice.chunks_exact_mut(L);
    for (chunk, out) in (&mut chunks).zip(&mut out_chunks) {
        let v = Simd::<u8, L>::from_slice(chunk);
        let idx: Simd<usize, L> = v.cast();
        let mapped = Simd::<u8, L>::gather_or_default(map, idx);
        for i in 0..L {
            out[i] = mapped[i];
        }
    }
    // tail
    for (i, &b) in chunks.remainder().iter().enumerate() {
        dst_slice[len - chunks.remainder().len() + i] = map[b as usize];
    }
}

#[unsafe(no_mangle)]
pub unsafe extern "C" fn map_u8_lut32(src: *const u8, len: usize, dst: *mut u8, map: *const u8) {
    if len == 0 || src.is_null() || dst.is_null() || map.is_null() {
        return;
    }
    map_u8_lut_impl::<32>(src, len, dst, map);
}

#[unsafe(no_mangle)]
pub unsafe extern "C" fn map_u8_lut64(src: *const u8, len: usize, dst: *mut u8, map: *const u8) {
    if len == 0 || src.is_null() || dst.is_null() || map.is_null() {
        return;
    }
    map_u8_lut_impl::<64>(src, len, dst, map);
}

// legacy alias removed

// -----------------------------------------------------------------------------

// FFI helper: no-op function to measure call overhead -------------------------
#[unsafe(no_mangle)]
pub unsafe extern "C" fn noop() {
    // deliberately does nothing
}

#[cfg(test)]
mod tests {
    #[test]
    fn test_sum_u8() {
        let data: Vec<u8> = (0u8..=255u8).collect();
        let expected: u32 = data.iter().map(|&b| b as u32).sum();
        let result = unsafe { super::sum_u8_64(data.as_ptr(), data.len()) };
        assert_eq!(result, expected);
    }

    #[test]
    fn test_sum_u8_wrap() {
        // Number of 0xFF bytes needed to overflow u32.
        const LEN: usize = 16_843_010; // ceil(2^32 / 255)
        let data = vec![0xFFu8; LEN];
        // Expected result is (255 * LEN) mod 2^32.
        let expected = ((255u64 * LEN as u64) & 0xFFFF_FFFF) as u32;
        let result = unsafe { super::sum_u8_64(data.as_ptr(), data.len()) };
        assert_eq!(result, expected);
    }

    #[test]
    fn test_sum_u8_various_lengths() {
        // Stress a variety of lengths to make sure chunk/remainder logic works.
        let mut lengths = vec![
            0usize,
            1,
            64 - 1,
            64,
            64 + 1,
            2 * 64 - 3,
            2 * 64,
            4096,
            10_000,
        ];
        lengths.push(123_456); // an arbitrary large size

        for len in lengths {
            // Create data pattern 0,1,2,3,... wrapping every 256.
            let data: Vec<u8> = (0..len as u32).map(|i| (i % 256) as u8).collect();

            let expected: u32 = data.iter().map(|&b| b as u32).sum();
            let got = unsafe { super::sum_u8_64(data.as_ptr(), data.len()) };
            assert_eq!(got, expected, "failed at len={}", len);
        }
    }

    #[test]
    fn test_is_ascii() {
        let ascii = b"Hello, world!";
        let non_ascii = [0x48u8, 0x80u8, 0x49u8];

        unsafe {
            assert_eq!(super::is_ascii32(ascii.as_ptr(), ascii.len()), 1);
            assert_eq!(super::is_ascii64(non_ascii.as_ptr(), non_ascii.len()), 0);
            assert_eq!(super::is_ascii32(core::ptr::null(), 0), 1);
        }
    }
}
