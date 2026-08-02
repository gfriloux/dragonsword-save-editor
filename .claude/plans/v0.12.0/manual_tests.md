# Manual tests — v0.12.0

UI is not auto-tested (PROCEDURE §4). Run the app and check the flow visually.

```bash
# fresh first-run: wipe the remembered config to see the folder prompt
rm -f "${XDG_CONFIG_HOME:-$HOME/.config}/dsa-save-editor/config.json"
nix develop --command go run ./cmd/dsa-save-editor -no-open
# open the printed URL in a browser
```

## Phase 1 + 2 — first-run, picker, shell

- [ ] First run shows **Accueil** with an empty "Dossier du jeu" prompt (no save path needed on the CLI).
- [ ] Entering a valid game folder (contains `DS/Content/Paks` or `DS/Saved/SaveGames`) and validating lists the save slots; an invalid folder toasts an error.
- [ ] Each slot card shows SLOT n, the account id, the **screenshot** (when present), the last-played time and the path.
- [ ] The remembered folder persists across restarts; "Changer" lets you pick another.
- [ ] `needs-save` tabs are dimmed until a slot is opened; clicking a slot opens it and jumps to **Inventaire**.
- [ ] Top bar: brand, save path (mono), FR/EN toggle, "N modif." badge, "Écrire dans la save" CTA. The CTA + badge appear only once a save is open.
- [ ] Nav bar is on its own line and scrolls horizontally in a narrow window (CTA never wraps).

## Screens (with a save open)

- [ ] **Monnaies** — cards with a stepper; gold steps by 10 000, others by 50; edits bump the modif badge.
- [ ] **Inventaire** — category rail + **grid of rarity-bordered cells** + **detail panel**: click a cell → stepper + presets (0/99/999/MAX); cells show quantity, dim at 0, and an **amber dot** when changed from the save; the detail shows an `ancienne → nouvelle` diff with "Annuler cette modification"; "Remplir" sets a category; the amber dots clear after a write; FR/EN switches names + the rarity legend.
- [ ] **Personnages** / **Équipe** — read-only cards / slots render.
- [ ] **Équipement** / **Gemmes** — enchant/XP/lock editable; stat chips read-only.
- [ ] **Cuisine** — "Tout débloquer" works (recipe material details come in Phase 3).
- [ ] **SQL brut** — table list filter, paging, double-click cell edit; edited cell flashes amber.
- [ ] **Écrire dans la save** — toasts success, resets the modif badge; a `.bak` is created; the game must be closed.

## Authentic icons (Phase 6)

- [ ] With a game folder set, item icons across all screens render the **real
      in-game art** (via `/api/icon?cid=`), extracted from the user's own paks and
      cached under the OS cache dir (`~/.cache/dsa-save-editor/icons/`). First load
      of a screen builds them (a brief pop-in); later loads are instant from cache.
- [ ] Items with no authentic icon (or before a game folder is set) fall back to
      the th.gl sprite, then a category dot — no broken images.

## Notes / known gaps (later phases)

- Recipe material details (Phase 3) not yet wired into Cuisine.
