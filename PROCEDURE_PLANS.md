# PROCEDURE_PLANS.md — Working procedures

> This document defines the process to follow **before any work on the code**.
> Every change is decomposed into atomic steps, each independently testable and
> committable.
>
> **Core rule: no code without a validated plan.**
> And: read [`DESIGN.md`](DESIGN.md) before any change. If an idea does not fit the
> invariants, the answer is no.

---

## 1. Plan before anything

As soon as a version or a feature is on the table, create:

```
.claude/plans/v{X.Y.Z}/
  plan.md           ← context, scope, phases, decisions, files touched
  manual_tests.md   ← manual tests (grown during dev, run at validation)
  phase0_results.md ← real state of the repo before coding (see §2)
```

Plans live **under `.claude/plans/`**, never at the repository root. Tooling and
infrastructure work (CI, changelog, release, these procedures) goes under
`.claude/plans/release/` instead of a `vX.Y.Z/` directory. An obsolete plan is
**deleted**, never duplicated as `_v2` / `_v3`. See
[`.claude/plans/README.md`](.claude/plans/README.md).

### Minimal `plan.md` content

- **Context**: where we start from, why.
- **Objective**: what we want to reach.
- **Scope**: explicit in-scope / out-of-scope.
- **Layer(s) involved**: `sqlcipher` | `save` | `web` | `cmd` | `nix` | `docs`.
- **Working-tree state**: what already exists, what must disappear.
- **Ordered atomic steps**: each with its verification and commit message.
- **Technical decisions**: choices and their justification.

---

## 2. Phase 0 — mandatory audit

**Before touching the code**, check the real state — never assume the tree is
clean:

```bash
nix develop --command just ci
```

Record the result in `.claude/plans/v{X.Y.Z}/phase0_results.md`.

---

## 3. Git policy — hybrid

- Claude works on a **dedicated branch** (`feat/…`, `fix/…`, `refactor/…`,
  `perf/…`, `docs/…`, `chore/…`, `ci/…`, `build/…`), never directly on `main`.
- Claude **commits atomically**: one logical change = one commit, in
  [Conventional Commits](#commit-convention). Each commit passes the gates on its
  own.
- Claude **never** runs `merge`, `push` or `tag`. The user reviews, merges onto
  `main` and pushes.
- **A plan always ends with a merge onto `main`.** When the gates are green, the
  user merges the plan branch and pushes **before** the next plan starts. A finished
  plan branch is never left unmerged: every plan starts from an up-to-date `main`.

### Commit convention

```
type(scope): short imperative message
```

- **type**: `feat`, `fix`, `refactor`, `perf`, `test`, `docs`, `build`, `ci`,
  `chore`.
- **scope**: `sqlcipher`, `save`, `web`, `cmd`, `nix`, or the package touched.

Documentation is updated **in the same commit** as the code it describes. A
structural change committed without updating `DESIGN.md` / `README.md` leaves the
docs stale — that is a defect.

---

## 4. Test discipline — golden on the deterministic layers

The `sqlcipher` and `save` layers are deterministic: a given encrypted file and key
produce an exact plaintext, and an edit round-trips back to a valid, game-readable
save. Tests exercise these against a real save, gated on an env var so they are
skippable in a clean checkout:

```bash
go test ./...                                   # unit tests only
DSA_SAVE=/path/to/6144_Slot1.db go test ./...   # also decrypt/edit/re-encrypt a real save
```

- A behaviour change must be visible in a test.
- **We automate**: the SQLCipher codec (decrypt/encrypt/round-trip), key derivation,
  and the edit → save → re-open cycle. Cross-checking against the reference
  `sqlcipher` CLI confirms game compatibility.
- **We do not automate**: the browser UI rendering, launching against the real
  running game, and anything requiring the game process. Those go into
  `manual_tests.md`.
- **No personal save data in the repo** (see `DESIGN.md` invariant 4). Fixtures are
  local; real saves are referenced via `DSA_SAVE`, never committed.

---

## 5. Change types & recipes

| Type | Layer | Steps (each = 1 commit) |
|---|---|---|
| SQLCipher / format change | `sqlcipher` | 1. test (round-trip / vector) → 2. impl (`feat(sqlcipher): …`) → 3. doc |
| New edit capability | `save` | 1. test against real save → 2. impl (`feat(save): …`) → 3. doc |
| UI / API change | `web` | 1. impl (`feat(web): …`) → 2. `manual_tests.md` updated → visual review |
| CLI change | `cmd` | 1. impl (`feat(cmd): …`) → 2. doc / usage |
| Bug fix | affected layer | 1. failing regression test (`test: reproduce …`) → 2. fix (`fix(scope): …`) |
| Refactor | affected layer | 1. refactor with no test change (`refactor(scope): …`). If a test result moves, it was not a refactor. |
| Docs only | — | `docs: …` |

---

## 6. Plan template

```markdown
## Plan: [Title]

**Type:** [format | edit capability | UI | CLI | bug | refactor | docs]
**Objective:** ...
**Why:** ...
**Layer(s):** [sqlcipher | save | web | cmd | nix | docs]

### Files touched
- [ ] `internal/...`
- [ ] `*_test.go`
- [ ] `DESIGN.md` / `README.md`

### Atomic steps
#### Step 1: [Title]
**Description:** ...
**Verification:** `just ci` (or the relevant recipe)
**Commit:** `type(scope): message`

### Quality gates
- [ ] `just ci` passes
- [ ] Real-save round-trip still green (`DSA_SAVE=… just test`)
- [ ] Docs synced (same commit)
- [ ] Atomic commits on a dedicated branch
```

---

## 7. Quality gates

Every change passes these gates before it is considered done. **One definition:**
the `Justfile`. pre-commit and CI call it.

```bash
just fmt-check   # gofmt — no unformatted file
just vet         # go vet — no findings
just lint        # staticcheck — no warnings tolerated
just test        # go test ./...
just build       # go build ./...
just ci          # all of the above, in order
```

Run non-interactively in the Nix dev shell: `nix develop --command just ci`.

The Windows cross-compile is a release-time gate: `just build-windows` must produce
a single static `.exe`.

---

## 8. What does not change between versions

- **DESIGN.md is authoritative.** Outside the invariants → no.
- **Pure Go, no CGO**: hard invariant (cross-compilation to Windows). See DESIGN.md.
- **Never corrupt a save**: read-only on open, atomic write, automatic backup, game
  closed before writing.
- **The passphrase / SQLCipher parameters** are the game's; changing the targeted
  format is a DESIGN decision, not an implicit code change.
- **Hybrid git**: branch + atomic commits by Claude; merge/push/tag by the user.
- **Nix**: always `nix develop --command …` for non-interactive commands.
- **`tmp/`**: uncommitted scratch (notes, work output, handoffs).
- **No personal save data** committed, ever.

---

**Last updated:** 2026-08-01
**Status:** Active
