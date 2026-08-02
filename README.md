# DragonSword Awakening — Save Editor

A small, self-contained editor for **DragonSword Awakening (2026)** save files.
It decrypts the game's SQLCipher save database, lets you browse and edit any table
in a local browser UI, and writes it back in a format the game reads unchanged.

Pure Go, **no CGO** — a single static binary that cross-compiles to Windows and
Linux.

## Quick start

```sh
# from a dev shell (Nix): nix develop
go run ./cmd/dsa-save-editor /path/to/<accountId>_Slot<N>.db
```

or with Nix, no checkout of toolchains needed:

```sh
nix run .#default -- /path/to/6144_Slot1.db      # Linux
nix build .#windows                              # produces a Windows .exe
```

The editor opens `http://127.0.0.1:<port>/` in your browser with two tabs:

- **Editor** — a friendly, game-oriented view:
  - **Currency** and **Consumables** (potions, cooked food, materials) with quantity
    steppers; add materials from the catalog by name and fill every stack at once.
  - **Characters** and **Team** — read-only reference views (levels, squads).
  - **Equipment** and **Gems** — edit enchant level, item XP and lock; stat
    references are shown read-only.
  - **Cooking** — one click to unlock all normal cooking recipes.

  Item, character and equipment **names and icons** come from a bundled catalog
  datamined by [th.gl](https://dragonswordawakening.th.gl) — names in **French &
  English**, switchable with the FR/EN toggle in the header — plus category inference
  and your own labels (✎).
- **Database (advanced)** — the raw `tb_*` table browser: double-click any non-key
  cell to edit it.

Make changes in either tab, then **Write to save file**.

> **Close the game before writing.** The running game keeps the database open and
> would overwrite your changes. A timestamped backup of the original
> (`<save>.<timestamp>.bak`) is created automatically the first time you write.

Save files live at:

```
.../DragonSword Awakening/DS/Saved/SaveGames/<accountId>/<accountId>_Slot<N>.db
```

## How the save format works

The `.db` is an **SQLCipher v4** encrypted SQLite database
(`PRAGMA cipher_compatibility = 4`):

| parameter        | value                                   |
| ---------------- | --------------------------------------- |
| cipher           | AES-256-CBC                             |
| page size        | 4096 (first 16 bytes = salt)           |
| KDF              | PBKDF2-HMAC-SHA512, 256000 iterations   |
| page integrity   | HMAC-SHA512 (16-byte IV + 64-byte MAC per page) |

The passphrase is a **hardcoded constant, identical for every save**:

```
13314374259236352028
```

The client embeds the string
`x'1f86483d109b44e81d31181f14b2bba014cf8a174b4fa7f3fe666f075c2ae6c0'`, hashes its
UTF-16 code units with **FNV-1a-64** (offset basis `0xcbf29ce484222325`, prime
`0x100000001b3`), and renders the result as a decimal string via `%llu` — the game
then calls `PRAGMA key = '<that number>'`. `internal/sqlcipher/key.go` reproduces
this derivation, so the key is documented rather than merely pasted.

## Layout

```
internal/sqlcipher/   pure-Go SQLCipher v4 decrypt/encrypt + key derivation
internal/save/        decrypt → edit via modernc.org/sqlite → re-encrypt (+ backup)
internal/domain/      typed game view (currencies, consumables) + item catalog
internal/web/         embedded browser UI + JSON API (/api/*, /api/game/*)
cmd/dsa-save-editor/  CLI entry point (opens the save, serves the UI)
```

The crypto layer only depends on the Go standard library; editing uses the pure-Go
[`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) driver. No component
requires CGO.

## Development

```sh
nix develop            # dev shell: go, staticcheck, just, nixfmt, pre-commit…
just                   # list recipes
just ci                # gofmt + vet + staticcheck + go test + go build
just check             # nix flake check (build + go test + gofmt)
pre-commit install     # run the gates automatically before each commit
```

`just` is the single source of truth for the gates; pre-commit and `nix flake
check` reuse them. (No GitHub CI by choice.)

## Tests

```sh
go test ./...                              # unit tests
DSA_SAVE=/path/to/6144_Slot1.db go test ./...   # also exercises a real save
```

The `DSA_SAVE`-gated tests decrypt a real file, verify the plaintext is valid
SQLite, and confirm an edit round-trips back through the encrypted format.

## Credits

Item, character and equipment **names and icons** are datamined by
[The Hidden Gaming Lair (th.gl)](https://dragonswordawakening.th.gl) and bundled as
`internal/domain/data/items.json` (names + icon positions) and
`internal/web/static/sprite.webp` (the icon sprite sheet). Regenerate both with
`just gen-catalog` (`cmd/gen-catalog`). All names and icons are © their respective
owners.

## Disclaimer

For personal, offline single-player use. Editing saves can corrupt progress and
may violate the game's terms of service; keep the automatic backups.
