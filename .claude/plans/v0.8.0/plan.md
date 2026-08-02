# Plan: v0.8.0 — community save-format documentation

**Type:** docs
**Objective:** Document, for the community, everything we've determined **with
certainty** about the DragonSword Awakening save format: file layout, encryption,
database schema, the content-ID (CID) scheme, and the switch/recipe encoding.
**Why:** This knowledge is scattered across code and memory; a clear reference helps
others and anchors our own work. It is fed by every future plan that changes it.
**Layer(s):** docs.

## Scope

**In scope** — a browsable `docs/` markdown set:
- `docs/README.md` — index, scope, disclaimer, credits (th.gl).
- `docs/save-files.md` — the on-disk files (`<id>_Slot<N>.db`, `SPack_Slot<N>.sav`,
  `ScreenShot_<N>.png`) and the account folder.
- `docs/encryption.md` — SQLCipher v4 parameters, the constant passphrase and its
  FNV-1a-64 derivation, and how to open the DB with standard tools.
- `docs/database.md` — schema overview, the gameplay tables documented in detail
  (columns + meaning), and the full 105-table list.
- `docs/content-ids.md` — the CID scheme (ranges → categories; the 142·method·tier·
  variant cooked-dish structure) and how names/icons map via th.gl.
- `docs/switches.md` — the `tb_switch` bitmask system and the recipe encoding
  (categories 15–60).
- `docs/paks.md` — why item data comes from th.gl and not the game's paks (custom
  encryption wall).

**Out of scope:** anything not confirmed (marked as "unknown"/"not documented" where
partial). No docs site framework — plain markdown for now.

## Process note

Add to `PROCEDURE_PLANS.md`: a structural finding (new table understood, CID range,
switch category, etc.) updates `docs/` **in the same change** — like DESIGN/README.

## Atomic steps (each = 1 commit)

1. `docs: add reference index, save files and encryption`.
2. `docs: document the database schema`.
3. `docs: document the content-ID scheme and th.gl mapping`.
4. `docs: document the switch/recipe bitmask system and paks`.
5. `docs: link the reference from README and note the doc-sync rule`.

## Quality gates

- [ ] `just ci` (docs don't affect it, but run for hygiene) / `nixfmt` clean.
- [ ] Only confirmed facts; uncertainties clearly flagged.
- [ ] th.gl credited; disclaimer present.
- [ ] Atomic commits on `docs/save-format-reference`; user merges.

## Decisions

- **Plain markdown under `docs/`** (GitHub-browsable). Can grow into a docs site later.
- **Certainty first**: document what we proved; flag the rest.
