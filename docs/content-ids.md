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
| Character XP   | `141xxxx` (**except** `14102xx`) | `1410002` Manuel des bases du combat | `tb_stackable_item.ITEM_CID` |
| Gear XP (mana) | `14102xx`              | `1410202`/`1410203`/`1410204` | `tb_stackable_item.ITEM_CID` |
| Material      | `100x`,`131x`,`143x`–`147x`,`19xx`,`200x` | `1000500`, `1450001` … | `tb_stackable_item.ITEM_CID` |
| Mount         | `132xxxx`              | `1320000` … `1320033`         | `tb_vehicle.VEHICLE_CID` |
| Costume       | `999xxxx`              | `9990008` … `9990031`         | `tb_costume.COSTUME_CID` |
| Title         | `210xxxx`              | `2100000` … `2104xxx` (108)   | `tb_title` bit (`CATEGORY = id/64`) |

Notes:
- The `1000xxx` prefix is shared by currencies (`tb_currency`) and some stackable
  consumables (`tb_stackable_item`): the **table** decides the kind, not the CID alone.
- `stat` references (`MAIN_STAT_CID`, `SUB_STAT_CID*`, `STAT_INFO_CID`) are a separate CID
  space (small 4–5 digit ids); their meaning is **not yet documented**.
- **Titles** (`210xxxx`) are not one-row-per-title: unlocking is a **bit** in `tb_title`
  (`CATEGORY = id/64`, `bit = id%64`), the same bitmask shape as recipes. The 108 titles,
  names, font colours and stat bonuses come from `AccountTitleData.xml` + `StringData.xml`
  in the paks (see `cmd/pak-titles`). Ranges observed: 7 categories `32812`–`32876`.
- The `999xxxx` **costume** space bundles both **outfits and weapon skins** (all
  `ItemType=COSTUME`, indistinguishable by type); they pair even/odd — an outfit `999xxx0`
  with its matching weapon skin `999xxx1` (confirmed by the icon paths, e.g.
  `…_Costume_Cerese_Unique_01` vs `…_Cerese_Unique_Weapon_01`). A character equips both at
  once. `VEHICLE_CID 1320033` is currently untranslated in the game data (Korean
  `흉악한 새끼 용`), so its catalog name is the raw Korean until a locale provides one.
- The `141xxxx` range is **not** potions: `1410002`–`1410105` are **character-XP** training
  texts (`ItemType=CHARACTER_EXP`, e.g. `1410002` "Manuel des bases du combat"), while
  `14102xx` are the three **mana** upgrade mats (`ItemType=EQUIPMENT_EXP`: `1410202` Fragment
  de mana, `1410203` Cristal de mana, `1410204` Minéral de mana). This was settled by the
  paks' own data (`ItemType`), not guessed.
- **Item names and types now come from the game's own data** (`GameItemData.xml` +
  `StringData.xml`, extracted from the paks — see [Paks](paks.md)), merged into
  `items.json` by `cmd/pak-catalog`. Each item carries its authoritative `ItemType`
  (COOKING, COOKING_INGREDIENT, EQUIPMENT, EQUIPMENT_EXP, CHARACTER_EXP, GEM,
  CHARACTER_MASTER_SOUL, COMMON…), which will drive the functional categories below.

## Consumable functional categories (from the game's own data)

The Consumables panel groups items by the **game's own item categories**, datamined from
the paks (`GameItemData.ns:Category` → `GameItemCategoryData` → `StringData`) by
`cmd/pak-catalog` into `internal/domain/data/item_categories.json`. Each item carries its
category id in `Item.group`; the sidebar shows the localized category name. This replaced
an earlier hand-curated CID map — the game taxonomy is authoritative and already localized.

Each game category has a **CategoryType** (parent) and a numeric id. Only the
consumable-relevant types are surfaced in the panel; equipment / vehicles / costumes /
characters have their own panels (their items carry no consumable group).

| CategoryType      | Example categories (id → name)                                   |
| ----------------- | ---------------------------------------------------------------- |
| `NORMAL_MATERIAL` | 1700 Viande, 1701 Poisson, 1704 Légumes, 1705 Champignons…       |
| `GROW_MATERIAL`   | 1600 Amélioration des héros, 1602 …d'équipement, 1603 Matériaux de monstres, 1604 Ressources minières, 1605 Ressources de collecte, 1607 Pierres de caractéristiques, 1608 Fabrication d'équipement |
| `GEM`             | 1500–1505 Runes (Détermination, Protection, Vitalité…)           |
| `KARMA`           | 1400–1408 Karma (Amplitude, Écrasement, Rafale…)                 |
| `COOK`            | 1200 Griller, 1201 Bouillir, 1202 Découper…                      |
| `VALUABLE`        | 1800 Pierre d'invocation, 1802 Matériaux d'échange, 1803 Clés de trésor, 1900–1902 boîtes |

Item-level facts settled by this data (things earlier guessed from CID prefixes):
- `141xxxx` is **not** potions — `1410002`–`1410105` are **character-XP** texts
  (`ItemType=CHARACTER_EXP`, e.g. `1410002` "Manuel des bases du combat"), `14102xx` are
  the mana upgrade mats (`ItemType=EQUIPMENT_EXP`). Their game category is 1600 / 1602.
- The awakening gem `1450410` ("Gemme d'éveil: Stigmate", `CHARACTER_MASTER_SOUL`) is filed
  under 1802 "Matériaux d'échange" by the game.
- The `1450101`–`1450127` "Pierre de …" stones are 1607 "Pierres de caractéristiques";
  crystals `1460101`+ are 1604 "Ressources minières"; monster parts `1450001`+ are 1603.

`items_extra.json` (a hand-curated bilingual name supplement) is retained as a fallback for
the rare item the datamine can't name, but is now essentially empty — `pak-catalog` names
virtually everything.

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
