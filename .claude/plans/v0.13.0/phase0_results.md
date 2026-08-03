# Phase 0 — audit (v0.13.0)

_To fill at execution start._

## `nix develop --command just ci`
- [x] fmt-check
- [x] vet
- [x] lint
- [x] test
- [x] build

Result: **green** (exit 0) on `feat/cooking-recipes`, tree clean apart from the
uncommitted `tmp/_spike/` throwaway (excluded from `./...` by the `_` prefix).

## Phase 1 spike — switch-key formula validation (user-run, in-game)

Formula under test: `category = CookBook_SwitchData / 64`, `bit = CookBook_SwitchData % 64`.

- [~] Check #1 — single recipe: not run separately (superseded by #2, which is stronger:
      9 distinct keys, all previously unknown).
- [x] Check #2 — category 62 tail (`SwitchData` 4000–4008 → cat 62, bits 32–40): helper
      reported all 9 `was_known=false -> now_known=true`; USER_DBID=1000; backup written.
      **User confirmed in-game: the recipes are unlocked.**

Decision: **PASS.** The mapping `category = key/64, bit = key%64` is validated in-game, and
the old blanket 15–60 provably missed these 9 category-62 recipes. Proceed to Phase 2.
