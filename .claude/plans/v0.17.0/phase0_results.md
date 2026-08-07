# Phase 0 — audit (v0.17.0)

Date: 2026-08-07

- Branch base: `main` at tag `v0.16.1`.
- Working tree: clean (only untracked `.envrc`, local dev only).
- `nix develop --command just ci`: **green**
  - fmt-check, vet, staticcheck: no findings
  - `go test ./...`: all packages OK (config, domain, save, sqlcipher, web, …)
  - `go build ./...`: OK

Tree is clean and ready. Work proceeds on `feat/inventory-supercategories`.
