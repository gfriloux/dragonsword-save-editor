# Plan: v0.6.0 — item icons (th.gl sprite sheet)

**Type:** UX + tooling
**Objective:** Show each item/character's real icon in the Editor, sourced from th.gl.
**Why:** Names alone are dry; icons make the inventory/equipment/character panels
readable at a glance.
**Layer(s):** tooling, domain, web, docs.

## How th.gl icons work

Every row references a single **sprite sheet** (`cdn.th.gl/.../icons.<hash>.webp`,
~1489×1586, ~1 MB) and shows a 64×64 cell via CSS `object-position:<x>px <y>px`
(scaled with `zoom`). The sprite hash rotates when th.gl rebuilds, so the sprite and
the per-item positions must be captured in the **same** generator run.

## Scope

**In scope:**
- Extend `cmd/gen-catalog` to also parse each item's `object-position` (x, y) and to
  download the current sprite to `internal/web/static/sprite.webp`.
- Store `x`, `y` per item (and a global cell size) in `items.json`.
- Serve the sprite (already covered by the static file server) and render icons in the
  Editor rows/cards, falling back to the category dot when an item has no icon.
- Bundle the sprite (~1 MB) with attribution to th.gl.

**Out of scope:**
- Extracting individual icon files (th.gl only ships the sheet).
- Icons for CIDs th.gl lacks (they keep the category dot).

## Design

- `items.json`: add `"iconSize": 64` and per item `"x"`, `"y"` (the object-position
  pixel offsets, negative).
- `catalog.go`: `seedEntry` gains X, Y; `Item` gains `IconX`, `IconY`, `Icon` (true when
  the item comes from the seed, i.e. has a sprite cell). Expose `IconSize()`.
- `internal/web`: the embedded static FS already serves `/sprite.webp`. `/api/info`
  returns `iconSize`. Item carries icon fields.
- Frontend: an `icon(it)` helper renders `<img src="/sprite.webp">` with
  `object-fit:none; object-position:<x>px <y>px; width/height:<cell>px; zoom:<size/cell>`;
  used in every editor row/card in place of (or beside) the category dot.

## Atomic steps (each = 1 commit, `just ci` green)

1. `feat(tooling): scrape icon positions and download the th.gl sprite`.
2. `feat(domain): regenerate items.json with icon positions + bundle sprite`.
3. `feat(domain): expose icon position and size from the catalog`.
4. `feat(web): render item icons in the editor panels`.
5. `docs: document item icons and sprite regeneration`.

## Quality gates

- [ ] `just ci` at each step; catalog tests updated for the icon fields.
- [ ] The sprite loads (`/sprite.webp` 200) and icons render (user check).
- [ ] Fallback to the category dot for un-iconed CIDs.
- [ ] Attribution to th.gl for the bundled sprite.
- [ ] Atomic commits on `feat/item-icons`; user merges.
