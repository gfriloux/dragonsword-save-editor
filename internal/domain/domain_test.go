package domain

import (
	"os"
	"testing"

	"github.com/gfriloux/dragonsword-save-editor/internal/save"
)

// openGame copies the real save (DSA_SAVE) to a temp file and returns an editable
// Game over it. Skips when DSA_SAVE is unset.
func openGame(t *testing.T) *Game {
	t.Helper()
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
	s, err := save.Open(work, "")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	cat, err := LoadCatalog("")
	if err != nil {
		t.Fatal(err)
	}
	return New(s, cat)
}

func TestCurrenciesResolve(t *testing.T) {
	g := openGame(t)
	cur, err := g.Currencies()
	if err != nil {
		t.Fatal(err)
	}
	if len(cur) == 0 {
		t.Fatal("no currencies found")
	}
	for _, c := range cur {
		if c.Category != "currency" {
			t.Errorf("currency %d resolved to category %q", c.CID, c.Category)
		}
		if c.NameFR == "" {
			t.Errorf("currency %d has empty name", c.CID)
		}
	}
}

func TestSetCurrencyRoundTrip(t *testing.T) {
	g := openGame(t)
	cur, err := g.Currencies()
	if err != nil || len(cur) == 0 {
		t.Fatalf("currencies: %v", err)
	}
	target := cur[0]
	const want = 777001
	if err := g.SetCurrency(target.CID, want); err != nil {
		t.Fatalf("set: %v", err)
	}
	cur2, err := g.Currencies()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cur2 {
		if c.CID == target.CID && c.Amount != want {
			t.Fatalf("amount = %d, want %d", c.Amount, want)
		}
	}
}

func TestConsumablesAndSetStack(t *testing.T) {
	g := openGame(t)
	items, err := g.Consumables()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) == 0 {
		t.Fatal("no consumables found")
	}
	var haveStackable, haveCook bool
	for _, it := range items {
		switch it.Kind {
		case KindStackable:
			haveStackable = true
		case KindCook:
			haveCook = true
		default:
			t.Fatalf("unexpected kind %q", it.Kind)
		}
	}
	if !haveStackable || !haveCook {
		t.Fatalf("expected both kinds, stackable=%v cook=%v", haveStackable, haveCook)
	}

	first := items[0]
	if err := g.SetStack(first.Kind, first.ID, 99); err != nil {
		t.Fatalf("set stack: %v", err)
	}
	items2, err := g.Consumables()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range items2 {
		if it.Kind == first.Kind && it.ID == first.ID && it.Count != 99 {
			t.Fatalf("count = %d, want 99", it.Count)
		}
	}
}

func TestCharactersRead(t *testing.T) {
	g := openGame(t)
	chars, err := g.Characters()
	if err != nil {
		t.Fatal(err)
	}
	if len(chars) == 0 {
		t.Fatal("no characters")
	}
	for _, c := range chars {
		if c.Category != "character" {
			t.Errorf("char %d category %q, want character", c.CID, c.Category)
		}
		if c.Level < 1 {
			t.Errorf("char %d level %d", c.CID, c.Level)
		}
	}
}

func TestTeamsRead(t *testing.T) {
	g := openGame(t)
	teams, err := g.Teams()
	if err != nil {
		t.Fatal(err)
	}
	if len(teams) == 0 {
		t.Fatal("no team pages")
	}
	for _, p := range teams {
		if len(p.Slots) != 3 {
			t.Fatalf("page %d has %d slots, want 3", p.PageID, len(p.Slots))
		}
		for _, s := range p.Slots {
			if !s.Empty && s.Category != "character" {
				t.Errorf("slot category %q, want character", s.Category)
			}
		}
	}
}

func TestEquipmentReadAndEdit(t *testing.T) {
	g := openGame(t)
	eq, err := g.Equipments()
	if err != nil {
		t.Fatal(err)
	}
	if len(eq) == 0 {
		t.Fatal("no equipment")
	}
	for _, e := range eq {
		if e.Category != "gear" {
			t.Errorf("equip %d category %q, want gear", e.CID, e.Category)
		}
	}
	target := eq[0]
	if err := g.SetEnchant(target.DBID, 12); err != nil {
		t.Fatalf("set enchant: %v", err)
	}
	if err := g.SetEquipLock(target.DBID, !target.IsLock); err != nil {
		t.Fatalf("set lock: %v", err)
	}
	eq2, err := g.Equipments()
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range eq2 {
		if e.DBID == target.DBID {
			if e.EnchantLevel != 12 {
				t.Fatalf("enchant = %d, want 12", e.EnchantLevel)
			}
			if e.IsLock == target.IsLock {
				t.Fatalf("lock did not toggle")
			}
		}
	}
}

func TestGemsEmptyOK(t *testing.T) {
	g := openGame(t)
	gems, err := g.Gems()
	if err != nil {
		t.Fatalf("gems: %v", err)
	}
	// The reference save has none; the accessor must simply return an empty slice.
	for _, gm := range gems {
		if gm.DBID == 0 {
			t.Error("gem with zero DBID")
		}
	}
}

func TestAddOrSetStackable(t *testing.T) {
	g := openGame(t)
	const cid = 1459999 // not present in the reference save
	if err := g.AddOrSetStackable(cid, 42); err != nil {
		t.Fatalf("add: %v", err)
	}
	find := func() *Stack {
		items, err := g.Consumables()
		if err != nil {
			t.Fatal(err)
		}
		for i := range items {
			if items[i].Kind == KindStackable && items[i].ID == cid {
				return &items[i]
			}
		}
		return nil
	}
	if s := find(); s == nil || s.Count != 42 {
		t.Fatalf("after add: %+v", s)
	}
	// Upsert: setting again replaces the quantity.
	if err := g.AddOrSetStackable(cid, 7); err != nil {
		t.Fatal(err)
	}
	if s := find(); s == nil || s.Count != 7 {
		t.Fatalf("after upsert: %+v", s)
	}
}

func TestFillStackables(t *testing.T) {
	g := openGame(t)
	if err := g.FillStackables(500); err != nil {
		t.Fatal(err)
	}
	items, err := g.Consumables()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range items {
		if it.Kind == KindStackable && it.Count != 500 {
			t.Fatalf("stackable %d count %d, want 500", it.ID, it.Count)
		}
	}
}
