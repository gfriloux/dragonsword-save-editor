# Plan: v0.3.0 — Characters & Team (read-only), Equipment & Gems (editable)

**Type:** UI + domain accessors
**Objective:** Extend the Editor with four panels: Characters and Team as
**read-only** reference views, Equipment and Gems as **editable** panels.
**Why:** Characters/team give context (who you have, current squads) without the risk
of editing referential data; equipment/gems is where safe, high-value edits live
(enchant level, item XP, lock).
**Layer(s):** domain, web, docs.

## Scope

**In scope:**
- Read-only accessors + panels: **Characters** (`tb_character`), **Team** (`tb_team`).
- Editable accessors + panels: **Equipment** (`tb_equipment`) and **Gems** (`tb_gem`).
  - Equipment editable fields: `ENCHANT_LEVEL`, `EXP`, `IS_LOCK`. Stat CIDs
    (`MAIN_STAT_CID`, `SUB_STAT_CID1..5`) are shown **read-only** (names unknown).
  - Gem editable field: `IS_LOCK`. (The reference save has 0 gems — handle empty.)
- Catalog: resolve character names via a "character" context and equipment items via a
  "gear" category (CID 13x); character CIDs are 5-digit.

**Out of scope (later):**
- Editing character stats or team composition (kept read-only by request).
- Editing equipment/gem **stat CIDs** (needs a stat-name catalog → future, with pak data).
- Adding/removing rows.

## Working-tree state

Clean (see `phase0_results.md`). `internal/domain` has currencies + consumables.

## Design

- `internal/domain/domain.go`
  - `Characters() []Character` — read-only. `Character{Item, Level, Exp, Ascend, HP,
    Transcend, SoldierGrade int64}` (Item resolved with "character" context).
  - `Teams() []TeamPage` — read-only. `TeamPage{PageID int64; Slots [3]TeamSlot}`,
    `TeamSlot{Item, Level int64}` (slot CID resolved + character level for context;
    empty slot when CID is 0 / unknown).
  - `Equipments() []Equipment` — `Equipment{Item, DBID, EnchantLevel, Exp, IsLock,
    GemDBID, MainStatCID int64; SubStatCIDs []int64}` (active rows only).
  - `SetEnchant(dbid, level)`, `SetEquipExp(dbid, exp)`, `SetEquipLock(dbid, locked)`.
  - `Gems() []Gem` — `Gem{Item, DBID, StatInfoCID, IsLock int64}`.
  - `SetGemLock(dbid, locked)`.
- `internal/domain/catalog.go` — add "gear" inference for CID prefix 13x; "character"
  and "gear" recognised as categories (dot colours).
- `internal/web` — new endpoints:
  - `GET /api/game/characters`, `GET /api/game/teams` (read-only).
  - `GET /api/game/equipment`, `POST /api/game/equipment` `{dbid, field, value}`
    (field ∈ enchant|exp|lock).
  - `GET /api/game/gems`, `POST /api/game/gem` `{dbid, locked}`.
- Frontend — four new Editor sections. Characters and Team render read-only cards/
  tables (no inputs). Equipment renders editable rows (enchant/exp steppers + lock
  toggle, stat CIDs shown muted). Gems renders a lock toggle (empty-state message when
  none). Labels (✎) still allowed for names.

## Atomic steps (each = 1 commit, `just ci` green)

1. `feat(domain): add read-only character and team accessors` — + "character" context, tests.
2. `feat(domain): add equipment and gem accessors with enchant/exp/lock edits` — + "gear" inference, tests (round-trip on a real save).
3. `feat(web): add /api/game endpoints for characters, team, equipment and gems`.
4. `feat(web): add read-only Characters and Team panels`.
5. `feat(web): add editable Equipment and Gems panels`.
6. `docs: document the v0.3.0 panels` — README/DESIGN + `manual_tests.md`.

## Quality gates

- [ ] `nix develop --command just ci` at each step.
- [ ] `DSA_SAVE=… just test` green: equipment enchant/exp/lock round-trip; characters/
      team read; gems handle empty.
- [ ] Manual (`manual_tests.md`): Characters/Team display correctly and are **not**
      editable; equipment edits persist through Write + reopen (+ in-game check by user).
- [ ] Docs updated in the same commits.
- [ ] Atomic commits on `feat/editor-v0.3.0`; user merges.

## Decisions

- **Characters and Team are read-only** (user request) — no inputs rendered, no
  write endpoints for them.
- **Equipment/Gems editable** limited to safe scalar/bool fields; stat references stay
  read-only until a stat-name catalog exists.
- Gem panel must render a clean empty state (0 gems in the reference save).
