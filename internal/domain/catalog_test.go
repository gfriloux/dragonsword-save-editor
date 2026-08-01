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
		1420204: "food",
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
	if it.Category != "currency" || it.Name != "Currency 1000001" {
		t.Fatalf("LookupCtx currency: category=%q name=%q", it.Category, it.Name)
	}
}

func TestLookupFallback(t *testing.T) {
	c, err := LoadCatalog("")
	if err != nil {
		t.Fatal(err)
	}
	it := c.Lookup(1420204)
	if it.Category != "food" {
		t.Errorf("category = %q, want food", it.Category)
	}
	if it.Known {
		t.Error("expected Known=false for an unseeded CID")
	}
	if it.Name != "Food 1420204" {
		t.Errorf("fallback name = %q, want %q", it.Name, "Food 1420204")
	}
}

func TestLabelPrecedenceAndPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "labels.json")
	c, err := LoadCatalog(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.SetLabel(1420204, "Grilled Meat"); err != nil {
		t.Fatal(err)
	}
	it := c.Lookup(1420204)
	if it.Name != "Grilled Meat" || !it.Known {
		t.Fatalf("after label: name=%q known=%v", it.Name, it.Known)
	}

	// A fresh catalog loading the same overrides file sees the label.
	c2, err := LoadCatalog(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := c2.Lookup(1420204).Name; got != "Grilled Meat" {
		t.Fatalf("persisted label = %q, want %q", got, "Grilled Meat")
	}

	// Clearing the label falls back to inference.
	if err := c2.SetLabel(1420204, ""); err != nil {
		t.Fatal(err)
	}
	if got := c2.Lookup(1420204); got.Known || got.Name != "Food 1420204" {
		t.Fatalf("after clear: name=%q known=%v", got.Name, got.Known)
	}
}
