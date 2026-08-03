// Command pak-catalog merges the game's own item data — extracted from the encrypted
// .pak archives as raw XML — into the bundled catalog internal/domain/data/items.json.
//
// The game ships its full server dataset as plain XML inside the paks (no .usmap
// needed). Two files matter here:
//
//	GameItemData.xml  one row per item: ID (=CID), ItemType, Category, Grade,
//	                  Name/Desc = string keys, IconName.
//	StringData.xml    one row per string key: all languages, incl. Fr and En.
//
// See .claude/plans/release/pak-extraction/ for how the XML is extracted (a small
// CUE4Parse step, offline like gen-catalog). This tool is pure Go and only reads the
// already-extracted XML fixtures (game-copyright — they live under tmp/, never committed).
//
// It MERGES: pak FR/EN names override/augment the catalog, new CIDs are added, and
// entries the paks do not list (characters live in a separate table) are preserved.
// th.gl icon positions are kept; each item also gains its authoritative ItemType and a
// functional `group` (the game's item-category id). It also emits item_categories.json —
// the localized, ordered consumable-category list the Consumables sidebar renders.
//
//	go run ./cmd/pak-catalog   # defaults to tmp/pak/*.xml + internal/domain/data/*.json
package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const sourceNote = "Item names datamined from the game's own GameItemData/StringData (paks) and merged over the th.gl scrape; icons still courtesy of The Hidden Gaming Lair (th.gl). Regenerate with `just gen-catalog` (icons) then `go run ./cmd/pak-catalog` (names)."

type item struct {
	FR       string `json:"fr,omitempty"`
	EN       string `json:"en,omitempty"`
	Category string `json:"category"`
	Type     string `json:"type,omitempty"`
	Grade    string `json:"grade,omitempty"`    // rarity: normal|rare|superior|epic|legendary
	Group    string `json:"group,omitempty"`    // game item-category id (functional grouping)
	Icon     string `json:"iconPath,omitempty"` // pak-relative UTexture2D asset path (no extension)
	X        int    `json:"x"`
	Y        int    `json:"y"`
}

type catalog struct {
	Source   string          `json:"_source"`
	IconSize int             `json:"iconSize"`
	Items    map[string]item `json:"items"`
}

// consumableCatColor maps the game's item-CategoryType to a UI dot colour. Only these
// types are surfaced as Consumables-panel categories; items of other types (equipment,
// costumes, vehicles, characters…) live in their own panels.
var consumableCatColor = map[string]string{
	"COOK":            "#cf8f6f",
	"NORMAL_MATERIAL": "#6fcf7f",
	"GROW_MATERIAL":   "#e08a5a",
	"KARMA":           "#d75f8f",
	"GEM":             "#b98ce0",
	"VALUABLE":        "#e0a44a",
}

// consumableCategory is one entry of the emitted item_categories.json (consumed by
// internal/domain to build the Consumables sidebar).
type consumableCategory struct {
	Key     string `json:"key"`
	LabelFR string `json:"labelFr"`
	LabelEN string `json:"labelEn"`
	Color   string `json:"color"`
	id      int    // for ordering only
}

type categoriesFile struct {
	Source     string               `json:"_source"`
	Categories []consumableCategory `json:"categories"`
}

type catRow struct {
	ID           string `xml:"ID,attr"`
	CategoryType string `xml:"CategoryType,attr"`
	Name         string `xml:"Name,attr"`
	Memo         string `xml:"Memo,attr"`
}

// coarseCategory maps an item's ItemType to the catalog's coarse category (used for
// the icon-dot colour and CID inference). The finer functional grouping is derived
// separately (see internal/domain/consumable_category.go).
func coarseCategory(itemType string) string {
	switch itemType {
	case "COOKING":
		return "food"
	case "EQUIPMENT":
		return "gear"
	case "VEHICLE":
		return "mount"
	case "COSTUME":
		return "costume"
	case "CHARACTER":
		return "character"
	case "PAY_GOLD", "PAY_CASH", "PAY_EMERALD", "PAY_PLAY_POINT", "PAY_PLAY_POINT_SUB",
		"PAY_ADV_EXP", "PAY_CHARACTER_EXP", "PAY_REMINISCENCE", "PLAY_POINT_KEEP", "PLAY_POINT_RECHARGE":
		return "currency"
	default: // COMMON, COOKING_INGREDIENT(_SPECIAL), GEM, *_SOUL, *_EXP, KARMA, QUEST, MAP, boxes…
		return "material"
	}
}

type strRow struct {
	ID string `xml:"ID,attr"`
	FR string `xml:"Fr,attr"`
	EN string `xml:"En,attr"`
}

type itemRow struct {
	ID       string `xml:"ID,attr"`
	Name     string `xml:"Name,attr"`
	ItemType string `xml:"ItemType,attr"`
	Category string `xml:"Category,attr"`
	Grade    string `xml:"Grade,attr"`
	IconName string `xml:"IconName,attr"`
}

// --- Cooking recipes (CookRecipeData.xml + CookToolData.xml) --------------------
//
// Whether a recipe is *known* is a tb_switch flag: category = CookBook_SwitchData/64,
// bit = CookBook_SwitchData%64 (validated in-game, see docs/switches.md). recipes.json
// stays purely structural — ingredient item-CIDs and category ids resolve to names via
// the existing items.json / item_categories.json at runtime.

type toolRow struct {
	ID       string `xml:"ID,attr"`
	ToolType string `xml:"ToolType,attr"`
	Name     string `xml:"Name,attr"` // StringData key for the localized action label
}

type recipeRow struct {
	SwitchData string `xml:"CookBook_SwitchData,attr"`
	ToolType   string `xml:"ToolType,attr"`
	Cond1Type  string `xml:"IngredientCond1_Type,attr"`
	Cond1Value string `xml:"IngredientCond1_Value,attr"`
	Cond2Type  string `xml:"IngredientCond2_Type,attr"`
	Cond2Value string `xml:"IngredientCond2_Value,attr"`
	Cond3Type  string `xml:"IngredientCond3_Type,attr"`
	Cond3Value string `xml:"IngredientCond3_Value,attr"`
	Cook1      string `xml:"Cook_ID1,attr"`
	Cook2      string `xml:"Cook_ID2,attr"`
	Cook3      string `xml:"Cook_ID3,attr"`
	Cook4      string `xml:"Cook_ID4,attr"`
	Cook5      string `xml:"Cook_ID5,attr"`
}

type toolOut struct {
	Type    string `json:"type"`
	LabelFR string `json:"labelFr"`
	LabelEN string `json:"labelEn"`
}

// recipeIngredient is one required ingredient slot: a whole game item-category
// ("type", id ∈ item_categories.json) or a specific item ("item", id ∈ items.json).
// A repeated slot means "×2/×3" of the same ingredient.
type recipeIngredient struct {
	Kind string `json:"kind"` // "type" | "item"
	ID   int    `json:"id"`
}

type recipeOut struct {
	Key         int                `json:"key"` // CookBook_SwitchData; cat=Key/64, bit=Key%64
	Tool        string             `json:"tool"`
	Ingredients []recipeIngredient `json:"ingredients"`
	Dishes      []int              `json:"dishes"` // Cook_ID1..5: the dish CID per quality tier
}

type recipesFile struct {
	Source  string      `json:"_source"`
	Tools   []toolOut   `json:"tools"`
	Recipes []recipeOut `json:"recipes"`
}

// ingredient converts one (Type, Value) condition pair into a recipeIngredient, or
// (,false) when the slot is empty.
func ingredient(condType, condValue string) (recipeIngredient, bool) {
	v, err := strconv.Atoi(condValue)
	if err != nil {
		return recipeIngredient{}, false
	}
	switch condType {
	case "INGREDIENT_TYPE":
		return recipeIngredient{Kind: "type", ID: v}, true
	case "INGREDIENT_ID":
		return recipeIngredient{Kind: "item", ID: v}, true
	default:
		return recipeIngredient{}, false
	}
}

// iconGameRe pulls the /Game/... asset path (without extension) out of an
// IconName like `…Texture2D'/Game/Art/UI/…/Icon_Item_Common_Gold.Icon_…'`.
var iconGameRe = regexp.MustCompile(`/Game/([^.'"]+)`)

// iconPath maps an IconName to the pak-relative asset path (no extension), e.g.
// "Art/UI/InGame/Icon_Item/Common/Icon_Item_Common_Gold". Empty if none.
func iconPath(iconName string) string {
	m := iconGameRe.FindStringSubmatch(iconName)
	if m == nil {
		return ""
	}
	return m[1]
}

// normGrade maps the game's item Grade to a normalized rarity token, or "".
func normGrade(g string) string {
	switch g {
	case "NORMAL":
		return "normal"
	case "RARE":
		return "rare"
	case "SUPERIOR":
		return "superior"
	case "EPIC":
		return "epic"
	case "LEGENDARY":
		return "legendary"
	default:
		return ""
	}
}

// streamRows decodes every element whose local name is `local` from an XML file into
// T, invoking fn for each. Namespace prefixes (ns1:) are ignored — attributes match by
// local name.
func streamRows[T any](path, local string, fn func(T)) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := xml.NewDecoder(f)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == local {
			var row T
			if err := dec.DecodeElement(&row, &se); err != nil {
				return err
			}
			fn(row)
		}
	}
}

func main() {
	log.SetFlags(0)
	itemsXML := flag.String("items", "tmp/pak/GameItemData.xml", "GameItemData.xml extracted from the paks")
	stringsXML := flag.String("strings", "tmp/pak/StringData.xml", "StringData.xml extracted from the paks")
	categoriesXML := flag.String("categories", "tmp/pak/GameItemCategoryData.xml", "GameItemCategoryData.xml extracted from the paks")
	catalogPath := flag.String("catalog", "internal/domain/data/items.json", "items.json to merge into (in place)")
	categoriesOut := flag.String("categories-out", "internal/domain/data/item_categories.json", "item_categories.json to write")
	recipesXML := flag.String("recipes", "tmp/pak/CookRecipeData.xml", "CookRecipeData.xml extracted from the paks")
	toolsXML := flag.String("tools", "tmp/pak/CookToolData.xml", "CookToolData.xml extracted from the paks")
	recipesOut := flag.String("recipes-out", "internal/domain/data/recipes.json", "recipes.json to write")
	flag.Parse()

	raw, err := os.ReadFile(*catalogPath)
	if err != nil {
		log.Fatalf("read catalog: %v", err)
	}
	var cat catalog
	if err := json.Unmarshal(raw, &cat); err != nil {
		log.Fatalf("parse catalog: %v", err)
	}
	if cat.Items == nil {
		cat.Items = map[string]item{}
	}

	strs := map[string]strRow{}
	if err := streamRows(*stringsXML, "StringData", func(r strRow) { strs[r.ID] = r }); err != nil {
		log.Fatalf("parse strings: %v", err)
	}
	log.Printf("strings: %d", len(strs))

	// Game item categories → the Consumables sidebar (localized, colour by type). Only
	// consumable-relevant CategoryTypes are kept; others (equipment, costume…) are elided.
	var cats []consumableCategory
	consumable := map[string]bool{} // category id -> is a surfaced consumable category
	err = streamRows(*categoriesXML, "GameItemCategoryData", func(r catRow) {
		color, ok := consumableCatColor[r.CategoryType]
		if !ok {
			return
		}
		if strings.Contains(r.Memo, "임시") { // placeholder "temporary data" categories
			return
		}
		s := strs[r.Name]
		if s.FR == "" && s.EN == "" {
			return
		}
		id, _ := strconv.Atoi(r.ID)
		cats = append(cats, consumableCategory{Key: r.ID, LabelFR: s.FR, LabelEN: s.EN, Color: color, id: id})
		consumable[r.ID] = true
	})
	if err != nil {
		log.Fatalf("parse categories: %v", err)
	}
	sort.Slice(cats, func(i, j int) bool { return cats[i].id < cats[j].id })
	cats = append(cats, consumableCategory{Key: "unsorted", LabelFR: "Non trié", LabelEN: "Unsorted", Color: "#8a93a6"})

	var added, updated, noName int
	err = streamRows(*itemsXML, "GameItemData", func(r itemRow) {
		s := strs[r.Name]
		it, ok := cat.Items[r.ID]
		if !ok {
			it = item{Category: coarseCategory(r.ItemType)}
			added++
		} else {
			updated++
		}
		it.Type = r.ItemType
		it.Grade = normGrade(r.Grade)
		it.Icon = iconPath(r.IconName)
		switch r.ItemType {
		case "EQUIPMENT", "VEHICLE", "COSTUME", "CHARACTER":
			it.Group = "" // instance items — shown in their own panels, not Consumables
		default:
			if consumable[r.Category] {
				it.Group = r.Category
			} else {
				it.Group = "unsorted"
			}
		}
		if s.FR != "" {
			it.FR = s.FR
		}
		if s.EN != "" {
			it.EN = s.EN
		}
		if s.FR == "" && s.EN == "" {
			noName++
		}
		cat.Items[r.ID] = it
	})
	if err != nil {
		log.Fatalf("parse items: %v", err)
	}

	cat.Source = sourceNote
	writeJSON(*catalogPath, cat)
	log.Printf("wrote %s: %d items (%d added, %d updated, %d without a resolvable name)",
		*catalogPath, len(cat.Items), added, updated, noName)

	writeJSON(*categoriesOut, categoriesFile{
		Source:     "Item categories datamined from the game's GameItemCategoryData/StringData (paks). Regenerate with `go run ./cmd/pak-catalog`.",
		Categories: cats,
	})
	log.Printf("wrote %s: %d consumable categories", *categoriesOut, len(cats))

	emitRecipes(*recipesXML, *toolsXML, *recipesOut, strs)
}

// emitRecipes parses the cooking tables and writes recipes.json (tools + per-recipe
// switch key, tool, ingredient slots and dish tiers). Tool labels resolve through the
// already-loaded StringData; everything else stays as raw ids resolved at runtime.
func emitRecipes(recipesXML, toolsXML, out string, strs map[string]strRow) {
	var tools []toolOut
	if err := streamRows(toolsXML, "CookToolData", func(r toolRow) {
		s := strs[r.Name]
		tools = append(tools, toolOut{Type: r.ToolType, LabelFR: s.FR, LabelEN: s.EN})
	}); err != nil {
		log.Fatalf("parse cook tools: %v", err)
	}

	var recipes []recipeOut
	err := streamRows(recipesXML, "CookRecipeData", func(r recipeRow) {
		key, err := strconv.Atoi(r.SwitchData)
		if err != nil {
			log.Fatalf("recipe with non-numeric CookBook_SwitchData %q", r.SwitchData)
		}
		var ings []recipeIngredient
		for _, c := range [][2]string{
			{r.Cond1Type, r.Cond1Value}, {r.Cond2Type, r.Cond2Value}, {r.Cond3Type, r.Cond3Value},
		} {
			if ing, ok := ingredient(c[0], c[1]); ok {
				ings = append(ings, ing)
			}
		}
		var dishes []int
		for _, d := range []string{r.Cook1, r.Cook2, r.Cook3, r.Cook4, r.Cook5} {
			if v, err := strconv.Atoi(d); err == nil {
				dishes = append(dishes, v)
			}
		}
		recipes = append(recipes, recipeOut{Key: key, Tool: r.ToolType, Ingredients: ings, Dishes: dishes})
	})
	if err != nil {
		log.Fatalf("parse cook recipes: %v", err)
	}
	sort.Slice(recipes, func(i, j int) bool { return recipes[i].Key < recipes[j].Key })

	writeJSON(out, recipesFile{
		Source:  "Cooking recipes datamined from the game's CookRecipeData/CookToolData (paks). category=key/64, bit=key%64. Regenerate with `go run ./cmd/pak-catalog`.",
		Tools:   tools,
		Recipes: recipes,
	})
	log.Printf("wrote %s: %d recipes, %d tools", out, len(recipes), len(tools))
}

func writeJSON(path string, v any) {
	buf, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(path, append(buf, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
}
