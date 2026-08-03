# Manual tests — v0.13.0 (Cuisine recipe details)

_Grown during dev, run at validation. UI/game-facing checks the automated gates cannot cover._

## Cuisine screen
- [ ] Recipe list renders; filter by tool (Poêle / Marmite / Couteau) works.
- [ ] Known / unknown badge matches the opened save's actual state.
- [ ] Required ingredients show resolved names + icons; specific ingredients show
      owned/needed counts from the save.
- [ ] Produced dish name/icon shown.

## Per-recipe unlock (round-trip, on a COPY of a save, game closed)
- [ ] Toggle one unknown recipe → known, save, re-open in-game → that recipe (and only it)
      is known.
- [ ] "Tout débloquer" → every recipe known in-game, including the category-62 tail that
      the old blanket missed.
- [ ] Re-open the edited save in the editor → state persisted, no corruption, backup written.
