# Phase 0 — audit (v0.6.0)

Date: 2026-08-01. Branch `feat/item-icons` off `main` @ `e1d9d98` (v0.5.0 merged).
`just ci` green. Tree clean.

th.gl icons confirmed: single sprite `icons.<hash>.webp` (~1489×1586, ~1 MB, the hash
rotates), per-item `object-position:<x>px <y>px`, 64px cells, `zoom:0.3125`. Positions
parse cleanly (e.g. 1360001 -> -1322,-860). Sprite must be downloaded in the same run
as the positions.
