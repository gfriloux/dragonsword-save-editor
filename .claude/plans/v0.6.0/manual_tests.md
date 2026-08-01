# Manual tests — v0.6.0 (item icons)

The catalog + sprite are validated by regeneration and the API smoke test (icon
fields present; `/sprite.webp` serves 200). The visual result needs a browser.

Run: `go run ./cmd/dsa-save-editor <copy of a save>.db` (copy, game closed).

## Icons

- [ ] Consumables, Equipment, Characters and Team rows/cards show the item's **real
      icon** (from the sprite), not just a coloured dot.
- [ ] Items th.gl lacks an icon for (~10, e.g. some currencies/potions) fall back to
      the category dot — no broken image.
- [ ] Icons stay correct after switching FR/EN and after re-rendering (add/fill).
- [ ] The Add-material datalist still works (names + CID).

## Regeneration

- [ ] `just gen-catalog` refreshes `items.json` (positions) **and**
      `internal/web/static/sprite.webp` together (the sprite hash rotates); icons still
      line up afterwards.

## Regression

- [ ] All editing (currency, consumables, equipment, add/fill), the language switch and
      the Database tab still work.
