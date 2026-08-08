# Changelog

All notable changes to the DragonSword Awakening save editor. Format inspired by
[Keep a Changelog], versions follow [SemVer]. Generated from Conventional Commits
by git-cliff.

[Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
[SemVer]: https://semver.org/

## [0.17.1] - 2026-08-08

### Documentation

- Document the release workflow
- **plans**: Plan the release CI
- Changelog for v0.17.0

### Build & tooling

- Stamp the version into the release binaries

### Continuous integration

- Publish a GitHub release on v* tags

## [0.17.0] - 2026-08-07

### Features

- **web**: Group the inventory rail under super-category headers
- **web**: Expose consumable category groups via the API
- **cmd**: Tag consumable categories with their CategoryType group

### Documentation

- Plan v0.17.0 (inventory super-categories)
- Changelog for v0.16.1

## [0.16.1] - 2026-08-06

### Bug fixes

- **web**: Don't reveal the inventory screen when writing the save

### Documentation

- Changelog for v0.16.0

## [0.16.0] - 2026-08-06

### Features

- **web**: Titles panel
- **web**: Titles endpoints
- **domain**: Unlock titles (per-title and unlock-all)
- **domain**: Read titles with unlocked state from tb_title
- **cmd**: Pak-titles converter → titles.json

### Documentation

- Document tb_title and the titles screen
- **plans**: Add v0.16.0 titles plan
- Changelog for v0.15.0

## [0.15.0] - 2026-08-05

### Features

- **web**: Costumes & familiers panels
- **web**: Costumes & familiers endpoints
- **domain**: Unlock and equip vehicles (familiers)
- **domain**: Read owned vehicles + per-character mounts
- **domain**: Unlock and equip costumes
- **domain**: Read owned costumes + costume catalog
- **domain**: Mint unique instance DBIDs for inserts

### Documentation

- **plans**: Add v0.15.0 costumes & familiers plan
- Document costume & vehicle tables and screens
- Changelog for v0.14.0

## [0.14.0] - 2026-08-05

### Features

- **web**: Themed lock toggle in equipment/gems
- **web**: Reusable themed modal (confirm/prompt)

### Bug fixes

- **web**: Themed scrollbars
- **web**: Unified input focus & editable-field borders
- **web**: Drop native number-input spinners

### Refactoring

- **web**: Replace native prompt/confirm with themed modal

### Documentation

- V0.14.0 theme polish
- **plan**: V0.14.0 theme polish (point C, T2–T6)

## [0.13.0] - 2026-08-03

### Features

- **web**: Show an icon for category ingredients
- **domain**: Representative icon for category ingredients
- **cmd**: Recipes.json maps ingredient categories to a representative icon
- **web**: Cuisine recipe grid + detail panel with effects
- **domain**: Expose recipe dish effects
- **cmd**: Recipes.json carries dish effects (ContentsBuff)
- **cmd**: Pure-Go pak XML extractor (pak-dump)
- **web**: Cuisine recipe list + per-recipe unlock
- **domain**: Per-recipe cooking known-state + key-accurate unlock
- **cmd**: Pak-catalog emits recipes.json

### Documentation

- Changelog for v0.13.0
- Cooking recipe details (v0.13.0)
- **plan**: V0.13.0 cooking recipe details + per-recipe unlock

## [0.12.0] - 2026-08-02

### Features

- **web**: Render authentic game icons with sprite fallback
- **web**: Authentic item icons extracted from the user's paks (cached)
- **domain**: Carry each item's icon asset path from the paks
- **texture**: Decode cooked DXT5 UTexture2D icons to images
- **oodle**: Pure-Go Oodle (Kraken) decoder via embedded WASM + wazero
- **pak**: Expose per-block compressed data (CompressedBlocks)
- **pak**: Pure-Go reader for the custom version-101 paks
- **web**: Inventaire grid of rarity cells + item detail panel
- **domain**: Carry item rarity (grade) from the paks
- **web**: Embed the Sang & acier fonts (offline)
- **web**: Sang & acier refonte — first-run picker + themed screens
- **web**: First-run config + save-picker; open a slot on demand
- **save**: Discover save slots + screenshots under a game folder
- **config**: Persist the game folder in the OS config dir

### Bug fixes

- **web**: Widen the Inventaire category rail

### Documentation

- DESIGN + changelog for v0.12.0
- **plans**: V0.12.0 chantier — UI refonte + save picker + authentic icons

### Build & tooling

- **nix**: Update vendorHash for the wazero dependency

## [0.11.0] - 2026-08-02

### Features

- **web**: Group Consumables by the game's item categories
- **domain**: Data-driven consumable categories from the game
- **cmd**: Pak-catalog emits game item categories + per-item group
- **cmd**: Add pak-catalog to merge native item names from the paks

### Documentation

- Update changelog for v0.11.0
- **plans**: Mark v0.11.0 Phase 3 delivered (data-driven categories)
- **plans**: Pak-extraction chantier + plan v0.11.0 (native item names)

### Miscellaneous

- **domain**: Regenerate items.json with native names from the paks
- Ignore /tmp scratch (extracted pak fixtures are game-copyright)

## [0.10.0] - 2026-08-02

### Features

- **domain**: Curated bilingual names for off-th.gl items (items_extra.json)
- **domain**: Fold all crystals into Crafting, drop the Crystals category
- **domain**: Add Crafting consumable category
- **domain**: Add Awakening consumable category (gemme d'éveil)
- **domain**: Add Skill and Exchange consumable categories (curated)
- **web**: Category-sidebar Consumables panel (direction B)
- **web**: Expose curated consumable categories in the API
- **domain**: Classify consumables into curated functional categories

### Documentation

- **plans**: Mark v0.10.0 delivered (final taxonomy, gates, version map)
- **plans**: Plan v0.10.0 (curated consumable categories, direction-B panel)

## [0.9.0] - 2026-08-02

### Features

- **web**: List the full th.gl stackable catalog with editable counts

### Documentation

- **plans**: Mark v0.9.0 plan validated and log it in the version map
- **plans**: Plan v0.9.0 (browse the full stackable catalog)

## [0.8.0] - 2026-08-02

### Documentation

- Link the save-format reference and note the doc-sync rule
- Document the switch/recipe bitmask system and paks
- Document the content-ID scheme and th.gl mapping
- Document the database schema
- Add save-format reference index, save files and encryption
- **plans**: Plan v0.8.0 (community save-format reference)

## [0.7.0] - 2026-08-02

### Features

- **web**: Add a Cooking panel with unlock-all-recipes
- **web**: Add /api/game/recipes/unlock-all
- **domain**: Add UnlockAllRecipes (tb_switch categories 15-60)

### Documentation

- Document unlocking all recipes
- **plans**: Plan v0.7.0 (unlock all recipes)

## [0.6.0] - 2026-08-01

### Features

- **web**: Render item icons in the editor panels
- **domain**: Expose icon position and size from the catalog
- **domain**: Regenerate items.json with icon positions and bundle the sprite
- **tooling**: Scrape icon positions and download the th.gl sprite

### Documentation

- Document item icons and sprite regeneration
- **plans**: Plan v0.6.0 (item icons from th.gl sprite)

## [0.5.0] - 2026-08-01

### Features

- **web**: Add an item picker and fill-all to the Consumables panel
- **web**: Add /api/game endpoints for catalog and stackable add/fill
- **domain**: Add stackable upsert, fill-all and catalog listing

### Documentation

- Document adding items and fill-all
- **plans**: Plan v0.5.0 (add stackable items)

## [0.4.0] - 2026-08-01

### Features

- **web**: Add an FR/EN language switch to the editor
- **domain**: Generate bilingual items.json from th.gl (890 names)
- **domain**: Make the item catalog bilingual (fr/en names)
- **tooling**: Add cmd/gen-catalog to fetch th.gl item names

### Documentation

- Document the th.gl item catalog and language switch
- **plans**: Plan v0.4.0 (bilingual item catalog from th.gl)
- Document checks, pre-commit and the dev shell
- **plans**: Plan for dev checks and pre-commit

### Build & tooling

- Add pre-commit config running the Just gates
- Add nix recipes to the Justfile (check, fmt-nix)
- **nix**: Add flake checks (build + gofmt) and dev-shell tooling
- **nix**: Format flake.nix with nixfmt

## [0.3.0] - 2026-08-01

### Features

- **web**: Clarify the Gems panel (socketed vs inventory gems)
- **web**: Add editable Equipment and Gems panels
- **web**: Add read-only Characters and Team panels
- **web**: Add /api/game endpoints for characters, team, equipment and gems
- **domain**: Add equipment and gem accessors with enchant/exp/lock edits
- **domain**: Add read-only character and team accessors
- **web**: Add game Editor view (tabs, currency and consumables panels)
- **web**: Add /api/game endpoints for the editor view
- **domain**: Add typed currency and consumables accessors
- **domain**: Add item catalog (embedded names + prefix inference + user labels)
- **all**: First version

### Bug fixes

- **web**: Transport 64-bit item ids as JSON strings so browser edits hit the right row

### Refactoring

- **domain**: Resolve currency/food category by context, not CID prefix
- Rename Go module to github.com/gfriloux/dragonsword-save-editor

### Documentation

- Document the v0.3.0 panels
- **plans**: Plan v0.3.0 (characters/team read-only, equipment/gems editable)
- Document the domain layer and the Editor/Database views
- **plans**: Plan v0.2.0 (Editor view — currency + consumables)
- **changelog**: Generate initial CHANGELOG.md
- Add CLAUDE.md (project instructions)
- Add PROCEDURE_PLANS.md (working procedures)
- Add DESIGN.md (architecture and invariants)
- **plans**: Add plans layout and working-procedures plan

### Build & tooling

- Enable github remote in cliff.toml
- Add git-cliff changelog configuration
- Add Justfile quality gates and dev tooling

<!-- generated by git-cliff -->
