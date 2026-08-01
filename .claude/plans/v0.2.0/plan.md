# Plan: v0.2.0 — friendly Editor view (domain socle + Currency + Consumables)

**Type:** UI + new domain layer
**Objective:** Make the tool usable by non-power users: add a friendly, game-oriented
"Editor" view alongside the existing raw table browser, starting with the two simplest
high-value panels — Currency and Consumables.
**Why:** The generic `tb_*` grid works but exposes opaque tables/CIDs. A domain view
with names, categories and proper widgets makes editing (money, food/potions) obvious.
**Layer(s):** domain (new), web, docs.

## Scope

**In scope:**
- New `internal/domain` layer: typed accessors over `save` + an item catalog.
- Item catalog (Option A): embedded `data/items.json` (CID → name/category) + category
  inference from CID prefix + user-editable labels persisted to a local overrides file.
- `/api/game/*` endpoints.
- UI split: top-level **Editor | Database** tabs. "Database" = the current grid,
  unchanged. "Editor" = new panels.
- Two Editor panels: **Currency** (`tb_currency.AMOUNT`) and **Consumables**
  (`tb_stackable_item.STACK_CNT` + `tb_cook_item.STACK_CNT`, grouped by category).

**Out of scope (later slices / plans):**
- Characters, Team, Equipment, Gems, Costumes panels.
- Adding brand-new items (insert rows), presets, "max all", import/export, diff view.
- Full pak-based item catalog (separate RE chantier — see memory `dsa-pak-protection`).

## Working-tree state

Clean (see `phase0_results.md`). No `internal/domain`, no `data/items.json`. The web
layer currently serves only the generic grid.

## Design

```
internal/sqlcipher → internal/save → internal/domain → internal/web
                                        └─ catalog: data/items.json (go:embed) + inference + user overrides
```

- `internal/domain/catalog.go` — `Item{CID, Name, Category, Known}`; `Catalog` loads the
  embedded seed + optional user overrides; `Lookup(cid)` resolves name via
  override → items.json → category inference (prefix map); `SetLabel(cid, name)` writes
  the overrides file (in the user config dir).
- `internal/domain/domain.go` — `Game{save, catalog}` with `Currencies()`,
  `SetCurrency(cid, amount)`, `Consumables()` (stackable + cook, tagged by category),
  `SetStack(kind, id, count)`.
- `data/items.json` — small seed (known currencies, a few obvious items); grows over time.
- Category inference (CID 3-digit prefix): 1000=currency, 141=potions, 142=cooked food,
  143/145/146/147=materials/misc, 151=misc. Unknown → "Item <cid>".
- `internal/web` — `/api/game/currency` (GET/POST), `/api/game/consumables` (GET),
  `/api/game/stack` (POST), `/api/game/label` (POST); frontend gains an Editor|Database
  tab switch; Editor renders Currency + Consumables panels.

## Atomic steps (each = 1 commit, `just ci` green)

1. **Catalog** — `internal/domain/catalog.go` + `data/items.json` + `catalog_test.go`
   (inference + override precedence). Commit: `feat(domain): add item catalog (embedded names + prefix inference + user labels)`.
2. **Game accessors** — `internal/domain/domain.go` + tests against a real save (`DSA_SAVE`).
   Commit: `feat(domain): add typed currency and consumables accessors`.
3. **Game API** — `/api/game/*` handlers in `internal/web`. Commit: `feat(web): add /api/game endpoints for the editor view`.
4. **UI shell** — Editor|Database tab switch; keep the grid as "Database". Commit: `feat(web): split UI into Editor and Database tabs`.
5. **Currency panel** — list of currencies (name + amount input). Commit: `feat(web): add currency editor panel`.
6. **Consumables panel** — grouped stackable + cook items with quantity steppers and label editing. Commit: `feat(web): add consumables editor panel`.
7. **Docs** — update `README.md` and `DESIGN.md` (new `internal/domain` layer + the two views). Commit: `docs: document the domain layer and the Editor/Database views`.

## Quality gates

- [ ] `nix develop --command just ci` passes at each step.
- [ ] `DSA_SAVE=… just test` green (domain accessors round-trip against a real save).
- [ ] Manual: Editor loads, Currency edit persists (verified via reopen), Consumables
      qty edit persists, Database tab still works. Recorded in `manual_tests.md`.
- [ ] Docs updated in the same commits as the code they describe.
- [ ] Atomic commits on `feat/editor-mvp`; user merges to `main`.

## Decisions

- **Option A for names** (embedded json + inference + labels), not pak extraction.
- **Editor is the default tab**, Database kept as the advanced escape hatch — the
  generic layer is not thrown away.
- **User label overrides** live in the OS config dir (e.g. `~/.config/dsa-save-editor/`),
  not in the repo and not in the save.
