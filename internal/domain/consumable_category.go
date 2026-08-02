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
	CatAwakening    = "awakening"    // character awakening (éveil) materials
	CatSkill        = "skill"        // skill-upgrade materials
	CatCraft        = "craft"        // equipment-fabrication materials (incl. crystals)
	CatRune         = "rune"         // equipment runes
	CatEquipXP      = "equip_xp"     // equipment upgrade (mana) materials
	CatExchange     = "exchange"     // character-exchange tokens (buy characters)
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
	{CatAwakening, "Éveil", "Awakening", "#b0d04a"},
	{CatSkill, "Compétences", "Skill materials", "#d75f8f"},
	{CatCraft, "Fabrication", "Crafting", "#d0b06a"},
	{CatRune, "Runes", "Runes", "#b98ce0"},
	{CatEquipXP, "XP équipement", "Gear XP", "#e0c14a"},
	{CatExchange, "Échange", "Exchange", "#e0a44a"},
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
	// Explicit CID sets (confirmed in-game) take precedence over range rules.
	case cid == 1000500: // Invitation du Destin — buy characters via the exchange
		return CatExchange
	case cid == 1450410: // Gemme d'éveil — character awakening material
		return CatAwakening
	case cid >= 1450101 && cid <= 1450127: // all "Pierre de …" equipment-craft stones
		return CatCraft
	case cid == 1450820 || cid == 1450821: // Souvenirs.../Oubli... entremêlés
		return CatCraft
	case cid >= 1460101 && cid <= 1460199: // Cristal de cohésion / coin / veine
		return CatCraft
	case cid == 1450811 || cid == 1450812 || cid == 1450815 || cid == 1450816 || cid == 1450823:
		// Chronique de combat / Grimoire du guerrier / Nageoire épineuse /
		// Crochet ensanglanté / Sang du sage — skill-upgrade materials.
		return CatSkill
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
	case cid >= 1310000 && cid <= 1319999:
		return CatRune
	default:
		return CatUnsorted
	}
}
