package domain

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

//go:embed data/items.json data/items_extra.json data/item_categories.json
var seedFS embed.FS

// Item is a resolved catalog entry for a content id (CID). Names are provided in
// both languages so the UI can switch without a round-trip.
type Item struct {
	CID      int64  `json:"cid"`
	NameFR   string `json:"nameFr"`
	NameEN   string `json:"nameEn"`
	Category string `json:"category"` // currency | potion | food | material | gear | character | costume | mount | misc
	Group    string `json:"group"`    // game item-category id (functional consumable grouping); "unsorted" if none
	Grade    string `json:"grade"`    // rarity: normal | rare | superior | epic | legendary; "" if unknown
	IconPath string `json:"iconPath"` // pak-relative UTexture2D asset path (no extension); "" if none
	Known    bool   `json:"known"`    // true if a seed name or a user label exists
	Icon     bool   `json:"icon"`     // true if the item has a sprite cell
	IconX    int    `json:"iconX"`    // sprite object-position (px), negative
	IconY    int    `json:"iconY"`
}

type seedEntry struct {
	FR       string `json:"fr"`
	EN       string `json:"en"`
	Category string `json:"category"`
	Group    string `json:"group"`
	Grade    string `json:"grade"`
	IconPath string `json:"iconPath"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
}

type seedFile struct {
	IconSize int                  `json:"iconSize"`
	Items    map[string]seedEntry `json:"items"`
}

// Catalog resolves CIDs to bilingual names and categories, combining an embedded
// seed (generated from th.gl), user-provided labels, and CID-based inference.
type Catalog struct {
	seed      map[string]seedEntry
	overrides map[string]string // cid -> user label (applies to both languages)
	ovrPath   string
	iconSize  int
}

// IconSize is the sprite cell size in pixels (the icons live in sprite.webp).
func (c *Catalog) IconSize() int { return c.iconSize }

// LoadCatalog reads the embedded seed and, if present, the user overrides file at
// overridesPath. A missing overrides file is not an error; pass "" to keep labels
// in memory only.
func LoadCatalog(overridesPath string) (*Catalog, error) {
	raw, err := seedFS.ReadFile("data/items.json")
	if err != nil {
		return nil, err
	}
	var sf seedFile
	if err := json.Unmarshal(raw, &sf); err != nil {
		return nil, fmt.Errorf("domain: parsing embedded items.json: %w", err)
	}
	if sf.Items == nil {
		sf.Items = map[string]seedEntry{}
	}
	// Merge curated bilingual names for off-th.gl items (items_extra.json), only for
	// CIDs the th.gl seed does not already cover so a regen never clobbers them.
	if er, err := seedFS.ReadFile("data/items_extra.json"); err == nil {
		var ef seedFile
		if err := json.Unmarshal(er, &ef); err != nil {
			return nil, fmt.Errorf("domain: parsing embedded items_extra.json: %w", err)
		}
		for cid, e := range ef.Items {
			if _, ok := sf.Items[cid]; !ok {
				sf.Items[cid] = e
			}
		}
	}
	iconSize := sf.IconSize
	if iconSize == 0 {
		iconSize = 64
	}
	c := &Catalog{seed: sf.Items, overrides: map[string]string{}, ovrPath: overridesPath, iconSize: iconSize}
	if overridesPath != "" {
		if b, err := os.ReadFile(overridesPath); err == nil {
			if err := json.Unmarshal(b, &c.overrides); err != nil {
				return nil, fmt.Errorf("domain: parsing overrides %q: %w", overridesPath, err)
			}
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return c, nil
}

// Lookup resolves a CID with category from the seed or CID inference.
func (c *Catalog) Lookup(cid int64) Item { return c.LookupCtx(cid, "") }

// LookupCtx resolves a CID, letting the caller assert a category it knows from
// context (e.g. a row from tb_currency is a currency regardless of its CID).
// Per language, name precedence is: user label > seed name > generated fallback.
// Category precedence: seed category > categoryHint > CID inference.
func (c *Catalog) LookupCtx(cid int64, categoryHint string) Item {
	key := strconv.FormatInt(cid, 10)
	seed, hasSeed := c.seed[key]
	category := seed.Category
	if category == "" {
		category = categoryHint
	}
	if category == "" {
		category = inferCategory(cid)
	}
	label, hasLabel := c.overrides[key]

	pick := func(seedName string) string {
		if label != "" {
			return label
		}
		if seedName != "" {
			return seedName
		}
		return fallbackName(category, cid)
	}
	return Item{
		CID:      cid,
		NameFR:   pick(seed.FR),
		NameEN:   pick(seed.EN),
		Category: category,
		Group:    seed.Group, // "" for non-consumables / items without a game category
		Grade:    seed.Grade,
		IconPath: seed.IconPath,
		Known:    hasLabel || (hasSeed && (seed.FR != "" || seed.EN != "")),
		Icon:     hasSeed && (seed.X != 0 || seed.Y != 0),
		IconX:    seed.X,
		IconY:    seed.Y,
	}
}

// Entries returns every seeded catalog item (resolved), sorted by CID — used to
// populate the "add item" picker.
func (c *Catalog) Entries() []Item {
	out := make([]Item, 0, len(c.seed))
	for k := range c.seed {
		if cid, err := strconv.ParseInt(k, 10, 64); err == nil {
			out = append(out, c.Lookup(cid))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CID < out[j].CID })
	return out
}

// SetLabel records a user-provided name for a CID (both languages) and persists it
// if the catalog was loaded with an overrides path.
func (c *Catalog) SetLabel(cid int64, name string) error {
	key := strconv.FormatInt(cid, 10)
	if name == "" {
		delete(c.overrides, key)
	} else {
		c.overrides[key] = name
	}
	if c.ovrPath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(c.ovrPath), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c.overrides, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.ovrPath, b, 0o644)
}

// inferCategory maps the leading digits of a CID to a coarse category. "currency"
// and "food" are asserted from context by the accessors, not inferred here — the
// 100x prefix is shared by currencies and stackable consumables.
func inferCategory(cid int64) string {
	switch prefix3(cid) {
	case "141":
		return "potion"
	case "142":
		return "food"
	case "143", "144", "145", "146", "147":
		return "material"
	}
	if prefixN(cid, 2) == "13" { // weapons/armour item CIDs
		return "gear"
	}
	return "misc"
}

func fallbackName(category string, cid int64) string {
	title := map[string]string{
		"currency":  "Currency",
		"potion":    "Potion",
		"food":      "Food",
		"material":  "Material",
		"gear":      "Gear",
		"character": "Character",
		"costume":   "Costume",
		"mount":     "Mount",
		"misc":      "Item",
	}[category]
	if title == "" {
		title = "Item"
	}
	return fmt.Sprintf("%s %d", title, cid)
}

func prefix3(cid int64) string { return prefixN(cid, 3) }

func prefixN(cid int64, n int) string {
	s := strconv.FormatInt(cid, 10)
	if len(s) >= n {
		return s[:n]
	}
	return s
}

// DefaultOverridesPath returns the per-user labels file location, or "" if the
// user config dir cannot be determined.
func DefaultOverridesPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "dsa-save-editor", "labels.json")
}
