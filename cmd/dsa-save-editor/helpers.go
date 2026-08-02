package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"

	"github.com/gfriloux/dragonsword-save-editor/internal/config"
	"github.com/gfriloux/dragonsword-save-editor/internal/save"
)

// buildVersion is overridden at build time with -ldflags "-X main.buildVersion=...".
var buildVersion = "dev"

func waitForSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

// autoDetectSave returns the most recently modified save. It prefers the
// remembered game folder's save tree, then falls back to DSA_SAVE_DIR and the
// current directory. Returns "" when nothing is found.
func autoDetectSave() string {
	if cfg, err := config.Load(); err == nil && cfg.GameDir != "" {
		if slots, err := save.Discover(cfg.GameDir); err == nil && len(slots) > 0 {
			return slots[0].Path // Discover sorts most-recent first
		}
	}

	var dirs []string
	if d := os.Getenv("DSA_SAVE_DIR"); d != "" {
		dirs = append(dirs, d)
	}
	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, cwd)
	}

	type cand struct {
		path string
		mod  int64
	}
	var found []cand
	for _, d := range dirs {
		matches, _ := filepath.Glob(filepath.Join(d, "*_Slot*.db"))
		for _, m := range matches {
			if fi, err := os.Stat(m); err == nil {
				found = append(found, cand{m, fi.ModTime().UnixNano()})
			}
		}
	}
	if len(found) == 0 {
		return ""
	}
	sort.Slice(found, func(i, j int) bool { return found[i].mod > found[j].mod })
	return found[0].path
}
