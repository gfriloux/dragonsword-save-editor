## Plan: Curated consumable categories + Category-sidebar Consumables panel (direction B)

**Type:** UI + edit-view (domain classification + web)
**Objective:** Replace the flat Consumables list with a **category-sidebar** layout
(mockup "direction B"): a second nav rail lists curated functional categories
(Ingrédients, Percée, Runes, Cristaux, XP équipement, Potions, Plats cuisinés, Non
trié…); selecting one shows just that category's items (owned + not-owned from the
th.gl catalog), each editable 0 → X.
**Why:** Since v0.9.0 the panel lists ~179 materials in one flat "material" group —
too dense to navigate. The user wants functional grouping to find and top up items
fast. Categorisation is **curated** (chosen over data-driven: th.gl exposes no item
sub-typing) and **prudent** (anything uncertain goes to "Non trié", never mis-sorted).
**Layer(s):** domain, web, docs

### Context / findings (audit + real-save harvest)

Harvested the real save (`6144_Slot1.db`, 144 owned consumable rows; dump kept at
`tmp/dump/out.txt`) and cross-referenced `internal/domain/data/items.json`:

- th.gl provides **no** functional sub-typing of materials → categories must be
  **curated by us** (decision: "curated fonctionnel").
- The `145x` prefix is a **mixed bag**, NOT uniformly "percée": monster parts
  (`14500xx`) and plants (`14505xx`) are breakthrough, but `14501xx` are weapon
  enhancement stones ("Pierre de …"), `14502xx` faction badges, `14504xx`
  grimoires/memory, `14506xx` reminiscence crystals, `14508xx` mixed boss drops.
  ⇒ classification must be **explicit ranges + CID sets**, not a coarse prefix.
- **New confirmed save-format facts** (to document):
  - `14102xx` inside the `141x` range are **mana upgrade materials**, not potions:
    `1410202` Fragment de mana, `1410203` Cristal de mana, `1410204` Minéral de mana
    (confirmed by the user's counts 9999 / 9994 / 8683). The rest of `141x`
    (`1410002`–`1410105`) are real potions.
  - `14505xx` plants (`1450501` Fruit de la vitalité, `1450502` Graine primordiale,
    `1450503` Goutte de pureté, `1450504` Feuille de vigueur) are breakthrough mats.
  - Off-th.gl items exist and have **no catalog name** (only CIDs): `1000800/801/802/804`
    are the likely character-XP books (3 manuals + Livre du Héros), but this is
    **unverified** → they stay "Non trié" for now (user decision, prudence).

### Locked taxonomy (v0.10.0)

Confident categories (encoded); everything not matched falls through to **Non trié**:

| Key           | FR / EN label              | Rule (CID)                                   |
| ------------- | -------------------------- | -------------------------------------------- |
| `cooked`      | Plats cuisinés / Cooked    | `142xxxx` (cook instances, owned)            |
| `potion`      | Potions / Potions          | `141xxxx` **except** `14102xx`               |
| `equip_xp`    | XP équipement / Gear XP    | `1410202`, `1410203`, `1410204` (mana)       |
| `ingredient`  | Ingrédients / Ingredients  | `143xxxx`, `144xxxx`                          |
| `breakthrough`| Percée / Breakthrough      | `1450001`–`1450018`, `1450501`–`1450504`     |
| `crystal`     | Cristaux / Crystals        | `146xxxx`                                    |
| `rune`        | Runes / Runes              | `131xxxx`                                    |
| `unsorted`    | Non trié / Unsorted        | default (everything else)                    |

Not created yet (no confirmed members): **XP personnage** — pending the user
identifying the book CIDs in-game. Adding it later is a one-line curated change.

### Scope

**In scope**
- A curated classifier `domain.ClassifyConsumable(cid) → category key`, plus the
  ordered category list (key + FR/EN label + display order + color), with a unit test.
- API: attach the functional category key to each consumable/catalog item and expose
  the category list.
- UI: direction-B Consumables panel — category sidebar with per-category owned/total
  counts; the item pane shows the selected category's rows (owned + not-owned),
  editable 0 → X via the existing `/api/game/stackable` (upsert) / `/api/game/stack`.
- Docs: `docs/content-ids.md` (mana range, `145x` sub-families, XP-book candidates),
  a short "consumable categories" note, `README.md` capability line.

**Out of scope**
- Data-driven categories (th.gl has none), currencies, instance items.
- The XP-personnage category (pending user verification of book CIDs).
- Per-item user category overrides (possible later; not needed now).
- Mockup extras not in direction B (search-first grid, compact rows) — a later UX pass.

### Files touched
- [ ] `internal/domain/consumable_category.go` (+ `_test.go`)
- [ ] `internal/domain/domain.go` (attach category key to `Stack`/catalog items if needed)
- [ ] `internal/web/game.go` (category list endpoint + category on payloads)
- [ ] `internal/web/web.go` (route)
- [ ] `internal/web/static/app.js`, `index.html`, `style.css` (direction-B panel)
- [ ] `docs/content-ids.md`, `README.md`
- [ ] `.claude/plans/v0.10.0/manual_tests.md`

### Atomic steps

#### Step 1: Curated classifier in domain (+ doc of the new facts)
**Description:** Add `internal/domain/consumable_category.go`: the ordered category
table (key, LabelFR, LabelEN, Color) and `ClassifyConsumable(cid int64) string`
implementing the locked taxonomy (explicit ranges + the mana/breakthrough CID sets),
defaulting to `unsorted`. Table-driven unit test covering each rule and the
`14102xx`-vs-potion and `145x` sub-family edge cases. Update `docs/content-ids.md`
with the mana range, the `145x` sub-family map and the (unverified) XP-book CIDs.
**Verification:** `just ci` (unit test green).
**Commit:** `feat(domain): classify consumables into curated functional categories`

#### Step 2: Expose categories in the API
**Description:** In `internal/web/game.go`, add `GET /api/game/consumable-categories`
returning the ordered category list (key/labels/color). Attach the functional
category key to items returned by `/api/game/consumables` and `/api/game/catalog`
(or compute client-side from a shared cid→key — but server-side keeps one source of
truth). Register the route in `web.go`.
**Verification:** `just ci`; hit the endpoint manually.
**Commit:** `feat(web): expose curated consumable categories in the API`

#### Step 3: Direction-B Consumables panel (category sidebar)
**Description:** Rebuild `renderConsumables()` per mockup B: a category rail (colored
dot + label + `owned N / total`), a selected-category item pane with its own toolbar
(category title, owned/total, a per-category Fill, optional owned/missing filter).
Merge owned counts (`/api/game/consumables`) with the full catalog
(`/api/game/catalog`), group by the functional category key, sort owned-first,
dim not-owned rows. Reuse `stepperRow`, icons, `cat-dot`; add minimal CSS for the
rail. Keep the app header/tabs/editor-nav chrome untouched.
**Verification:** `just ci` + manual tests M1–M5 on the real save.
**Commit:** `feat(web): category-sidebar Consumables panel (direction B)`
(README capability line updated in this commit — doc-in-same-commit.)

### Technical decisions
- **Explicit ranges + CID sets, default `unsorted`.** The `145x` mess makes prefix
  rules unsafe; prudence beats coverage (user: "plutôt prudent que mal trié").
- **Classifier lives in `domain`.** It's game knowledge; the web layer only renders.
  One source of truth; unit-tested deterministically (no save needed).
- **Reuse existing upsert/update endpoints.** No save-write change; the panel is a
  new view over the same edit paths (`/api/game/stackable`, `/api/game/stack`).
- **Direction B** chosen by the user from the three mockups (`tmp/mockups/`).
- **Curated map is intentionally incomplete.** Unknowns surface in "Non trié" where
  the user can still edit them; the map grows as items are identified in-game.

### Quality gates
- [ ] `just ci` passes (incl. new domain unit test)
- [ ] Real-save round-trip green (`DSA_SAVE=… just test`)
- [ ] Docs synced in the same commits (content-ids facts with Step 1; README with Step 3)
- [ ] Atomic commits on `feat/consumable-categories`
- [ ] Manual tests M1–M5 pass (see `manual_tests.md`)
