package domain

import "testing"

func TestRecipeCatalogLoaded(t *testing.T) {
	if len(recipeTools) != 3 {
		t.Fatalf("tools = %d, want 3", len(recipeTools))
	}
	if len(recipeSeeds) < 1000 {
		t.Fatalf("recipes = %d, want >= 1000", len(recipeSeeds))
	}
	minKey, maxKey := recipeSeeds[0].Key, recipeSeeds[0].Key
	for _, r := range recipeSeeds {
		if r.Key < minKey {
			minKey = r.Key
		}
		if r.Key > maxKey {
			maxKey = r.Key
		}
		if len(r.Dishes) == 0 {
			t.Fatalf("recipe %d has no dishes", r.Key)
		}
	}
	if cat := maxKey / 64; cat < 61 {
		t.Fatalf("max category = %d, want the tail past the old blanket (>= 61)", cat)
	}
	t.Logf("keys %d-%d, categories %d-%d", minKey, maxKey, minKey/64, maxKey/64)
}

func TestRecipeEffects(t *testing.T) {
	var withEffect int
	for _, r := range recipeSeeds {
		if len(r.Effects) > 0 {
			withEffect++
		}
	}
	if withEffect == 0 {
		t.Fatal("no recipe carries an effect; the ContentsBuff resolution is broken")
	}
	// The first recipe (key 1001, grilled meat) restores HP across its tiers.
	r := recipeSeeds[0]
	if r.EffectName == nil || r.EffectName.FR == "" {
		t.Fatalf("recipe %d has no effect name", r.Key)
	}
	if len(r.Effects) != len(r.Dishes) {
		t.Fatalf("recipe %d: %d effects for %d dish tiers (must be parallel)", r.Key, len(r.Effects), len(r.Dishes))
	}
	t.Logf("%d/%d recipes carry an effect; recipe %d effect: %q", withEffect, len(recipeSeeds), r.Key, r.Effects[0].FR)
}

func TestSwitchPos(t *testing.T) {
	// Ground truth from docs/switches.md: "Poisson grillé" key 1002 -> cat 15, bit 42.
	if cat, bit := switchPos(1002); cat != 15 || bit != 42 {
		t.Fatalf("switchPos(1002) = (%d,%d), want (15,42)", cat, bit)
	}
}

// TestUnlockAllMaskCoversCat62 asserts the key-accurate unlock reaches the category-62
// tail (keys 4000-4008) that the old blanket 15-60 missed.
func TestUnlockAllMaskCoversCat62(t *testing.T) {
	masks := map[int]uint64{}
	for _, rs := range recipeSeeds {
		cat, bit := switchPos(rs.Key)
		masks[cat] |= 1 << uint(bit)
	}
	if masks[62] == 0 {
		t.Fatal("no category-62 bits in the unlock mask")
	}
	if masks[62]&(1<<40) == 0 { // key 4008 -> cat 62, bit 40
		t.Fatal("key 4008 (cat 62 bit 40) not in the mask")
	}
}

func TestResolveIngredientsCollapse(t *testing.T) {
	c, err := LoadCatalog("")
	if err != nil {
		t.Fatal(err)
	}
	// The same type twice must collapse to one ingredient with Qty 2.
	seeds := []ingredientSeed{{Kind: "type", ID: 1700}, {Kind: "type", ID: 1700}}
	got := resolveIngredients(c, seeds, map[int64]int64{}, map[int64]int64{1700: 5}, consumableCategoryLabels())
	if len(got) != 1 {
		t.Fatalf("ingredients = %d, want 1 (collapsed)", len(got))
	}
	if got[0].Qty != 2 {
		t.Fatalf("qty = %d, want 2", got[0].Qty)
	}
	if got[0].Owned != 5 {
		t.Fatalf("owned = %d, want 5 (category total)", got[0].Owned)
	}
	if got[0].NameFR == "" {
		t.Fatal("type ingredient resolved to an empty FR label")
	}
}

// --- real-save gated ------------------------------------------------------------

func TestRecipesReadAndToggle(t *testing.T) {
	g := openGame(t)
	recipes, err := g.Recipes()
	if err != nil {
		t.Fatal(err)
	}
	if len(recipes) != len(recipeSeeds) {
		t.Fatalf("resolved %d recipes, want %d", len(recipes), len(recipeSeeds))
	}

	// Pick a currently-unknown recipe and flip it, asserting exactly that bit moves.
	var target Recipe
	for _, r := range recipes {
		if !r.Known {
			target = r
			break
		}
	}
	if target.Key == 0 {
		t.Skip("every recipe already known in this save")
	}
	if err := g.SetRecipeKnown(target.Key, true); err != nil {
		t.Fatal(err)
	}
	after, err := g.Recipes()
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range after {
		want := r.Known
		for _, b := range recipes {
			if b.Key == r.Key {
				want = b.Known
				break
			}
		}
		if r.Key == target.Key {
			want = true
		}
		if r.Known != want {
			t.Fatalf("recipe %d known=%v, want %v (only the target should change)", r.Key, r.Known, want)
		}
	}

	// And it round-trips back off.
	if err := g.SetRecipeKnown(target.Key, false); err != nil {
		t.Fatal(err)
	}
	back, err := g.Recipes()
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range back {
		if r.Key == target.Key && r.Known {
			t.Fatalf("recipe %d still known after locking", target.Key)
		}
	}
}

func TestUnlockAllRecipesCoversCat62(t *testing.T) {
	g := openGame(t)
	if err := g.UnlockAllRecipes(); err != nil {
		t.Fatal(err)
	}
	recipes, err := g.Recipes()
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range recipes {
		if !r.Known {
			t.Fatalf("recipe %d (cat %d bit %d) not known after UnlockAllRecipes", r.Key, r.Category, r.Bit)
		}
	}
}
