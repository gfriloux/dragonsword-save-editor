# Content IDs (CIDs)

A **CID** is a stable numeric id for a piece of static game content — an item, a
character, a cooked dish, a cosmetic. The save stores CIDs in `*_CID` columns
([Database schema](database.md)); the same CIDs key the community data on th.gl, which
is how this editor resolves names and icons.

## Ranges

Confirmed from real saves and the th.gl database:

| Kind          | CID pattern            | Examples / range              | Save column(s) |
| ------------- | ---------------------- | ----------------------------- | -------------- |
| Character     | 5-digit, `10xxx`       | `10001` (Eileen) … `10040`    | `tb_character.CHARACTER_CID`, team slots |
| Currency      | `1000xxx`              | `1000001`, `1000002`          | `tb_currency.ITEM_CID` |
| Equipment     | `136xxxx`–`139xxxx`    | `1360001` … `1390155`         | `tb_equipment.ITEM_CID` |
| Cooked dish   | `142xxxx`              | see below                     | `tb_cook_item.ITEM_CID` |
| Potion        | `141xxxx` (**except** `14102xx`) | `1410002` … `1410105`  | `tb_stackable_item.ITEM_CID` |
| Mana upgrade   | `14102xx`              | `1410202`/`1410203`/`1410204` | `tb_stackable_item.ITEM_CID` |
| Material      | `100x`,`131x`,`143x`–`147x`,`19xx`,`200x` | `1000500`, `1450001` … | `tb_stackable_item.ITEM_CID` |
| Mount         | `132xxxx`              | `1320000` … `1320033`         | `tb_vehicle.VEHICLE_CID` |
| Costume       | `999xxxx`              | `9990008` … `9990031`         | `tb_costume.COSTUME_CID` |

Notes:
- The `1000xxx` prefix is shared by currencies (`tb_currency`) and some stackable
  consumables (`tb_stackable_item`): the **table** decides the kind, not the CID alone.
- `stat` references (`MAIN_STAT_CID`, `SUB_STAT_CID*`, `STAT_INFO_CID`) are a separate CID
  space (small 4–5 digit ids); their meaning is **not yet documented**.
- The `141xxxx` range mixes two kinds: real potions (`1410002`–`1410105`) and the three
  **mana upgrade** materials `14102xx` (`1410202` Fragment de mana, `1410203` Cristal de
  mana, `1410204` Minéral de mana — confirmed against a real save's counts). The mana
  ids are equipment-upgrade fuel, not potions.

## Consumable functional categories (curated)

th.gl publishes no functional sub-type for items, so the editor groups stackable
consumables with a **hand-curated** map (`internal/domain/consumable_category.go`),
kept deliberately prudent: only CIDs confirmed from a real save are assigned; the rest
fall through to **Non trié / Unsorted**. Confirmed groups:

| Category      | Rule (CID)                                | Notes |
| ------------- | ----------------------------------------- | ----- |
| Ingredients   | `143xxxx`, `144xxxx`                       | cooking ingredients |
| Breakthrough  | `1450001`–`1450018`, `1450501`–`1450504`  | monster parts + plants (Fruit/Graine/Goutte/Feuille) |
| Crystals      | `146xxxx`                                 | crafting crystals |
| Runes         | `131xxxx`                                 | equipment runes |
| Gear XP       | `1410202`–`1410204`                       | mana upgrade mats |
| Potions       | `141xxxx` except `14102xx`                 | recovery potions |
| Cooked food   | `142xxxx`                                 | cooked dishes |

The `145xxxx` prefix is a **mixed bag**, not a single kind — only two of its sub-blocks
are breakthrough. The rest stay Unsorted until confirmed:

| `145`·B·B | Contents (examples)                                   | Kind (unconfirmed) |
| --------- | ----------------------------------------------------- | ------------------ |
| `00`      | Essence/Peau/Os/Carapace/Molaire/Griffe de monstre    | **breakthrough** ✓ |
| `01`      | Pierre d'amplification / de rafale / hématite / glace | weapon enhancement stones |
| `02`      | Insignes (Orbis, Organa, mercenaires…)                | faction badges |
| `04`      | Grimoires, Fragments de mémoire                        | skill/awakening? |
| `05`      | Fruit de la vitalité, Graine primordiale, Goutte…     | **breakthrough** ✓ |
| `06`      | Cristal de la mémoire / du souvenir / réminiscence    | reminiscence |
| `08`      | Chronique de combat, Grimoire du guerrier, boss drops | mixed |

Off-th.gl items carry no catalog name (only a CID). The four misc ids
`1000800`/`1000801`/`1000802`/`1000804` are the likely character-XP books (three combat
manuals + Livre du Héros) but remain **unverified**, so they stay Unsorted for now.

## Cooked-dish CIDs (`142·M·T·VV`)

Cooking recipes / dishes encode method, tier and variant in the CID:

```
1 4 2 M T V V
      │ │ └─┴── variant (01–29)
      │ └────── tier: 1 Normal · 2 Superior · 3 Rare · 4 Epic · 5 Legendary
      └──────── method: 0 grilled · 1 boiled · 2 sliced
```

- Grilled (`1420xxx`) — e.g. `1420102` = "Poisson grillé (Normal)".
- Boiled  (`1421xxx`).
- Sliced  (`1422xxx`).

A few special dishes (iced drinks, "unknown origin" dish…) use atypical CIDs outside
`1420xxx–1422xxx`. See [Switches & recipes](switches.md) for how *knowing* a recipe is
stored (it is **not** the dish CID).

## Names & icons (th.gl)

The game's item/character names and icons live in DataTables inside the encrypted paks
([Paks](paks.md)), which we can't read. Instead, [th.gl](https://dragonswordawakening.th.gl)
publishes the same data keyed by the same CIDs. Its database pages expose each entry as
`db_<category>_<cid>`:

| th.gl route        | CID kind      |
| ------------------ | ------------- |
| `db/characters`    | characters    |
| `db/equipment`     | equipment     |
| `db/recipes`       | cooked dishes |
| `db/materials_db`  | materials     |
| `db/ingredients_db`| ingredients   |
| `db/costumes`      | costumes      |
| `db/mounts`        | mounts        |

Each row carries the localized name (FR/EN) and an icon rectangle in a shared sprite
sheet (`cdn.th.gl/.../icons.<hash>.webp`, positioned via CSS). This editor scrapes that
into `internal/domain/data/items.json` + a bundled sprite via `cmd/gen-catalog`. Names
and icons are © their respective owners.
