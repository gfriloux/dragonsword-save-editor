package domain

import "testing"

func TestTitleCatalogLoaded(t *testing.T) {
	if len(titleSeeds) != 108 {
		t.Fatalf("titles = %d, want 108", len(titleSeeds))
	}
	cats := map[int]bool{}
	for _, ts := range titleSeeds {
		if ts.NameFR == "" && ts.NameEN == "" {
			t.Fatalf("title %d has no resolvable name", ts.ID)
		}
		cat, _ := titlePos(ts.ID)
		cats[cat] = true
	}
	for _, want := range []int{32812, 32828, 32843, 32844, 32859, 32875, 32876} {
		if !cats[want] {
			t.Fatalf("category %d missing from the catalog", want)
		}
	}
	if len(cats) != 7 {
		t.Fatalf("categories = %d, want 7", len(cats))
	}
}

func TestTitlePos(t *testing.T) {
	// Ground truth from the real save: title 2100004 -> cat 32812 bit 36; 2104007 -> 32875,7.
	if cat, bit := titlePos(2100004); cat != 32812 || bit != 36 {
		t.Fatalf("titlePos(2100004) = (%d,%d), want (32812,36)", cat, bit)
	}
	if cat, bit := titlePos(2104007); cat != 32875 || bit != 7 {
		t.Fatalf("titlePos(2104007) = (%d,%d), want (32875,7)", cat, bit)
	}
}

// TestTitleUnlockAllMaskHighBit asserts the unlock mask reaches category 32843's high
// bits (past 62), which only round-trip through tb_title's INTEGER as a negative int64.
func TestTitleUnlockAllMaskHighBit(t *testing.T) {
	masks := map[int]uint64{}
	for _, ts := range titleSeeds {
		cat, bit := titlePos(ts.ID)
		masks[cat] |= 1 << uint(bit)
	}
	if masks[32843] == 0 {
		t.Fatal("no category-32843 bits in the unlock mask")
	}
	if int64(masks[32843]) >= 0 {
		t.Fatalf("category 32843 mask %#x has no high bit set (want a negative int64)", masks[32843])
	}
}

// --- real-save gated ------------------------------------------------------------

func TestTitlesRead(t *testing.T) {
	g := openGame(t)
	titles, err := g.Titles()
	if err != nil {
		t.Fatal(err)
	}
	if len(titles) != len(titleSeeds) {
		t.Fatalf("resolved %d titles, want %d", len(titles), len(titleSeeds))
	}
	// The fixture 6144_Slot1.db has exactly these six titles unlocked.
	wantUnlocked := map[int64]bool{
		2100004: true, 2101000: true, 2104007: true, 2104012: true, 2104013: true, 2104100: true,
	}
	got := map[int64]bool{}
	for _, ti := range titles {
		if ti.Unlocked {
			got[ti.ID] = true
		}
	}
	if len(got) != len(wantUnlocked) {
		t.Fatalf("unlocked = %d titles %v, want %d %v", len(got), got, len(wantUnlocked), wantUnlocked)
	}
	for id := range wantUnlocked {
		if !got[id] {
			t.Fatalf("title %d expected unlocked, but is not", id)
		}
	}
}

func TestTitlesToggle(t *testing.T) {
	g := openGame(t)
	before, err := g.Titles()
	if err != nil {
		t.Fatal(err)
	}
	var target Title
	for _, ti := range before {
		if !ti.Unlocked {
			target = ti
			break
		}
	}
	if target.ID == 0 {
		t.Skip("every title already unlocked in this save")
	}
	if err := g.SetTitleUnlocked(target.ID, true); err != nil {
		t.Fatal(err)
	}
	after, err := g.Titles()
	if err != nil {
		t.Fatal(err)
	}
	prev := map[int64]bool{}
	for _, ti := range before {
		prev[ti.ID] = ti.Unlocked
	}
	for _, ti := range after {
		want := prev[ti.ID]
		if ti.ID == target.ID {
			want = true
		}
		if ti.Unlocked != want {
			t.Fatalf("title %d unlocked=%v, want %v (only the target should change)", ti.ID, ti.Unlocked, want)
		}
	}
	// And it round-trips back off.
	if err := g.SetTitleUnlocked(target.ID, false); err != nil {
		t.Fatal(err)
	}
	back, err := g.Titles()
	if err != nil {
		t.Fatal(err)
	}
	for _, ti := range back {
		if ti.ID == target.ID && ti.Unlocked {
			t.Fatalf("title %d still unlocked after locking", target.ID)
		}
	}
}

func TestUnlockAllTitles(t *testing.T) {
	g := openGame(t)
	if err := g.UnlockAllTitles(); err != nil {
		t.Fatal(err)
	}
	titles, err := g.Titles()
	if err != nil {
		t.Fatal(err)
	}
	for _, ti := range titles {
		if !ti.Unlocked {
			t.Fatalf("title %d (cat %d bit %d) not unlocked after UnlockAllTitles", ti.ID, ti.Category, ti.Bit)
		}
	}
}

// TestSetTitlePreservesFav proves the write leaves FAV_BIT_FIELD (the displayed title)
// intact — an INSERT OR REPLACE would have wiped it.
func TestSetTitlePreservesFav(t *testing.T) {
	g := openGame(t)
	uid, err := g.UserID()
	if err != nil {
		t.Fatal(err)
	}
	// Seed a non-zero favourite in category 32812 (present in the fixture), then unlock a
	// different title in that category and assert the favourite is untouched.
	const cat, fav = 32812, int64(1 << 5)
	if _, err := g.s.Exec(
		`INSERT INTO tb_title (USER_DBID, CATEGORY, BIT_FIELD, FAV_BIT_FIELD) VALUES (?,?,0,?)
		 ON CONFLICT(USER_DBID, CATEGORY) DO UPDATE SET FAV_BIT_FIELD=excluded.FAV_BIT_FIELD`,
		uid, cat, fav); err != nil {
		t.Fatal(err)
	}
	if err := g.SetTitleUnlocked(int64(cat)*64+10, true); err != nil {
		t.Fatal(err)
	}
	var gotFav int64
	if err := g.s.DB().QueryRow(
		`SELECT FAV_BIT_FIELD FROM tb_title WHERE USER_DBID=? AND CATEGORY=?`, uid, cat).Scan(&gotFav); err != nil {
		t.Fatal(err)
	}
	if gotFav != fav {
		t.Fatalf("FAV_BIT_FIELD = %d after unlock, want %d (preserved)", gotFav, fav)
	}
}
