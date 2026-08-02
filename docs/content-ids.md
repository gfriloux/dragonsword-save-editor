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
| Potion        | `141xxxx`              | `1410002` …                   | `tb_stackable_item.ITEM_CID` |
| Material      | `100x`,`131x`,`143x`–`147x`,`19xx`,`200x` | `1000500`, `1450001` … | `tb_stackable_item.ITEM_CID` |
| Mount         | `132xxxx`              | `1320000` … `1320033`         | `tb_vehicle.VEHICLE_CID` |
| Costume       | `999xxxx`              | `9990008` … `9990031`         | `tb_costume.COSTUME_CID` |

Notes:
- The `1000xxx` prefix is shared by currencies (`tb_currency`) and some stackable
  consumables (`tb_stackable_item`): the **table** decides the kind, not the CID alone.
- `stat` references (`MAIN_STAT_CID`, `SUB_STAT_CID*`, `STAT_INFO_CID`) are a separate CID
  space (small 4–5 digit ids); their meaning is **not yet documented**.

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
