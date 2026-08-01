package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"
)

// buildVersion is overridden at build time with -ldflags "-X main.buildVersion=...".
var buildVersion = "dev"

func waitForSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

// autoDetectSave looks for a "<id>_Slot<N>.db" file in the current directory and
// in an optional DSA_SAVE_DIR. It returns the most recently modified match, or "".
func autoDetectSave() string {
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
