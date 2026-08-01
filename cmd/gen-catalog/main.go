// Command gen-catalog builds internal/domain/data/items.json by scraping the
// community datamine at th.gl (The Hidden Gaming Lair), which exposes DragonSword
// Awakening's game data keyed by the same content ids (CIDs) the save uses.
//
// It fetches each database list page in French and English, extracts the
// CID→name rows, maps th.gl's categories to ours, and writes a bilingual catalog.
// Run it manually to (re)generate the bundled catalog:
//
//	just gen-catalog        # or: go run ./cmd/gen-catalog
//
// Names are © their respective owners; the mapping is courtesy of th.gl.
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
	"sort"
	"strconv"
	"time"
)

const (
	base       = "https://dragonswordawakening.th.gl"
	userAgent  = "dsa-save-editor gen-catalog (+https://github.com/gfriloux/dragonsword-save-editor)"
	outPath    = "internal/domain/data/items.json"
	sourceNote = "Item and character names datamined by The Hidden Gaming Lair (th.gl). Regenerate with `just gen-catalog`."
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

// Matches a table row link: .../db_<cat>_<cid>"> [<img>] <span ...>Name</span>.
var rowRE = regexp.MustCompile(`db_[a-z_]+?_(\d+)"[^>]*>(?:<img[^>]*>)?<span[^>]*>([^<]+)</span>`)

type item struct {
	FR       string `json:"fr,omitempty"`
	EN       string `json:"en,omitempty"`
	Category string `json:"category"`
}

type catalog struct {
	Source string          `json:"_source"`
	Items  map[string]item `json:"items"`
}

func fetchNames(lang, route string) (map[int64]string, error) {
	url := fmt.Sprintf("%s/%s/db/%s", base, lang, route)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", userAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: HTTP %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	out := map[int64]string{}
	for _, m := range rowRE.FindAllSubmatch(body, -1) {
		cid, err := strconv.ParseInt(string(m[1]), 10, 64)
		if err != nil {
			continue
		}
		out[cid] = html.UnescapeString(string(m[2]))
	}
	return out, nil
}

func main() {
	log.SetFlags(0)
	items := map[int64]item{}

	for route, category := range routes {
		fr, err := fetchNames("fr", route)
		if err != nil {
			log.Fatalf("fetch fr/%s: %v", route, err)
		}
		time.Sleep(400 * time.Millisecond)
		en, err := fetchNames("en", route)
		if err != nil {
			log.Fatalf("fetch en/%s: %v", route, err)
		}
		time.Sleep(400 * time.Millisecond)

		cids := map[int64]bool{}
		for c := range fr {
			cids[c] = true
		}
		for c := range en {
			cids[c] = true
		}
		for c := range cids {
			it := items[c]
			it.Category = category
			if v, ok := fr[c]; ok {
				it.FR = v
			}
			if v, ok := en[c]; ok {
				it.EN = v
			}
			items[c] = it
		}
		log.Printf("%-14s fr=%d en=%d", route, len(fr), len(en))
	}

	out := catalog{Source: sourceNote, Items: make(map[string]item, len(items))}
	for c, it := range items {
		out.Items[strconv.FormatInt(c, 10)] = it
	}
	buf, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(outPath, append(buf, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
	// Report a stable, sorted summary.
	keys := make([]int64, 0, len(items))
	for c := range items {
		keys = append(keys, c)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	log.Printf("wrote %s: %d items", outPath, len(items))
}
