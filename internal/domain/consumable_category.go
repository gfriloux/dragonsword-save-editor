package domain

// Curated functional categories for consumable / stackable items.
//
// th.gl exposes no functional sub-typing of items, so this grouping is maintained
// by hand from confirmed facts (a real save cross-referenced with the th.gl
// catalog; see docs/content-ids.md and .claude/plans/v0.10.0/). The guiding rule is
// prudence: only CIDs we are confident about are assigned a category; everything
// else falls through to "unsorted" rather than risk mis-sorting. The map grows as
// more items are identified in-game.

// Consumable category keys.
const (
	CatIngredient   = "ingredient"   // cooking ingredients (meat, fish, vegetables…)
	CatBreakthrough = "breakthrough" // character breakthrough (percée) materials
	CatCrystal      = "crystal"      // crafting crystals
	CatRune         = "rune"         // equipment runes
	CatEquipXP      = "equip_xp"     // equipment upgrade (mana) materials
	CatPotion       = "potion"       // potions
	CatCooked       = "cooked"       // cooked dishes
	CatUnsorted     = "unsorted"     // anything not yet curated
)

// ConsumableCategory is one curated functional group, with bilingual labels and a
// display colour (hex) for the UI category dot.
type ConsumableCategory struct {
	Key     string `json:"key"`
	LabelFR string `json:"labelFr"`
	LabelEN string `json:"labelEn"`
	Color   string `json:"color"`
}

// consumableCategories is the ordered list shown in the Consumables sidebar. Order
// here is the display order; "unsorted" is always last.
var consumableCategories = []ConsumableCategory{
	{CatIngredient, "Ingrédients", "Ingredients", "#6fcf7f"},
	{CatBreakthrough, "Percée", "Breakthrough", "#e08a5a"},
	{CatCrystal, "Cristaux", "Crystals", "#5aa9e0"},
	{CatRune, "Runes", "Runes", "#b98ce0"},
	{CatEquipXP, "XP équipement", "Gear XP", "#e0c14a"},
	{CatPotion, "Potions", "Potions", "#5ad0c0"},
	{CatCooked, "Plats cuisinés", "Cooked food", "#cf8f6f"},
	{CatUnsorted, "Non trié", "Unsorted", "#8a93a6"},
}

// ConsumableCategories returns the ordered curated category list (a copy).
func ConsumableCategories() []ConsumableCategory {
	out := make([]ConsumableCategory, len(consumableCategories))
	copy(out, consumableCategories)
	return out
}

// ClassifyConsumable maps a content id to its curated functional category key.
// Order matters: narrower rules (mana inside the 141x range, breakthrough inside
// the mixed 145x range) are checked before the broader range they sit in. Unknown
// CIDs return CatUnsorted.
func ClassifyConsumable(cid int64) string {
	switch {
	case cid >= 1410202 && cid <= 1410204: // Fragment/Cristal/Minéral de mana
		return CatEquipXP
	case cid >= 1410000 && cid <= 1419999:
		return CatPotion
	case cid >= 1420000 && cid <= 1429999:
		return CatCooked
	case cid >= 1430000 && cid <= 1449999:
		return CatIngredient
	case cid >= 1450001 && cid <= 1450018: // monster parts (essence/hide/bone/carapace/molar/claw)
		return CatBreakthrough
	case cid >= 1450501 && cid <= 1450504: // Fruit/Graine/Goutte/Feuille
		return CatBreakthrough
	case cid >= 1460000 && cid <= 1469999:
		return CatCrystal
	case cid >= 1310000 && cid <= 1319999:
		return CatRune
	default:
		return CatUnsorted
	}
}
