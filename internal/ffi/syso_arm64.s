//go:build arm64
// +build arm64

#include "textflag.h"

// func noop_raw()
TEXT ·noop_raw(SB), NOSPLIT, $0-0
    BL      noop(SB)
    RET 

// func sum_u8_32_raw(ptr *byte, n uintptr) uint32
TEXT ·sum_u8_32_raw(SB), NOSPLIT, $0-20
    MOVD    ptr+0(FP), R0
    MOVD    n+8(FP),  R1
    BL      sum_u8_32(SB)
    MOVW    R0, ret+16(FP)
    RET

// func sum_u8_64_raw(ptr *byte, n uintptr) uint32
TEXT ·sum_u8_64_raw(SB), NOSPLIT, $0-20
    MOVD    ptr+0(FP), R0
    MOVD    n+8(FP),  R1
    BL      sum_u8_64(SB)
    MOVW    R0, ret+16(FP)
    RET

// func is_ascii32_raw(ptr *byte, n uintptr) uint8
TEXT ·is_ascii32_raw(SB), NOSPLIT, $0-17
    MOVD    ptr+0(FP), R0
    MOVD    n+8(FP),  R1
    BL      is_ascii32(SB)
    MOVBU   R0, ret+16(FP)
    RET

// func is_ascii64_raw(ptr *byte, n uintptr) uint8
TEXT ·is_ascii64_raw(SB), NOSPLIT, $0-17
    MOVD    ptr+0(FP), R0
    MOVD    n+8(FP),  R1
    BL      is_ascii64(SB)
    MOVBU   R0, ret+16(FP)
    RET

// func validate_u8_lut32_raw(ptr *byte, n uintptr, lut *byte) uint8
TEXT ·validate_u8_lut32_raw(SB), NOSPLIT, $0-25
    MOVD    ptr+0(FP), R0
    MOVD    n+8(FP),  R1
    MOVD    lut+16(FP), R2
    BL      validate_u8_lut32(SB)
    MOVBU   R0, ret+24(FP)
    RET

// func validate_u8_lut64_raw(ptr *byte, n uintptr, lut *byte) uint8
TEXT ·validate_u8_lut64_raw(SB), NOSPLIT, $0-25
    MOVD    ptr+0(FP), R0
    MOVD    n+8(FP),  R1
    MOVD    lut+16(FP), R2
    BL      validate_u8_lut64(SB)
    MOVBU   R0, ret+24(FP)
    RET

// func map_u8_lut32_raw(src *byte, n uintptr, dst *byte, lut *byte)
TEXT ·map_u8_lut32_raw(SB), NOSPLIT, $0-32
    MOVD    src+0(FP), R0
    MOVD    n+8(FP),  R1
    MOVD    dst+16(FP), R2
    MOVD    lut+24(FP), R3
    BL      map_u8_lut32(SB)
    RET

// func map_u8_lut64_raw(src *byte, n uintptr, dst *byte, lut *byte)
TEXT ·map_u8_lut64_raw(SB), NOSPLIT, $0-32
    MOVD    src+0(FP), R0
    MOVD    n+8(FP),  R1
    MOVD    dst+16(FP), R2
    MOVD    lut+24(FP), R3
    BL      map_u8_lut64(SB)
    RET 

// end 