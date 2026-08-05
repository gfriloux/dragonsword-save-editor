# Phase 0 — audit (v0.14.0)

Date : 2026-08-04. Branche : `feat/theme-polish` (partie de `main` @ `701dc84`).

## `nix develop --command just ci`

Vert. `go vet`, `staticcheck`, `go test ./...`, `go build ./...` passent tous.
Seul avertissement : « Git tree is dirty » — attendu, dû au fix T1 (spinner)
déjà présent dans l'arbre avant le lancement (`internal/web/static/style.css`).

```
go vet ./...
staticcheck ./...
go test ./...        → tous ok (config, domain, save, web, oodle, pak, sqlcipher, texture)
go build ./...       → ok
```

## État de l'arbre au départ

- `internal/web/static/style.css` : **modifié** (T1 — suppression du spinner natif
  `input[type=number]`, déjà appliqué et validé comme trivial avant ce plan).
- `.claude/plans/BACKLOG.md`, `.envrc` : untracked (non concernés par ce plan).

Rien à faire disparaître. Le tree est sain, ci verte : on peut coder.
