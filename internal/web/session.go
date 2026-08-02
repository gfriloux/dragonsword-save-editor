package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gfriloux/dragonsword-save-editor/internal/config"
	"github.com/gfriloux/dragonsword-save-editor/internal/domain"
	"github.com/gfriloux/dragonsword-save-editor/internal/save"
)

// Open decrypts the save at path and makes it the current session, replacing and
// closing any previously open save. Exposed so the CLI can honour a save path
// given as a power-user override.
func (s *Server) Open(path string) error {
	sv, err := save.Open(path, "")
	if err != nil {
		return err
	}
	g := domain.New(sv, s.cat)
	s.mu.Lock()
	old := s.sv
	s.sv, s.g = sv, g
	s.mu.Unlock()
	if old != nil {
		_ = old.Close()
	}
	return nil
}

// Close closes the current session's save, if any.
func (s *Server) Close() error {
	s.mu.Lock()
	sv := s.sv
	s.sv, s.g = nil, nil
	if s.iconSvc != nil {
		s.iconSvc.Close()
		s.iconSvc = nil
	}
	s.mu.Unlock()
	if sv != nil {
		return sv.Close()
	}
	return nil
}

// looksLikeGameDir reports whether dir contains the DragonSword layout we need
// (either the paks or the save tree).
func looksLikeGameDir(dir string) bool {
	for _, sub := range []string{
		filepath.Join(dir, "DS", "Content", "Paks"),
		filepath.Join(dir, "DS", "Saved", "SaveGames"),
	} {
		if fi, err := os.Stat(sub); err == nil && fi.IsDir() {
			return true
		}
	}
	return false
}

// handleConfig reports the current configuration and session state.
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	cfg, _ := config.Load()
	s.mu.Lock()
	sv := s.sv
	s.mu.Unlock()
	resp := map[string]any{
		"gameDir":  cfg.GameDir,
		"saveOpen": sv != nil,
	}
	if sv != nil {
		resp["savePath"] = sv.Path()
	}
	writeJSON(w, http.StatusOK, resp)
}

type gameDirReq struct {
	Dir string `json:"dir"`
}

// handleSetGameDir validates and persists the game folder.
func (s *Server) handleSetGameDir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, fmt.Errorf("POST only"))
		return
	}
	var req gameDirReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	dir := strings.TrimSpace(req.Dir)
	if dir == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("empty path"))
		return
	}
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("not a folder: %s", dir))
		return
	}
	if !looksLikeGameDir(dir) {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("this folder has no DS/Content/Paks or DS/Saved/SaveGames"))
		return
	}
	if err := (config.Config{GameDir: dir}).Store(); err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"gameDir": dir})
}

// handleSaves lists the slots discovered under the configured game folder.
func (s *Server) handleSaves(w http.ResponseWriter, r *http.Request) {
	cfg, _ := config.Load()
	if cfg.GameDir == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("no game folder configured"))
		return
	}
	slots, err := save.Discover(cfg.GameDir)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"gameDir": cfg.GameDir, "slots": slots})
}

// handleScreenshot serves a slot screenshot, constrained to ScreenShot_*.png
// files under the configured save tree (no arbitrary file read).
func (s *Server) handleScreenshot(w http.ResponseWriter, r *http.Request) {
	cfg, _ := config.Load()
	if cfg.GameDir == "" {
		http.NotFound(w, r)
		return
	}
	p := r.URL.Query().Get("path")
	abs, err := filepath.Abs(p)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	root := filepath.Clean(save.SaveGamesDir(cfg.GameDir)) + string(os.PathSeparator)
	base := filepath.Base(abs)
	if !strings.HasPrefix(abs, root) ||
		!strings.HasPrefix(base, "ScreenShot_") || !strings.HasSuffix(base, ".png") {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, abs)
}

type openReq struct {
	Path string `json:"path"`
}

// handleOpen opens a chosen slot, constrained to *_Slot*.db files under the
// configured save tree.
func (s *Server) handleOpen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, fmt.Errorf("POST only"))
		return
	}
	var req openReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	cfg, _ := config.Load()
	if cfg.GameDir == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("no game folder configured"))
		return
	}
	abs, err := filepath.Abs(req.Path)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	root := filepath.Clean(save.SaveGamesDir(cfg.GameDir)) + string(os.PathSeparator)
	if !strings.HasPrefix(abs, root) || !strings.HasSuffix(filepath.Base(abs), ".db") {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("not a save under this game folder"))
		return
	}
	if err := s.Open(abs); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "path": abs})
}
