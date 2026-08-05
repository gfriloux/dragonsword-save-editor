package domain

import (
	"database/sql"
	"fmt"
)

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

// UnlockVehicle adds an owned vehicle for cid, minting a fresh DBID. It is a
// no-op if already owned. CREATED_DATE falls back to the game's column default.
func (g *Game) UnlockVehicle(cid int64) error {
	if g.cat.Lookup(cid).Category != "mount" {
		return fmt.Errorf("domain: CID %d is not a vehicle", cid)
	}
	uid, err := g.UserID()
	if err != nil {
		return err
	}
	var n int
	if err := g.s.DB().QueryRow(
		`SELECT COUNT(*) FROM tb_vehicle WHERE USER_DBID=? AND VEHICLE_CID=?`, uid, cid).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil // already owned
	}
	dbid, err := g.mintDBID("tb_vehicle", "VEHICLE_DBID")
	if err != nil {
		return err
	}
	_, err = g.s.Exec(
		`INSERT INTO tb_vehicle (VEHICLE_DBID, USER_DBID, VEHICLE_CID) VALUES (?,?,?)`,
		dbid, uid, cid)
	return err
}

// SetMount equips vehicleDBID (an owned vehicle) on a character; vehicleDBID 0
// removes the mount. The character's equip-mount row is created if absent
// (other accessory columns keep their 0 defaults), or updated in place so the
// equipped accessories are preserved.
func (g *Game) SetMount(characterCID, vehicleDBID int64) error {
	uid, err := g.UserID()
	if err != nil {
		return err
	}
	if vehicleDBID != 0 {
		var n int
		if err := g.s.DB().QueryRow(
			`SELECT COUNT(*) FROM tb_vehicle WHERE USER_DBID=? AND VEHICLE_DBID=?`, uid, vehicleDBID).Scan(&n); err != nil {
			return err
		}
		if n == 0 {
			return fmt.Errorf("domain: vehicle DBID %d is not owned", vehicleDBID)
		}
	}
	_, err = g.s.Exec(
		`INSERT INTO tb_equip_mount (USER_DBID, CHARACTER_CID, VEHICLE) VALUES (?,?,?)
		 ON CONFLICT(USER_DBID, CHARACTER_CID) DO UPDATE SET VEHICLE=excluded.VEHICLE`,
		uid, characterCID, vehicleDBID)
	return err
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
