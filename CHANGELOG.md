# Changelog

All notable changes to the DragonSword Awakening save editor. Format inspired by
[Keep a Changelog], versions follow [SemVer]. Generated from Conventional Commits
by git-cliff.

[Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
[SemVer]: https://semver.org/

## [Unreleased]

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
