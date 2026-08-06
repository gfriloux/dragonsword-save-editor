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
