package domain

import "testing"

func TestCostumesResolve(t *testing.T) {
	g := openGame(t) // skips unless DSA_SAVE is set
	costumes, err := g.Costumes()
	if err != nil {
		t.Fatal(err)
	}
	if len(costumes) == 0 {
		t.Fatal("no costumes owned in the save")
	}
	for _, c := range costumes {
		if c.DBID == 0 {
			t.Errorf("costume CID %d has a zero DBID", c.CID)
		}
		if c.Category != "costume" {
			t.Errorf("costume CID %d resolved to category %q, want costume", c.CID, c.Category)
		}
		if c.NameFR == "" {
			t.Errorf("costume CID %d has no FR name", c.CID)
		}
	}
}

func TestCostumeCatalogOwnedFlag(t *testing.T) {
	g := openGame(t)
	entries, err := g.CostumeCatalog()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("costume catalog is empty")
	}

	// Every catalog entry is a costume.
	for _, e := range entries {
		if e.Category != "costume" {
			t.Errorf("catalog entry CID %d is category %q, want costume", e.CID, e.Category)
		}
	}

	// The owned flags must match the distinct owned CIDs exactly.
	costumes, err := g.Costumes()
	if err != nil {
		t.Fatal(err)
	}
	ownedCIDs := map[int64]bool{}
	for _, c := range costumes {
		ownedCIDs[c.CID] = true
	}
	flagged := map[int64]bool{}
	for _, e := range entries {
		if e.Owned {
			flagged[e.CID] = true
		}
	}
	if len(flagged) != len(ownedCIDs) {
		t.Fatalf("catalog flags %d owned costumes, save owns %d distinct CIDs", len(flagged), len(ownedCIDs))
	}
	for cid := range ownedCIDs {
		if !flagged[cid] {
			t.Errorf("owned costume CID %d not flagged as owned in the catalog", cid)
		}
	}
}
