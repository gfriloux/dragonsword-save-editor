# DESIGN.md — DragonSword Awakening save editor

> This document defines the spirit, the structure and the **invariants** of the
> project. Before adding anything, check that it fits here. If it does not, the
> answer is no.

## Purpose

A small, self-contained editor for **DragonSword Awakening (2026)** save files, for
personal offline single-player use. It decrypts the game's SQLCipher database, lets
the user inspect and edit any table through a local browser UI, and writes it back
in a format the game reads unchanged.

## Invariants

These are hard constraints. Changing one is a DESIGN decision, not an incidental
implementation change.

1. **Pure Go, no CGO.** The whole program must cross-compile to Windows and Linux
   with `CGO_ENABLED=0` and no external toolchain. Never introduce a dependency
   that requires CGO. Editing uses `modernc.org/sqlite` (pure Go); the crypto layer
   uses only the standard library. The game's Oodle-compressed pak assets are
   decoded by running an embedded WebAssembly build of the `ooz` decompressor on
   the pure-Go `wazero` runtime (`internal/oodle`) — so even that stays CGO-free and
   ships inside the single static binary.

2. **The crypto layer targets exactly the game's format.** SQLCipher v4
   (`cipher_compatibility = 4`): AES-256-CBC, 4096-byte pages, PBKDF2-HMAC-SHA512
   with 256000 iterations, per-page HMAC-SHA512. The passphrase is the game's fixed
   constant, derived in code from the embedded seed via FNV-1a-64 rather than
   pasted blindly. We do not implement other SQLCipher configurations until a save
   needs them.

3. **Never corrupt a save.**
   - Reading a save never mutates the original file.
   - Writing is atomic (write to a temp file in the same directory, then rename).
   - A timestamped backup of the original is made the first time a save is written.
   - The editor must not write while the game holds the database open — the user is
     told to close the game first.

4. **No personal data — and no game art — in the repository.** Real saves,
   `SPack_*.sav`, screenshots and backups are local fixtures, never committed (see
   `.gitignore`). Authentic item icons are **extracted per-user at runtime** from the
   player's own paks and cached in the OS cache dir; the copyrighted art is never
   bundled or committed. Datamined *data* (names, categories, icon paths) is committed
   (`items.json`) — it is fact, not art.

5. **Layering is one-directional.** Lower layers never import higher ones:

   ```
   internal/sqlcipher   crypto codec + key derivation (stdlib only)
   internal/oodle       Oodle (Kraken) decompressor — embedded ooz.wasm on wazero
   internal/pak         pure-Go reader for the custom (version-101) UE5 paks
   internal/texture     cooked DXT5 UTexture2D → image
        ▲   (the four above are low-level codecs; pak uses oodle)
   internal/save        decrypt → edit via modernc.org/sqlite → re-encrypt (+ backup)
        ▲
   internal/config      persisted user settings (the remembered game folder)
   internal/domain      typed game view (currencies, consumables…) + item catalog
   internal/icons       extract + cache authentic icons (pak → oodle → texture → PNG)
        ▲
   internal/web         embedded browser UI + JSON API
        ▲
   cmd/dsa-save-editor  CLI entry point
   ```

## Structure

| Path                     | Responsibility                                             |
| ------------------------ | --------------------------------------------------------- |
| `internal/sqlcipher/`    | Pure-Go SQLCipher v4 decrypt/encrypt and key derivation.  |
| `internal/oodle/`        | Oodle/Kraken decompressor (embedded `ooz.wasm` run on wazero). |
| `internal/pak/`          | Pure-Go reader for the game's custom (version-101) UE5 paks. |
| `internal/texture/`      | Decode cooked DXT5 `UTexture2D` icons to images.          |
| `internal/save/`         | Open/edit/persist a save; schema introspection; slot discovery. |
| `internal/config/`       | Persist the remembered game folder (OS config dir).       |
| `internal/domain/`       | Typed game view (currencies, consumables) + item catalog. |
| `internal/icons/`        | Extract + cache authentic item icons (pak → oodle → texture → PNG). |
| `internal/web/`          | Embedded static UI (`go:embed`) + JSON API (`/api/*`, `/api/game/*`). |
| `cmd/dsa-save-editor/`   | Flags, first-run config, HTTP server, browser launch.     |
| `cmd/pak-catalog/`       | Offline: merge names/categories/grade/icon paths into `items.json`; emit `recipes.json` (cooking recipes + dish effects). |
| `cmd/pak-dump/`          | Offline: extract named data XML from the paks in pure Go (feeds `pak-catalog`). |
| `flake.nix`              | Dev shell + `packages.default` (Linux) + `packages.windows`. |
| `.claude/plans/`         | Planning documents (see `PROCEDURE_PLANS.md`).            |

## The UI

The app is player-facing. On first run it asks for the **game folder** (remembered
in the OS config dir), lists the existing **saves** with their screenshots, and — once
a slot is opened — shows the "Sang & acier" editor: an own-line tab bar over game
screens (Accueil · Monnaies · Inventaire · Personnages · Équipe · Équipement · Gemmes ·
Cuisine · **SQL brut**). Item names come from a bundled bilingual catalog
(`internal/domain/data/items.json`, names/categories/grade/icon paths from the paks,
merged over th.gl), switchable FR/EN, with user-editable labels. Item **icons** are the
authentic in-game art, extracted per-user from the paks and cached (falling back to the
th.gl sprite, then a category dot). **Cuisine** is a recipe book: a grid of dish cards
(known/locked) and a detail panel showing the eat-effect, the required ingredients
(resolved to icons, with owned/required counts), and a per-recipe known/lock toggle.
**SQL brut** is the generic `tb_*` table browser, kept as the power user's escape hatch.

## Reverse-engineered facts

The save format was recovered from `DSClient-Win64-Shipping.exe`. The full findings
(SQLCipher parameters, the constant passphrase `13314374259236352028`, and its
FNV-1a-64 derivation from the embedded seed) are documented in
`internal/sqlcipher/key.go` and the project `README.md`.

The **paks** are read in pure Go (`internal/pak`): a standard UE5 pak whose footer
version is obfuscated to 101, its AES-256-ECB index (public key) carrying two extra
XOR layers, reverse-engineered from CUE4Parse's DragonSword profile. `cmd/pak-dump`
extracts the game's data XML; `cmd/pak-catalog` mines names/categories/grade/icon paths
into `items.json` and cooking data into `recipes.json`; icon textures are decoded on
demand at runtime.

**Cooking recipes** live in `CookRecipeData.xml`: each recipe's `CookBook_SwitchData`
key maps to a `tb_switch` flag (`category = key/64`, `bit = key%64`, validated in-game —
see `docs/switches.md`), and its dish `Value1` chains through `ContentsBuffData` to the
localized eat-effect. Both are committed as *data* (`recipes.json`), not art.

## Out of scope (for now)

- Server/online features — the save is treated as local single-player state only.
- Editing the companion `SPack_*.sav` (UE GVAS slot metadata); only the `.db` is
  edited (its screenshot is read for the save picker).
- A handful of **special recipes** (iced drinks `1999xxx`, `1423001`, `1430920`) that
  are not in `CookRecipeData.xml` — a different table, not covered by the recipe unlock.
