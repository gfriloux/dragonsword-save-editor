## Plan: Player-facing UI refonte + first-run save picker + authentic game icons

**Type:** UI + CLI + new codec layer (pak/oodle) + data (recipes)
**Objective:** Turn the editor from a "database tool" into a player-facing app:
(1) first run asks for the **game folder** (no more save path on the CLI), remembers it,
lets the user change it; (2) it lists the existing saves with their **screenshot**;
(3) once a save is picked it shows a redesigned, game-oriented UI ("Sang & acier");
plus **authentic in-game icons** extracted from the user's own paks (pure-Go, cached
locally, never committed) and **cooking recipe details** (required materials + owned
counts).
**Why:** the current UI exposes `tb_*` tables as the primary concept and passes the save
path as an argument — unfit for a non-technical player. Design handoff in
`tmp/ui/design_handoff_save_editor_ui/` (7 screens, high-fidelity, same static stack).
**Layer(s):** cmd, web, domain, **new `internal/pak` + `internal/oodle`**, nix, docs

---

### Context — spike findings (2026-08-02), all verified this session

Pure-Go reading of the game's custom paks is **proven end-to-end** (Go probe in
scratchpad `pakprobe/`, ran clean on all 54 paks). Details persisted in memory
`dsa-pak-pure-go`. Summary that this plan relies on:

- **Pak footer** is standard UE5 (magic `0x5A6F12E1` at EOF-204); only the version dword
  is obfuscated to `101` (real 11). repak/UnrealPak fail *only* because they reject 101.
- **Index crypto** = AES-256-**ECB** with the public key `0x263479C4…01AD`, then a
  **2-layer XOR deobfuscation** (reversed from CUE4Parse `GameTypes/DragonSword`):
  index buffer XOR `decrypted[2]`; each FString XOR its own last byte; encoded-entries
  blob XOR its byte `[5]`; UE bool in the index = int32 (4 bytes).
- **Icons**: `GameItemData.xml` gives every item an `IconName` → a `Texture2D` path. All
  **608 icons live in `pakchunk0_s24-WindowsClient.pak`**. Per icon: `.uasset` = stored
  (cmi=0), **`.uexp` = Oodle** (cmi=1), not encrypted, 2 blocks, uncompressed **65680 B ≈
  128×128×4 + 144** → almost certainly **uncompressed BGRA8** (PNG encode trivial once
  decompressed). Compression method table = `[Oodle]` only.
- **The one hard piece = Oodle-Kraken decompression** of the `.uexp`. No pure-Go Oodle
  exists; `oriath-net/gooz` wraps C `ooz` via **cgo** (forbidden by DESIGN invariant #1).
  **Decision (user): port Kraken to pure Go** (native, single static binary, no wasm/cgo).
- **Recipes**: `CookRecipeData.xml` + `CookToolData.xml` are in the paks (the old
  "cooking wall" from memory `dsa-cook-recipes` is gone). They give per-recipe **materials
  + quantities** and the `CookRecipeKey` → `(tb_switch category, bit)` mapping (also fixes
  the ~3 special recipes still locked).
- **Save layout** (relative to the game folder the user provides):
  `…/DS/Content/Paks/*.pak` (icons) and `…/DS/Saved/SaveGames/<accountId>/<id>_Slot<N>.db`
  with siblings `SPack_Slot<N>.sav` (GVAS slot meta) and **`ScreenShot_<N>.png`** (the slot
  thumbnail the design wants). One game folder covers both saves and icons.

---

### Scope

**In:**
- First-run flow: prompt for game folder, persist it in the OS config dir, allow changing;
  no save path on the CLI (kept as an optional override/`--game-dir` flag).
- Save discovery + picker screen (slots, level/playtime from the db/SPack, screenshot).
- Full "Sang & acier" UI refonte in the existing static stack (HTML/CSS/vanilla JS,
  `go:embed`, no bundler/framework), embedded woff2 fonts, JSON API unchanged.
- Cooking details: dev-side extraction of `recipes.json` (materials + effect + stars +
  key→cat/bit) committed like `items.json`; domain surfaces materials + owned counts;
  Cuisine screen renders them; special recipes unlocked from the real key map.
- Authentic icons: `internal/pak` (pure-Go reader) + `internal/oodle` (pure-Go Kraken) +
  UTexture2D→PNG decode; a first-run/"refresh" extraction that caches PNGs (or a packed
  atlas) in the OS **cache** dir; web serves cached icons, falls back to th.gl sprite.

**Out:**
- Editing `SPack_*.sav` (GVAS) — still read-only, only for slot metadata/screenshot.
- Mermaid/Leviathan/Selkie/LZNA/BitKnit codecs unless a needed pak asset uses them
  (icons are Kraken; port those only if a required asset needs them).
- Online/server features. Character/equipment stat-table expansion beyond today.

---

### New architecture / layering

```
internal/sqlcipher   (unchanged)
internal/oodle       NEW — pure-Go Kraken decompressor (stdlib only, no cgo)
internal/pak         NEW — pure-Go custom-pak reader (footer, AES-ECB, XOR deobf,
                     dir index, entry decode, block extract) → uses internal/oodle;
                     + UTexture2D → image decode
internal/save        (unchanged)
internal/domain      + recipes (materials/owned), + icon-cache resolution
internal/web         full UI refonte; serves cached icons; API unchanged
cmd/dsa-save-editor  first-run config, game-dir memory, save discovery, icon-cache build
cmd/pak-catalog      + emit recipes.json (dev-side, like items.json)
```

`internal/pak`/`internal/oodle` are low-level codecs (peers of `sqlcipher`); they must not
import `save`/`domain`/`web`. The runtime pak read is **pure Go, CGO_ENABLED=0** — invariant
#1 holds. `gooz` (cgo) is used **only** to generate golden test hashes dev-side; it is
**not** a module dependency.

---

### Files touched (high level)
- [ ] `internal/oodle/*.go` (+ `*_test.go`, golden gated on `DSA_GAME_DIR`)
- [ ] `internal/pak/*.go` (+ `*_test.go`)
- [ ] `internal/domain/` recipes + icon-cache resolver (+ tests)
- [ ] `internal/domain/data/recipes.json` (generated, committed — data, not art)
- [ ] `cmd/pak-catalog/main.go` (emit recipes.json)
- [ ] `cmd/dsa-save-editor/` first-run, config, save discovery, icon-cache build
- [ ] `internal/web/static/{index.html,style.css,app.js}` + embedded woff2 + icon serving
- [ ] `internal/web/*.go` (config/game-dir/save-list/icon endpoints)
- [ ] `DESIGN.md`, `README.md`, `docs/paks.md`, `docs/content-ids.md`, `Justfile`, `flake.nix`

---

### Phases (each phase = independently committable; risky Oodle phase isolated last)

#### Phase 0 — audit
`nix develop --command just ci` → record in `phase0_results.md`.

#### Phase 1 — First-run config + save discovery + picker (cmd + web)
- Config store in the OS config dir (reuse `domain.DefaultOverridesPath()` sibling):
  remembered game folder; `--game-dir` flag override; the positional save path stays as a
  power-user override.
- Discover saves: glob `<gameDir>/DS/Saved/SaveGames/*/*_Slot*.db`; per slot resolve
  `ScreenShot_<N>.png`, `SPack_Slot<N>.sav`; read level/name/playtime (from the db
  `tb_user`, and/or SPack GVAS — investigate which is cheapest/read-only).
- API: `GET /api/config`, `POST /api/config/game-dir`, `GET /api/saves`,
  `POST /api/open` (select a save → the rest of the app operates on it),
  `GET /api/saves/<id>/screenshot`.
- Minimal Accueil screen (see design §1) usable before the full theme lands.
**Commits:** `feat(cmd): first-run game-folder config + save discovery`,
`feat(web): save-picker (Accueil) screen + screenshot endpoint`.

#### Phase 2 — UI refonte "Sang & acier" (web)
Reimplement the shell (2 bars + tab nav + toast), design tokens, embedded woff2
(Marcellus/Karla/JetBrains Mono), and re-skin every existing panel per the handoff:
Monnaies, Inventaire (3 columns), Personnages, Équipe, SQL brut. JSON API unchanged;
dirty-state stays derived. `manual_tests.md` grown here (visual review).
**Commits:** `feat(web): Sang & acier shell + design tokens + fonts`, then one per screen.

#### Phase 3 — Cooking recipe details (cmd + domain + web)
- `cmd/pak-catalog`: parse `CookRecipeData.xml` (+ CookTool, + StringData for names) →
  `recipes.json` = per recipe: dish CID, effect, stars, **materials [{cid, qty}]**,
  `CookRecipeKey` → `(category, bit)`. Commit the JSON (data, not art).
- domain: load recipes; expose per recipe its materials with **owned counts** (from the
  save's stackables); rewrite `UnlockAllRecipes` to use the real key→(cat,bit) map (covers
  the special recipes) instead of blanket categories 15–60.
- web: Cuisine detail panel shows "Matériaux requis" with owned/required per material.
**Commits:** `feat(cmd): pak-catalog emits recipes.json`,
`feat(domain): recipe materials + owned counts + key-accurate unlock`,
`feat(web): Cuisine shows required materials and owned counts`.

#### Phase 4 — Pure-Go pak reader (internal/pak), no Oodle yet
Footer parse, AES-ECB + 2-layer XOR deobf, primary + full-directory index, encoded-entry
decode, per-block offset/size extraction, raw block read. Test (gated `DSA_GAME_DIR`):
mount points, file counts, locate `Icon_Item_*`, read a **stored** `.uasset` byte-for-byte.
**Commit:** `feat(pak): pure-Go reader for the custom version-101 paks`.

#### Phase 5 — Pure-Go Oodle-Kraken (internal/oodle)  ← highest effort/risk
Port Kraken decode from `ooz` (kraken.cpp) to Go, x86 intrinsics → scalar Go. Entry:
`Decompress(in, outSize) ([]byte, error)`. **Golden strategy:** generate expected SHA-256
of real icon `.uexp` blocks once via `gooz` (cgo, dev-side, throwaway), commit **only the
hashes** (no game bytes); test decompresses live from `DSA_GAME_DIR` and compares SHA.
**Commit:** `feat(oodle): pure-Go Kraken decompressor`.

#### Phase 6 — UTexture2D decode + icon cache pipeline (pak + cmd + web)
- `internal/pak`: parse UTexture2D from `.uasset`+decompressed `.uexp` → confirm BGRA8
  128×128 → `image.NRGBA` → PNG. (Confirm format on first real decode; handle BC7 only if
  the assumption breaks.)
- cmd: build the icon cache on first run / on demand ("Rafraîchir les icônes"): read the
  committed `IconName` map, extract+decompress+decode each of the 608 icons from the user's
  paks, write to the OS **cache** dir (PNG files or a packed atlas + offsets). Never
  committed (copyright).
- web: serve cached icons across all screens; fall back to th.gl `sprite.webp` if absent.
**Commits:** `feat(pak): UTexture2D → image decode`,
`feat(cmd): extract + cache authentic item icons from the paks`,
`feat(web): render authentic cached icons with th.gl fallback`.

#### Phase 7 — Docs + release
DESIGN.md (new layers; "icons extracted per-user at runtime, never committed" model;
restate CGO invariant holds), README, `docs/paks.md` (deobf recipe), `docs/content-ids.md`
(recipes), `Justfile`/`flake.nix` if a `pak-catalog`/icon recipe is added, CHANGELOG via
`just changelog`. `just build-windows` static `.exe` gate.
**Commit:** `docs: …` (+ any in-same-commit doc updates done inline above).

---

### Technical decisions
- **Recipes = committed data** (extracted dev-side like `items.json`); **icons = per-user
  runtime extraction** (copyright art, cached in OS cache dir, never committed). This is the
  key split that keeps the repo clean and the binary self-contained.
- **Kraken ported to Go, not wasm/cgo** (user's choice) → single static binary, no runtime,
  debuggable. gooz stays dev-side only (golden hashes), never a dependency.
- **One game folder** input covers saves and paks; saves discovered under
  `DS/Saved/SaveGames`, paks under `DS/Content/Paks`.
- API kept **unchanged** where it exists; new endpoints are additive (config/saves/icons).
- th.gl `sprite.webp` retained as the **fallback** until the icon cache is built.

### Open investigations (resolve within their phase, not blocking the plan)
- Cheapest read-only source for slot level/name/playtime (db `tb_user` vs SPack GVAS).
- Final icon pixel format confirmation on first real decode (expected BGRA8; BC7 fallback).
- Whether a packed atlas (one file + offsets, mirrors today's sprite model) beats 608 PNGs
  for serving/perf.

### Risk register
- **Kraken port** is the dominant risk/effort. Mitigated by the ooz reference + golden-hash
  tests + isolation in Phase 5 (nothing else depends on Oodle; Phases 1–3 ship value with
  th.gl icons if Phase 5/6 slip). If it stalls, the release can still be cut before Phase 6.
- Game updates could re-pack icons into a different pak / change the XOR seed — the reader
  is data-driven (reads the footer/methods per pak), so only asset paths would move.

### Quality gates
- [ ] `nix develop --command just ci` green at each commit
- [ ] `DSA_SAVE=… just test` round-trip still green
- [ ] `DSA_GAME_DIR=… go test ./internal/pak/... ./internal/oodle/...` (gated) green
- [ ] `just build-windows` produces a single static `.exe` (CGO_ENABLED=0)
- [ ] No game bytes committed (icons, textures, XML fixtures stay in `tmp/`/cache)
- [ ] Docs synced in the same commit; atomic commits on a dedicated branch
