# Plan: v0.5.0 — add stackable items to the inventory

**Type:** domain + web
**Objective:** Let the user **add** stackable items (materials/consumables) they don't
own yet, picked by name from the th.gl catalog, and set a quantity — plus a
"fill all" shortcut for existing stacks.
**Why:** Editing existing quantities isn't enough; players want to grant themselves
materials/consumables they lack.
**Layer(s):** domain, web, docs.

## Scope

**In scope:**
- Add / set a stackable item: upsert `tb_stackable_item (USER_DBID, ITEM_CID, STACK_CNT)`.
- "Fill all" stackables to a chosen amount.
- A catalog-fed picker: search the bundled catalog by name (FR/EN), pick an item, set qty.
- Restrict the picker to stackable-appropriate categories (material, potion, misc) to
  avoid inserting characters/gear/food into the wrong table.

**Out of scope (later / risky):**
- Adding cook items or equipment (need fabricated 64-bit `ITEM_DBID`s — unsafe for now).
- Adding currencies/characters/costumes.
- Deleting items.

## Design

- `internal/domain`
  - `AddOrSetStackable(cid, count int64) error` — `INSERT OR REPLACE INTO tb_stackable_item`.
  - `FillStackables(count int64) error` — `UPDATE tb_stackable_item SET STACK_CNT=?`.
  - `Catalog.Entries() []Item` — all seed items (resolved), for the picker; the API
    filters/returns them so the frontend can search client-side.
- `internal/web`
  - `GET /api/game/catalog` → the catalog entries (cid, nameFr, nameEn, category).
  - `POST /api/game/stackable` `{cid, count}` → add/set.
  - `POST /api/game/stackable/fill` `{count}` → fill all.
- Frontend (Consumables panel): an "Add item" row with a search box (client-side over
  the catalog, filtered to material/potion/misc), a quantity field and an Add button;
  and a "Set all to N" control. Re-render after changes.

## Atomic steps (each = 1 commit, `just ci` green)

1. `feat(domain): add stackable upsert, fill-all and catalog listing` + tests (real save).
2. `feat(web): add /api/game endpoints for catalog and stackable add/fill`.
3. `feat(web): add an item picker and fill-all to the Consumables panel`.
4. `docs: document adding items`.

## Quality gates

- [ ] `just ci` at each step; `DSA_SAVE=… just test` green (add/fill round-trip; the
      added row reads back and survives Write + reopen; verified against `sqlcipher`).
- [ ] Manual: search the catalog, add a material, set its qty; it appears in Consumables
      and persists in-game (user check). "Fill all" works. In `manual_tests.md`.
- [ ] Picker limited to safe categories.
- [ ] Atomic commits on `feat/add-items`; user merges.

## Decisions

- **Stackable only** for now (ITEM_CID-keyed, no fabricated DBID) — safe.
- **Upsert semantics**: adding an item you already own sets its quantity (predictable),
  rather than incrementing.
