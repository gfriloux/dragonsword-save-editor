package domain

import (
	"encoding/json"
	"fmt"
)

// Titles are datamined into data/titles.json (see cmd/pak-titles). Whether a title is
// *unlocked* is a flag in the tb_title bitmask table, keyed by the account:
//
//	category = id / 64      bit = id % 64
//
// the same shape as the cooking-recipe switches (tb_switch). titles.json is purely
// structural — id, bilingual name, font colour and stat bonuses; (category, bit) derive
// here. tb_title also carries FAV_BIT_FIELD (the displayed/favourite title); we never
// touch it (see the writes in SetTitleUnlocked / UnlockAllTitles).

// TitleStat is one stat bonus a title grants (MaxHP / Attack / Defence).
type TitleStat struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

// titleSeed is one raw title as stored in titles.json.
type titleSeed struct {
	ID     int64       `json:"id"`
	NameFR string      `json:"nameFr"`
	NameEN string      `json:"nameEn"`
	Color  string      `json:"color"`
	Stats  []TitleStat `json:"stats"`
}

var titleSeeds []titleSeed

func init() {
	raw, err := seedFS.ReadFile("data/titles.json")
	if err != nil {
		panic(fmt.Sprintf("domain: reading titles.json: %v", err))
	}
	var f struct {
		Titles []titleSeed `json:"titles"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		panic(fmt.Sprintf("domain: parsing titles.json: %v", err))
	}
	titleSeeds = f.Titles
}

// titlePos maps a title id to its (category, bit) in tb_title.
func titlePos(id int64) (category, bit int) { return int(id / 64), int(id % 64) }

// Title is a game title resolved against a save.
type Title struct {
	ID       int64       `json:"id"`
	Category int         `json:"category"`
	Bit      int         `json:"bit"`
	NameFR   string      `json:"nameFr"`
	NameEN   string      `json:"nameEn"`
	Color    string      `json:"color,omitempty"`
	Stats    []TitleStat `json:"stats,omitempty"`
	Unlocked bool        `json:"unlocked"`
}

// Titles lists every game title resolved against the save: unlocked state read from
// tb_title, plus the localized name, font colour and stat bonuses from the catalog.
func (g *Game) Titles() ([]Title, error) {
	uid, err := g.UserID()
	if err != nil {
		return nil, err
	}
	unlocked, err := g.titleUnlockedBits(uid)
	if err != nil {
		return nil, err
	}
	out := make([]Title, 0, len(titleSeeds))
	for _, ts := range titleSeeds {
		cat, bit := titlePos(ts.ID)
		out = append(out, Title{
			ID:       ts.ID,
			Category: cat,
			Bit:      bit,
			NameFR:   ts.NameFR,
			NameEN:   ts.NameEN,
			Color:    ts.Color,
			Stats:    ts.Stats,
			Unlocked: unlocked[cat]>>uint(bit)&1 == 1,
		})
	}
	return out, nil
}

// titleUnlockedBits reads the tb_title categories into a cat→bitfield map.
func (g *Game) titleUnlockedBits(uid int64) (map[int]uint64, error) {
	rows, err := g.s.DB().Query(`SELECT CATEGORY, BIT_FIELD FROM tb_title WHERE USER_DBID=?`, uid)
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
