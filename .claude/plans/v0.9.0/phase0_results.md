# Phase 0 — audit (v0.9.0)

Date: 2026-08-02. Branch `feat/browse-stackable-catalog` off `main` @ `365ffa5`
(v0.8.0 merged).

`nix develop --command just ci` → **green** (fmt-check, vet, staticcheck, test, build
all pass). Working tree clean except an untracked local `.envrc` (direnv; not part of
this change, not committed).

## State relevant to this plan

- `internal/domain` already has `AddOrSetStackable(cid, count)` (upsert into
  `tb_stackable_item`), `Consumables()` (owned stacks + cook), `Catalog().Entries()`
  (all seeded items). `tb_stackable_item` = `(USER_DBID, ITEM_CID)` PK + `STACK_CNT`,
  no triggers → upsert safe (confirmed v0.5.0 phase0).
- `internal/web` exposes `POST /api/game/stackable` (→ `AddOrSetStackable`),
  `GET /api/game/catalog` (all catalog items) and `GET /api/game/consumables`
  (owned stacks + cook). `app.js` `renderConsumables()` lists owned rows and a
  datalist "Add a material" toolbar filtered to `category === "material"`.

## Catalog / th.gl coverage (measured this session)

- th.gl DB routes (index): `characters, costumes, equipment, ingredients_db,
  materials_db, monsters, mounts, npcs, quests, recipes` — no potions/currency page.
- `materials_db` = 131 unique CIDs, `ingredients_db` = 48 → 179 total, which matches
  exactly the 179 `category:"material"` entries in `items.json`. Scrape is complete;
  no pagination dropped.
- Catalog holds **0** potions (`141xxxx`) and **0** currencies (`1000xxx`) — th.gl
  does not publish them. Confirmed the reason to stay th.gl-only for this plan.

Conclusion: the change is UI-only (surface the already-complete stackable catalog as
editable rows). No domain/save/API/gen-catalog work needed.
