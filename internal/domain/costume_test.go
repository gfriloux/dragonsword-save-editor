package domain

import "testing"

// findCostume returns the first owned costume whose CID equals cid.
func findCostume(t *testing.T, g *Game, cid int64) (Costume, bool) {
	t.Helper()
	costumes, err := g.Costumes()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range costumes {
		if c.CID == cid {
			return c, true
		}
	}
	return Costume{}, false
}

func TestUnlockCostumeRoundTrip(t *testing.T) {
	g := openGame(t) // skips unless DSA_SAVE is set

	// Pick a costume the save does not own yet.
	cat, err := g.CostumeCatalog()
	if err != nil {
		t.Fatal(err)
	}
	var target int64
	for _, e := range cat {
		if !e.Owned {
			target = e.CID
			break
		}
	}
	if target == 0 {
		t.Skip("save already owns every costume; nothing to unlock")
	}

	if err := g.UnlockCostume(target); err != nil {
		t.Fatalf("unlock: %v", err)
	}
	c, ok := findCostume(t, g, target)
	if !ok {
		t.Fatalf("costume %d not owned after unlock", target)
	}
	if c.DBID == 0 {
		t.Fatal("unlocked costume has a zero DBID")
	}
	if c.EquipCharacterCID != 0 {
		t.Errorf("freshly unlocked costume is equipped (char %d), want none", c.EquipCharacterCID)
	}

	// Unlocking again is a no-op: still exactly one row for that CID.
	if err := g.UnlockCostume(target); err != nil {
		t.Fatalf("re-unlock: %v", err)
	}
	all, err := g.Costumes()
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, x := range all {
		if x.CID == target {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("owned rows for CID %d = %d after double unlock, want 1", target, count)
	}
}

func TestSetCostumeEquipRoundTrip(t *testing.T) {
	g := openGame(t)

	// Any character to be the wearer.
	chars, err := g.Characters()
	if err != nil || len(chars) == 0 {
		t.Fatalf("characters: %v", err)
	}
	wearer := chars[0].CID

	// Any owned costume.
	costumes, err := g.Costumes()
	if err != nil || len(costumes) == 0 {
		t.Fatalf("costumes: %v", err)
	}
	dbid := costumes[0].DBID

	if err := g.SetCostumeEquip(dbid, wearer); err != nil {
		t.Fatalf("equip: %v", err)
	}
	got := reloadCostume(t, g, dbid)
	if got.EquipCharacterCID != wearer {
		t.Fatalf("after equip, wearer = %d, want %d", got.EquipCharacterCID, wearer)
	}

	if err := g.SetCostumeEquip(dbid, 0); err != nil {
		t.Fatalf("unequip: %v", err)
	}
	got = reloadCostume(t, g, dbid)
	if got.EquipCharacterCID != 0 {
		t.Fatalf("after unequip, wearer = %d, want 0", got.EquipCharacterCID)
	}
}

// reloadCostume re-reads the owned costume with the given DBID.
func reloadCostume(t *testing.T, g *Game, dbid int64) Costume {
	t.Helper()
	costumes, err := g.Costumes()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range costumes {
		if c.DBID == dbid {
			return c
		}
	}
	t.Fatalf("costume DBID %d not found", dbid)
	return Costume{}
}

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
