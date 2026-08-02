package web

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gfriloux/dragonsword-save-editor/internal/config"
	"github.com/gfriloux/dragonsword-save-editor/internal/icons"
)

// iconService lazily builds (or rebuilds, if the game folder changed) the icon
// extraction service for the configured game folder. Returns nil if no folder
// is set.
func (s *Server) iconService() *icons.Service {
	cfg, _ := config.Load()
	if cfg.GameDir == "" {
		return nil
	}
	paks := filepath.Join(cfg.GameDir, "DS", "Content", "Paks")
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.iconSvc == nil || s.iconSvc.PaksDir() != paks {
		if s.iconSvc != nil {
			s.iconSvc.Close()
		}
		s.iconSvc = icons.New(cfg.GameDir)
	}
	return s.iconSvc
}

// handleIcon serves an item's authentic icon PNG, extracted from the user's
// paks and cached. A 404 (no game folder, unknown CID, no icon, extraction
// failure) lets the UI fall back to the th.gl sprite.
func (s *Server) handleIcon(w http.ResponseWriter, r *http.Request) {
	cid, err := strconv.ParseInt(r.URL.Query().Get("cid"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	iconPath := s.cat.Lookup(cid).IconPath
	if iconPath == "" {
		http.NotFound(w, r)
		return
	}
	svc := s.iconService()
	if svc == nil {
		http.NotFound(w, r)
		return
	}
	png, err := svc.PNG(cid, iconPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "max-age=86400")
	w.Write(png)
}
