//go:build amd64
// +build amd64

#include "textflag.h"

// Helper macro: move ret if needed? but simpler explicit.

// func noop_raw()
TEXT ·noop_raw(SB), NOSPLIT, $0-0
    CALL noop(SB)
    RET

// end 