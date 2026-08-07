## Plan: Inventory super-categories (grouped category rail)

**Type:** UI (+ cmd data pipeline)
**Objective:** Group the inventory (Consommables) category rail under a handful of
super-category headers, so the flat list of ~50 categories reads as ordered
sections ("Cuisine", "Runes", "Effets", "Ingrédients", "Matériaux", "Objets de
valeur", "Autres").
**Why:** The rail lists every game item-category flat. The user cannot tell at a
glance that "Griller / Bouillir / Découper" are all cooking, or that the five rune
categories belong together. The grouping already exists in the game data — it is
just discarded.
**Layer(s):** `cmd`, `web`, `docs`

### Key finding

Every `GameItemCategoryData` row carries a `CategoryType`. That field **is** the
super-category. Today `cmd/pak-catalog` (`main.go:326`) reads it only to pick a UI
colour (`consumableCatColor`, `main.go:60`) and throws the type itself away. There
are exactly **6** surfaced types, and each maps **1:1 to a colour** already:

| CategoryType | colour | super-category (FR / EN) |
|---|---|---|
| `COOK` | `#cf8f6f` | Cuisine / Cooking |
| `GEM` | `#b98ce0` | Runes / Runes |
| `KARMA` | `#d75f8f` | Effets / Effects |
| `NORMAL_MATERIAL` | `#6fcf7f` | Ingrédients / Ingredients |
| `GROW_MATERIAL` | `#e08a5a` | Matériaux / Materials |
| `VALUABLE` | `#e0a44a` | Objets de valeur / Valuables |
| `unsorted` (`#8a93a6`) | — | Autres / Other (trailing) |

**Decision — COOK ≠ NORMAL_MATERIAL:** kept as two distinct groups. COOK items are
consumable ("potion-like": heal / buffs); NORMAL_MATERIAL is raw food that is not
usable until cooked. Different game type, different group.

### Working-tree state

- Clean on `main` at `v0.16.1` (phase 0: `just ci` green — see `phase0_results.md`).
- `item_categories.json` has no `group` field yet; the rail renders flat.

### Technical decisions

1. **`group` is derived from the paks, authoritative in the generator.**
   `cmd/pak-catalog` emits the `CategoryType` as a new `group` field on each
   category. The generator stays the source of truth.
2. **Regeneration needs local paks (not committed).** `go run ./cmd/pak-catalog`
   consumes decrypted XML fixtures the repo does not carry. So the committed
   `item_categories.json` is filled **deterministically via the inverse of the
   colour→type map** — bit-for-bit identical to what a real regen produces, since
   the 7 colours are distinct. A future real regen reproduces it exactly.
3. **The 6 super-category labels + their order are hand-authored, in Go.** This is
   the only non-derived piece; it also carries the display order. Lives in
   `internal/domain/consumable_category.go` as a fixed ordered slice. The enum keys
   (`COOK`…) are English-technical and the paks do not localize them, hence the
   hand table. Labels are provisional — revisited in manual test.
4. **Headers are static** (not collapsible/clickable): the ask is purely visual.
5. **Trailing "Autres" group** collects any category whose `group` is not one of
   the six known types (today: only `unsorted`).

### Files touched
- [ ] `cmd/pak-catalog/main.go` — emit `group` on each emitted category
- [ ] `internal/domain/data/item_categories.json` — regenerated with `group`
- [ ] `internal/domain/consumable_category.go` — `Group` field + `ConsumableGroup`
      type + ordered list + `ConsumableGroups()` accessor
- [ ] `internal/domain/consumable_category_test.go` — group coverage assertions
- [ ] `internal/web/game.go` — handler returns `groups` alongside `categories`
- [ ] `internal/web/static/app.js` — grouped rail with headers
- [ ] `internal/web/static/style.css` — `.cat-group-header`
- [ ] `DESIGN.md` — note the category→super-category grouping
- [ ] `.claude/plans/v0.17.0/manual_tests.md`

### Atomic steps

#### Step 1: Tag categories with their CategoryType group (generator + data)
**Description:** Add `Group string \`json:"group"\`` to `consumableCategory` in
`cmd/pak-catalog`; set it from `r.CategoryType` at emit, `"unsorted"` for the
fallback. Fill the committed `item_categories.json` accordingly (colour→type
inverse). Data + generator in the same commit.
**Verification:** `nix develop --command just ci`; spot-check the JSON groups match
the colour table above.
**Commit:** `feat(cmd): tag consumable categories with their CategoryType group`

#### Step 2: Expose super-category groups from the domain + API
**Description:** Add `Group` to `domain.ConsumableCategory`; add `ConsumableGroup`
{Key, LabelFR, LabelEN} + the ordered slice + `ConsumableGroups()`. Handler
`handleConsumableCategories` returns `{categories, groups}`. Add a domain test:
every category's `group` is a known group key or `unsorted`; the six known groups
are present and ordered.
**Verification:** `nix develop --command just ci`.
**Commit:** `feat(web): expose consumable category groups via the API`

#### Step 3: Render the rail under super-category headers
**Description:** In `RENDER.inv()`, iterate `groups` in order; for each group with
≥1 non-empty category, emit a `.cat-group-header` (group label, tinted by the
group colour) then that group's `.cat-link` buttons. Ungrouped categories fall
under a trailing "Autres / Other" header. Style `.cat-group-header`. Update
`manual_tests.md` and the `DESIGN.md` note.
**Verification:** `nix develop --command just ci`; manual visual review in the app.
**Commit:** `feat(web): group the inventory rail under super-category headers`

### Quality gates
- [ ] `just ci` passes on each commit
- [ ] Real-save round-trip still green (`DSA_SAVE=… just test`) — unaffected, but run
- [ ] Docs synced (DESIGN.md note in step 3's commit)
- [ ] Atomic commits on `feat/inventory-supercategories`
- [ ] Manual visual review of the grouped rail (headers, order, "Autres")
