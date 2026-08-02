# Phase 0 — audit (v0.12.0)

Date: 2026-08-02. Branch: `feat/ui-refonte` (off `main`).

## `nix develop --command just ci`

Green. All gates pass on a clean tree:

- `gofmt` — no unformatted file
- `go vet ./...` — no findings
- `staticcheck ./...` — no warnings
- `go test ./...` — `internal/{domain,save,sqlcipher}` ok; `cmd/*` and `internal/web` have
  no test files
- `go build ./...` — ok

## Working-tree state at start

- Untracked: `.claude/plans/v0.12.0/` (this plan), `.envrc` (local, ignored/irrelevant).
- No pending code changes. Baseline is clean — any test/gate movement from here is caused
  by this plan's work.

## Notes carried into the phases

- `internal/web` has no Go tests (UI is manual-tested per PROCEDURE §4) — the refonte's
  validation goes in `manual_tests.md`.
- New layers `internal/pak` + `internal/oodle` will introduce the project's first codec
  tests gated on `DSA_GAME_DIR` (mirroring the `DSA_SAVE` pattern).
