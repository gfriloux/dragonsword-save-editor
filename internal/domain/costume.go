package domain

// Costume is a resolved owned costume row. The 999xxxx "costume" space bundles
// both outfits and weapon skins (indistinguishable by type); they are listed
// flat. Equip state lives only here — a character wears a costume when its
// EquipCharacterCID holds that character's CID, and the relation is NOT
// exclusive (a character wears several rows at once, e.g. outfit + weapon skin).
type Costume struct {
	Item
	DBID              int64 `json:"dbid,string"` // instance id; string to stay uniform with other instances
	EquipCharacterCID int64 `json:"equipCharacterCid"`
	PartsOn           int64 `json:"partsOn"`
	IsNew             bool  `json:"isNew"`
}

// CatalogEntry is a catalog item annotated with whether the save already owns it.
// Shared by the costume and vehicle "unlock" pickers.
type CatalogEntry struct {
	Item
	Owned bool `json:"owned"`
}

// Costumes lists the owned costumes, resolved through the catalog.
func (g *Game) Costumes() ([]Costume, error) {
	uid, err := g.UserID()
	if err != nil {
		return nil, err
	}
	rows, err := g.s.DB().Query(
		`SELECT COSTUME_DBID, COSTUME_CID, EQUIP_CHARACTER_CID, PARTS_ON, IS_NEW
		   FROM tb_costume WHERE USER_DBID=? ORDER BY COSTUME_CID`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Costume{}
	for rows.Next() {
		var (
			c     Costume
			cid   int64
			isNew int64
		)
		if err := rows.Scan(&c.DBID, &cid, &c.EquipCharacterCID, &c.PartsOn, &isNew); err != nil {
			return nil, err
		}
		c.Item = g.cat.LookupCtx(cid, "costume")
		c.IsNew = isNew != 0
		out = append(out, c)
	}
	return out, rows.Err()
}

// CostumeCatalog lists every catalog costume with an owned flag, for the unlock
// picker.
func (g *Game) CostumeCatalog() ([]CatalogEntry, error) {
	return g.categoryCatalog("costume", "tb_costume", "COSTUME_CID")
}

// ownedCIDs returns the set of CIDs the save holds in table.col for this account.
func (g *Game) ownedCIDs(table, col string) (map[int64]bool, error) {
	uid, err := g.UserID()
	if err != nil {
		return nil, err
	}
	// table and col are fixed internal constants, never user input.
	rows, err := g.s.DB().Query("SELECT "+col+" FROM "+table+" WHERE USER_DBID=?", uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	owned := map[int64]bool{}
	for rows.Next() {
		var cid int64
		if err := rows.Scan(&cid); err != nil {
			return nil, err
		}
		owned[cid] = true
	}
	return owned, rows.Err()
}

// categoryCatalog lists every catalog item of the given category, flagged with
// whether the save already owns it (ownership read from table.col).
func (g *Game) categoryCatalog(category, table, col string) ([]CatalogEntry, error) {
	owned, err := g.ownedCIDs(table, col)
	if err != nil {
		return nil, err
	}
	out := []CatalogEntry{}
	for _, it := range g.cat.Entries() {
		if it.Category != category {
			continue
		}
		out = append(out, CatalogEntry{Item: it, Owned: owned[it.CID]})
	}
	return out, nil
}
