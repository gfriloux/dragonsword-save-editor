# Manual tests — v0.5.0 (add stackable items)

Automated: `internal/domain` tests (upsert, fill-all, catalog listing round-trip on a
real save; verified against `sqlcipher` in the dev smoke test). Below needs a browser.

Run: `go run ./cmd/dsa-save-editor <copy of a save>.db` (copy, game closed).

## Add a material

- [ ] In **Consumables**, the toolbar shows an "Add a material by name…" box backed by
      the catalog (type to filter; entries read "Name (CID)").
- [ ] Picking a material not owned + a quantity + **Add** inserts it; it appears in the
      material group with that count.
- [ ] Adding one already owned **sets** its quantity (upsert, not increment).
- [ ] After **Write to save file**, reopening shows the new item; it shows up in-game
      (user check).

## Fill all

- [ ] "Set all stacks to N" + **Fill** (with confirm) sets every stackable to N.
- [ ] Change persists through Write + reopen.

## Regression

- [ ] Existing quantity edits, currency, equipment, the language switch and the
      Database tab still work.
