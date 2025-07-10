package algo

// Recent benchmarks with the syso trampoline show a per-call overhead of
// just ~0.3 ns on an Apple M2 Max, meaning a **quarter** of a cache line
// (16 bytes) is already competitive.  We drop the crossover to 16 B to favour
// SIMD more aggressively on modern CPUs.
const simdThreshold = 16
