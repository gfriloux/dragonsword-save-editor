# Phase 0 — audit (v0.4.0)

Date: 2026-08-01. Branch `feat/item-catalog-thgl` off `main` @ `8124ae9`.

`nix develop --command just ci` green (gofmt/vet/staticcheck/test/build).

Working tree clean. `internal/domain/data/items.json` is currently an empty seed
(`{"items":{}}`); the catalog resolves names via inference + user labels only.

th.gl coverage validated ahead of the plan: 164/187 (87%) of this save's distinct
CIDs have names on th.gl, keyed by the same CIDs (e.g. 10001=Eileen,
1360001="Casque d'élève de l'Académie de magie"). Gaps: potions 141x, currencies
1000001/2, a few 145x/147x/151x.
