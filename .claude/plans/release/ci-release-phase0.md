# Phase 0 — state of the tree before the release CI

Run on 2026-08-08, from `main` at `50f2379` (`docs: changelog for v0.17.0`).

```
$ nix develop --command just ci
gofmt-check  ok
go vet ./...  ok
staticcheck ./...  ok
go test ./...  all packages ok (DSA_SAVE unset → unit tests only)
go build ./...  ok
```

Green. Working tree clean apart from an untracked `.envrc` (local direnv, not committed).

Relevant existing state:

- No `.github/` directory at all — the "release workflow" mentioned in `CLAUDE.md`
  never existed.
- `Justfile` has `build` (host, no version stamping) and `build-windows`
  (`dist/dsa-save-editor.exe`, no version stamping). No Linux release recipe.
- `cmd/dsa-save-editor/helpers.go:15` — `var buildVersion = "dev"`, overridden with
  `-ldflags "-X main.buildVersion=…"`; printed by `--version` (`main.go:35`).
- `flake.nix` hard-codes `version = "0.1.0"` and stamps it into `buildVersion`, so
  `nix run .#default -- --version` reports `0.1.0` at tag v0.17.0. Pre-existing drift,
  out of scope (see plan).
- `CHANGELOG.md` sections are `## [0.17.0] - 2026-08-07`, i.e. the tag without its
  leading `v`.
