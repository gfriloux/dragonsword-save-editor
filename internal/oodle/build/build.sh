#!/usr/bin/env bash
# Rebuild internal/oodle/ooz.wasm — the pure-Go Oodle decoder's WebAssembly core.
#
# Source: oriath-net/gooz (bundles the open-source ooz Kraken/Mermaid/Leviathan/
# LZNA/BitKnit decompressor). Pinned commit: 3eae6c61d925562077b7bbb053c05fdd5a1050d5
#
# ooz uses only SSE2 (no AVX), so it maps cleanly to wasm SIMD128. We swap its
# <immintrin.h>/<x86intrin.h> includes for <emmintrin.h> and provide _rotl
# (an MSVC intrinsic clang lacks) via shim.h. Everything else (SSE2 emulation,
# libc, malloc) comes from emscripten.
set -euo pipefail
here=$(dirname "$(realpath "$0")")
out="$here/../ooz.wasm"
src=$(mktemp -d)
git clone https://github.com/oriath-net/gooz "$src"
git -C "$src" checkout 3eae6c61d925562077b7bbb053c05fdd5a1050d5
sed -i 's|#include <immintrin.h>|#include <emmintrin.h>|; s|#include <x86intrin.h>||'   "$src"/kraken.cpp "$src"/lzna.cpp
cp "$here/shim.h" "$src"/shim.h
( cd "$src" && emcc kraken.cpp lzna.cpp bitknit.cpp     -O2 -msimd128 -msse -msse2 -include shim.h -fno-exceptions -fno-rtti     --no-entry -sSTANDALONE_WASM -sALLOW_MEMORY_GROWTH=1     -sEXPORTED_FUNCTIONS=_Kraken_Decompress,_malloc,_free     -sERROR_ON_UNDEFINED_SYMBOLS=0 -o "$out" )
echo "wrote $out"
