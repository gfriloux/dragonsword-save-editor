## Plan: Cooking recipe details + per-recipe unlock (Cuisine)

**Type:** edit capability + UI + data (deferred Phase 3 of v0.12.0)
**Objective:** Replace the blanket "Tout débloquer" with a real Cuisine screen: every
recipe listed (tool, required ingredients resolved to names/icons, produced dishes),
its **known / unknown** state read from the save, and a **per-recipe** unlock/lock plus a
key-accurate "unlock all". The blanket switch stays as a fallback only.
**Why:** DESIGN.md flags this as the next increment; the data (`CookRecipeData.xml`,
`CookToolData.xml`) is now pure-Go extractable. The current blanket unlock is also
**incorrect**: it writes `tb_switch` categories 15–60, but 9 real recipes
(`CookBook_SwitchData` 4000–4008) map to **category 62** and are never unlocked. A
key-accurate map fixes that.
**Layer(s):** cmd, domain, web, docs (no sqlcipher/save/pak/oodle change)

---

### Context — data findings (this session, from `tmp/pak/CookRecipeData.xml`)

- **1025 recipes**, each row carries: `CookBook_SwitchData` (the recipe's switch key),
  `ToolType` (FRYINGPAN / POT / KNIFE — 364 / 330 / 331), up to **3 ingredient
  conditions** (`IngredientCondN_Type` ∈ {`INGREDIENT_TYPE`, `INGREDIENT_ID`} +
  `IngredientCondN_Value`), and `Cook_ID1..5` — the 5 dish CIDs (grade/star tiers).
- **`CookToolData.xml`** = the 3 tools (id, ToolType, Korean memo, `Name` string key, icon).
- Dish CIDs (`Cook_ID*`, e.g. `1420101`) and specific-ingredient CIDs
  (`INGREDIENT_ID`, e.g. `1430603`) already **resolve in `items.json`** → real names/icons.

**Switch-key formula (working hypothesis, matches the single ground truth):**
```
category = CookBook_SwitchData / 64        (integer division)
bit      = CookBook_SwitchData % 64
```
For "Poisson grillé" (`CookBook_SwitchData=1002`) this gives **category 15, bit 42** — the
one in-game before/after data point from `docs/switches.md`. This is *algebraically the
same* mapping the v0.12.0 plan already anticipated (960 = 15×64); it is **still a single
data point** and MUST be validated in-game before the UI is built (Phase 1).

**Two facts this chantier acts on:**
- `CookBook_SwitchData` spans 1001–4008 → categories **15–62**. The current blanket
  (15–60) **misses 9 recipes** at category 62 (`SwitchData` 4000–4008). Precise unlock
  covers them.
- The 3 "special" dishes from `docs/switches.md` (iced drinks `1999xxx`, `1423001`,
  `1430920`) do **not** appear as `Cook_ID*` in `CookRecipeData.xml` → they live in a
  different table and stay **out of scope** here (documented, not silently dropped).

---

### Scope

**In:**
- Dev-side extraction: `cmd/pak-catalog` emits `internal/domain/data/recipes.json`
  (committed data, like `items.json`) — per recipe: switch key + derived (category, bit),
  tool, ingredient conditions (kind + value), dish CIDs per tier.
- domain: load recipes; read per-recipe **known/unknown** from `tb_switch`; **per-recipe**
  unlock + lock; key-accurate "unlock all" (covers category 62); resolve ingredients to
  bilingual names and, for `INGREDIENT_ID`, **owned counts** from the save's stackables.
- web: Cuisine screen — recipe list filterable by tool, known/unknown badges, required
  ingredients (name/icon + owned/needed for specific ones), produced dish, per-recipe
  toggle + "Tout débloquer" (now key-accurate).
- docs: correct `docs/switches.md`, update DESIGN.md/README.md, CHANGELOG.

**Out:**
- The 3 special recipes (`1999xxx`, `1423001`, `1430920`) — not in this table; a separate
  investigation, documented as uncovered.
- Editing recipe *contents* or dish stats; cooking simulation. We only flip the
  known/unknown switch and display the (read-only) recipe data.
- Any change to `sqlcipher` / `save` / `pak` / `oodle` (icons already work via `/api/icon`).

---

### Files touched
- [ ] `cmd/pak-catalog/main.go` (parse CookRecipe/CookTool XML → emit `recipes.json`)
- [ ] `internal/domain/data/recipes.json` (generated, committed — data, not art)
- [ ] `internal/domain/recipes.go` (+ `recipes_test.go`) — catalog load + resolution
- [ ] `internal/domain/domain.go` — per-recipe known/unlock/lock; key-accurate UnlockAll
- [ ] `internal/domain/*_test.go` — real-save gated tests (`DSA_SAVE`)
- [ ] `internal/web/game.go` + `web.go` — recipe list + per-recipe endpoints
- [ ] `internal/web/static/{index.html,style.css,app.js}` — Cuisine screen
- [ ] `internal/web/manual_tests.md` → `.claude/plans/v0.13.0/manual_tests.md`
- [ ] `docs/switches.md`, `DESIGN.md`, `README.md`, `CHANGELOG.md`

---

### Atomic steps

#### Phase 0 — audit
`nix develop --command just ci` → record in `phase0_results.md`. Confirm the tree is green
before touching anything.

#### Phase 1 — Validate the switch-key formula in-game (spike, user-run) ← gate
Before any UI/domain code, confirm the mapping on a **test save** (never the real save in
place; work on a copy). Two checks, each a throwaway save the user opens in-game:
1. **Single recipe:** set exactly `tb_switch` (category = key/64, bit = key%64) for **one**
   currently-unknown recipe (e.g. `SwitchData=1002`, "Poisson grillé"). In-game: exactly
   that recipe becomes known, nothing else.
2. **Category 62 tail:** set the 9 bits for `SwitchData` 4000–4008 (category 62). In-game:
   those recipes appear as known, no side effects.

Deliver as a tiny throwaway CLI/SQL under `tmp/` (not committed). **The user runs the game
and reports.** Record the outcome in `phase0_results.md`. If the formula does **not** hold,
stop and re-plan (fall back to blanket + document); do not build the UI on a wrong map.
**No commit** (spike only) unless a reusable helper is worth keeping.

#### Phase 2 — Extract recipes.json (cmd + data)
Extend `cmd/pak-catalog` (or a sibling parser it already owns) to read
`CookRecipeData.xml` + `CookToolData.xml` from `tmp/pak/` and emit
`internal/domain/data/recipes.json`. Per recipe:
```json
{ "key": 1002, "category": 15, "bit": 42, "tool": "FRYINGPAN",
  "ingredients": [ {"kind":"type","value":1701} ],
  "dishes": [1420102,1420202,1420302,1420402,1420502] }
```
Plus a small `tools` block (id, ToolType, name key) and, if resolvable, an
`ingredientTypes` map (INGREDIENT_TYPE value → label — see open investigations). Commit the
JSON (game *data*, like items.json; not copyright art).
**Commit:** `feat(cmd): pak-catalog emits recipes.json`.

#### Phase 3 — domain: recipe catalog + known state + precise unlock
- `internal/domain/recipes.go`: `go:embed` recipes.json; typed `Recipe`, `RecipeCatalog`;
  resolve tool label, ingredient names (via existing `Catalog.Lookup`), dish name.
- Read **known/unknown** per recipe from `tb_switch` (one batched read of the recipe
  categories, test each recipe's bit).
- `SetRecipeKnown(key, known bool)` — flip one bit in the right category row (read-modify-
  write `BIT_FIELD`, upsert). `UnlockAllRecipes` rewritten to OR the exact bits of every
  known recipe key (covers category 62); keep the blanket only as an internal fallback if
  recipes.json is somehow empty.
- For `INGREDIENT_ID` conditions, surface **owned counts** from the save's stackables
  (reuse the consumables path). `INGREDIENT_TYPE` conditions resolve to a category label.
- Tests gated on `DSA_SAVE`: round-trip a single-recipe unlock, assert the exact
  (category,bit) changed and re-open is valid; assert `UnlockAllRecipes` sets category 62.
**Commit:** `feat(domain): per-recipe cooking known-state + key-accurate unlock`.

#### Phase 4 — web: Cuisine screen
- API (additive): `GET /api/game/recipes` (list: key, tool, known, ingredients resolved,
  dish name/icon), `POST /api/game/recipes/known` (`{key, known}`). Keep
  `/api/game/recipes/unlock-all` (now key-accurate).
- UI: recipe list filterable by tool (3 tabs / segmented control), known/unknown badge,
  required ingredients (name + `/api/icon?cid=` + owned/needed for specific ones), produced
  dish, per-recipe toggle, "Tout débloquer". Dirty-state stays derived; JSON style matches
  existing panels. Grow `manual_tests.md` (visual + toggle round-trip).
**Commits:** `feat(web): Cuisine recipe list + per-recipe unlock`, then refinements.

#### Phase 5 — docs + release
- `docs/switches.md`: correct the recipe mapping to `category = SwitchData/64,
  bit = SwitchData%64`; record the **15–62** span and the blanket-15–60 gap it fixes;
  state the 3 specials are a different table (uncovered).
- DESIGN.md: move cooking details out of "Out of scope"; README Cuisine bullet updated.
- `CHANGELOG.md` via `just changelog`. `just build-windows` static `.exe` gate.
**Commit:** `docs: cooking recipe details + corrected switch mapping` (plus in-same-commit
doc edits done inline in earlier phases).

---

### Technical decisions
- **Recipes = committed data** (extracted dev-side like `items.json`), consistent with the
  v0.12.0 split (data committed, icons per-user). No game *art* committed.
- **Per-recipe unlock via exact (category,bit)**, replacing the blanket range — it is both
  finer-grained *and* strictly more correct (the 15–60 blanket misses category 62).
- Blanket "unlock all" kept as UX ("Tout débloquer") but reimplemented over the real key
  set; the old fixed 15–60 loop is removed.
- API additive; existing endpoints unchanged. Icons reuse `/api/icon?cid=`.

### Open investigations (resolve within their phase, not blocking the plan)
- **INGREDIENT_TYPE → label**: what the values 1700–1707 mean (meat/fish/seafood/egg/
  vegetable/mushroom/rice/herb?). Check whether `GameItemData.xml` Category / a StringData
  key resolves them; if not, ship a small curated bilingual map (like item categories) and
  note the source. (Phase 2/3.)
- **Ingredient quantity**: the XML has up to 3 condition *slots*, no explicit qty field; a
  repeated type/id likely means "×2/×3". Confirm the intended reading; display counts
  accordingly. (Phase 3.)
- **Dish tiers / stars**: `Cook_ID1..5` are almost certainly the grade tiers of the dish;
  confirm and decide whether the UI shows one dish or the tier ladder. (Phase 4.)
- **The 3 specials**: locate their real table (drinks?) for a possible follow-up; out of
  scope here. (Docs only.)

### Risk register
- **Formula is one data point.** Phase 1 is the gate; nothing downstream is built until the
  user validates in-game. If it fails, fall back to blanket + document, no UI churn.
- Category 62 write is new territory for the blanket — Phase 1 check #2 de-risks it.
- Game update could renumber `CookBook_SwitchData`; recipes.json is regenerated from the
  paks, so only a re-extract would be needed.

### Quality gates
- [ ] `nix develop --command just ci` green at each commit
- [ ] `DSA_SAVE=… just test` round-trip still green (incl. new recipe-unlock test)
- [ ] Phase 1 formula validated in-game by the user before Phase 3 lands
- [ ] `just build-windows` produces a single static `.exe` (CGO_ENABLED=0)
- [ ] No game art / no personal save committed; recipes.json is data only
- [ ] Docs synced in the same commit; atomic commits on `feat/cooking-recipes`
