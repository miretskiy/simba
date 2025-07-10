//go:build amd64 && syso_experiment
// +build amd64,syso_experiment

#include "textflag.h"

// func AddU32(a, b uint32) uint32
TEXT ·AddU32(SB), NOSPLIT, $0-0
    // The first two uint32 parameters are already in RDI and RSI according to
    // the Go AMD64 internal ABI (they are zero-extended to 64 bits).  The C
    // function add_u32 has the same calling convention, so we can tail-call
    // it directly.
    CALL add_u32(SB)
    // Result is already in EAX; the Go ABI returns integers in AX, so just RET.
    RET 