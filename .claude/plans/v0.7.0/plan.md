# Plan: v0.7.0 — unlock all cooking recipes

**Type:** domain + web
**Objective:** Add a one-click "unlock all cooking recipes" action to the editor.
**Why:** Known recipes gate what you can cook; the game exposes no bulk unlock.
**Layer(s):** domain, web, docs.

## How recipes are stored (reverse-engineered, see memory `dsa-cook-recipes`)

Known recipes are flags in **`tb_switch`** — a generic bitmask table
`(USER_DBID, CATEGORY, BIT_FIELD)`. A recipe's `CookRecipeKey` maps to
`(CATEGORY = 15 + key/64, bit = key%64)`. Empirically (before/after diff of unlocking
one recipe, then progressive test saves validated in-game), the normal recipes
(grilled / boiled / sliced, CIDs 1420xxx–1422xxx) occupy categories **15–60**.
Setting those categories' bits unlocks all normal recipes with no side effects
(validated in-game). A handful of special recipes (odd CIDs, e.g. "Plat d'origine
inconnue", iced drinks) sit outside this range and are deferred (need their own diff).

## Scope

**In scope:**
- `UnlockAllRecipes()`: upsert `tb_switch` categories 15..60 to all-bits for the user.
- API endpoint + a "Cooking" section in the Editor with an "Unlock all recipes" button
  (with confirmation).

**Out of scope:**
- The ~3 special recipes outside categories 15–60 (future, once diffed).
- Per-recipe unlock UI, or reading the exact known/total counts.

## Design

- `internal/domain/domain.go`: `const (recipeSwitchLo=15; recipeSwitchHi=60)`;
  `UnlockAllRecipes() error` loops the range and
  `INSERT OR REPLACE INTO tb_switch (USER_DBID, CATEGORY, BIT_FIELD) VALUES (?, cat, -1)`
  (all 64 bits; bits beyond real recipes are inert — the game only checks keys that
  exist in its recipe table).
- `internal/web`: `POST /api/game/recipes/unlock-all`.
- Frontend: a "Cooking" Editor section with a short explanation and an "Unlock all
  recipes" button (confirm dialog), reporting success.

## Atomic steps (each = 1 commit, `just ci` green)

1. `feat(domain): add UnlockAllRecipes (tb_switch categories 15-60)` + test (real save).
2. `feat(web): add /api/game/recipes/unlock-all`.
3. `feat(web): add a Cooking panel with unlock-all-recipes`.
4. `docs: document unlocking all recipes`.

## Quality gates

- [ ] `just ci`; `DSA_SAVE=… just test` (unlock flips a known-locked recipe bit; the
      set persists through Write + reopen; verified vs `sqlcipher`).
- [ ] Manual: click unlock, Write, load in-game → all normal recipes unlocked, nothing
      else changed (already validated with the equivalent test save).
- [ ] Atomic commits on `feat/unlock-recipes`; user merges.

## Decisions

- **Categories 15–60, all bits** — the validated range; extra bits are inert.
- Setting `BIT_FIELD = -1` (uint64 all-ones) per category; OR-ing would give the same.
