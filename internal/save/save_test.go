package save

import (
	"os"
	"testing"
)

// TestEditRoundTrip opens a copy of a real save, edits an item quantity, writes
// it back, and verifies the change persists through a fresh decrypt.
func TestEditRoundTrip(t *testing.T) {
	src := os.Getenv("DSA_SAVE")
	if src == "" {
		t.Skip("set DSA_SAVE to a real .db to run this test")
	}
	raw, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	work := t.TempDir() + "/save.db"
	if err := os.WriteFile(work, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := Open(work, "")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	tables, err := s.Tables()
	if err != nil || len(tables) == 0 {
		t.Fatalf("tables: %v (%d)", err, len(tables))
	}

	const sentinel = 12345
	if _, err := s.Exec(`UPDATE tb_stackable_item SET STACK_CNT=? WHERE ITEM_CID=1000501`, sentinel); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := s.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	s.Close()

	s2, err := Open(work, "")
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer s2.Close()
	var got int
	if err := s2.DB().QueryRow(`SELECT STACK_CNT FROM tb_stackable_item WHERE ITEM_CID=1000501`).Scan(&got); err != nil {
		t.Fatalf("scan: %v", err)
	}
	if got != sentinel {
		t.Fatalf("STACK_CNT = %d, want %d", got, sentinel)
	}

	if out := os.Getenv("DSA_OUT"); out != "" {
		raw2, _ := os.ReadFile(work)
		os.WriteFile(out, raw2, 0o644)
	}
}
