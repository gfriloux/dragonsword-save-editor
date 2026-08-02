## Plan: Full item catalog via pak extraction (CUE4Parse)

**Type:** tooling / datamine (offline, decoupled from the pure-Go editor)
**Objective:** Extract the game's item DataTable(s) + FR/EN localization from the `.pak`
archives to produce a **complete** `items.json` — covering the **off-th.gl** items we
currently hand-curate (potions, mana, XP books, awakening gems, …) and, ideally, an item
"type" field that can **drive the functional categories automatically**.
**Why:** th.gl publishes only a subset (10 routes, no potions/currencies/mana/…). The item
data lives in the paks; the wall we hit before is gone.
**Layer(s):** new offline tool (C#/CUE4Parse) + `cmd/` converter (Go) + `internal/domain/data` + docs.

### Breakthrough (research + headless spike, 2026-08-02)
- The pak **AES key is static & public**: `0x263479C442D45B7EEDE7B3A36BBB3C3B39EF9178A2F82AB694FB410AB15E01AD`
  (in `luk-gg/UnrealExporter/aes.txt`, `zibildak/MemoFastv`). Our old M1 failure was from
  trying to *derive* it dynamically — wrong approach; it's a fixed value.
- **CUE4Parse has a dedicated profile `GAME_DragonSwordAwakening = GAME_UE5_3 + 10`**
  (EGame.cs). That's why FModel/CUE4Parse reads these paks. `repak 0.2.3` fails on every
  standard version (footer magic absent) → the footer is custom ("version 101"), handled
  only by that CUE4Parse profile. Paks: 54 classic `.pak`, ~19 GB, no IoStore.
- Community already datamines it (Nexus mod 88 "Fmodel .usmap for Modders"; "More Materials"
  mods) → item data is inside the paks and extractable.

**PROVEN END-TO-END, headless on Linux (2026-08-02):** a tiny custom CUE4Parse extractor
(dotnet-sdk_10 + CUE4Parse master, EGame `GAME_DragonSwordAwakening`, the AES key) mounted all
54 paks (322,941 files) and pulled files raw. **The game ships its entire server dataset as
plain XML inside the paks** — so **NO `.usmap` is needed** for the item data:
- `…/Server/XML/GameData/GameItemData.xml` — **1212 items**; per row: `ID`(=CID), `ItemType`
  (COOKING, COOKING_INGREDIENT, EQUIPMENT, EQUIPMENT_EXP, CHARACTER_EXP, GEM,
  GEM_SOCKET_MATERIAL, CHARACTER_MASTER_SOUL, COMMON…), `Category` (numeric), `Grade`,
  `Name`/`Desc` = string keys, `IconName`.
- `…/StringData.xml` — string key → **all 11 languages incl. `Fr` AND `En`** (the game ships
  native French; the L10N/ folders are only voice overrides).
- Also `GameItemCategoryData.xml`, `GameItemTypeDefineData.xml`, `GameItemGradeData.xml`, and a
  full `…/Server/dsgamedb_sqlite.sql`.
- **Validated:** `1450410` → Name `901450410` → Fr "Gemme d'éveil: Stigmate" / En "Stigma
  Awaken Stone" (matches the user); `1410202` → EQUIPMENT_EXP, "Fragment de mana"/"Mana Shard".
- Working artifacts + recipe saved under `tmp/pak/` (gitignored).

**Invariant note:** the extractor is a small C#/CUE4Parse step — **offline, like
`cmd/gen-catalog`**, NOT linked into the editor. The editor stays pure Go, no CGO. The parser
(XML → catalog) is Go.

### Scope
**In:** extract the game-data XML (GameItemData + StringData + category/type/grade defs) raw
from the paks; a Go converter → full `items.json` (CID → FR/EN name + type/category/grade for
all 1212 items) reconciled with the existing th.gl **icons/sprite**; auto-derive the
functional consumable categories from `ItemType`/`Category`; shrink `items_extra.json`; docs.
**Out:** shipping CUE4Parse inside the editor; extracting models/audio/textures; a usmap
(not needed); a live overlay.

### Phases
#### Phase 1 — Extractor *(DONE — proven)*
A ~20-line C# program referencing CUE4Parse core (net10) mounts the paks (profile + AES key)
and `SaveAsset()`s the raw XML. Recipe + artifacts in `tmp/pak/`. No usmap.
**Deliverable (have):** GameItemData.xml, StringData.xml, GameItemTypeDefineData.xml,
GameItemCategoryData.xml + the full file list.

#### Phase 2 — Go converter (`cmd/pak-catalog`)
Parse the XML in Go: for each `GameItemData` row, CID = `ID`; look up `Name` in StringData →
`Fr`/`En`; capture `ItemType`, `Category`, `Grade`. Merge with th.gl icons/sprite (keep the
sprite; the XML gives names+CIDs th.gl lacked). Emit `internal/domain/data/items.json`
(now covering off-th.gl items). Decide FR source: prefer the game's native `Fr`, keep th.gl
as fallback. Unit-check known CIDs (1450410, 1410202, 1000500…).
**Note:** wire the extractor recipe into the tool's doc (regen story like `gen-catalog`);
the C# step stays offline. Raw XML fixtures live in `tmp/` (game-copyright; uncommitted).

#### Phase 3 — Auto-derive categories + integrate
Map `ItemType`/`Category` → functional category (COOKING→cooked, COOKING_INGREDIENT→ingredient,
EQUIPMENT_EXP→equip_xp, CHARACTER_EXP→char_xp, GEM/GEM_SOCKET_MATERIAL→…, CHARACTER_MASTER_SOUL
→awakening, EQUIPMENT→gear…). `COMMON` (106) still needs `Category`/`Grade` to split
breakthrough/craft/skill — cross-check against the hand-curated map in
[[dsa-consumable-curation]] and reconcile (data wins where confident; keep `unsorted` fallback).
Retire most of `ClassifyConsumable`'s hand rules + `items_extra.json`. Update docs
(`docs/paks.md`: no longer a wall; `content-ids.md`). Ship as a normal `vX.Y.Z` catalog refresh.

### Risks / unknowns (much reduced)
- The C# extractor needs a one-off CUE4Parse-master build (dotnet 10). Recorded recipe;
  re-run only when the game patches its data.
- `COMMON`/`Category` semantics for the breakthrough/craft/skill split — resolve via
  `GameItemCategoryData.xml` + the existing curated map.
- Confirm the XML `ID` space == the save's `ITEM_CID` (spot-checked: 1450410/1410202 match).
- FR provenance: use the game's native `Fr` (authoritative) over th.gl.

### Verification / gates
- Editor stays **pure Go** (`just ci` unchanged; the extractor is external tooling).
- Round-trip: a handful of known save CIDs resolve to correct FR/EN names post-import.
- Docs synced with any newly confirmed format facts.
