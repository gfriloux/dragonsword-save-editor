# Phase 0 — audit (v0.11.0)

Date: 2026-08-02. Branch `feat/pak-catalog` off `main` @ `4c7402a` (v0.10.0 merged).
`nix develop --command just ci` → **green**. Working tree: untracked `.envrc`, `tmp/`,
and the `.claude/plans/release/pak-extraction/` + `v0.11.0/` plan docs. Added `/tmp/` to
`.gitignore` (scratch + game-copyright pak fixtures must never be committed).

Extraction (Phase 1) is done and its artifacts live in `tmp/pak/` (GameItemData.xml,
StringData.xml, type/category defs, extractor source + recipe). This chantier only adds the
Go converter that turns those into a refreshed `items.json`. See
`.claude/plans/release/pak-extraction/plan.md`.
