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
// display colour (hex) for the UI category dot.
type ConsumableCategory struct {
	Key     string `json:"key"`
	LabelFR string `json:"labelFr"`
	LabelEN string `json:"labelEn"`
	Color   string `json:"color"`
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
