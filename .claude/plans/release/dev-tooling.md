# Plan: dev tooling — flake checks, pre-commit, Justfile

**Type:** tooling / infrastructure
**Objective:** Round out the developer workflow: `nix flake check`, a pre-commit hook,
and Nix recipes in the Justfile. No GitHub CI (user opted out).
**Why:** The gates already exist in the Justfile; wire them so they run automatically
(pre-commit) and at the Nix level (`nix flake check`), and make the dev shell provide
the tools.
**Layer(s):** nix, docs.

## Scope

**In scope:**
- `flake.nix`: add `checks.<system>` (build+test via the package, plus a `gofmt`
  formatting check); extend the dev shell with `pre-commit` and a Nix formatter
  (`nixfmt`); format `flake.nix` itself with nixfmt.
- `.pre-commit-config.yaml`: local hooks that call the Just gates (`just ci`) and
  check Nix formatting — one definition (the Justfile), three triggers.
- `Justfile`: add `check` (`nix flake check`) and `fmt-nix` recipes.
- Docs: note the workflow in `CLAUDE.md` / `README.md`.

**Out of scope:**
- GitHub Actions CI (deferred by request) and renovate.
- Nix-level staticcheck (kept in `just lint` / pre-commit to avoid sandbox-dep
  fragility; `nix flake check` still builds + tests + gofmt-checks).

## Atomic steps (each = 1 commit, gates green)

1. `docs(plans)` — this plan + phase0.
2. `build(nix): format flake.nix with nixfmt`.
3. `build(nix): add flake checks and dev-shell tooling` — `checks` (build + gofmt),
   dev shell gains `pre-commit` + `nixfmt`.
4. `build: add nix recipes to the Justfile (check, fmt-nix)`.
5. `build: add pre-commit config running the Just gates`.
6. `docs: document checks, pre-commit and the dev shell`.

## Quality gates

- [ ] `nix develop --command just ci` green.
- [ ] `nix flake check` green (build + tests + gofmt).
- [ ] `nixfmt --check flake.nix` clean.
- [ ] `pre-commit run --all-files` green (after `pre-commit install`).
- [ ] Atomic commits on `chore/dev-tooling`; user merges.
