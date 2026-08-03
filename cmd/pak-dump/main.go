// Command pak-dump extracts named data files from the game's custom paks in pure Go,
// writing each to an output directory. It is the offline datamine step feeding
// cmd/pak-catalog (like gen-catalog): the game ships its whole server dataset as plain
// XML inside the paks (…/Server/XML/GameData/*.xml), so no .usmap is needed.
//
// It reuses the editor's own pure-Go pak reader (internal/pak) + Oodle decoder
// (internal/oodle) — no CGO, no external tooling. Extracted XML is game-copyright and
// lives under tmp/ (gitignored), never committed.
//
//	go run ./cmd/pak-dump -game "<gameDir>" GameItemData.xml StringData.xml CookRecipeData.xml \
//	    CookToolData.xml GameItemCategoryData.xml GameItemTypeDefineData.xml ContentsBuffData.xml
//
// Matching is by file base name (case-insensitive); every pak entry whose base name is in
// the requested set is written to -out (default tmp/pak).
package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gfriloux/dragonsword-save-editor/internal/oodle"
	"github.com/gfriloux/dragonsword-save-editor/internal/pak"
)

func main() {
	log.SetFlags(0)
	game := flag.String("game", "", "game folder (the one containing DS/Content/Paks)")
	out := flag.String("out", "tmp/pak", "output directory")
	flag.Parse()
	if *game == "" || flag.NArg() == 0 {
		log.Fatal("usage: pak-dump -game <dir> Name1.xml Name2.xml …")
	}

	pv, err := pak.OpenDir(filepath.Join(*game, "DS", "Content", "Paks"))
	if err != nil {
		log.Fatalf("open paks: %v", err)
	}
	dec, err := oodle.New()
	if err != nil {
		log.Fatalf("oodle: %v", err)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatalf("mkdir: %v", err)
	}

	want := map[string]bool{}
	for _, n := range flag.Args() {
		want[strings.ToLower(n)] = true
	}
	found := map[string]bool{}
	for _, f := range pv.Files() {
		base := strings.ToLower(filepath.Base(f))
		if !want[base] || found[base] {
			continue
		}
		e := pv.Find(f)
		if e == nil {
			continue
		}
		data, err := e.Read(dec)
		if err != nil {
			log.Printf("read %s: %v", f, err)
			continue
		}
		dst := filepath.Join(*out, filepath.Base(f))
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			log.Fatalf("write %s: %v", dst, err)
		}
		log.Printf("wrote %s (%d bytes)", dst, len(data))
		found[base] = true
	}
	for n := range want {
		if !found[n] {
			log.Printf("NOT FOUND in paks: %s", n)
		}
	}
}
