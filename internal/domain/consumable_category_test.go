package domain

import "testing"

func TestClassifyConsumable(t *testing.T) {
	cases := []struct {
		cid  int64
		want string
		note string
	}{
		// Mana upgrade mats sit inside the 141x range but are NOT potions.
		{1410202, CatEquipXP, "Fragment de mana"},
		{1410203, CatEquipXP, "Cristal de mana"},
		{1410204, CatEquipXP, "Minéral de mana"},
		// Real potions, on either side of the mana block.
		{1410002, CatPotion, "potion below mana"},
		{1410105, CatPotion, "potion below mana"},
		{1410201, CatPotion, "141x just below mana → potion"},
		{1410205, CatPotion, "141x just above mana → potion"},
		// Cooked dishes.
		{1420102, CatCooked, "Poisson grillé"},
		{1422314, CatCooked, "Rouleaux de printemps"},
		// Ingredients (143x/144x).
		{1430003, CatIngredient, "Viande de bête"},
		{1440802, CatIngredient, "Aged Game Meat"},
		// Breakthrough: monster parts and plants only.
		{1450001, CatBreakthrough, "Essence de monstre"},
		{1450018, CatBreakthrough, "Griffe de monstre (last)"},
		{1450003, CatBreakthrough, "Essence de monstre intermédiaire (percée)"},
		{1450501, CatBreakthrough, "Fruit de la vitalité"},
		{1450504, CatBreakthrough, "Feuille de vigueur"},
		// Equipment-craft stones (whole 14501xx block) + entremêlés + Cristal de veine.
		{1450101, CatCraft, "Pierre d'amplification"},
		{1450104, CatCraft, "Pierre de concassage"},
		{1450122, CatCraft, "Pierre de foudre"},
		{1450127, CatCraft, "Pierre trouble (last stone)"},
		{1450820, CatCraft, "Souvenirs et mémoires entremêlés"},
		{1450821, CatCraft, "Oubli et réminiscences entremêlés"},
		{1460103, CatCraft, "Cristal de veine"},
		// The rest of the mixed 145x block stays unsorted.
		{1450202, CatUnsorted, "Insigne d'Orbis (faction badge)"},
		{1450401, CatUnsorted, "Grimoire de la perspicacité"},
		{1450601, CatUnsorted, "Cristal de la mémoire (reminiscence)"},
		// Skill-upgrade materials (confirmed in-game), curated explicitly.
		{1450811, CatSkill, "Chronique de combat"},
		{1450812, CatSkill, "Grimoire du guerrier"},
		{1450815, CatSkill, "Nageoire épineuse"},
		{1450816, CatSkill, "Crochet ensanglanté"},
		{1450823, CatSkill, "Sang du sage"},
		// Awakening material (off-th.gl, identified by count in a real save).
		{1450410, CatAwakening, "Gemme d'éveil"},
		{1450409, CatUnsorted, "neighbour of the awakening gem, unconfirmed"},
		// Exchange token: buy characters.
		{1000500, CatExchange, "Invitation du Destin"},
		{1000501, CatUnsorted, "Fragment de lumière sacrée (unconfirmed)"},
		// Crystals and runes.
		{1460101, CatCrystal, "Cristal de cohésion"},
		{1310001, CatRune, "Damaged Rune of Determination"},
		// Off-th.gl XP-book candidates stay unsorted (unverified).
		{1000800, CatUnsorted, "XP-book candidate, unverified"},
		{1000802, CatUnsorted, "XP-book candidate, unverified"},
		// Other unknowns.
		{1470103, CatUnsorted, "147x"},
		{1510018, CatUnsorted, "151x"},
	}
	for _, c := range cases {
		if got := ClassifyConsumable(c.cid); got != c.want {
			t.Errorf("ClassifyConsumable(%d) [%s] = %q, want %q", c.cid, c.note, got, c.want)
		}
	}
}

func TestConsumableCategoriesShape(t *testing.T) {
	cats := ConsumableCategories()
	if len(cats) == 0 {
		t.Fatal("no categories")
	}
	if cats[len(cats)-1].Key != CatUnsorted {
		t.Errorf("last category = %q, want %q (unsorted is always last)", cats[len(cats)-1].Key, CatUnsorted)
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
