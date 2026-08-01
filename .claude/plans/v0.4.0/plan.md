# Plan: v0.4.0 — real item names (bilingual) from th.gl

**Type:** data + UI
**Objective:** Give the Editor real item/character names by seeding the catalog from
the community datamine at **th.gl**, in **French and English**, with a language
switch in the UI.
**Why:** Reversing the custom paks is a wall (see memory `dsa-pak-protection`); th.gl
already exposes the game's data keyed by the same CIDs (validated: 87% of this save's
CIDs covered). It's the pragmatic, reliable path.
**Layer(s):** tooling, domain, web, docs.

## Scope

**In scope:**
- A generator (`cmd/gen-catalog`) that fetches th.gl DB pages (`/fr/` and `/en/`) for
  characters, equipment, costumes, mounts, materials, ingredients and recipes, parses
  the CID→name rows, and writes `internal/domain/data/items.json`.
- New catalog format: `{cid: {fr, en, category}}`; the catalog resolves both names.
- Item carries `nameFr` + `nameEn`; the UI has an FR/EN toggle (client-side, persisted)
  that re-renders names. User labels (✎) still override both languages.
- Attribution to th.gl (README/DESIGN + a header in items.json and the generator).

**Out of scope:**
- Names not on th.gl (potions `141x`, currencies `1000001/2`, a few misc) — they keep
  category inference + manual labels.
- Icons (th.gl uses a sprite sheet; not worth it now).
- Runtime fetching: the generator is a manual `just gen-catalog`; `items.json` is
  committed and used offline.

## Design

- `internal/domain/data/items.json`: `{ "_source": "...th.gl...", "items": { "<cid>":
  {"fr": "...", "en": "...", "category": "..."} } }`.
- Category map th.gl→ours: characters→character, equipment→gear, recipes→food,
  materials_db/ingredients_db→material, costumes→costume, mounts→mount.
- `catalog.go`: parse the new seed; `Lookup`/`LookupCtx` fill `NameFR`/`NameEN`
  (override applies to both; fallback per language uses the shared category name).
- `Item`: replace `Name` with `NameFR`/`NameEN` (JSON `nameFr`/`nameEn`); `Known` true
  if any seed name or a label exists.
- `internal/web` static UI: a language toggle in the header; `displayName(it)` picks the
  active language; toggling re-renders the panels. Persist choice in localStorage.

## Atomic steps (each = 1 commit, `just ci` green)

1. `feat(tooling): add cmd/gen-catalog to fetch th.gl names`.
2. `feat(domain): generate bilingual items.json from th.gl` (the data + attribution).
3. `feat(domain): make the item catalog bilingual (fr/en names)` — catalog + Item + tests.
4. `feat(web): return both language names from /api/game`.
5. `feat(web): add an FR/EN language switch to the editor`.
6. `docs: document the item catalog, th.gl attribution and regeneration`.

## Quality gates

- [ ] `nix develop --command just ci` at each step.
- [ ] `DSA_SAVE=… just test` green (accessors resolve fr/en names).
- [ ] Manual: names show in FR; the toggle switches to EN; ✎ labels still override;
      unknown CIDs fall back sanely. In `manual_tests.md`.
- [ ] Attribution to th.gl present.
- [ ] Atomic commits on `feat/item-catalog-thgl`; user merges.

## Decisions

- **Both languages, client-side toggle** (user choice). The API returns both names so
  switching needs no round-trip semantics beyond a re-render.
- **Bundle items.json + attribution** (user choice): committed, offline; th.gl credited.
- Coverage gaps stay on inference + labels; the generator can be re-run to refresh.
