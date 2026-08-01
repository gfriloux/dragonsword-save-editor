# Manual tests — v0.2.0 (Editor view)

Automated coverage lives in `internal/domain/*_test.go` (catalog inference/labels,
currency & consumables accessors round-tripping a real save). The items below are
the parts that need a human / a browser and are not automated.

Run: `go run ./cmd/dsa-save-editor <copy of a save>.db` (use a **copy**, game closed).

## Editor view

- [ ] App opens on the **Editor** tab; header shows the save path.
- [ ] **Currency** section lists currencies; each shows a category dot, name
      (fallback "Currency <cid>" when unseeded), CID, and an amount field.
- [ ] Editing an amount (Enter / blur / ± buttons) flashes green; **Write to save
      file** then reports "Saved ✓" and creates a `<save>.<ts>.bak` backup.
- [ ] Reopening the written file shows the new amount (or verify with the reference
      `sqlcipher` CLI).
- [ ] **Consumables** section groups items by category (food / potion / material /
      misc) with counts; 100x stackable items are **not** shown as "currency".
- [ ] ± / typed quantity edits persist through Write + reopen.
- [ ] **✎ name** on a row prompts for a label; after saving it, the name updates and
      persists across restarts (stored under the OS config dir, not the save).

## Database view (regression)

- [ ] The **Database (advanced)** tab still lists all tables, pages, and edits raw
      cells exactly as before.

## Cross-checks (smoke, automatable-ish)

- [ ] `/api/game/currency` GET returns resolved currencies; POST updates one.
- [ ] `/api/game/consumables` GET returns both `stackable` and `cook` kinds.
- [ ] `/api/game/stack` POST updates the right table; verified via `sqlcipher` after
      `/api/save` (done in the v0.2.0 dev smoke test).
