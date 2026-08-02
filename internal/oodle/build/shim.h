#pragma once
#include <stdint.h>
static inline uint32_t _rotl(uint32_t x, int n){ return __builtin_rotateleft32(x, (uint32_t)n); }
