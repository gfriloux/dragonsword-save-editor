package domain

import "testing"

func TestRandIDNonZeroAndVaries(t *testing.T) {
	seen := map[int64]bool{}
	for i := 0; i < 1000; i++ {
		id, err := randID()
		if err != nil {
			t.Fatal(err)
		}
		if id == 0 {
			t.Fatal("randID returned 0, which is reserved for none/unset")
		}
		if id < 0 || id > 0xFFFFFFFF {
			t.Fatalf("randID out of uint32 range: %d", id)
		}
		seen[id] = true
	}
	// 1000 draws over 2^32 should essentially never collide.
	if len(seen) < 999 {
		t.Fatalf("suspiciously many collisions: %d unique of 1000", len(seen))
	}
}

func TestMintIDFromSkipsTakenIDs(t *testing.T) {
	// The first two draws collide with existing rows; the loop must retry and
	// return the first free id (333), exercising the real retry logic.
	taken := map[int64]bool{111: true, 222: true}
	seq := []int64{111, 222, 333, 444}
	i := 0
	draw := func() (int64, error) { id := seq[i]; i++; return id, nil }
	exists := func(id int64) (bool, error) { return taken[id], nil }

	got, err := mintIDFrom(draw, exists)
	if err != nil {
		t.Fatal(err)
	}
	if got != 333 {
		t.Fatalf("mintIDFrom = %d, want 333 (first free id)", got)
	}
}

func TestMintIDFromExhausts(t *testing.T) {
	// Every id is taken: the loop must give up rather than spin forever.
	draw := func() (int64, error) { return 1, nil }
	exists := func(int64) (bool, error) { return true, nil }
	if _, err := mintIDFrom(draw, exists); err == nil {
		t.Fatal("expected exhaustion error when all ids are taken")
	}
}

func TestMintIDNonZero(t *testing.T) {
	// Nothing taken: mintID must return a usable non-zero id from the RNG.
	id, err := mintID(func(int64) (bool, error) { return false, nil })
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Fatal("mintID returned 0")
	}
}

func TestMintDBIDAgainstRealSave(t *testing.T) {
	g := openGame(t) // skips unless DSA_SAVE is set
	id, err := g.mintDBID("tb_costume", "COSTUME_DBID")
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Fatal("mintDBID returned 0")
	}
	var n int
	if err := g.s.DB().QueryRow(
		"SELECT COUNT(*) FROM tb_costume WHERE COSTUME_DBID=?", id).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("mintDBID returned an id already present: %d", id)
	}
}
