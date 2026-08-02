## Plan: Browse the full th.gl stackable catalog and set any count 0 → X

**Type:** UI
**Objective:** In the **Consumables** panel, let the user browse **every th.gl-known
stackable item** (materials + ingredients), owned or not, and set any count from
`0` to `X` — instead of only listing owned stacks and searching a material by name
to add it.
**Why:** The user hasn't finished the game, so their save only holds the stackables
they've already picked up. The bundled catalog, however, is the **complete** th.gl
scrape (179 materials/ingredients). Today the panel only shows *owned* stacks, and
adding an unowned one means knowing its name and typing it into a datalist. There is
no way to *see* what you don't own yet and give it to yourself. This surfaces the
whole catalog as editable rows.
**Layer(s):** web (+ docs)

### Context / findings (audit)

- The catalog (`internal/domain/data/items.json`, 890 entries) is already the full
  th.gl scrape, **not** derived from the save. th.gl exposes exactly 10 DB routes
  (`characters, costumes, equipment, ingredients_db, materials_db, monsters, mounts,
  npcs, quests, recipes`); `cmd/gen-catalog` already scrapes the 7 item-bearing ones.
- **Stackable coverage is complete:** `materials_db` (131) + `ingredients_db` (48)
  = 179 = every `category:"material"` entry in the catalog. No pagination is dropped.
- th.gl publishes **no** potions (`141xxxx`) nor currencies (`1000xxx`) page, so
  "sync everything from th.gl" cannot surface them. **Decision (user):** stay
  **th.gl-only** and **stackables-only** — we do not mine or curate potion/currency
  lists, do not add currencies, and do not touch instance items (gear/mounts/costumes/
  characters, which would need synthesised 64-bit DBIDs).
- The support already exists end-to-end: `domain.AddOrSetStackable` (upsert into
  `tb_stackable_item`), `POST /api/game/stackable`, and the "Add a material" toolbar.
  `tb_stackable_item` is `(USER_DBID, ITEM_CID)` PK + `STACK_CNT`, 3 columns, no
  triggers → the existing `INSERT OR REPLACE` upsert is safe (confirmed v0.5.0).
- **Conclusion:** this is a pure **web-layer surfacing** change. No domain, API, save,
  or catalog change is required — the frontend already fetches `/api/game/catalog`
  and `/api/game/consumables` and can merge them client-side.

### Scope

**In scope**
- The Consumables panel **always** lists every `category:"material"` catalog entry
  (owned + not owned) as an editable stepper row, value pre-filled with the owned
  count (or `0`), committing via the existing `POST /api/game/stackable` (upsert
  0 → X). No toggle — the full list is shown all the time.
- Owned potions/food keep showing as they do today (owned rows).
- Deliberately **no UX polish** here (no filter/search, no collapsing, no owned/
  not-owned styling beyond what falls out naturally). A follow-up plan handles UX.

**Out of scope** (per user decisions)
- Potions / currencies absent from th.gl (no catalog source; no mining/curation).
- Adding currencies, cooked food (instances), or instance items (gear/mounts/
  costumes/characters).
- **UX** (search/filter, collapsing, owned/not-owned visual treatment) — a separate
  follow-up plan.
- Any `cmd/gen-catalog`, `internal/domain`, `internal/save` or API change.

### Files touched
- [ ] `internal/web/static/app.js` — browse-all-stackables view + toggle
- [ ] `internal/web/static/style.css` — minor styling if needed (owned/unowned hint)
- [ ] `.claude/plans/v0.9.0/manual_tests.md`
- [ ] `docs/` — no format fact changes; none expected (UI-only). Confirm at the end.

### Atomic steps

#### Step 1: Always list the full stackable catalog in the Consumables panel
**Description:** In `renderConsumables()`, replace the owned-only **material** group
with the **full** catalog: render every catalog entry with `category === "material"`,
sorted by name, each as a `stepperRow`:
- `value` = owned `STACK_CNT` if the CID is present in `/api/game/consumables`
  (`kind === "stackable"`), else `0`.
- `onCommit(v)` → `postJSON("/api/game/stackable", { cid, count: v })` then re-render.
- Reuse `displayName`, icons and the `known` flag exactly like existing rows.
Build the owned-count map from the existing `/api/game/consumables` response; reuse
`catalog()` (cached). No new network endpoint, no toggle. The **food** and **potion**
groups keep listing owned instances only (unchanged). The "Add a material" datalist
toolbar becomes redundant (all materials are now rows) and is **removed**; the
"set all stacks to N" fill control is kept.
**Verification:** `nix develop --command just ci` (JS is static-embedded, so mainly
fmt/vet/build) + manual tests M1–M3.
**Commit:** `feat(web): list the full th.gl stackable catalog with editable counts`

#### Step 2: Docs sync check
**Description:** UI-only change reveals no *new* save-format fact, so `docs/` should
not need edits. Re-read `docs/content-ids.md` / `docs/database.md`; if the stackable
"owned vs catalog" distinction is worth a sentence, add it in the same spirit.
Update `README.md` feature list if it enumerates editor capabilities.
**Verification:** `just ci`.
**Commit:** `docs: note browsing the full stackable catalog` (only if a doc changes)

### Technical decisions
- **Web-only, client-side merge.** All data is already served by
  `/api/game/catalog` + `/api/game/consumables`; adding a domain method or endpoint
  would duplicate logic for no gain. KISS.
- **Reuse `/api/game/stackable` (upsert).** 0 → X and X → Y both go through the
  safe `AddOrSetStackable` upsert; owned rows edited from this view therefore also
  route through it (equivalent result, 3-column table, no data loss).
- **Always-on, no toggle.** Per user: show the full stackable catalog all the time;
  managing the resulting density (filter/collapse) is deferred to a UX follow-up plan.
- **Stackable = `category:"material"`.** In the catalog only materials/ingredients
  map to `tb_stackable_item`; food is cook-instance, the rest are instance items.
  Owned potions still appear via the normal owned-rows path.

### Quality gates
- [ ] `just ci` passes
- [ ] Real-save round-trip still green (`DSA_SAVE=… just test`) — unchanged layers,
      but re-run to confirm no regression
- [ ] Docs synced (same commit) — or confirmed no doc change needed
- [ ] Atomic commits on `feat/browse-stackable-catalog`
- [ ] Manual tests M1–M3 pass (see `manual_tests.md`)
