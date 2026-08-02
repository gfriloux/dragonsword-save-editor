# Save files

## Location

DragonSword Awakening is an Unreal Engine 5 game (internal project name `DS`). Saves
live under the game's `Saved` directory:

```
.../DragonSword Awakening/DS/Saved/SaveGames/<accountId>/
    <accountId>_Slot<N>.db          ← the save database (encrypted)
    SPack_Slot<N>.sav               ← slot metadata (Unreal GVAS)
    ScreenShot_<N>.png              ← the slot thumbnail
```

- `<accountId>` — the numeric account/user id; it is also the folder name (e.g. `6144`).
- `<N>` — the save slot number (`1`, `2`, `3`, `4`).

## `<accountId>_Slot<N>.db` — the save database

An **encrypted SQLite database** (SQLCipher v4). This is where all editable state
lives: characters, inventory, currencies, equipment, cooking, progression, etc. Its
size is always a multiple of 4096 bytes. See [Encryption](encryption.md) to open it and
[Database schema](database.md) for its contents.

## `SPack_Slot<N>.sav` — slot metadata

An **Unreal Engine GVAS** save (starts with the ASCII magic `GVAS`). It holds the
slot-summary shown in the load menu (playtime, level, location text…). It is **not
encrypted** but uses Unreal's binary property serialization. The editor does not touch
it; editing item/character data in the `.db` does not require changing it.

## `ScreenShot_<N>.png` — thumbnail

A regular PNG screenshot used as the slot's thumbnail. Cosmetic; unrelated to the data.

## Notes

- The `.rne` files next to the game binaries are a copy/variant of the `.exe`; unrelated
  to saves.
- Backups: the game itself creates transient `*.bak` / journal files next to the `.db`
  while saving. Make your own backup of `<accountId>_Slot<N>.db` before editing.
