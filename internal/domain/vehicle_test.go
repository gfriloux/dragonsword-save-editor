package domain

import "testing"

func TestVehiclesResolve(t *testing.T) {
	g := openGame(t) // skips unless DSA_SAVE is set
	vehicles, err := g.Vehicles()
	if err != nil {
		t.Fatal(err)
	}
	if len(vehicles) == 0 {
		t.Fatal("no vehicles owned in the save")
	}
	for _, v := range vehicles {
		if v.DBID == 0 {
			t.Errorf("vehicle CID %d has a zero DBID", v.CID)
		}
		if v.Category != "mount" {
			t.Errorf("vehicle CID %d resolved to category %q, want mount", v.CID, v.Category)
		}
		if v.NameFR == "" {
			t.Errorf("vehicle CID %d has no FR name", v.CID)
		}
	}
}

func TestVehicleCatalogOwnedFlag(t *testing.T) {
	g := openGame(t)
	entries, err := g.VehicleCatalog()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("vehicle catalog is empty")
	}
	for _, e := range entries {
		if e.Category != "mount" {
			t.Errorf("catalog entry CID %d is category %q, want mount", e.CID, e.Category)
		}
	}

	vehicles, err := g.Vehicles()
	if err != nil {
		t.Fatal(err)
	}
	ownedCIDs := map[int64]bool{}
	for _, v := range vehicles {
		ownedCIDs[v.CID] = true
	}
	flagged := map[int64]bool{}
	for _, e := range entries {
		if e.Owned {
			flagged[e.CID] = true
		}
	}
	if len(flagged) != len(ownedCIDs) {
		t.Fatalf("catalog flags %d owned vehicles, save owns %d distinct CIDs", len(flagged), len(ownedCIDs))
	}
	for cid := range ownedCIDs {
		if !flagged[cid] {
			t.Errorf("owned vehicle CID %d not flagged as owned in the catalog", cid)
		}
	}
}

func TestMountsResolve(t *testing.T) {
	g := openGame(t)
	mounts, err := g.Mounts()
	if err != nil {
		t.Fatal(err)
	}
	if len(mounts) == 0 {
		t.Fatal("no equip-mount rows in the save")
	}
	ridden := 0
	for _, m := range mounts {
		if m.CharacterCID == 0 {
			t.Error("mount row with a zero character CID")
		}
		// Vehicle presence must agree with VehicleDBID.
		if m.Vehicle != nil {
			ridden++
			if m.VehicleDBID == 0 {
				t.Errorf("char %d has a resolved vehicle but a zero DBID", m.CharacterCID)
			}
			if m.Vehicle.Category != "mount" {
				t.Errorf("char %d rides a non-mount category %q", m.CharacterCID, m.Vehicle.Category)
			}
		}
	}
	if ridden == 0 {
		t.Fatal("no character rides a mount; expected at least one")
	}
}
