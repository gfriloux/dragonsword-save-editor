// Package domain provides a game-oriented, typed view over a raw save: an item
// catalog (names/categories) and accessors for currencies, consumables, etc. It
// sits above internal/save and is consumed by internal/web's editor view.
package domain

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

//go:embed data/items.json
var seedFS embed.FS

// Item is a resolved catalog entry for a content id (CID).
type Item struct {
	CID      int64  `json:"cid"`
	Name     string `json:"name"`     // human name, or a generated fallback
	Category string `json:"category"` // currency | potion | food | material | misc
	Known    bool   `json:"known"`    // true if the name comes from the seed or a user label
}

type seedEntry struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

type seedFile struct {
	Items map[string]seedEntry `json:"items"`
}

// Catalog resolves CIDs to names and categories, combining an embedded seed,
// user-provided labels, and category inference from the CID structure.
type Catalog struct {
	seed      map[string]seedEntry
	overrides map[string]string // cid -> user label
	ovrPath   string
}

// LoadCatalog reads the embedded seed and, if present, the user overrides file at
// overridesPath. A missing overrides file is not an error; pass "" to disable
// persistence (labels are then kept in memory only).
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
	c := &Catalog{seed: sf.Items, overrides: map[string]string{}, ovrPath: overridesPath}
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

// Lookup resolves a CID. Name precedence: user label > seed name > generated
// fallback ("<Category> <cid>"). Category precedence: seed category > inference.
func (c *Catalog) Lookup(cid int64) Item {
	key := strconv.FormatInt(cid, 10)
	seed, hasSeed := c.seed[key]
	category := seed.Category
	if category == "" {
		category = inferCategory(cid)
	}
	label, hasLabel := c.overrides[key]

	name := label
	if name == "" {
		name = seed.Name
	}
	known := hasLabel || (hasSeed && seed.Name != "")
	if name == "" {
		name = fallbackName(category, cid)
	}
	return Item{CID: cid, Name: name, Category: category, Known: known}
}

// SetLabel records a user-provided name for a CID and persists it if the catalog
// was loaded with an overrides path.
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

// inferCategory maps the leading digits of a CID to a coarse category. Derived
// from the save's item tables (see .claude/plans/v0.2.0/plan.md).
func inferCategory(cid int64) string {
	switch prefix3(cid) {
	case "100":
		return "currency"
	case "141":
		return "potion"
	case "142":
		return "food"
	case "143", "144", "145", "146", "147":
		return "material"
	default:
		return "misc"
	}
}

func fallbackName(category string, cid int64) string {
	title := map[string]string{
		"currency": "Currency",
		"potion":   "Potion",
		"food":     "Food",
		"material": "Material",
		"misc":     "Item",
	}[category]
	if title == "" {
		title = "Item"
	}
	return fmt.Sprintf("%s %d", title, cid)
}

func prefix3(cid int64) string {
	s := strconv.FormatInt(cid, 10)
	if len(s) >= 3 {
		return s[:3]
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
