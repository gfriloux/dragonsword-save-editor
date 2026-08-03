package domain

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
)

// Cooking recipes are datamined into data/recipes.json (see cmd/pak-catalog). Whether a
// recipe is *known* is a flag in the generic tb_switch bitmask table:
//
//	category = key / 64      bit = key % 64
//
// where key is the recipe's CookBook_SwitchData. Validated in-game (docs/switches.md);
// keys span 1001-4008 → categories 15-62. recipes.json is purely structural: ingredient
// category ids and dish/item CIDs resolve to names through the item catalog here.

// RecipeTool is a cooking tool with localized action labels.
type RecipeTool struct {
	Type    string `json:"type"`
	LabelFR string `json:"labelFr"`
	LabelEN string `json:"labelEn"`
}

// EffectText is one localized effect line (the dish's eat-effect, via ContentsBuffData).
type EffectText struct {
	FR string `json:"fr,omitempty"`
	EN string `json:"en,omitempty"`
}

// recipeSeed is one raw recipe as stored in recipes.json.
type recipeSeed struct {
	Key         int              `json:"key"`
	Tool        string           `json:"tool"`
	Ingredients []ingredientSeed `json:"ingredients"`
	Dishes      []int64          `json:"dishes"`
	EffectName  *EffectText      `json:"effectName"`
	Effects     []EffectText     `json:"effects"` // parallel to Dishes; scales with tier
}

type ingredientSeed struct {
	Kind string `json:"kind"` // "type" | "item"
	ID   int64  `json:"id"`
}

var (
	recipeTools []RecipeTool
	recipeSeeds []recipeSeed
)

func init() {
	raw, err := seedFS.ReadFile("data/recipes.json")
	if err != nil {
		panic(fmt.Sprintf("domain: reading recipes.json: %v", err))
	}
	var f struct {
		Tools   []RecipeTool `json:"tools"`
		Recipes []recipeSeed `json:"recipes"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		panic(fmt.Sprintf("domain: parsing recipes.json: %v", err))
	}
	recipeTools = f.Tools
	recipeSeeds = f.Recipes
}

// RecipeTools returns the ordered cooking-tool list (a copy).
func RecipeTools() []RecipeTool {
	out := make([]RecipeTool, len(recipeTools))
	copy(out, recipeTools)
	return out
}

// switchPos maps a recipe key to its (category, bit) in tb_switch.
func switchPos(key int) (category, bit int) { return key / 64, key % 64 }

// RecipeIngredient is one required ingredient slot, resolved for display. For an item it
// is a specific CID; for a type it is a whole game item-category. Qty collapses repeated
// slots (a recipe listing the same ingredient twice needs ×2). Owned is what the save
// holds: the exact stackable count for an item, or the sum over the category for a type.
type RecipeIngredient struct {
	Kind     string `json:"kind"` // "type" | "item"
	ID       int64  `json:"id"`
	NameFR   string `json:"nameFr"`
	NameEN   string `json:"nameEn"`
	Qty      int    `json:"qty"`
	Owned    int64  `json:"owned"`
	IconPath string `json:"iconPath,omitempty"`
}

// Recipe is a cooking recipe resolved against a save.
type Recipe struct {
	Key         int                `json:"key"`
	Category    int                `json:"category"`
	Bit         int                `json:"bit"`
	Tool        string             `json:"tool"`
	Known       bool               `json:"known"`
	Ingredients []RecipeIngredient `json:"ingredients"`
	Dish        Item               `json:"dish"`                 // representative dish (tier 1) for name/icon
	Dishes      []int64            `json:"dishes"`               // the dish CID at each of the 5 quality tiers
	EffectName  *EffectText        `json:"effectName,omitempty"` // the eat-effect's category label
	Effects     []EffectText       `json:"effects,omitempty"`    // effect text per dish tier
}

// Recipes lists every cooking recipe resolved against the save: known/unknown state read
// from tb_switch, ingredients resolved to names with owned counts, and the produced dish.
func (g *Game) Recipes() ([]Recipe, error) {
	uid, err := g.UserID()
	if err != nil {
		return nil, err
	}
	known, err := g.recipeKnownBits(uid)
	if err != nil {
		return nil, err
	}
	ownedByCID, ownedByGroup, err := g.stackableOwnership(uid)
	if err != nil {
		return nil, err
	}
	catLabels := consumableCategoryLabels()

	out := make([]Recipe, 0, len(recipeSeeds))
	for _, rs := range recipeSeeds {
		cat, bit := switchPos(rs.Key)
		r := Recipe{
			Key:        rs.Key,
			Category:   cat,
			Bit:        bit,
			Tool:       rs.Tool,
			Known:      known[cat]>>uint(bit)&1 == 1,
			Dishes:     rs.Dishes,
			EffectName: rs.EffectName,
			Effects:    rs.Effects,
		}
		if len(rs.Dishes) > 0 {
			r.Dish = g.cat.LookupCtx(rs.Dishes[0], "food")
		}
		r.Ingredients = resolveIngredients(g.cat, rs.Ingredients, ownedByCID, ownedByGroup, catLabels)
		out = append(out, r)
	}
	return out, nil
}

// resolveIngredients collapses repeated slots and resolves each to display names + owned.
func resolveIngredients(cat *Catalog, seeds []ingredientSeed, ownedByCID, ownedByGroup map[int64]int64, catLabels map[string][2]string) []RecipeIngredient {
	var out []RecipeIngredient
	index := map[string]int{} // kind:id -> position in out (to collapse repeats)
	for _, s := range seeds {
		k := s.Kind + ":" + strconv.FormatInt(s.ID, 10)
		if i, ok := index[k]; ok {
			out[i].Qty++
			continue
		}
		ing := RecipeIngredient{Kind: s.Kind, ID: s.ID, Qty: 1}
		switch s.Kind {
		case "item":
			it := cat.Lookup(s.ID)
			ing.NameFR, ing.NameEN, ing.IconPath = it.NameFR, it.NameEN, it.IconPath
			ing.Owned = ownedByCID[s.ID]
		case "type":
			if l, ok := catLabels[strconv.FormatInt(s.ID, 10)]; ok {
				ing.NameFR, ing.NameEN = l[0], l[1]
			}
			ing.Owned = ownedByGroup[s.ID]
		}
		index[k] = len(out)
		out = append(out, ing)
	}
	return out
}

// recipeKnownBits reads the recipe-bearing tb_switch categories into a cat→bitfield map.
func (g *Game) recipeKnownBits(uid int64) (map[int]uint64, error) {
	rows, err := g.s.DB().Query(`SELECT CATEGORY, BIT_FIELD FROM tb_switch WHERE USER_DBID=?`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[int]uint64{}
	for rows.Next() {
		var cat, field int64
		if err := rows.Scan(&cat, &field); err != nil {
			return nil, err
		}
		m[int(cat)] = uint64(field)
	}
	return m, rows.Err()
}

// stackableOwnership returns owned stackable counts by CID and, folded through the
// catalog's functional group, by game item-category id (for "type" ingredients).
func (g *Game) stackableOwnership(uid int64) (byCID, byGroup map[int64]int64, err error) {
	rows, err := g.s.DB().Query(
		`SELECT ITEM_CID, STACK_CNT FROM tb_stackable_item WHERE USER_DBID=?`, uid)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	byCID, byGroup = map[int64]int64{}, map[int64]int64{}
	for rows.Next() {
		var cid, cnt int64
		if err := rows.Scan(&cid, &cnt); err != nil {
			return nil, nil, err
		}
		byCID[cid] = cnt
		if gid, err := strconv.ParseInt(g.cat.Lookup(cid).Group, 10, 64); err == nil {
			byGroup[gid] += cnt
		}
	}
	return byCID, byGroup, rows.Err()
}

// consumableCategoryLabels maps a category key to {FR, EN} labels.
func consumableCategoryLabels() map[string][2]string {
	m := map[string][2]string{}
	for _, c := range consumableCategories {
		m[c.Key] = [2]string{c.LabelFR, c.LabelEN}
	}
	return m
}

// SetRecipeKnown flips one recipe's known flag: it read-modify-writes the recipe's bit in
// its tb_switch category, preserving the other bits.
func (g *Game) SetRecipeKnown(key int, known bool) error {
	uid, err := g.UserID()
	if err != nil {
		return err
	}
	cat, bit := switchPos(key)
	var cur int64
	err = g.s.DB().QueryRow(
		`SELECT BIT_FIELD FROM tb_switch WHERE USER_DBID=? AND CATEGORY=?`, uid, cat).Scan(&cur)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	field := uint64(cur)
	if known {
		field |= 1 << uint(bit)
	} else {
		field &^= 1 << uint(bit)
	}
	_, err = g.s.Exec(
		`INSERT OR REPLACE INTO tb_switch (USER_DBID, CATEGORY, BIT_FIELD) VALUES (?,?,?)`,
		uid, cat, int64(field))
	return err
}

// UnlockAllRecipes marks every known cooking recipe as known by OR-ing the exact bit of
// each recipe key into its tb_switch category. Unlike the old blanket (categories 15-60),
// this covers every real recipe — including the category-62 tail (keys 4000-4008) the
// blanket silently missed — and touches no non-recipe bit.
func (g *Game) UnlockAllRecipes() error {
	uid, err := g.UserID()
	if err != nil {
		return err
	}
	masks := map[int]uint64{}
	for _, rs := range recipeSeeds {
		cat, bit := switchPos(rs.Key)
		masks[cat] |= 1 << uint(bit)
	}
	tx, err := g.s.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for cat, mask := range masks {
		var cur int64
		err := tx.QueryRow(`SELECT BIT_FIELD FROM tb_switch WHERE USER_DBID=? AND CATEGORY=?`, uid, cat).Scan(&cur)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		field := int64(uint64(cur) | mask)
		if _, err := tx.Exec(
			`INSERT OR REPLACE INTO tb_switch (USER_DBID, CATEGORY, BIT_FIELD) VALUES (?,?,?)`,
			uid, cat, field); err != nil {
			return err
		}
	}
	return tx.Commit()
}
