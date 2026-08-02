package save

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDiscover(t *testing.T) {
	game := t.TempDir()
	acc := filepath.Join(SaveGamesDir(game), "6144")
	if err := os.MkdirAll(acc, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name string) string {
		p := filepath.Join(acc, name)
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	s1 := write("6144_Slot1.db")
	write("ScreenShot_1.png")
	s2 := write("6144_Slot2.db") // no screenshot
	// Make Slot2 newer so it sorts first.
	newer := time.Now().Add(time.Hour)
	if err := os.Chtimes(s2, newer, newer); err != nil {
		t.Fatal(err)
	}

	slots, err := Discover(game)
	if err != nil {
		t.Fatal(err)
	}
	if len(slots) != 2 {
		t.Fatalf("want 2 slots, got %d", len(slots))
	}
	if slots[0].Path != s2 || slots[0].Slot != 2 || slots[0].Screenshot != "" {
		t.Fatalf("slot0 unexpected: %+v", slots[0])
	}
	if slots[1].Path != s1 || slots[1].Slot != 1 || slots[1].AccountID != "6144" ||
		filepath.Base(slots[1].Screenshot) != "ScreenShot_1.png" {
		t.Fatalf("slot1 unexpected: %+v", slots[1])
	}
}

func TestDiscoverEmpty(t *testing.T) {
	slots, err := Discover(t.TempDir())
	if err != nil {
		t.Fatalf("empty game dir should not error: %v", err)
	}
	if len(slots) != 0 {
		t.Fatalf("want 0 slots, got %d", len(slots))
	}
}
