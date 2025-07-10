//go:build amd64
// +build amd64

#include "textflag.h"

// Helper macro: move ret if needed? but simpler explicit.

// func sum_u8_raw(ptr *byte, n uintptr) uint32
TEXT ·sum_u8_raw(SB), NOSPLIT, $0-20
    MOVQ ptr+0(FP), DI
    MOVQ n+8(FP), SI
    CALL sum_u8(SB)
    MOVL AX, ret+16(FP)
    RET

// func is_ascii_raw(ptr *byte, n uintptr) uint8
TEXT ·is_ascii_raw(SB), NOSPLIT, $0-17
    MOVQ ptr+0(FP), DI
    MOVQ n+8(FP), SI
    CALL is_ascii(SB)
    MOVB AL, ret+16(FP)
    RET

// func validate_u8_lut_raw(ptr *byte, n uintptr, lut *byte) uint8
TEXT ·validate_u8_lut_raw(SB), NOSPLIT, $0-25
    MOVQ ptr+0(FP), DI
    MOVQ n+8(FP), SI
    MOVQ lut+16(FP), DX
    CALL validate_u8_lut(SB)
    MOVB AL, ret+24(FP)
    RET

// func map_u8_lut_raw(src *byte, n uintptr, dst *byte, lut *byte)
TEXT ·map_u8_lut_raw(SB), NOSPLIT, $0-32
    MOVQ src+0(FP), DI
    MOVQ n+8(FP), SI
    MOVQ dst+16(FP), DX
    MOVQ lut+24(FP), CX
    CALL map_u8_lut(SB)
    RET

// func noop_raw()
TEXT ·noop_raw(SB), NOSPLIT, $0-0
    CALL noop(SB)
    RET

// end 