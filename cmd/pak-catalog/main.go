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
// th.gl icon positions are kept; each item also gains its authoritative ItemType.
//
//	go run ./cmd/pak-catalog   # defaults to tmp/pak/*.xml + internal/domain/data/items.json
package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"io"
	"log"
	"os"
)

const sourceNote = "Item names datamined from the game's own GameItemData/StringData (paks) and merged over the th.gl scrape; icons still courtesy of The Hidden Gaming Lair (th.gl). Regenerate with `just gen-catalog` (icons) then `go run ./cmd/pak-catalog` (names)."

type item struct {
	FR       string `json:"fr,omitempty"`
	EN       string `json:"en,omitempty"`
	Category string `json:"category"`
	Type     string `json:"type,omitempty"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
}

type catalog struct {
	Source   string          `json:"_source"`
	IconSize int             `json:"iconSize"`
	Items    map[string]item `json:"items"`
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
	catalogPath := flag.String("catalog", "internal/domain/data/items.json", "items.json to merge into (in place)")
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

	strings := map[string]strRow{}
	if err := streamRows(*stringsXML, "StringData", func(r strRow) { strings[r.ID] = r }); err != nil {
		log.Fatalf("parse strings: %v", err)
	}
	log.Printf("strings: %d", len(strings))

	var added, updated, noName int
	err = streamRows(*itemsXML, "GameItemData", func(r itemRow) {
		s := strings[r.Name]
		it, ok := cat.Items[r.ID]
		if !ok {
			it = item{Category: coarseCategory(r.ItemType)}
			added++
		} else {
			updated++
		}
		it.Type = r.ItemType
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
	buf, err := json.MarshalIndent(cat, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*catalogPath, append(buf, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s: %d items (%d added, %d updated, %d without a resolvable name)",
		*catalogPath, len(cat.Items), added, updated, noName)
}
