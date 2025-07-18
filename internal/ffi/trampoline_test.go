package ffi

import (
	"math"
	"math/rand"
	"testing"
	"unsafe"
)

func mix(h, v uint64) uint64 {
	return h ^ (v * 0x100000001b3)
}

func refHash(ptrVal uintptr, length uintptr, v32 uint32, v8 uint8, v64 uint64, f64bits uint64, f32bits uint32) uintptr {
	h := uint64(0xcbf29ce484222325)
	h = mix(h, uint64(ptrVal))
	h = mix(h, uint64(length))
	h = mix(h, uint64(v32))
	h = mix(h, uint64(v8))
	h = mix(h, v64)
	fb64 := f64bits & 0x7fffffffffffffff
	fb32 := uint64(f32bits & 0x7fffffff)
	h = mix(h, fb64)
	h = mix(h, fb32)
	return uintptr(h)
}

func TestTrampolineSanity(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 2000; i++ {
		// choose slice variant
		var slice []byte
		switch i % 3 {
		case 0:
			slice = nil
		case 1:
			slice = []byte{}
		default:
			slice = []byte{1, 2, 3, 4}
		}

		var ptr *byte
		var length uintptr
		if len(slice) > 0 {
			ptr = &slice[0]
			length = uintptr(len(slice))
		}

		v32 := rng.Uint32()
		v8 := uint8(rng.Intn(256))
		v64 := rng.Uint64()
		// ensure 64-bit extremes are exercised
		if i%5 == 0 {
			v64 = math.MaxUint64
		}
		f64 := rng.Float64()
		f32 := rng.Float32()
		if i%11 == 0 {
			f64 = math.Inf(1)
		} else if i%13 == 0 {
			f64 = math.NaN()
		}
		if i%17 == 0 {
			f32 = float32(math.Inf(-1))
		} else if i%19 == 0 {
			f32 = float32(math.NaN())
		}
		f64bits := math.Float64bits(f64)
		f32bits := uint32(math.Float32bits(f32))

		got := TrampolineSanityHash(ptr, length, v32, v8, v64, f64bits, f32bits)
		want := refHash(uintptr(unsafe.Pointer(ptr)), length, v32, v8, v64, f64bits, f32bits)
		if got != want {
			// obtain per-field diff to ease debugging
			echo := TrampolineEcho(ptr, length, v32, v8, v64, f64bits, f32bits)
			t.Fatalf("hash mismatch on iter %d\nwant=%#v\n gotEcho=%#v", i, Echo{Ptr: uintptr(unsafe.Pointer(ptr)), Len: length, V32: v32, V8: v8, V64: v64, F64Bits: f64bits, F32Bits: f32bits}, echo)
		}
	}
}
