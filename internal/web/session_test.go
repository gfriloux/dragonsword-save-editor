package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gfriloux/dragonsword-save-editor/internal/save"
)

// newTestServer returns a server with no catalog (session-agnostic endpoints do
// not need one) and an isolated config dir.
func newTestServer(t *testing.T) *Server {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	return New(nil)
}

func get(t *testing.T, s *Server, path string) *httptest.ResponseRecorder {
	t.Helper()
	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
	return rr
}

func TestConfigAndSavesFlow(t *testing.T) {
	s := newTestServer(t)

	// Fresh: no game dir, no save.
	rr := get(t, s, "/api/config")
	if rr.Code != 200 {
		t.Fatalf("config: %d", rr.Code)
	}
	var cfg map[string]any
	json.Unmarshal(rr.Body.Bytes(), &cfg)
	if cfg["gameDir"] != "" || cfg["saveOpen"] != false {
		t.Fatalf("unexpected fresh config: %v", cfg)
	}

	// A save-scoped endpoint 409s before any save is open.
	if rr := get(t, s, "/api/info"); rr.Code != http.StatusConflict {
		t.Fatalf("info without save: want 409, got %d", rr.Code)
	}

	// Saves with no game dir → 400.
	if rr := get(t, s, "/api/saves"); rr.Code != http.StatusBadRequest {
		t.Fatalf("saves without game dir: want 400, got %d", rr.Code)
	}

	// Build a fake game folder with one slot + screenshot.
	game := t.TempDir()
	acc := filepath.Join(save.SaveGamesDir(game), "6144")
	os.MkdirAll(acc, 0o755)
	os.WriteFile(filepath.Join(acc, "6144_Slot1.db"), []byte("x"), 0o644)
	shot := filepath.Join(acc, "ScreenShot_1.png")
	os.WriteFile(shot, []byte("PNGDATA"), 0o644)

	// Set the game dir.
	rr = post(t, s, "/api/config/game-dir", `{"dir":`+jsonStr(game)+`}`)
	if rr.Code != 200 {
		t.Fatalf("set game-dir: %d (%s)", rr.Code, rr.Body)
	}

	// Now saves lists the slot.
	rr = get(t, s, "/api/saves")
	if rr.Code != 200 {
		t.Fatalf("saves: %d", rr.Code)
	}
	var sv struct {
		Slots []save.Slot `json:"slots"`
	}
	json.Unmarshal(rr.Body.Bytes(), &sv)
	if len(sv.Slots) != 1 || sv.Slots[0].Slot != 1 {
		t.Fatalf("unexpected slots: %+v", sv.Slots)
	}

	// Screenshot inside the tree is served; a traversal attempt is refused.
	if rr := get(t, s, "/api/screenshot?path="+sv.Slots[0].Screenshot); rr.Code != 200 || rr.Body.String() != "PNGDATA" {
		t.Fatalf("screenshot: %d body=%q", rr.Code, rr.Body.String())
	}
	if rr := get(t, s, "/api/screenshot?path=/etc/passwd"); rr.Code != http.StatusNotFound {
		t.Fatalf("screenshot traversal: want 404, got %d", rr.Code)
	}

	// Opening a path outside the save tree is refused.
	if rr := post(t, s, "/api/open", `{"path":"/etc/passwd"}`); rr.Code != http.StatusBadRequest {
		t.Fatalf("open outside tree: want 400, got %d", rr.Code)
	}
}

func post(t *testing.T, s *Server, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, path, strings.NewReader(body)))
	return rr
}

func jsonStr(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
