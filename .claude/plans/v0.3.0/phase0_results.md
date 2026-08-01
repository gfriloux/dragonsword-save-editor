# Phase 0 — audit (v0.3.0)

Date: 2026-08-01
Branch: `feat/editor-v0.3.0` off `main` @ `f2ce732` (v0.2.0 merged).

## `nix develop --command just ci`

Green — `gofmt`, `go vet`, `staticcheck`, `go test` (`internal/domain`, `internal/save`,
`internal/sqlcipher` ok; real-save tests skipped without `DSA_SAVE`), `go build`.

## Data shape (reference save)

- `tb_character`: 8 rows. Cols used: CHARACTER_CID, LEVEL, EXP, ASCEND, HP, TRANSCEND,
  SOLDIER_GRADE. Character CIDs are 5-digit (e.g. 10002).
- `tb_team`: 3 pages (PAGE_ID 0/1/2), each SLOT1/2/3_CHARACTER_CID → a character.
- `tb_equipment`: 34 active (DELETED_DATE=0). Editable-looking: ENCHANT_LEVEL, EXP,
  IS_LOCK. Stats (MAIN_STAT_CID, SUB_STAT_CID1..5) are CID references (names unknown →
  display-only). Item CIDs 136x/137x. PK ITEM_DBID.
- `tb_gem`: **0 rows** in this save. Cols: ITEM_DBID, ITEM_CID, STAT_INFO_CID, IS_LOCK.
  → the gem panel/accessor must handle an empty set gracefully.

## Working tree

Clean. `internal/domain` currently exposes currencies + consumables only.
