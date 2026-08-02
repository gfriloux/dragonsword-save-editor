# Manual tests — v0.7.0 (unlock all recipes)

Automated: `internal/domain` test (UnlockAllRecipes sets categories 15–60 to all-bits,
leaves category 0 untouched; round-trips a real save). The in-game effect was already
validated with an equivalent hand-made test save.

Run: `go run ./cmd/dsa-save-editor <copy of a save>.db` (copy, game closed).

- [ ] The **Cooking** section shows an "Unlock all recipes" button.
- [ ] Clicking it (confirm) reports success; **Write to save file** persists it.
- [ ] Reopen (or `sqlcipher`): `tb_switch` categories 15–60 are all `-1`.
- [ ] In-game: every normal grilled / boiled / sliced recipe is known; nothing else
      changed. (The ~3 special-CID recipes remain locked — expected, deferred.)

## Regression
- [ ] All other panels (currency, consumables, add/fill, equipment, language, icons)
      and the Database tab still work.
