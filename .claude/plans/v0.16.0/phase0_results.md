# Phase 0 — audit avant code (v0.16.0)

> À renseigner **au démarrage de l'implémentation**, sur un `main` propre et à jour,
> avant la première ligne de code (voir `PROCEDURE_PLANS.md` §2).

**Date :** 2026-08-06
**Branche de départ :** `main` (dernier commit `6ad606b docs: changelog for v0.15.0`)

## `nix develop --command just ci` → **vert**

```
go vet ./...
staticcheck ./...
go test ./...
?   cmd/dsa-save-editor        [no test files]
?   cmd/gen-catalog            [no test files]
?   cmd/pak-catalog            [no test files]
?   cmd/pak-dump               [no test files]
ok  internal/config            0.005s
ok  internal/domain            0.129s
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
- Non suivis présents : `.claude/plans/v0.16.0/` (ce plan), `.envrc`, plus `.claude/plans/BACKLOG.md`
  modifié (item Titres marqué planifié) — tout part sur la branche `feat/titles`. `tmp/` (fixtures
  pak) hors suivi, non touché.
- Round-trip vrai save via une **copie** pointée par `DSA_SAVE`, jamais l'original.
- Pré-requis converter : `tmp/pak/AccountTitleData.xml` et `tmp/pak/StringData.xml` présents
  (déjà extraits le 2026-08-06). Sinon rejouer l'extracteur CUE4Parse (voir `tmp/pak/README.md`).
