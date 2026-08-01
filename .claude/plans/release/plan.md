# Plan: Working procedures & project tooling

**Type:** tooling / infrastructure
**Objective:** Give the project the same working method as the sibling projects
`auspex` and `stc`: plan-first, hybrid git, a single quality-gate definition, and
changelog tooling.
**Why:** The first version landed in one commit. Before growing the editor we want a
documented, repeatable process so future work is atomic, testable and traceable.
**Layer(s):** docs, nix.

## Scope

**In scope (core first):**
- Planning layout under `.claude/plans/` + this plan.
- `Justfile` as the single source of truth for the quality gates (Go).
- `git-cliff` changelog config + generated `CHANGELOG.md`.
- `DESIGN.md` (architecture + invariants), `PROCEDURE_PLANS.md`, `CLAUDE.md`.
- Dev shell updated with `just`, `git-cliff`, `staticcheck`.

**Out of scope (deferred):**
- CI workflows (`.github/workflows/ci.yml`, `release.yml`).
- `renovate.json`.
- `.pre-commit-config.yaml`.

These are the natural next tooling step once the core process is in place.

## Working-tree state

Branch `chore/working-procedures` off `main` (`c6d4e6f`). The Go module, flake and
README already exist; no CI/changelog/process docs yet.

## Atomic steps

1. **Plans layout + this plan** — `.claude/plans/README.md`, `.claude/plans/release/plan.md`.
   Verify: files present. Commit: `docs(plans): add plans layout and working-procedures plan`.
2. **Quality gates** — `Justfile` + flake dev shell (`just`, `git-cliff`, `staticcheck`).
   Verify: `nix develop --command just ci`. Commit: `build: add Justfile quality gates and dev tooling`.
3. **Changelog config** — `cliff.toml`. Commit: `build: add git-cliff changelog configuration`.
4. **Generate changelog** — `CHANGELOG.md` via `just changelog`.
   Commit: `docs(changelog): generate initial CHANGELOG.md`.
5. **Architecture doc** — `DESIGN.md`. Commit: `docs: add DESIGN.md (architecture and invariants)`.
6. **Procedures doc** — `PROCEDURE_PLANS.md`. Commit: `docs: add PROCEDURE_PLANS.md (working procedures)`.
7. **Project instructions** — `CLAUDE.md`. Commit: `docs: add CLAUDE.md (project instructions)`.

## Quality gates

- [ ] `nix develop --command just ci` passes.
- [ ] `just changelog` produces a clean `CHANGELOG.md` (reviewed diff).
- [ ] Atomic commits on `chore/working-procedures`; user merges to `main`.

## Decisions

- **Docs in English** to match the Go code and existing `README.md` (sibling `stc`
  also documents in English). `git-cliff` groups are English accordingly.
- **Core first**: CI, renovate and pre-commit are deferred to a follow-up so the
  process itself lands reviewable and small.
- **Gates are Go-native**: `gofmt`, `go vet`, `staticcheck`, `go test`, `go build`,
  with `build-windows` as a release-time cross-compile check.
