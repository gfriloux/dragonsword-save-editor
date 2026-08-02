# Plans

This directory holds every planning document, following the convention shared with
the sibling projects `auspex` and `stc`.

## Layout

- `vX.Y.Z/` — the plan that produced release `vX.Y.Z`. Each directory holds at least
  `plan.md` (context, scope, atomic steps, decisions), and usually `manual_tests.md`
  and `phase0_results.md`. A release built from more than one chantier keeps one file
  per chantier.
- `release/` — plans about tooling and project infrastructure (changelog, CI,
  release, working methods) rather than a single product feature.

## Rules

- Plans always live here, **never at the repository root**.
- An obsolete plan is **deleted**, never duplicated as `_v2` / `_v3`.
- A plan is authored before any code (see [`PROCEDURE_PLANS.md`](../../PROCEDURE_PLANS.md)).

## Version map

| Tag | Commit | Date | Content | Plan |
|-----|--------|------|---------|------|
| _(untagged)_ | `c6d4e6f` | 2026-08-01 | First version: pure-Go SQLCipher codec, save-editing layer, embedded web UI, Nix flake with Windows cross-compilation | — |
| _(pending)_ | — | 2026-08-01 | Working procedures: plans layout, Justfile gates, git-cliff, DESIGN/PROCEDURE/CLAUDE docs | [`release/`](release/plan.md) |
| _(pending)_ | — | 2026-08-02 | Consumables panel lists the full th.gl stackable catalog (owned + not owned) with editable counts (0 → X) | [`v0.9.0/`](v0.9.0/plan.md) |
| _(pending)_ | — | 2026-08-02 | Curated functional consumable categories + category-sidebar panel (direction B); curated bilingual names for off-th.gl items | [`v0.10.0/`](v0.10.0/plan.md) |
