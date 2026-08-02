# Paks (reading item data from the game)

The game ships all content in Unreal Engine `.pak` archives under `DS/Content/Paks/`
(54 chunks, ~19 GB). They are AES-encrypted with a **non-standard footer** (reports
"version 101"), so stock tools that assume the standard layout (`repak`, `UnrealPak`)
can't read them.

## They *are* readable — with the right tool

- **The AES key is a fixed, public constant:**
  `0x263479C442D45B7EEDE7B3A36BBB3C3B39EF9178A2F82AB694FB410AB15E01AD`.
  (An earlier attempt to *derive* it from the running game's memory failed precisely
  because it is static, not dynamically computed.)
- **CUE4Parse** (the library behind FModel) carries a dedicated game profile
  `GAME_DragonSwordAwakening` (base UE 5.3) that understands the custom footer. With the
  key + that profile it mounts all 54 paks and reads files — confirmed headless on Linux
  (dotnet + a ~20-line CUE4Parse program). No `.usmap` is needed for the data below.

## The data is plain XML (server dump inside the paks)

Under `DS/Content/__GeneratedGameData__/Server/`, the game bundles its **entire server
dataset**: a `dsgamedb_sqlite.sql` dump and one XML per table in `XML/GameData/`. The two
that matter for the catalog:

- **`GameItemData.xml`** — one row per item: `ID` (= the [CID](content-ids.md)),
  `ItemType`, `Category` (numeric), `Grade`, `Name`/`Desc` (string keys), `IconName`.
  ~1212 items — far more than th.gl's 890, and including everything th.gl omits
  (potions/manuals, mana, awakening gems…).
- **`StringData.xml`** — one row per string key with **all 11 languages, including `Fr`
  and `En`** (the game ships native French; the `L10N/` folders are only voice overrides).

## How this project uses it

`cmd/pak-catalog` (pure Go) parses those two XML files and merges the native FR/EN names
(and each item's `ItemType`) into `internal/domain/data/items.json`, over the th.gl
scrape (which still provides the **icons**/sprite via `cmd/gen-catalog`). Extracting the
XML from the paks is a small offline CUE4Parse step (like `gen-catalog`'s scraping) — the
extractor and recipe live under `tmp/pak/` (the extracted XML are game-copyright fixtures,
never committed). See `.claude/plans/release/pak-extraction/`.

(The **save** database uses standard SQLCipher with a recoverable constant key — see
[Encryption](encryption.md) — which is why editing saves is tractable.)
