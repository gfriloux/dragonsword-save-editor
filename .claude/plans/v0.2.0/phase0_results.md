# Phase 0 — audit (v0.2.0)

Date: 2026-08-01
Branch: `feat/editor-mvp` off `main` @ `0a449c7` (working procedures merged).

## `nix develop --command just ci`

Green:

- `gofmt` — clean.
- `go vet ./...` — no findings.
- `staticcheck ./...` — no warnings.
- `go test ./...` — `internal/save` ok, `internal/sqlcipher` ok (real-save tests skipped: `DSA_SAVE` unset).
- `go build ./...` — ok.

## Working tree

Clean. Existing layers: `internal/sqlcipher`, `internal/save`, `internal/web`,
`cmd/dsa-save-editor`. The web UI is the generic table browser (`/api/table`,
`/api/update`, `/api/save`). No `internal/domain` yet, no `data/items.json`.

Starting point is sound; no cleanup needed before coding.
