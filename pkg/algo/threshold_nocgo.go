package algo

// simdThreshold crossover for builds that do NOT use cgo (pure-Go FFI
// gateway, e.g. the syso + assembly shim path).  The fixed overhead of a
// syso call is ~2–3 ns on Apple M2, so even a single 32-byte cache line of
// work fully amortises it.  We keep the threshold at 64 B for now to stay on
// the conservative side and match the cgo setting on older Intel/Zen CPUs.
const simdThreshold = 64
