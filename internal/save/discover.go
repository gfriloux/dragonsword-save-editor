package save

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"
)

// Slot describes a save file discovered on disk, without decrypting it. It is
// pure filesystem metadata so listing saves is cheap and never touches the
// SQLCipher layer (opening/decrypting happens only when a slot is chosen).
type Slot struct {
	Path       string    `json:"path"`       // absolute path to <id>_Slot<N>.db
	AccountID  string    `json:"accountId"`  // the numeric SaveGames subfolder
	Slot       int       `json:"slot"`       // N in _Slot<N>
	Screenshot string    `json:"screenshot"` // sibling ScreenShot_<N>.png, or ""
	ModTime    time.Time `json:"modTime"`    // last modified (≈ last played)
}

var slotRe = regexp.MustCompile(`_Slot(\d+)\.db$`)

// SaveGamesDir returns the game's save folder for a given game installation
// folder.
func SaveGamesDir(gameDir string) string {
	return filepath.Join(gameDir, "DS", "Saved", "SaveGames")
}

// Discover lists the save slots found under a game folder, most-recent first.
// A game folder that has no SaveGames tree yields an empty list, not an error.
func Discover(gameDir string) ([]Slot, error) {
	matches, err := filepath.Glob(filepath.Join(SaveGamesDir(gameDir), "*", "*_Slot*.db"))
	if err != nil {
		return nil, err
	}
	slots := make([]Slot, 0, len(matches))
	for _, m := range matches {
		fi, err := os.Stat(m)
		if err != nil {
			continue
		}
		n := 0
		if g := slotRe.FindStringSubmatch(filepath.Base(m)); g != nil {
			n, _ = strconv.Atoi(g[1])
		}
		shot := filepath.Join(filepath.Dir(m), "ScreenShot_"+strconv.Itoa(n)+".png")
		if _, err := os.Stat(shot); err != nil {
			shot = ""
		}
		slots = append(slots, Slot{
			Path:       m,
			AccountID:  filepath.Base(filepath.Dir(m)),
			Slot:       n,
			Screenshot: shot,
			ModTime:    fi.ModTime(),
		})
	}
	sort.Slice(slots, func(i, j int) bool { return slots[i].ModTime.After(slots[j].ModTime) })
	return slots, nil
}
