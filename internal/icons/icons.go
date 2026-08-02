// Package icons builds and caches authentic in-game item icons on demand. Given
// an item's pak-relative texture path, it extracts the texture from the user's
// own paks (pure Go: pak → oodle → texture), encodes a PNG, and caches it in the
// OS cache dir. Nothing is bundled or committed — the art stays on the machine
// that owns the game.
package icons

import (
	"bytes"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/gfriloux/dragonsword-save-editor/internal/oodle"
	"github.com/gfriloux/dragonsword-save-editor/internal/pak"
	"github.com/gfriloux/dragonsword-save-editor/internal/texture"
)

// Service extracts and caches icons for one game folder.
type Service struct {
	paksDir  string
	cacheDir string

	mu   sync.Mutex
	pv   *pak.Provider
	dec  *oodle.Decoder
	init bool
	err  error
}

// New returns a service for the given game folder. It does not touch the paks
// until the first PNG request.
func New(gameDir string) *Service {
	cache, _ := os.UserCacheDir()
	return &Service{
		paksDir:  filepath.Join(gameDir, "DS", "Content", "Paks"),
		cacheDir: filepath.Join(cache, "dsa-save-editor", "icons"),
	}
}

// GameDir folder used, for detecting a configuration change.
func (s *Service) PaksDir() string { return s.paksDir }

func (s *Service) ensure() error {
	if s.init {
		return s.err
	}
	s.init = true
	if err := os.MkdirAll(s.cacheDir, 0o755); err != nil {
		s.err = err
		return err
	}
	pv, err := pak.OpenDir(s.paksDir)
	if err != nil {
		s.err = fmt.Errorf("open paks: %w", err)
		return s.err
	}
	dec, err := oodle.New()
	if err != nil {
		s.err = fmt.Errorf("oodle: %w", err)
		return s.err
	}
	s.pv, s.dec = pv, dec
	return nil
}

// PNG returns the PNG bytes of the icon at iconPath (pak-relative, no
// extension), building and caching it on first request. It is safe for
// concurrent use.
func (s *Service) PNG(cid int64, iconPath string) ([]byte, error) {
	if iconPath == "" {
		return nil, fmt.Errorf("icons: no icon path for CID %d", cid)
	}
	cacheFile := filepath.Join(s.cacheDir, strconv.FormatInt(cid, 10)+".png")
	if b, err := os.ReadFile(cacheFile); err == nil {
		return b, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	// Re-check under lock in case a sibling built it while we waited.
	if b, err := os.ReadFile(cacheFile); err == nil {
		return b, nil
	}
	if err := s.ensure(); err != nil {
		return nil, err
	}

	e := s.pv.Find(iconPath + ".uexp")
	if e == nil {
		return nil, fmt.Errorf("icons: %s not found in paks", iconPath)
	}
	raw, err := e.Read(s.dec)
	if err != nil {
		return nil, fmt.Errorf("icons: read %s: %w", iconPath, err)
	}
	img, err := texture.DecodeUExp(raw)
	if err != nil {
		return nil, fmt.Errorf("icons: decode %s: %w", iconPath, err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	writeAtomic(cacheFile, buf.Bytes())
	return buf.Bytes(), nil
}

// Close releases the pak/oodle resources.
func (s *Service) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dec != nil {
		s.dec.Close()
		s.dec = nil
	}
}

func writeAtomic(path string, b []byte) {
	tmp := path + ".tmp"
	if os.WriteFile(tmp, b, 0o644) == nil {
		os.Rename(tmp, path)
	}
}
