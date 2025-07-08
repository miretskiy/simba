#ifndef SIMBA_H
#define SIMBA_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

// SIMD kernels
uint32_t sum_u8(const uint8_t *ptr, size_t len);
// Returns 1 if all bytes are ASCII (< 128), 0 otherwise.
uint8_t is_ascii(const uint8_t *ptr, size_t len);

#ifdef __cplusplus
}
#endif

#endif // SIMBA_H 