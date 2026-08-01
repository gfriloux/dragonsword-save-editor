# Phase 0 — audit (v0.5.0)

Date: 2026-08-01. Branch `feat/add-items` off `main` @ `00171a4` (v0.4.0 merged).
`nix develop --command just ci` green. Working tree clean.

`internal/domain` has read + edit accessors for currencies, consumables, characters,
team, equipment, gems, and a bilingual catalog. `tb_stackable_item` PK is
(USER_DBID, ITEM_CID) with STACK_CNT — no triggers, so INSERT OR REPLACE is safe.
