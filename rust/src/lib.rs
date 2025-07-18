//! Rust SIMD kernels for Simba FFI layer
#![feature(portable_simd)]
#![allow(unsafe_op_in_unsafe_fn)] // calls to unsafe APIs are audited and wrapped inside unsafe fns
use core::simd::prelude::{SimdPartialEq, SimdUint};
use core::simd::{LaneCount, Simd, SupportedLaneCount};
use crc32c::{crc32c_append, crc32c_combine};

// === CRC32C (Castagnoli) update & combine ====================================

// Go's hash/crc32 package expects CRCs to be in *finalised* form—i.e. the
// algorithm complements the accumulator before and after processing a buffer.
// The `crc32c_append` function, on the other hand, operates on the *raw*
// (un-finalised) value so that callers can chain updates cheaply.  Therefore we
// need to mirror Go’s semantics by XOR-ing with 0xFFFF_FFFF around the call.
fn crc32c_update(init_finalised: u32, data: &[u8]) -> u32 {
    crc32c_append(init_finalised, data)
}

#[inline(always)]
fn crc32c_combine_go(crc1_final: u32, crc2_final: u32, len2: usize) -> u32 {
    // `crc32c_combine` operates directly on *finalised* CRC digests, matching Go’s
    // semantics, so we can forward the values unmodified.
    crc32c_combine(crc1_final, crc2_final, len2)
}

macro_rules! export_crc32_update {
    ($name:ident) => {
        #[doc = "Update CRC32C (Castagnoli) with additional bytes.\n\n\
                # Safety\n\
                `ptr` must be null or valid for `len` bytes."]
        #[unsafe(no_mangle)]
        pub unsafe extern "C" fn $name(ptr: *const u8, len: usize, init: u32) -> u32 {
            if ptr.is_null() || len == 0 {
                return init;
            }
            let data = core::slice::from_raw_parts(ptr, len);
            crc32c_update(init, data)
        }
    };
}

// Export symbols expected by Go trampolines (…_raw suffix)
export_crc32_update!(crc32_update_32_raw);
export_crc32_update!(crc32_update_64_raw);

// Keep shorter aliases without `_raw` for potential direct calls (optional)
export_crc32_update!(crc32_update_32);
export_crc32_update!(crc32_update_64);

/// Combine two finalised CRC32C digests (Castagnoli) as per Go's semantics.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn crc32_combine_raw(crc1: u32, crc2: u32, len2: usize) -> u32 {
    crc32c_combine_go(crc1, crc2, len2)
}

// Optional alias without `_raw`.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn crc32_combine(crc1: u32, crc2: u32, len2: usize) -> u32 {
    crc32c_combine_go(crc1, crc2, len2)
}

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

// === FFI trampoline sanity helper ===========================================
/// Simple checksum over the arguments; used only by Go tests to verify that
/// assembly trampolines pass parameters with the correct width/order.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn trampoline_sanity(
    ptr: *const u8,
    len: usize,
    val32: u32,
    val8: u8,
    val64: u64,
    f64_bits: u64,
    f32_bits: u32,
) -> usize {
    // Mix everything into a 64-bit value using a cheap LCG-style hash.
    let mut h = 0xcbf29ce484222325u64; // FNV offset basis
    #[inline(always)]
    fn mix(h: u64, v: u64) -> u64 {
        h ^ v.wrapping_mul(0x100_0000_01b3)
    }
    h = mix(h, ptr as u64);
    h = mix(h, len as u64);
    h = mix(h, val32 as u64);
    h = mix(h, val8 as u64);
    h = mix(h, val64);
    let fb64 = f64_bits & 0x7fff_ffff_ffff_ffffu64; // ignore sign if provided
    let fb32 = (f32_bits & 0x7fff_ffffu32) as u64;
    h = mix(h, fb64);
    h = mix(h, fb32);
    h as usize
}

/// Echo structure for detailed trampoline debugging (test builds only).
#[repr(C)]
#[derive(Copy, Clone, Debug, PartialEq)]
pub struct Echo {
    pub ptr: usize,
    pub len: usize,
    pub v32: u32,
    pub v8: u8,
    pub v64: u64,
    pub f64bits: u64,
    pub f32bits: u32,
}

/// Bounce all parameters back to the caller; used by Go unit tests to pinpoint
/// which argument (if any) is mis-marshalled by the assembly trampolines.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn trampoline_echo(
    ptr: *const u8,
    len: usize,
    v32: u32,
    v8: u8,
    v64: u64,
    f64bits: u64,
    f32bits: u32,
    out: *mut Echo,
) {
    if out.is_null() {
        return;
    }
    *out = Echo {
        ptr: ptr as usize,
        len,
        v32,
        v8,
        v64,
        f64bits,
        f32bits,
    };
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

#[cfg(test)]
mod crc32c_tests {
    /// Known-good CRC32C values computed via Go's hash/crc32 package.
    const CRC1: u32 = 0xa016d052; // checksum of a single byte 0x01
    const CRC_HELLO: u32 = 0x9a71bb4c; // checksum of "hello"

    #[test]
    fn test_crc32c_single_byte() {
        let data = [0x01u8];
        let got = super::crc32c_update(0, &data);
        assert_eq!(
            got, CRC1,
            "crc32c_update mismatch for 0x01: {:08x} != {:08x}",
            got, CRC1
        );
    }

    #[test]
    fn test_crc32c_hello() {
        let data = b"hello";
        let got = super::crc32c_update(0, data);
        assert_eq!(
            got, CRC_HELLO,
            "crc32c_update mismatch for \"hello\": {:08x} != {:08x}",
            got, CRC_HELLO
        );
    }

    #[test]
    fn test_crc32c_reference_crc_crate() {
        const CRC32C: crc::Crc<u32> = crc::Crc::<u32>::new(&crc::CRC_32_ISCSI);
        let data1 = [0x01u8];
        let data2 = b"hello";
        let ref1 = CRC32C.checksum(&data1);
        let ref2 = CRC32C.checksum(data2);
        assert_eq!(ref1, CRC1);
        assert_eq!(ref2, CRC_HELLO);
    }

    #[test]
    fn test_crc32c_direct_function() {
        let data1 = [0x01u8];
        let data2 = b"hello";
        assert_eq!(crc32c::crc32c(&data1), 0xa016d052);
        assert_eq!(crc32c::crc32c(data2), 0x9a71bb4c);
    }

    #[test]
    fn test_crc32c_append_semantics() {
        let data = [0x01u8];
        let res = crc32c::crc32c_append(0, &data);
        assert_eq!(
            res, 0xa016d052,
            "crc32c_append with init 0 should give finalised CRC"
        );
    }

    #[test]
    fn test_crc32c_50_ab() {
        let data = vec![0xABu8; 50];
        let expect = 0xd64d26c9u32;
        let got = super::crc32c_update(0, &data);
        assert_eq!(got, expect);
    }

    #[test]
    fn test_crc32c_ffi_raw() {
        let data = [0x01u8];
        let got = unsafe { super::crc32_update_32_raw(data.as_ptr(), data.len(), 0) };
        assert_eq!(got, 0xa016d052);
    }

    #[test]
    fn test_crc32c_alias_function() {
        let data = [0x01u8];
        let got = unsafe { super::crc32_update_32(data.as_ptr(), data.len(), 0) };
        assert_eq!(got, 0xa016d052);
    }

    #[test]
    fn test_crc32c_combine() {
        use rand::{RngCore, SeedableRng};
        let mut rng = rand::rngs::StdRng::seed_from_u64(42);
        let len1 = 100usize;
        let len2 = 200usize;
        let mut buf1 = vec![0u8; len1];
        let mut buf2 = vec![0u8; len2];
        rng.fill_bytes(&mut buf1);
        rng.fill_bytes(&mut buf2);

        let crc1 = super::crc32c_update(0, &buf1);
        let crc2 = super::crc32c_update(0, &buf2);
        let mut concat = Vec::from(buf1);
        concat.extend_from_slice(&buf2);
        let expected_concat = super::crc32c_update(0, &concat);
        let combined = unsafe { super::crc32_combine_raw(crc1, crc2, len2) };
        assert_eq!(combined, expected_concat, "combine mismatch");
    }

    #[test]
    fn debug_combine_variants() {
        use crc32c::crc32c_combine;
        let buf1: &[u8] = b"hello";
        let buf2: &[u8] = b" world"; // len2=6
        let crc1 = super::crc32c_update(0, buf1);
        let crc2 = super::crc32c_update(0, buf2);
        let mut concat = Vec::from(buf1);
        concat.extend_from_slice(&buf2);
        let expected = super::crc32c_update(0, &concat);

        let raw1 = !crc1;
        let raw2 = !crc2;

        let v1 = !crc32c_combine(raw1, raw2, buf2.len());
        let v2 = crc32c_combine(raw1, raw2, buf2.len());
        let v3 = !crc32c_combine(crc1, crc2, buf2.len());
        let v4 = crc32c_combine(crc1, crc2, buf2.len());
        println!(
            "expected {:08x} v1 {:08x} v2 {:08x} v3 {:08x} v4 {:08x}",
            expected, v1, v2, v3, v4
        );
    }

    #[test]
    fn test_crc32c_combine_variants_randomised() {
        use rand::{RngCore, SeedableRng};

        const TRIALS: usize = 100;
        let mut rng = rand::rngs::StdRng::seed_from_u64(0x5eed);

        for _ in 0..TRIALS {
            // Generate two random buffers of random length up to 8 KiB.
            let len1 = (rng.next_u32() % 8192) as usize;
            let len2 = (rng.next_u32() % 8192) as usize;
            let mut buf1 = vec![0u8; len1];
            let mut buf2 = vec![0u8; len2];
            rng.fill_bytes(&mut buf1);
            rng.fill_bytes(&mut buf2);

            let crc1 = super::crc32c_update(0, &buf1);
            let crc2 = super::crc32c_update(0, &buf2);

            // Ground-truth CRC of the concatenation.
            let mut concat = buf1.clone();
            concat.extend_from_slice(&buf2);
            let expected = super::crc32c_update(0, &concat);

            // Variant implemented in Simba (direct finalised combine).
            let chosen = super::crc32c_combine_go(crc1, crc2, len2);
            assert_eq!(chosen, expected, "crc32c_combine_go produced wrong value");

            // Sanity-check other flip combinations — none should match.
            let raw1 = !crc1;
            let raw2 = !crc2;
            let variants = [
                // Previously-used implementation that flipped before + after
                !crc32c::crc32c_combine(raw1, raw2, len2),
                // Mix-and-match variants
                crc32c::crc32c_combine(raw1, raw2, len2),
                !crc32c::crc32c_combine(crc1, crc2, len2),
            ];
            for (i, &v) in variants.iter().enumerate() {
                assert_ne!(v, expected, "variant {} unexpectedly matched", i + 1);
            }
        }
    }
}
