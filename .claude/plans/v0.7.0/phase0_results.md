# Phase 0 — audit (v0.7.0)

Date: 2026-08-02. Branch `feat/unlock-recipes` off `main` @ `0c38500` (v0.6.0 merged).
`just ci` green. Tree clean.

Recipe storage reverse-engineered (see memory `dsa-cook-recipes`): known recipes are
`tb_switch` bitmask flags; normal recipes occupy categories 15–60, validated in-game
via progressive test saves. `tb_switch` PK (USER_DBID, CATEGORY), so INSERT OR REPLACE
is safe.
