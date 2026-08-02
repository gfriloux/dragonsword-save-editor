## Plan: Native item names from the paks (`cmd/pak-catalog`) — Phase 2 of pak-extraction

**Type:** tooling + data (catalog refresh)
**Objective:** Merge the game's own item data (extracted from the paks as raw XML) into
`internal/domain/data/items.json`, giving **native FR/EN names to all ~1212 items** — including
the off-th.gl ones (mana, XP manuals, awakening gems, …) — plus each item's authoritative
`ItemType` for later category work. Keep th.gl icons/sprite.
**Why:** th.gl covers only 890 items and no potions/mana/etc. The paks ship the full server
dataset as plain XML (see `.claude/plans/release/pak-extraction/plan.md`; extractor DONE).
**Layer(s):** cmd (+ domain/data), docs

### Context (from the extraction spike, artifacts in `tmp/pak/`)
- `GameItemData.xml` (1212 items): row = `ID`(=CID), `ItemType`, `Category`(numeric), `Grade`,
  `Name`/`Desc` = string keys, `IconName`.
- `StringData.xml`: string key → all 11 langs incl. **`Fr` and `En`** (native French).
- Validated: `1450410`→"Gemme d'éveil: Stigmate"/"Stigma Awaken Stone"; `1410002`→CHARACTER_EXP
  "Manuel des bases du combat" (so the "potions" `1410xxx` are actually **XP manuals**, not potions).
- **Characters (`10xxx`, e.g. Eileen) are NOT in GameItemData** (separate `PCCharacterData.xml`).
  ⇒ this must be a **merge** that preserves existing th.gl entries, not a replace.

### Scope
**In:** `cmd/pak-catalog` (Go) reads the extracted `GameItemData.xml` + `StringData.xml` and the
current `items.json`; writes a merged `items.json`: pak FR/EN override/augment names, new CIDs
added, **existing entries not in the paks (characters…) preserved**, th.gl icon `x/y` kept, a new
`type` field (ItemType) stored per item. Coarse `category` for new items mapped from `ItemType`.
Docs: `docs/paks.md` (no longer a wall — extraction recipe), note in `content-ids.md`.
**Out:** the functional-category auto-derivation + retiring `ClassifyConsumable`/`items_extra.json`
(that's **Phase 3** / next version); character/equipment stat tables; icons from paks (keep th.gl).

### Files touched
- [ ] `cmd/pak-catalog/main.go`
- [ ] `internal/domain/data/items.json` (regenerated, committed)
- [ ] `internal/domain/catalog.go` (only if we surface `type`; likely NO change — extra JSON
      field is ignored by the current parser)
- [ ] `docs/paks.md`, `docs/content-ids.md`
- [ ] `Justfile` (a `pak-catalog` recipe, optional)

### Atomic steps
#### Step 1: `cmd/pak-catalog` tool
Parse both XML (namespace-robust, match attr local names). Build stringKey→(Fr,En). For each
GameItemData row: resolve names; merge into the in-memory catalog (keep x/y + existing category
for known CIDs; add `type`; for new CIDs set category via an ItemType→coarse map and x=y=0). Write
`items.json` with the same shape/format as `gen-catalog`. Inputs via flags (default `tmp/pak/…`).
**Verification:** `go run ./cmd/pak-catalog`; spot-check known CIDs; `just ci` (build/vet/lint).
**Commit:** `feat(cmd): add pak-catalog to merge native item names from the paks`

#### Step 2: Regenerate items.json + doc the pipeline
Run the tool, commit the regenerated `items.json` (now ~1212 items with native FR/EN + `type`).
Update `docs/paks.md` (wall solved; the XML/extractor recipe) and `content-ids.md` (141xxxx =
CHARACTER_EXP manuals, mana = EQUIPMENT_EXP). Confirm `just ci` + the domain catalog tests pass
(esp. `TestEntries` 10001=Eileen still present; adjust only if a name legitimately changed).
**Commit:** `chore(domain): regenerate items.json with native names from the paks` (+ docs)

### Technical decisions
- **Merge, never replace.** Characters and any th.gl-only entries must survive.
- **Pak FR/EN authoritative** over th.gl (native, consistent). Keep th.gl **icons** (pak icon
  extraction is out of scope).
- **`type` stored now, used in Phase 3.** The current domain parser ignores unknown JSON fields,
  so no domain change is required yet.
- Extracted XML are **game-copyright fixtures** → stay in `tmp/` (gitignored), never committed.

### Quality gates
- [ ] `just ci` passes
- [ ] `TestEntries`/catalog tests green (characters preserved)
- [ ] Spot-check: 1450410, 1410002, 1410202, 1360001, 1000001 resolve to correct FR/EN
- [ ] Docs synced; atomic commits on `feat/pak-catalog`
