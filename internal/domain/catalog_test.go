package domain

import (
	"path/filepath"
	"testing"
)

func TestInferCategory(t *testing.T) {
	cases := map[int64]string{
		1000001: "misc", // currency is asserted from context, not inferred
		1000500: "misc", // stackable sharing the 100x prefix
		1410002: "potion",
		1429999: "food",
		1430003: "material",
		1450001: "material",
		1510018: "misc",
		9999999: "misc",
	}
	for cid, want := range cases {
		if got := inferCategory(cid); got != want {
			t.Errorf("inferCategory(%d) = %q, want %q", cid, got, want)
		}
	}
}

func TestLookupCtxCategory(t *testing.T) {
	c, err := LoadCatalog("")
	if err != nil {
		t.Fatal(err)
	}
	// Context overrides inference for the fallback name and category.
	it := c.LookupCtx(1000001, "currency")
	if it.Category != "currency" || it.NameFR != "Currency 1000001" {
		t.Fatalf("LookupCtx currency: category=%q name=%q", it.Category, it.NameFR)
	}
}

func TestLookupFallback(t *testing.T) {
	c, err := LoadCatalog("")
	if err != nil {
		t.Fatal(err)
	}
	it := c.Lookup(1429999)
	if it.Category != "food" {
		t.Errorf("category = %q, want food", it.Category)
	}
	if it.Known {
		t.Error("expected Known=false for an unseeded CID")
	}
	if it.NameFR != "Food 1429999" {
		t.Errorf("fallback name = %q, want %q", it.NameFR, "Food 1429999")
	}
}

func TestLabelPrecedenceAndPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "labels.json")
	c, err := LoadCatalog(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.SetLabel(1429999, "Grilled Meat"); err != nil {
		t.Fatal(err)
	}
	it := c.Lookup(1429999)
	if it.NameFR != "Grilled Meat" || !it.Known {
		t.Fatalf("after label: name=%q known=%v", it.NameFR, it.Known)
	}

	// A fresh catalog loading the same overrides file sees the label.
	c2, err := LoadCatalog(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := c2.Lookup(1429999).NameFR; got != "Grilled Meat" {
		t.Fatalf("persisted label = %q, want %q", got, "Grilled Meat")
	}

	// Clearing the label falls back to inference.
	if err := c2.SetLabel(1429999, ""); err != nil {
		t.Fatal(err)
	}
	if got := c2.Lookup(1429999); got.Known || got.NameFR != "Food 1429999" {
		t.Fatalf("after clear: name=%q known=%v", got.NameFR, got.Known)
	}
}

func TestEntries(t *testing.T) {
	c, err := LoadCatalog("")
	if err != nil {
		t.Fatal(err)
	}
	es := c.Entries()
	if len(es) < 100 {
		t.Fatalf("expected a populated catalog, got %d", len(es))
	}
	for i := 1; i < len(es); i++ {
		if es[i-1].CID > es[i].CID {
			t.Fatal("entries not sorted by CID")
		}
	}
	var eileen *Item
	for i := range es {
		if es[i].CID == 10001 {
			eileen = &es[i]
		}
	}
	if eileen == nil || eileen.NameFR != "Eileen" {
		t.Fatalf("expected 10001=Eileen, got %+v", eileen)
	}
}
