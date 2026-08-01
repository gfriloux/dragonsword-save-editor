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
   uses only the standard library.

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

4. **No personal data in the repository.** Real saves, `SPack_*.sav`, screenshots
   and backups are local fixtures only, never committed (see `.gitignore`).

5. **Layering is one-directional.** Lower layers never import higher ones:

   ```
   internal/sqlcipher   crypto codec + key derivation (stdlib only)
        ▲
   internal/save        decrypt → edit via modernc.org/sqlite → re-encrypt (+ backup)
        ▲
   internal/web         embedded browser UI + JSON API
        ▲
   cmd/dsa-save-editor  CLI entry point
   ```

## Structure

| Path                     | Responsibility                                             |
| ------------------------ | --------------------------------------------------------- |
| `internal/sqlcipher/`    | Pure-Go SQLCipher v4 decrypt/encrypt and key derivation.  |
| `internal/save/`         | Open/edit/persist a save; schema introspection helpers.   |
| `internal/web/`          | Embedded static UI (`go:embed`) + JSON API over a `Save`. |
| `cmd/dsa-save-editor/`   | Flags, save auto-detection, HTTP server, browser launch.  |
| `flake.nix`              | Dev shell + `packages.default` (Linux) + `packages.windows`. |
| `.claude/plans/`         | Planning documents (see `PROCEDURE_PLANS.md`).            |

## Reverse-engineered facts

The save format was recovered from `DSClient-Win64-Shipping.exe`. The full findings
(SQLCipher parameters, the constant passphrase `13314374259236352028`, and its
FNV-1a-64 derivation from the embedded seed) are documented in
`internal/sqlcipher/key.go` and the project `README.md`.

## Out of scope (for now)

- Server/online features — the save is treated as local single-player state only.
- Decoding item/character CID → human-readable names (lives in the game's cooked
  DataTables inside the `.pak` archives). A natural future feature, tracked as its
  own plan when picked up.
- Editing the companion `SPack_*.sav` (UE GVAS slot metadata); only the `.db` is
  edited.
