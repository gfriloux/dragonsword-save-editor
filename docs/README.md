# DragonSword Awakening — save format reference

Community documentation of the **DragonSword Awakening (2026)** save format, reverse
engineered while building this editor. Everything here is stated **with certainty**
(verified against real saves and the game); anything partial is flagged as such.

## Contents

- [Save files](save-files.md) — the on-disk files and the account folder.
- [Encryption](encryption.md) — SQLCipher parameters, the constant key, and how to
  open the database with standard tools.
- [Database schema](database.md) — the tables, the ones that matter for editing, and
  the full table list.
- [Content IDs (CIDs)](content-ids.md) — the item/character ID scheme and how names +
  icons map to it.
- [Switches & recipes](switches.md) — the `tb_switch` bitmask system and how cooking
  recipes are unlocked.
- [Paks](paks.md) — why game data (names/icons) can't be read from the game's own
  archives.

## Credits

Item, character and equipment **names and icons** are datamined by
[The Hidden Gaming Lair (th.gl)](https://dragonswordawakening.th.gl), keyed by the same
CIDs the save uses. See [Content IDs](content-ids.md).

## Disclaimer

For personal, offline single-player use and interoperability/research. Names and icons
are © their respective owners. Editing saves can corrupt progress — keep backups.
