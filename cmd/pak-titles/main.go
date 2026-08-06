// Command pak-titles builds the bundled title catalog internal/domain/data/titles.json
// from the game's own data — extracted from the encrypted .pak archives as raw XML.
//
// The game ships its full server dataset as plain XML inside the paks. Two files matter:
//
//	AccountTitleData.xml  one row per title: ID, Name (= string key), FontColor, and up to
//	                      three stat bonuses (StatType1..3 / StatValue1..3).
//	StringData.xml        one row per string key: all languages, incl. Fr and En.
//
// See .claude/plans/release/pak-extraction/ for how the XML is extracted (a small
// CUE4Parse step, offline like gen-catalog / pak-catalog). This tool is pure Go and only
// reads the already-extracted XML fixtures (game-copyright — they live under tmp/, never
// committed).
//
// Whether a title is *unlocked* is a flag in tb_title's bitmask: category = ID/64,
// bit = ID%64 (the same shape as tb_switch recipes). titles.json stays purely structural;
// (category, bit) are derived at runtime in internal/domain.
//
//	go run ./cmd/pak-titles   # defaults to tmp/pak/*.xml + internal/domain/data/titles.json
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
)

const sourceNote = "Titles datamined from the game's own AccountTitleData/StringData (paks). id/64 = tb_title CATEGORY, id%64 = bit. Regenerate with `go run ./cmd/pak-titles`."

type strRow struct {
	ID string `xml:"ID,attr"`
	FR string `xml:"Fr,attr"`
	EN string `xml:"En,attr"`
}

type titleRow struct {
	ID        string `xml:"ID,attr"`
	Name      string `xml:"Name,attr"` // StringData key for the localized title
	FontColor string `xml:"FontColor,attr"`
	StatType1 string `xml:"StatType1,attr"`
	StatVal1  string `xml:"StatValue1,attr"`
	StatType2 string `xml:"StatType2,attr"`
	StatVal2  string `xml:"StatValue2,attr"`
	StatType3 string `xml:"StatType3,attr"`
	StatVal3  string `xml:"StatValue3,attr"`
}

// titleStat is one stat bonus a title grants (MaxHP / Attack / Defence, etc.).
type titleStat struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

type titleOut struct {
	ID     int         `json:"id"`
	NameFR string      `json:"nameFr,omitempty"`
	NameEN string      `json:"nameEn,omitempty"`
	Color  string      `json:"color,omitempty"`
	Stats  []titleStat `json:"stats,omitempty"`
}

type titlesFile struct {
	Source string     `json:"_source"`
	Titles []titleOut `json:"titles"`
}

// stats collects the up to three (Type, Value) pairs into a slice, skipping empty slots.
func (r titleRow) stats() []titleStat {
	var out []titleStat
	for _, p := range [][2]string{
		{r.StatType1, r.StatVal1}, {r.StatType2, r.StatVal2}, {r.StatType3, r.StatVal3},
	} {
		if p[0] == "" {
			continue
		}
		v, err := strconv.Atoi(p[1])
		if err != nil {
			continue
		}
		out = append(out, titleStat{Type: p[0], Value: v})
	}
	return out
}

// streamRows decodes every element whose local name is `local` from an XML file into T,
// invoking fn for each. Namespace prefixes (ns3:) are ignored — attributes match by local
// name.
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
	titlesXML := flag.String("titles", "tmp/pak/AccountTitleData.xml", "AccountTitleData.xml extracted from the paks")
	stringsXML := flag.String("strings", "tmp/pak/StringData.xml", "StringData.xml extracted from the paks")
	out := flag.String("out", "internal/domain/data/titles.json", "titles.json to write")
	flag.Parse()

	strs := map[string]strRow{}
	if err := streamRows(*stringsXML, "StringData", func(r strRow) { strs[r.ID] = r }); err != nil {
		log.Fatalf("parse strings: %v", err)
	}
	log.Printf("strings: %d", len(strs))

	var titles []titleOut
	var noName int
	if err := streamRows(*titlesXML, "AccountTitleData", func(r titleRow) {
		id, err := strconv.Atoi(r.ID)
		if err != nil {
			log.Fatalf("title with non-numeric ID %q", r.ID)
		}
		s := strs[r.Name]
		if s.FR == "" && s.EN == "" {
			noName++
		}
		titles = append(titles, titleOut{
			ID: id, NameFR: s.FR, NameEN: s.EN, Color: r.FontColor, Stats: r.stats(),
		})
	}); err != nil {
		log.Fatalf("parse titles: %v", err)
	}
	sort.Slice(titles, func(i, j int) bool { return titles[i].ID < titles[j].ID })

	writeJSON(*out, titlesFile{Source: sourceNote, Titles: titles})
	log.Printf("wrote %s: %d titles (%d without a resolvable name)", *out, len(titles), noName)
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
