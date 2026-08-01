# Manual tests — v0.3.0 (Characters, Team, Equipment, Gems)

Automated coverage: `internal/domain/domain_test.go` (characters/team read; equipment
enchant/exp/lock round-trip; gems empty-safe). Below is what needs a browser / a human.

Run: `go run ./cmd/dsa-save-editor <copy of a save>.db` (copy, game closed).

## Characters (read-only)

- [ ] Characters section shows one card per owned character with Level / EXP / HP /
      Ascend / Transcend / Soldier.
- [ ] No editable inputs are present (display only). Names fall back to
      "Character <cid>"; ✎ can set a label.

## Team (read-only)

- [ ] Team section shows each saved page with three slots; occupied slots resolve to
      a character name + level; empty slots read "— empty —".
- [ ] No editable inputs.

## Equipment (editable)

- [ ] Equipment section lists active gear; each row has an **Enchant** and **XP**
      number field, a **lock** checkbox, and read-only stat chips.
- [ ] Editing Enchant/XP (Enter/blur) flashes green; toggling lock sticks.
- [ ] **Write to save file** → reopen (or reference `sqlcipher`) shows the new
      ENCHANT_LEVEL / IS_LOCK. Confirmed in the v0.3.0 dev smoke test (enchant=20,
      lock=1 persisted).
- [ ] In-game: the edited enchant/lock is reflected (user check).

## Gems (editable)

- [ ] With no gems, the section shows "No gems in this save." (no error).
- [ ] With gems present, each row shows a lock checkbox that persists.

## Regression

- [ ] Currency, Consumables, and the Database tab still work as in v0.2.0.
