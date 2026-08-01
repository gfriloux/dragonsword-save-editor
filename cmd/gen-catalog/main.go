// Command gen-catalog builds the bundled item catalog by scraping the community
// datamine at th.gl (The Hidden Gaming Lair), which exposes DragonSword Awakening's
// game data keyed by the same content ids (CIDs) the save uses.
//
// It fetches each database list page in French and English, extracts each item's
// CID, bilingual name and icon position, downloads the current icon sprite sheet,
// and writes:
//
//	internal/domain/data/items.json   names + categories + icon positions
//	internal/web/static/sprite.webp   the icon sprite sheet
//
// Run it manually to (re)generate both (the sprite hash rotates, so they must be
// captured together):
//
//	just gen-catalog        # or: go run ./cmd/gen-catalog
//
// Names and icons are © their respective owners; the mapping is courtesy of th.gl.
package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

const (
	base       = "https://dragonswordawakening.th.gl"
	userAgent  = "dsa-save-editor gen-catalog (+https://github.com/gfriloux/dragonsword-save-editor)"
	itemsPath  = "internal/domain/data/items.json"
	spritePath = "internal/web/static/sprite.webp"
	iconSize   = 64 // th.gl sprite cell size, in pixels
	sourceNote = "Item and character names + icons datamined by The Hidden Gaming Lair (th.gl). Regenerate with `just gen-catalog`."
)

// th.gl DB route -> our category.
var routes = map[string]string{
	"characters":     "character",
	"equipment":      "gear",
	"recipes":        "food",
	"materials_db":   "material",
	"ingredients_db": "material",
	"costumes":       "costume",
	"mounts":         "mount",
}

// A table row: .../db_<cat>_<cid>"> [<img ... object-position:<x>px <y>px ...>] <span>Name</span>.
// rowRE captures rows that carry an icon (with its sprite position); nameRE is a
// fallback that captures the name for rows without a parseable icon.
var rowRE = regexp.MustCompile(`db_[a-z_]+?_(\d+)"[^>]*><img[^>]*object-position:(-?\d+)px (-?\d+)px[^>]*>[^<]*<span[^>]*>([^<]+)</span>`)
var nameRE = regexp.MustCompile(`db_[a-z_]+?_(\d+)"[^>]*>(?:<img[^>]*>)?<span[^>]*>([^<]+)</span>`)
var spriteRE = regexp.MustCompile(`https://cdn\.th\.gl/[^"]+icons[^"]*\.webp`)

type row struct {
	name string
	x, y int
}

func fetch(url string) ([]byte, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", base+"/")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: HTTP %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func parseRows(body []byte) map[int64]row {
	out := map[int64]row{}
	for _, m := range rowRE.FindAllSubmatch(body, -1) {
		cid, err := strconv.ParseInt(string(m[1]), 10, 64)
		if err != nil {
			continue
		}
		x, _ := strconv.Atoi(string(m[2]))
		y, _ := strconv.Atoi(string(m[3]))
		out[cid] = row{name: html.UnescapeString(string(m[4])), x: x, y: y}
	}
	// Fallback: keep names for rows whose icon didn't parse (x=y=0 → no icon).
	for _, m := range nameRE.FindAllSubmatch(body, -1) {
		cid, err := strconv.ParseInt(string(m[1]), 10, 64)
		if err != nil {
			continue
		}
		if _, ok := out[cid]; !ok {
			out[cid] = row{name: html.UnescapeString(string(m[2]))}
		}
	}
	return out
}

type item struct {
	FR       string `json:"fr,omitempty"`
	EN       string `json:"en,omitempty"`
	Category string `json:"category"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
}

type catalog struct {
	Source   string          `json:"_source"`
	IconSize int             `json:"iconSize"`
	Items    map[string]item `json:"items"`
}

func main() {
	log.SetFlags(0)
	items := map[int64]item{}
	spriteURL := ""

	for route, category := range routes {
		frBody, err := fetch(fmt.Sprintf("%s/fr/db/%s", base, route))
		if err != nil {
			log.Fatalf("fetch fr/%s: %v", route, err)
		}
		if spriteURL == "" {
			if m := spriteRE.Find(frBody); m != nil {
				spriteURL = string(m)
			}
		}
		fr := parseRows(frBody)
		time.Sleep(400 * time.Millisecond)
		enBody, err := fetch(fmt.Sprintf("%s/en/db/%s", base, route))
		if err != nil {
			log.Fatalf("fetch en/%s: %v", route, err)
		}
		en := parseRows(enBody)
		time.Sleep(400 * time.Millisecond)

		for cid, r := range fr {
			it := items[cid]
			it.Category, it.FR, it.X, it.Y = category, r.name, r.x, r.y
			items[cid] = it
		}
		for cid, r := range en {
			it := items[cid]
			if it.Category == "" {
				it.Category, it.X, it.Y = category, r.x, r.y
			}
			it.EN = r.name
			items[cid] = it
		}
		log.Printf("%-14s fr=%d en=%d", route, len(fr), len(en))
	}

	if spriteURL == "" {
		log.Fatal("could not find the sprite sheet URL")
	}
	sprite, err := fetch(spriteURL)
	if err != nil {
		log.Fatalf("download sprite: %v", err)
	}
	if err := os.WriteFile(spritePath, sprite, 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s: %d bytes (%s)", spritePath, len(sprite), spriteURL)

	out := catalog{Source: sourceNote, IconSize: iconSize, Items: make(map[string]item, len(items))}
	for c, it := range items {
		out.Items[strconv.FormatInt(c, 10)] = it
	}
	buf, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(itemsPath, append(buf, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s: %d items", itemsPath, len(items))
}
