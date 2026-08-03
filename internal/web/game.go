package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gfriloux/dragonsword-save-editor/internal/domain"
)

func (s *Server) handleCurrency(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cur, err := s.g.Currencies()
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		writeJSON(w, 200, map[string]any{"currencies": cur})
	case http.MethodPost:
		var req struct {
			CID    int64 `json:"cid"`
			Amount int64 `json:"amount"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, 400, err)
			return
		}
		if err := s.g.SetCurrency(req.CID, req.Amount); err != nil {
			writeErr(w, 400, err)
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true})
	default:
		writeErr(w, 405, fmt.Errorf("GET or POST only"))
	}
}

func (s *Server) handleConsumables(w http.ResponseWriter, r *http.Request) {
	items, err := s.g.Consumables()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"items": items})
}

// handleConsumableCategories returns the ordered curated functional categories used
// to group the Consumables panel. Each item's category key is in its "group" field.
func (s *Server) handleConsumableCategories(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"categories": domain.ConsumableCategories()})
}

func (s *Server) handleStack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, fmt.Errorf("POST only"))
		return
	}
	var req struct {
		Kind  string `json:"kind"`
		ID    int64  `json:"id,string"`
		Count int64  `json:"count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if err := s.g.SetStack(req.Kind, req.ID, req.Count); err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (s *Server) handleLabel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, fmt.Errorf("POST only"))
		return
	}
	var req struct {
		CID  int64  `json:"cid"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if err := s.g.Catalog().SetLabel(req.CID, req.Name); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (s *Server) handleCharacters(w http.ResponseWriter, r *http.Request) {
	chars, err := s.g.Characters()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"characters": chars})
}

func (s *Server) handleTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := s.g.Teams()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"teams": teams})
}

func (s *Server) handleEquipment(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		eq, err := s.g.Equipments()
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		writeJSON(w, 200, map[string]any{"equipment": eq})
	case http.MethodPost:
		var req struct {
			DBID  int64  `json:"dbid,string"`
			Field string `json:"field"`
			Value int64  `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, 400, err)
			return
		}
		var err error
		switch req.Field {
		case "enchant":
			err = s.g.SetEnchant(req.DBID, req.Value)
		case "exp":
			err = s.g.SetEquipExp(req.DBID, req.Value)
		case "lock":
			err = s.g.SetEquipLock(req.DBID, req.Value != 0)
		default:
			err = fmt.Errorf("unknown field %q", req.Field)
		}
		if err != nil {
			writeErr(w, 400, err)
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true})
	default:
		writeErr(w, 405, fmt.Errorf("GET or POST only"))
	}
}

func (s *Server) handleGems(w http.ResponseWriter, r *http.Request) {
	gems, err := s.g.Gems()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"gems": gems})
}

func (s *Server) handleGem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, fmt.Errorf("POST only"))
		return
	}
	var req struct {
		DBID   int64 `json:"dbid,string"`
		Locked bool  `json:"locked"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if err := s.g.SetGemLock(req.DBID, req.Locked); err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (s *Server) handleCatalog(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"items": s.g.Catalog().Entries()})
}

func (s *Server) handleAddStackable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, fmt.Errorf("POST only"))
		return
	}
	var req struct {
		CID   int64 `json:"cid"`
		Count int64 `json:"count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if err := s.g.AddOrSetStackable(req.CID, req.Count); err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (s *Server) handleFillStackables(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, fmt.Errorf("POST only"))
		return
	}
	var req struct {
		Count int64 `json:"count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if err := s.g.FillStackables(req.Count); err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (s *Server) handleUnlockRecipes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, fmt.Errorf("POST only"))
		return
	}
	if err := s.g.UnlockAllRecipes(); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (s *Server) handleRecipes(w http.ResponseWriter, r *http.Request) {
	recipes, err := s.g.Recipes()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"recipes": recipes, "tools": domain.RecipeTools()})
}

func (s *Server) handleSetRecipeKnown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, fmt.Errorf("POST only"))
		return
	}
	var req struct {
		Key   int  `json:"key"`
		Known bool `json:"known"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if err := s.g.SetRecipeKnown(req.Key, req.Known); err != nil {
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}
