# Phase 0 — audit (v0.10.0)

Date: 2026-08-02. Branch `feat/consumable-categories` off `main` @ `95b82e9`
(v0.9.0 merged + tagged).

`nix develop --command just ci` → **green** (fmt-check, vet, staticcheck, test,
build). Working tree clean except untracked local `.envrc` and `tmp/` scratch
(gitignored; holds the save-dump reference `tmp/dump/out.txt`).

## State relevant to this plan

- v0.9.0 shipped: the Consumables panel already lists the full th.gl stackable
  catalog (owned + not-owned) as one flat "material" group via a client-side merge of
  `/api/game/catalog` + `/api/game/consumables`, editable through
  `/api/game/stackable` (upsert) and `/api/game/stack` (update).
- `internal/domain` resolves a coarse `Item.Category` (currency/potion/food/material/
  gear/character/…) from seed + CID inference. There is **no** finer functional
  grouping yet — this plan adds it (`ClassifyConsumable`).
- `tb_stackable_item` = `(USER_DBID, ITEM_CID)` PK + `STACK_CNT`, upsert-safe.

## Real-save harvest (this session)

Decrypted `6144_Slot1.db` via `internal/save` + `domain.Consumables()`; 144 owned
consumable rows. Findings feeding the taxonomy are recorded in `plan.md`
(§Findings / §Locked taxonomy). Key confirmed facts:
- `14102xx` = mana upgrade mats (`1410202/203/204`), confirmed by the user's counts
  9999/9994/8683 — distinct from real potions (`1410002`–`1410105`).
- `145x` is a mixed prefix (breakthrough + enhancement stones + faction + memory +
  boss drops) → explicit CID rules required.
- Off-th.gl items (`1000800/801/802/804`) are unnamed candidates for character-XP
  books; left "Non trié" pending in-game verification (user decision).

Conclusion: domain (curated classifier) + web (API + direction-B panel) + docs. No
save-format write change; deterministic unit test for the classifier.
