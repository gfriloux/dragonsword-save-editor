package domain

import "testing"

func TestConsumableCategoriesShape(t *testing.T) {
	cats := ConsumableCategories()
	if len(cats) == 0 {
		t.Fatal("no categories loaded from item_categories.json")
	}
	if cats[len(cats)-1].Key != groupUnsorted {
		t.Errorf("last category = %q, want %q (unsorted is always last)", cats[len(cats)-1].Key, groupUnsorted)
	}
	seen := map[string]bool{}
	for _, c := range cats {
		if c.Key == "" || c.LabelFR == "" || c.LabelEN == "" || c.Color == "" {
			t.Errorf("category %+v has an empty field", c)
		}
		if seen[c.Key] {
			t.Errorf("duplicate category key %q", c.Key)
		}
		seen[c.Key] = true
	}
	// Returned slice must be a copy (mutating it must not affect the source).
	cats[0].LabelFR = "MUTATED"
	if ConsumableCategories()[0].LabelFR == "MUTATED" {
		t.Error("ConsumableCategories() leaks its backing array")
	}
}

func TestConsumableGroups(t *testing.T) {
	groups := ConsumableGroups()
	known := map[string]bool{}
	for _, g := range groups {
		if g.Key == "" || g.LabelFR == "" || g.LabelEN == "" {
			t.Errorf("group %+v has an empty field", g)
		}
		known[g.Key] = true
	}
	// The six surfaced CategoryTypes must all be declared as headers.
	for _, want := range []string{"COOK", "GEM", "KARMA", "NORMAL_MATERIAL", "GROW_MATERIAL", "VALUABLE"} {
		if !known[want] {
			t.Errorf("missing super-category group %q", want)
		}
	}
	// Every category's Group is a known header key or the unsorted fallback, and is
	// never empty (the UI needs it to place the category under a header).
	for _, c := range ConsumableCategories() {
		if c.Group == "" {
			t.Errorf("category %q (%s) has no group", c.Key, c.LabelFR)
		}
		if !known[c.Group] && c.Group != groupUnsorted {
			t.Errorf("category %q has unknown group %q", c.Key, c.Group)
		}
	}
	// Returned slice must be a copy.
	groups[0].LabelFR = "MUTATED"
	if ConsumableGroups()[0].LabelFR == "MUTATED" {
		t.Error("ConsumableGroups() leaks its backing array")
	}
}

func TestItemGroupFromSeed(t *testing.T) {
	c, err := LoadCatalog("")
	if err != nil {
		t.Fatal(err)
	}
	// Groups come straight from the datamined seed (game item categories).
	cases := map[int64]string{
		1430003: "1700", // Viande de bête — cooking ingredient (meat)
		1450001: "1603", // Essence de monstre — monster material
		1310001: "1500", // rune — Détermination
		1000500: "1802", // Invitation du Destin — exchange material
		10001:   "",     // Eileen (character, th.gl-only, no game category) → no group
	}
	for cid, want := range cases {
		if got := c.Lookup(cid).Group; got != want {
			t.Errorf("Lookup(%d).Group = %q, want %q", cid, got, want)
		}
	}
}
