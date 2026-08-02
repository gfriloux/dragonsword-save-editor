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
	Grade    string `json:"grade,omitempty"` // rarity: normal|rare|superior|epic|legendary
	Group    string `json:"group,omitempty"` // game item-category id (functional grouping)
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
