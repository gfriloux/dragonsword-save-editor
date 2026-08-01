package web

import (
	"encoding/json"
	"fmt"
	"net/http"
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

func (s *Server) handleStack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, fmt.Errorf("POST only"))
		return
	}
	var req struct {
		Kind  string `json:"kind"`
		ID    int64  `json:"id"`
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
