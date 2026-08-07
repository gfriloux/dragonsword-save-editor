# Manual tests — v0.17.0 (inventory super-categories)

Run the editor against a real save, open the **Inventaire / Consommables** panel.

## Rail grouping
- [ ] The category rail shows **section headers** above the category buttons.
- [ ] Headers appear in this order (only those with ≥1 non-empty category shown):
      Cuisine → Runes → Effets → Ingrédients → Matériaux → Objets de valeur → Autres.
- [ ] Under **Cuisine**: Griller, Bouillir, Découper, … (cooking categories).
- [ ] Under **Runes**: Détermination, Protection, Vitalité, Concentration,
      Amplification, Coup fatal.
- [ ] Under **Ingrédients**: Viande, Poisson, Légumes, … (raw NORMAL_MATERIAL).
- [ ] COOK and Ingrédients are **separate** headers (not merged).
- [ ] `unsorted` (if any owned) sits under **Autres**, last.

## Behaviour unchanged
- [ ] Clicking a category still selects it and lists its items (headers are inert).
- [ ] Owned/total counts on each `.cat-link` are unchanged.
- [ ] Language toggle (FR/EN) switches both category labels **and** header labels.
- [ ] "Remplir" and the item detail panel behave as before.

## Label sanity (provisional — revisit)
- [ ] The six header labels read naturally in-context; note any to rename.
