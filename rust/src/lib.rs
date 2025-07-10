//! Rust SIMD kernels for Simba FFI layer
#![feature(portable_simd)]
#![allow(unsafe_op_in_unsafe_fn)] // calls to unsafe APIs are audited and wrapped inside unsafe fns
use core::simd::prelude::{SimdPartialEq, SimdUint};
use core::simd::{LaneCount, Simd, SupportedLaneCount};

// === Portable SIMD byte-sum ===================================================

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

/* ─── sum_u8 public exports generated via macro ──────────────────────────── */

/// Generates thin extern "C" wrappers around `sum_u8_impl` for a given lane
/// width. Mirrors style of `export_eq_masks!`.
macro_rules! export_sum_u8 {
    ($name:ident, $lanes:expr) => {
        #[doc = concat!(
            "Sum the bytes in `data` using a ", stringify!($lanes), "-lane portable SIMD kernel and return the running total modulo 2^32.\n\n",
            "# Safety\n",
            "• `ptr` must be either null or valid for `len` bytes.\n",
            "• The buffer must not be mutated for the duration of the call."
        )]
        #[unsafe(no_mangle)]
        pub unsafe extern "C" fn $name(ptr: *const u8, len: usize) -> u32 {
            if ptr.is_null() || len == 0 {
                return 0;
            }
            let data = core::slice::from_raw_parts(ptr, len);
            sum_u8_impl::<$lanes>(data)
        }
    };
}

export_sum_u8!(sum_u8_16, 16);
export_sum_u8!(sum_u8_32, 32);
export_sum_u8!(sum_u8_64, 64);

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

/* ─── is_ascii public exports via macro ─────────────────────────────────── */
macro_rules! export_is_ascii {
    ($name:ident, $lanes:expr) => {
        #[doc = concat!(
            "Return 1 if all bytes are ASCII (< 0x80) using a ", stringify!($lanes), "-lane SIMD kernel, 0 otherwise.\n\n",
            "# Safety\n",
            "Same as other FFI helpers: `ptr` must be null or valid for `len` bytes."
        )]
        #[unsafe(no_mangle)]
        pub unsafe extern "C" fn $name(ptr: *const u8, len: usize) -> u8 {
            if ptr.is_null() || len == 0 {
                return 1;
            }
            let data = core::slice::from_raw_parts(ptr, len);
            is_ascii_impl::<$lanes>(data) as u8
        }
    };
}
export_is_ascii!(is_ascii16, 16);
export_is_ascii!(is_ascii32, 32);
export_is_ascii!(is_ascii64, 64);

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

/* ─── validate_u8_lut exports via macro ─────────────────────────────────── */
macro_rules! export_validate_u8_lut {
    ($name:ident, $lanes:expr) => {
        #[doc = concat!(
            "Validate every byte against a 256-byte lookup table using a ", stringify!($lanes), "-lane SIMD kernel. Non-zero table entry marks valid byte. Returns 1 on success, 0 on first mismatch.\n\n",
            "# Safety\n",
            "• `ptr`/`lut` must be valid for `len`/256 bytes respectively."
        )]
        #[unsafe(no_mangle)]
        pub unsafe extern "C" fn $name(ptr: *const u8, len: usize, lut: *const u8) -> u8 {
            if ptr.is_null() || len == 0 {
                return 1;
            }
            let data = core::slice::from_raw_parts(ptr, len);
            let table = core::slice::from_raw_parts(lut, 256);
            validate_u8_lut_impl::<$lanes>(data, table) as u8
        }
    };
}
export_validate_u8_lut!(validate_u8_lut16, 16);
export_validate_u8_lut!(validate_u8_lut32, 32);
export_validate_u8_lut!(validate_u8_lut64, 64);

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

/* ─── map_u8_lut exports via macro ───────────────────────────────────────── */
macro_rules! export_map_u8_lut {
    ($name:ident, $lanes:expr) => {
        #[doc = concat!(
            "Map each source byte through a 256-byte translation table using a ", stringify!($lanes), "-lane SIMD kernel and write results to `dst`.\n\n",
            "# Safety\n",
            "All pointers must be non-null and valid for `len` bytes. Buffers may overlap."
        )]
        #[unsafe(no_mangle)]
        pub unsafe extern "C" fn $name(src: *const u8, len: usize, dst: *mut u8, map: *const u8) {
            if len == 0 || src.is_null() || dst.is_null() || map.is_null() {
                return;
            }
            map_u8_lut_impl::<$lanes>(src, len, dst, map);
        }
    };
}
export_map_u8_lut!(map_u8_lut16, 16);
export_map_u8_lut!(map_u8_lut32, 32);
export_map_u8_lut!(map_u8_lut64, 64);

// === Byte equality mask =====================================================

#[inline(always)]
unsafe fn eq_u8_masks_impl<const LANES: usize>(
    src: *const u8,
    len: usize,
    needle: u8,
    out: *mut u128, // storage large enough for any mask size, cast later
) -> usize
where
    LaneCount<LANES>: SupportedLaneCount,
{
    if len == 0 {
        return 0;
    }
    let chunks = len / LANES;
    let src_slice = core::slice::from_raw_parts(src, len);
    let out_slice = core::slice::from_raw_parts_mut(out as *mut u128, chunks);

    for (i, chunk) in src_slice.chunks_exact(LANES).enumerate() {
        if i == chunks {
            break;
        }
        let v = Simd::<u8, LANES>::from_slice(chunk);
        let mask = v.simd_eq(Simd::splat(needle));
        out_slice[i] = mask.to_bitmask() as u128;
    }
    chunks
}

// Generic helper: generates a thin extern "C" wrapper that validates pointers
// Export helper specific to eq_u8_masks kernels (16/32/64 lanes)
macro_rules! export_eq_masks {
    ($name:ident, $lanes:expr, $int:ty) => {
        #[doc = concat!(
            "Generate equality bitmasks comparing each byte to `needle` across chunks of ", stringify!($lanes), " lanes. The resulting mask words are stored in `out`. Returns number of mask words written.\n\n",
            "# Safety\n",
            "`src` and `out` must be valid for `len` and `len/", stringify!($lanes), "` elements respectively."
        )]
        #[unsafe(no_mangle)]
        pub unsafe extern "C" fn $name(
            src: *const u8,
            len: usize,
            needle: u8,
            out: *mut $int,
        ) -> usize {
            if src.is_null() || out.is_null() || len == 0 {
                return 0;
            }
            eq_u8_masks_impl::<$lanes>(src, len, needle, out as *mut u128)
        }
    };
}

export_eq_masks!(eq_u8_masks16, 16, u16);
export_eq_masks!(eq_u8_masks32, 32, u32);
export_eq_masks!(eq_u8_masks64, 64, u64);

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

    #[test]
    fn test_map_u8_lut_basic() {
        // Mapping table: invert each byte (x -> 255 - x)
        let map: Vec<u8> = (0..=255u16).map(|b| 255u8.wrapping_sub(b as u8)).collect();
        let src: Vec<u8> = (0..=255u16).map(|b| b as u8).collect();
        let mut dst16 = vec![0u8; src.len()];
        let mut dst32 = vec![0u8; src.len()];
        let mut dst64 = vec![0u8; src.len()];
        unsafe {
            super::map_u8_lut16(src.as_ptr(), src.len(), dst16.as_mut_ptr(), map.as_ptr());
            super::map_u8_lut32(src.as_ptr(), src.len(), dst32.as_mut_ptr(), map.as_ptr());
            super::map_u8_lut64(src.as_ptr(), src.len(), dst64.as_mut_ptr(), map.as_ptr());
        }
        let expected: Vec<u8> = src.iter().map(|&b| 255 - b).collect();
        assert_eq!(dst16, expected, "16-lane mapping failed");
        assert_eq!(dst32, expected, "32-lane mapping failed");
        assert_eq!(dst64, expected, "64-lane mapping failed");
    }

    #[test]
    fn test_map_u8_lut_various_lengths() {
        let map: Vec<u8> = (0..=255u16).map(|b| (b as u8).wrapping_add(1)).collect(); // simple +1 mapping
        let lengths = [0usize, 1, 15, 16, 17, 31, 32, 33, 63, 64, 65, 255, 1023];
        for &len in &lengths {
            let src: Vec<u8> = (0..len as u32).map(|i| (i % 256) as u8).collect();
            let mut dst = vec![0u8; len];
            unsafe {
                super::map_u8_lut64(src.as_ptr(), len, dst.as_mut_ptr(), map.as_ptr());
            }
            for i in 0..len {
                assert_eq!(dst[i], src[i].wrapping_add(1), "idx {} len {}", i, len);
            }
        }
    }
}

#[cfg(test)]
mod mask_tests {
    fn scalar_mask(chunk: &[u8], needle: u8) -> u128 {
        let mut m = 0u128;
        for (i, &b) in chunk.iter().enumerate() {
            if b == needle {
                m |= 1 << i;
            }
        }
        m
    }
    #[test]
    fn test_eq_u8_masks_basic() {
        let data: Vec<u8> = (0..128u16).map(|i| (i % 256) as u8).collect();
        let mut out16 = vec![0u16; data.len() / 16];
        let mut out32 = vec![0u32; data.len() / 32];
        let mut out64 = vec![0u64; data.len() / 64];
        unsafe {
            let c16 = super::eq_u8_masks16(data.as_ptr(), data.len(), 3u8, out16.as_mut_ptr());
            let c32 = super::eq_u8_masks32(data.as_ptr(), data.len(), 3u8, out32.as_mut_ptr());
            let c64 = super::eq_u8_masks64(data.as_ptr(), data.len(), 3u8, out64.as_mut_ptr());
            assert_eq!(c16, out16.len());
            assert_eq!(c32, out32.len());
            assert_eq!(c64, out64.len());
        }
        // validate
        for (i, &mask) in out16.iter().enumerate() {
            let start = i * 16;
            let chunk = &data[start..start + 16];
            assert_eq!(mask as u128, scalar_mask(chunk, 3));
        }
    }
}
