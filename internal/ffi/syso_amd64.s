//go:build amd64
// +build amd64

#include "textflag.h"

// Helper macro: move ret if needed? but simpler explicit.

// func noop_raw()
TEXT ·noop_raw(SB), NOSPLIT, $0-0
    CALL noop(SB)
    RET

// func sum_u8_32_raw(ptr *byte, n uintptr) uint32
TEXT ·sum_u8_32_raw(SB), NOSPLIT, $0-20
    MOVQ ptr+0(FP), DI
    MOVQ n+8(FP),  SI
    CALL sum_u8_32(SB)
    MOVL AX, ret+16(FP)
    RET

// func sum_u8_64_raw(ptr *byte, n uintptr) uint32
TEXT ·sum_u8_64_raw(SB), NOSPLIT, $0-20
    MOVQ ptr+0(FP), DI
    MOVQ n+8(FP),  SI
    CALL sum_u8_64(SB)
    MOVL AX, ret+16(FP)
    RET

// func is_ascii32_raw(ptr *byte, n uintptr) uint8
TEXT ·is_ascii32_raw(SB), NOSPLIT, $0-17
    MOVQ ptr+0(FP), DI
    MOVQ n+8(FP),  SI
    CALL is_ascii32(SB)
    MOVB AL, ret+16(FP)
    RET

// func is_ascii64_raw(ptr *byte, n uintptr) uint8
TEXT ·is_ascii64_raw(SB), NOSPLIT, $0-17
    MOVQ ptr+0(FP), DI
    MOVQ n+8(FP),  SI
    CALL is_ascii64(SB)
    MOVB AL, ret+16(FP)
    RET

// func validate_u8_lut32_raw(ptr *byte, n uintptr, lut *byte) uint8
TEXT ·validate_u8_lut32_raw(SB), NOSPLIT, $0-25
    MOVQ ptr+0(FP), DI
    MOVQ n+8(FP),  SI
    MOVQ lut+16(FP), DX
    CALL validate_u8_lut32(SB)
    MOVB AL, ret+24(FP)
    RET

// func validate_u8_lut64_raw(ptr *byte, n uintptr, lut *byte) uint8
TEXT ·validate_u8_lut64_raw(SB), NOSPLIT, $0-25
    MOVQ ptr+0(FP), DI
    MOVQ n+8(FP),  SI
    MOVQ lut+16(FP), DX
    CALL validate_u8_lut64(SB)
    MOVB AL, ret+24(FP)
    RET

// func map_u8_lut32_raw(src *byte, n uintptr, dst *byte, lut *byte)
TEXT ·map_u8_lut32_raw(SB), NOSPLIT, $0-32
    MOVQ src+0(FP), DI
    MOVQ n+8(FP),  SI
    MOVQ dst+16(FP), DX
    MOVQ lut+24(FP), CX
    CALL map_u8_lut32(SB)
    RET

// func map_u8_lut64_raw(src *byte, n uintptr, dst *byte, lut *byte)
TEXT ·map_u8_lut64_raw(SB), NOSPLIT, $0-32
    MOVQ src+0(FP), DI
    MOVQ n+8(FP),  SI
    MOVQ dst+16(FP), DX
    MOVQ lut+24(FP), CX
    CALL map_u8_lut64(SB)
    RET 