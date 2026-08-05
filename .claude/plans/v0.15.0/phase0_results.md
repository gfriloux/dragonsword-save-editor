# Phase 0 — audit avant code (v0.15.0)

**Date :** 2026-08-05
**Branche de départ :** `main` (dernier commit `998c1a6 docs: changelog for v0.14.0`)

## `nix develop --command just ci` → **vert**

```
go vet ./...
staticcheck ./...
go test ./...
?   cmd/dsa-save-editor        [no test files]
?   cmd/gen-catalog            [no test files]
?   cmd/pak-catalog            [no test files]
?   cmd/pak-dump               [no test files]
ok  internal/config            0.006s
ok  internal/domain            0.131s
?   internal/icons             [no test files]
ok  internal/oodle             (cached)
ok  internal/pak               (cached)
ok  internal/save              0.007s
ok  internal/sqlcipher         (cached)
ok  internal/texture           (cached)
ok  internal/web               0.049s
go build ./...
```

Exit 0 — fmt-check + vet + lint + test + build tous OK. Arbre sain, on peut coder.

## État de l'arbre
- Non suivis présents (`.envrc`, `.claude/plans/BACKLOG.md`) — hors périmètre, non touchés.
- Round-trip vrai save non exécuté ici (`DSA_SAVE` non défini) ; le save live de référence
  est sous
  `/run/media/kuri/…/DragonSword Awakening/DS/Saved/SaveGames/6144/6144_Slot1.db`.
  Le round-trip des steps utilisera un **copie** de ce fichier via `DSA_SAVE`, jamais
  l'original ni un fichier commité.
