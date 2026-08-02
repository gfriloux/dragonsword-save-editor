# Manual tests — v0.10.0 (curated categories, direction-B panel)

Against the real save via the running editor, game **closed** before writing:

```
DSA_SAVE=/path/to/6144_Slot1.db nix develop --command just run
```

## M1 — Category sidebar renders
1. Editor tab → Consumables. A category rail lists (delivered order): Ingrédients,
   Percée, Éveil, Compétences, Fabrication, Runes, XP équipement, Échange, Potions,
   Plats cuisinés, Non trié — each with a colored dot and `owned N / total`.
2. Selecting a category shows only its items in the item pane.

## M2 — Percée is correct and prudent
1. Open **Percée**: monster parts (essence/peau/os/carapace/molaire/griffe) +
   Fruit de la vitalité, Graine primordiale, Goutte de pureté, Feuille de vigueur.
2. It must **not** contain the "Pierre de …" stones or crystals (those are in
   **Fabrication**), skill books (**Compétences**), or the awakening gem (**Éveil**).

## M3 — Mana lands in XP équipement, not Potions
1. **XP équipement** contains Fragment/Cristal/Minéral de mana (`1410202/203/204`).
2. **Potions** contains the real potions (`1410002`–`1410105`) and **not** the mana.

## M4 — Set an unowned item 0 → X within a category
1. In **Ingrédients** (or any category), pick a not-owned item (count 0), set to 50,
   commit, **Write**, reopen → persists.

## M5 — Non trié is the safety net
1. **Non trié** holds everything uncategorised (enhancement stones, badges,
   `1000800/801/802/804`, `147x`, `151x`, …) and its rows are still editable 0 → X.

## Non-goals verified
- No XP-personnage category yet. No currencies/instance items here. Owned potions/
  cooked food still appear (owned only; th.gl lists no unowned potions).
