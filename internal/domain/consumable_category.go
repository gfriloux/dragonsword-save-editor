package domain

import (
	"encoding/json"
	"fmt"
)

// Functional consumable categories are the game's own item categories, datamined from
// the paks (GameItemCategoryData + StringData) by cmd/pak-catalog into
// data/item_categories.json. Each item carries its category id in Item.Group (from the
// seed); items without one fall to "unsorted". This replaced the earlier hand-curated
// CID rules — the game's taxonomy is authoritative and already localized.

const groupUnsorted = "unsorted"

// ConsumableCategory is one curated functional group, with bilingual labels and a
// display colour (hex) for the UI category dot. Group is the game CategoryType
// (COOK, GEM, …) — the super-category the inventory rail groups the category under;
// see ConsumableGroups. (Distinct from Item.Group, which holds a category Key.)
type ConsumableCategory struct {
	Key     string `json:"key"`
	LabelFR string `json:"labelFr"`
	LabelEN string `json:"labelEn"`
	Color   string `json:"color"`
	Group   string `json:"group"`
}

// ConsumableGroup is a super-category: it gathers several ConsumableCategory rows
// (all sharing the same game CategoryType) under one header in the inventory rail.
// The labels and ordering are hand-authored — the game enum keys are English-technical
// and unlocalized — and revisited against the running game.
type ConsumableGroup struct {
	Key     string `json:"key"` // game CategoryType, matches ConsumableCategory.Group
	LabelFR string `json:"labelFr"`
	LabelEN string `json:"labelEn"`
}

// consumableGroups is the display order of the super-category headers. Categories
// whose Group is none of these (today: only "unsorted") fall under a trailing
// "Autres / Other" header rendered by the UI.
var consumableGroups = []ConsumableGroup{
	{Key: "COOK", LabelFR: "Cuisine", LabelEN: "Cooking"},
	{Key: "GEM", LabelFR: "Runes", LabelEN: "Runes"},
	{Key: "KARMA", LabelFR: "Effets", LabelEN: "Effects"},
	{Key: "NORMAL_MATERIAL", LabelFR: "Ingrédients", LabelEN: "Ingredients"},
	{Key: "GROW_MATERIAL", LabelFR: "Matériaux", LabelEN: "Materials"},
	{Key: "VALUABLE", LabelFR: "Objets de valeur", LabelEN: "Valuables"},
}

var consumableCategories []ConsumableCategory

func init() {
	raw, err := seedFS.ReadFile("data/item_categories.json")
	if err != nil {
		panic(fmt.Sprintf("domain: reading item_categories.json: %v", err))
	}
	var f struct {
		Categories []ConsumableCategory `json:"categories"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		panic(fmt.Sprintf("domain: parsing item_categories.json: %v", err))
	}
	consumableCategories = f.Categories
}

// ConsumableCategories returns the ordered consumable category list (a copy).
func ConsumableCategories() []ConsumableCategory {
	out := make([]ConsumableCategory, len(consumableCategories))
	copy(out, consumableCategories)
	return out
}

// ConsumableGroups returns the ordered super-category headers (a copy).
func ConsumableGroups() []ConsumableGroup {
	out := make([]ConsumableGroup, len(consumableGroups))
	copy(out, consumableGroups)
	return out
}
