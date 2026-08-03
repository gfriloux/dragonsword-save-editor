# Switches & recipes

## The `tb_switch` bitmask system

Many "collection" style unlocks are stored as **bitmasks** rather than one row per item:

```sql
CREATE TABLE tb_switch (
  USER_DBID INTEGER,
  CATEGORY  INTEGER,       -- a switch group / word index
  BIT_FIELD INTEGER,       -- 64-bit bitmask (read as uint64)
  PRIMARY KEY (USER_DBID, CATEGORY)
);
```

Each `CATEGORY` holds up to 64 boolean flags in `BIT_FIELD`. A logical set of flags that
needs more than 64 entries spans several consecutive `CATEGORY` rows. Setting all 64 bits
is `BIT_FIELD = -1` (SQLite stores it signed; the game reads it as `0xFFFFFFFFFFFFFFFF`).

Sibling tables reuse the same shape for other content types and reset periods:
`tb_switch_day/week/month`, `tb_unexpected_switch_*`, `tb_episode_switch`,
`tb_reminiscence_switch`, `tb_event_rewarded_switch`, `tb_karma_collection_switch`,
`tb_mercenary_trainee_*_switch`.

## Cooking recipes (known / unlocked)

**Whether a cooking recipe is *known* is a flag in `tb_switch`** — not the presence of the
dish CID anywhere. Each recipe in `CookRecipeData.xml` (now pure-Go extractable from the
paks) carries a `CookBook_SwitchData` key, which maps to:

```
CATEGORY = CookBook_SwitchData / 64
bit      = CookBook_SwitchData % 64
```

### What we determined

- A before/after diff of unlocking one recipe in-game ("Poisson grillé",
  `CookBook_SwitchData=1002`) flipped exactly `tb_switch` **CATEGORY 15, bit 42** —
  `1002/64 = 15`, `1002%64 = 42`. First ground-truth data point.
- Re-validated (v0.13.0): a spike set the 9 keys `CookBook_SwitchData` **4000–4008**
  (→ **CATEGORY 62**, bits 32–40, all previously unknown); the user confirmed in-game that
  exactly those recipes became known. Second, stronger data point.
- The keys span **1001–4008 → categories 15–62** (1025 recipes; tools FryingPan / Pot /
  Knife). Categories `0` and `5` are non-recipe flags and must be left alone.

Since the exact `CookBook_SwitchData` key is known per recipe, the editor now reads/sets
each recipe's **own bit** (per-recipe known state, unlock and lock), and "unlock all" ORs
the exact bit of every recipe key into its category.

### Recipe data & dish effects

`CookRecipeData.xml` gives each recipe its `ToolType` (FryingPan / Pot / Knife), up to
three ingredient conditions (`IngredientCondN_Type` = `INGREDIENT_TYPE` for a whole item
category or `INGREDIENT_ID` for a specific item), and `Cook_ID1..5` — the produced dish
at each of the **5 quality grades** (Normal, Superior, Rare, Epic, Legendary; the grade is
encoded in the CID, e.g. `14201xx`→Normal, `14202xx`→Superior).

The dish's **eat-effect** is one hop further: a COOKING item's `Value1` is a
`ContentsBuffData` id whose `Desc` is the localized effect text (e.g. "restores 850 HP").
The effect scales with the grade, so each of the 5 tiers has its own text.

All of this is datamined into `internal/domain/data/recipes.json` (committed *data*): run
`cmd/pak-dump` to extract `CookRecipeData.xml`, `CookToolData.xml`, `GameItemData.xml`,
`StringData.xml` and `ContentsBuffData.xml` from the paks (pure Go, no CGO), then
`cmd/pak-catalog` to convert them.

### The old blanket, and the bug it hid

Earlier releases had no access to `CookBook_SwitchData` and unlocked by writing
`BIT_FIELD = -1` to categories **15–60**. That range was derived empirically and is
**one category short**: 9 real recipes (keys 4000–4008) live in **category 62** and were
never unlocked by the blanket. The key-accurate map fixes this.

### Caveats

- A clean linear `CookRecipeKey = CID − constant` formula over the *dish CID* does **not**
  fit; the mapping is over `CookBook_SwitchData`, not the CID.
- A handful of **special recipes** (iced drinks `1999xxx`, "Plat d'origine inconnue"
  `1423001`, "Champignons de l'ascension grillés" `1430920`) do **not** appear in
  `CookRecipeData.xml` at all — they live in a different table and are **not** covered by
  this map.
