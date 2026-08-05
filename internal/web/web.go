// Package web serves the embedded browser UI and a small JSON API on top of an
// open save.
package web

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"net/http"
	"sync"

	"github.com/gfriloux/dragonsword-save-editor/internal/domain"
	"github.com/gfriloux/dragonsword-save-editor/internal/icons"
	"github.com/gfriloux/dragonsword-save-editor/internal/save"
)

//go:embed static
var staticFS embed.FS

// errNoSave is returned by save-scoped endpoints when no save is open yet.
var errNoSave = errors.New("no save open — pick one first")

// Server serves the UI and JSON API. It starts without a save: the game folder
// is configured and a slot is picked through the API, at which point a session
// (an open Save + its domain Game) is created. The session can be replaced by
// opening another slot.
type Server struct {
	cat *domain.Catalog
	mux *http.ServeMux

	mu      sync.Mutex
	sv      *save.Save     // nil until a save is opened
	g       *domain.Game   // nil until a save is opened
	iconSvc *icons.Service // nil until the first icon request
}

// New returns an http.Handler serving the UI and API. cat is the (save-agnostic)
// item catalog, loaded once for the whole process.
func New(cat *domain.Catalog) *Server {
	s := &Server{cat: cat, mux: http.NewServeMux()}
	sub, _ := fs.Sub(staticFS, "static")
	s.mux.Handle("/", http.FileServer(http.FS(sub)))

	// Session-agnostic: configuration and slot selection.
	s.mux.HandleFunc("/api/config", s.handleConfig)
	s.mux.HandleFunc("/api/config/game-dir", s.handleSetGameDir)
	s.mux.HandleFunc("/api/saves", s.handleSaves)
	s.mux.HandleFunc("/api/screenshot", s.handleScreenshot)
	s.mux.HandleFunc("/api/open", s.handleOpen)
	s.mux.HandleFunc("/api/icon", s.handleIcon)

	// Generic database view (needs an open save).
	s.mux.HandleFunc("/api/info", s.needSave(s.handleInfo))
	s.mux.HandleFunc("/api/table", s.needSave(s.handleTable))
	s.mux.HandleFunc("/api/update", s.needSave(s.handleUpdate))
	s.mux.HandleFunc("/api/save", s.needSave(s.handleSave))
	// Game-oriented editor view (needs an open save).
	s.mux.HandleFunc("/api/game/currency", s.needSave(s.handleCurrency))
	s.mux.HandleFunc("/api/game/consumables", s.needSave(s.handleConsumables))
	s.mux.HandleFunc("/api/game/consumable-categories", s.needSave(s.handleConsumableCategories))
	s.mux.HandleFunc("/api/game/stack", s.needSave(s.handleStack))
	s.mux.HandleFunc("/api/game/label", s.needSave(s.handleLabel))
	s.mux.HandleFunc("/api/game/characters", s.needSave(s.handleCharacters))
	s.mux.HandleFunc("/api/game/teams", s.needSave(s.handleTeams))
	s.mux.HandleFunc("/api/game/equipment", s.needSave(s.handleEquipment))
	s.mux.HandleFunc("/api/game/gems", s.needSave(s.handleGems))
	s.mux.HandleFunc("/api/game/gem", s.needSave(s.handleGem))
	s.mux.HandleFunc("/api/game/catalog", s.needSave(s.handleCatalog))
	s.mux.HandleFunc("/api/game/stackable", s.needSave(s.handleAddStackable))
	s.mux.HandleFunc("/api/game/stackable/fill", s.needSave(s.handleFillStackables))
	s.mux.HandleFunc("/api/game/recipes", s.needSave(s.handleRecipes))
	s.mux.HandleFunc("/api/game/recipes/known", s.needSave(s.handleSetRecipeKnown))
	s.mux.HandleFunc("/api/game/recipes/unlock-all", s.needSave(s.handleUnlockRecipes))
	s.mux.HandleFunc("/api/game/costumes", s.needSave(s.handleCostumes))
	s.mux.HandleFunc("/api/game/costumes/unlock", s.needSave(s.handleCostumeUnlock))
	s.mux.HandleFunc("/api/game/costumes/equip", s.needSave(s.handleCostumeEquip))
	s.mux.HandleFunc("/api/game/familiers", s.needSave(s.handleFamiliers))
	s.mux.HandleFunc("/api/game/familiers/unlock", s.needSave(s.handleFamilierUnlock))
	s.mux.HandleFunc("/api/game/familiers/equip", s.needSave(s.handleFamilierEquip))
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

// needSave wraps a handler so it 409s when no save is open yet.
func (s *Server) needSave(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		ok := s.g != nil
		s.mu.Unlock()
		if !ok {
			writeErr(w, http.StatusConflict, errNoSave)
			return
		}
		h(w, r)
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]string{"error": err.Error()})
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	tables, err := s.sv.Tables()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{
		"path":     s.sv.Path(),
		"tables":   tables,
		"iconSize": s.g.Catalog().IconSize(),
	})
}

func (s *Server) handleTable(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	limit := atoiDefault(r.URL.Query().Get("limit"), 200)
	offset := atoiDefault(r.URL.Query().Get("offset"), 0)

	cols, err := s.sv.Columns(name)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	total, err := s.sv.RowCount(name)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	_, rows, err := s.sv.Rows(name, limit, offset)
	if err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{
		"columns": cols,
		"rows":    rows,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

type updateReq struct {
	Table  string         `json:"table"`
	PK     map[string]any `json:"pk"`
	Column string         `json:"column"`
	Value  any            `json:"value"`
}

func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, fmt.Errorf("POST only"))
		return
	}
	var req updateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	req.PK = normalizeMap(req.PK)
	if err := s.sv.UpdateCell(req.Table, req.PK, req.Column, normalizeVal(req.Value)); err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (s *Server) handleSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, fmt.Errorf("POST only"))
		return
	}
	if err := s.sv.Save(); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "path": s.sv.Path()})
}

// normalizeVal turns JSON numbers that are whole into int64 so integer columns
// don't get stored as "1.0".
func normalizeVal(v any) any {
	if f, ok := v.(float64); ok && f == math.Trunc(f) && !math.IsInf(f, 0) {
		return int64(f)
	}
	return v
}

func normalizeMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = normalizeVal(v)
	}
	return out
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return def
	}
	return n
}
