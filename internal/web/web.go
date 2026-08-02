// Package web serves the embedded browser UI and a small JSON API on top of an
// open save.
package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"net/http"

	"github.com/gfriloux/dragonsword-save-editor/internal/domain"
	"github.com/gfriloux/dragonsword-save-editor/internal/save"
)

//go:embed static
var staticFS embed.FS

// Server wires the API handlers to a Game (domain view) and its underlying Save
// (generic database view).
type Server struct {
	sv  *save.Save
	g   *domain.Game
	mux *http.ServeMux
}

// New returns an http.Handler serving the UI and API for g.
func New(g *domain.Game) *Server {
	s := &Server{sv: g.Save(), g: g, mux: http.NewServeMux()}
	sub, _ := fs.Sub(staticFS, "static")
	s.mux.Handle("/", http.FileServer(http.FS(sub)))
	// Generic database view.
	s.mux.HandleFunc("/api/info", s.handleInfo)
	s.mux.HandleFunc("/api/table", s.handleTable)
	s.mux.HandleFunc("/api/update", s.handleUpdate)
	s.mux.HandleFunc("/api/save", s.handleSave)
	// Game-oriented editor view.
	s.mux.HandleFunc("/api/game/currency", s.handleCurrency)
	s.mux.HandleFunc("/api/game/consumables", s.handleConsumables)
	s.mux.HandleFunc("/api/game/stack", s.handleStack)
	s.mux.HandleFunc("/api/game/label", s.handleLabel)
	s.mux.HandleFunc("/api/game/characters", s.handleCharacters)
	s.mux.HandleFunc("/api/game/teams", s.handleTeams)
	s.mux.HandleFunc("/api/game/equipment", s.handleEquipment)
	s.mux.HandleFunc("/api/game/gems", s.handleGems)
	s.mux.HandleFunc("/api/game/gem", s.handleGem)
	s.mux.HandleFunc("/api/game/catalog", s.handleCatalog)
	s.mux.HandleFunc("/api/game/stackable", s.handleAddStackable)
	s.mux.HandleFunc("/api/game/stackable/fill", s.handleFillStackables)
	s.mux.HandleFunc("/api/game/recipes/unlock-all", s.handleUnlockRecipes)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

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
