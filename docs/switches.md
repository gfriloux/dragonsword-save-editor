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
dish CID anywhere. The game's `DCookRecipeData` DataTable gives each recipe a
`CookRecipeKey`, which maps to:

```
CATEGORY = 15 + (CookRecipeKey / 64)
bit      =       CookRecipeKey % 64
```

### What we determined

- A before/after diff of unlocking one recipe in-game ("Poisson grillé") flipped exactly
  `tb_switch` **CATEGORY 15, bit 42** — the single ground-truth data point.
- Progressive test saves (setting whole category ranges, validated in-game) showed the
  **normal recipes** — grilled/boiled/sliced (`1420xxx`/`1421xxx`/`1422xxx`) — occupy
  **categories 15–60**. Categories `0` and `5` are non-recipe flags and must be left
  alone.
- Setting `BIT_FIELD = -1` for categories **15–60** marks every normal recipe as known,
  with no observed side effects. Bits beyond real recipes are inert (the game only checks
  keys that exist in its recipe table).

### Caveats

- A clean linear `CookRecipeKey = CID − constant` formula does **not** fit, so we unlock
  by the whole category range rather than per-recipe bits.
- A handful of **special recipes** with atypical CIDs (iced drinks `1999xxx`,
  "Plat d'origine inconnue" `1423001`, "Champignons de l'ascension grillés" `1430920`)
  live outside categories 15–60 and are **not** covered — they need their own diff.

This editor exposes it as a one-click "unlock all recipes" (`UnlockAllRecipes`, which
upserts categories 15–60 to `-1`).
