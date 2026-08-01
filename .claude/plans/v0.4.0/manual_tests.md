# Manual tests — v0.4.0 (bilingual item catalog)

Automated: `internal/domain` tests (catalog fallback/label precedence, accessors).
The catalog data itself is validated by regeneration (`just gen-catalog`) and the
API smoke test (character 10001=Eileen; equipment 1360021 FR/EN). Below needs a browser.

Run: `go run ./cmd/dsa-save-editor <copy of a save>.db` (copy, game closed).

## Names

- [ ] Characters, Equipment, Consumables (food/materials) show **real names**
      (e.g. Eileen; "Casque d'apprenti chevalier") instead of "Character 10001".
- [ ] CIDs th.gl lacks (potions 141x, currencies) still show a sensible fallback
      ("Potion 1410002") and can be labelled (✎).

## Language switch

- [ ] The **FR / EN** toggle in the header switches all names live (Eileen stays
      Eileen; "Casque d'apprenti chevalier" ↔ "Apprentice Knight Helmet").
- [ ] The choice persists across reloads (localStorage).
- [ ] A ✎ label overrides the name in **both** languages.

## Regeneration

- [ ] `just gen-catalog` refetches th.gl and rewrites `internal/domain/data/items.json`
      (review the diff); the build then embeds the new names.

## Regression

- [ ] Editing (currency, consumables, equipment) still persists through Write +
      reopen; the Database tab still works.
