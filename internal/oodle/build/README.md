# ooz.wasm — the Oodle decoder core

`internal/oodle` decodes the game's Oodle-compressed pak data in **pure Go** (no
CGO) by running a WebAssembly build of the open-source **ooz** decompressor on
the pure-Go **wazero** runtime. This keeps the editor a single static,
cross-compilable binary (DESIGN invariant #1) while still reading Oodle assets.

## Why WASM and not a hand port

ooz is ~4200 lines of dense C (bit readers, Huffman, TANS/rANS, SSE2). A faithful
hand port to Go is a multi-day, bug-prone effort; compiling the reference C to
WebAssembly gives a byte-exact decoder for a fraction of the effort and risk.
Provenance is auditable: same source, documented flags.

## Rebuild

Needs `emscripten` (`nix shell nixpkgs#emscripten`). Then:

    ./build.sh

It clones oriath-net/gooz (pinned to `3eae6c61d925562077b7bbb053c05fdd5a1050d5`), swaps the x86 intrinsic
headers for SSE2, adds `shim.h` (`_rotl`), and emits `../ooz.wasm`. Only SSE2 is
used, so `-msimd128` maps it to wasm SIMD.

## Verification

`internal/pak`'s gated `TestReadOodleIcon` (needs `DSA_GAME_DIR`) decompresses a
real icon and checks its SHA-256 against the ooz C reference output.
