package domain

import "database/sql"

// Vehicle is a resolved owned vehicle (familier / monture) row.
type Vehicle struct {
	Item
	DBID int64 `json:"dbid,string"` // instance id; string to stay uniform with other instances
}

// Mount is a character's equipped mount, resolved. Vehicle is nil when the
// character has no mount (tb_equip_mount.VEHICLE = 0, or a dangling reference).
type Mount struct {
	CharacterCID int64 `json:"characterCid"`
	VehicleDBID  int64 `json:"vehicleDbid,string"` // 0 = none
	Vehicle      *Item `json:"vehicle"`
}

// Vehicles lists the owned vehicles (familiers), resolved through the catalog.
func (g *Game) Vehicles() ([]Vehicle, error) {
	uid, err := g.UserID()
	if err != nil {
		return nil, err
	}
	rows, err := g.s.DB().Query(
		`SELECT VEHICLE_DBID, VEHICLE_CID FROM tb_vehicle WHERE USER_DBID=? ORDER BY VEHICLE_CID`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Vehicle{}
	for rows.Next() {
		var (
			v   Vehicle
			cid int64
		)
		if err := rows.Scan(&v.DBID, &cid); err != nil {
			return nil, err
		}
		v.Item = g.cat.LookupCtx(cid, "mount")
		out = append(out, v)
	}
	return out, rows.Err()
}

// VehicleCatalog lists every catalog vehicle with an owned flag, for the unlock
// picker.
func (g *Game) VehicleCatalog() ([]CatalogEntry, error) {
	return g.categoryCatalog("mount", "tb_vehicle", "VEHICLE_CID")
}

// Mounts lists, per character that has an equip-mount row, which vehicle it
// rides (resolved), or none.
func (g *Game) Mounts() ([]Mount, error) {
	uid, err := g.UserID()
	if err != nil {
		return nil, err
	}
	rows, err := g.s.DB().Query(
		`SELECT em.CHARACTER_CID, em.VEHICLE, v.VEHICLE_CID
		   FROM tb_equip_mount em
		   LEFT JOIN tb_vehicle v
		     ON v.VEHICLE_DBID = em.VEHICLE AND v.USER_DBID = em.USER_DBID
		  WHERE em.USER_DBID=? ORDER BY em.CHARACTER_CID`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Mount{}
	for rows.Next() {
		var (
			m    Mount
			vcid sql.NullInt64
		)
		if err := rows.Scan(&m.CharacterCID, &m.VehicleDBID, &vcid); err != nil {
			return nil, err
		}
		if m.VehicleDBID != 0 && vcid.Valid {
			it := g.cat.LookupCtx(vcid.Int64, "mount")
			m.Vehicle = &it
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
