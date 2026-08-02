# Manual tests — v0.9.0 (browse the full stackable catalog)

Run against a real save via the running editor (`just run` / the built binary on a
save copy). The game must be **closed** before writing.

## M1 — The full catalog is always listed
1. Open the editor, **Editor** tab, **Consumables** panel.
2. The **material** group lists **179** material/ingredient rows (owned + not owned,
   no toggle), each with name, icon (when available) and an editable stepper. Rows
   you don't own show count `0`.
3. The **food** and **potion** groups still list only what you own (unchanged).

## M2 — Give yourself an item you don't own (0 → X)
1. Pick a not-owned material (count `0`). Set it to e.g. `50`, commit (Enter).
2. **Write** the save. Reopen the editor → the item shows `50`. (Optional: load the
   save in-game and confirm the item is present.)

## M3 — Edit an item you already own
1. In the material list, find an owned material, change its count, commit, Write.
2. Reopen → new count persists. Owned potions/food are untouched.

## Non-goals verified absent
- No currency add, no cooked-food add, no gear/mount/costume/character add.
- No potion appears that isn't already owned (th.gl lists none).
