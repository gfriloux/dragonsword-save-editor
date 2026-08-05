package domain

import "testing"

// mountFor returns the equip-mount row for a character, if any.
func mountFor(t *testing.T, g *Game, characterCID int64) (Mount, bool) {
	t.Helper()
	mounts, err := g.Mounts()
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range mounts {
		if m.CharacterCID == characterCID {
			return m, true
		}
	}
	return Mount{}, false
}

func TestUnlockVehicleRoundTrip(t *testing.T) {
	g := openGame(t) // skips unless DSA_SAVE is set

	cat, err := g.VehicleCatalog()
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
		t.Skip("save already owns every vehicle; nothing to unlock")
	}

	if err := g.UnlockVehicle(target); err != nil {
		t.Fatalf("unlock: %v", err)
	}
	vehicles, err := g.Vehicles()
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, v := range vehicles {
		if v.CID == target {
			count++
			if v.DBID == 0 {
				t.Error("unlocked vehicle has a zero DBID")
			}
		}
	}
	if count != 1 {
		t.Fatalf("owned rows for CID %d = %d after unlock, want 1", target, count)
	}

	// No-op on re-unlock.
	if err := g.UnlockVehicle(target); err != nil {
		t.Fatalf("re-unlock: %v", err)
	}
	vehicles, _ = g.Vehicles()
	count = 0
	for _, v := range vehicles {
		if v.CID == target {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("owned rows for CID %d = %d after double unlock, want 1", target, count)
	}
}

func TestSetMountRoundTrip(t *testing.T) {
	g := openGame(t)

	vehicles, err := g.Vehicles()
	if err != nil || len(vehicles) == 0 {
		t.Fatalf("vehicles: %v", err)
	}
	owned := vehicles[0].DBID

	// A character that already has an equip-mount row: equip then unequip.
	mounts, err := g.Mounts()
	if err != nil || len(mounts) == 0 {
		t.Fatalf("mounts: %v", err)
	}
	existing := mounts[0].CharacterCID

	if err := g.SetMount(existing, owned); err != nil {
		t.Fatalf("equip existing: %v", err)
	}
	if m, _ := mountFor(t, g, existing); m.Vehicle == nil || m.VehicleDBID != owned {
		t.Fatalf("char %d did not get vehicle %d", existing, owned)
	}
	if err := g.SetMount(existing, 0); err != nil {
		t.Fatalf("unequip: %v", err)
	}
	if m, _ := mountFor(t, g, existing); m.VehicleDBID != 0 || m.Vehicle != nil {
		t.Fatalf("char %d still has a mount after unequip", existing)
	}

	// A character with NO equip-mount row: SetMount must insert one.
	chars, err := g.Characters()
	if err != nil {
		t.Fatal(err)
	}
	hasRow := map[int64]bool{}
	for _, m := range mounts {
		hasRow[m.CharacterCID] = true
	}
	var fresh int64
	for _, c := range chars {
		if !hasRow[c.CID] {
			fresh = c.CID
			break
		}
	}
	if fresh != 0 {
		if err := g.SetMount(fresh, owned); err != nil {
			t.Fatalf("equip fresh: %v", err)
		}
		m, ok := mountFor(t, g, fresh)
		if !ok || m.Vehicle == nil || m.VehicleDBID != owned {
			t.Fatalf("insert path failed: char %d has no resolved mount", fresh)
		}
	} else {
		t.Log("every character already has an equip-mount row; insert path not exercised")
	}
}

func TestSetMountRejectsUnownedVehicle(t *testing.T) {
	g := openGame(t)
	chars, err := g.Characters()
	if err != nil || len(chars) == 0 {
		t.Fatalf("characters: %v", err)
	}
	// A DBID that is (almost certainly) not an owned vehicle.
	if err := g.SetMount(chars[0].CID, 1); err == nil {
		t.Fatal("expected an error equipping an unowned vehicle DBID")
	}
}

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
