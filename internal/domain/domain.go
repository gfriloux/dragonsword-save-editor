package domain

import (
	"database/sql"
	"fmt"

	"github.com/gfriloux/dragonsword-save-editor/internal/save"
)

// Game is a typed, game-oriented view over an open save.
type Game struct {
	s      *save.Save
	cat    *Catalog
	userID int64
	uidSet bool
}

// New wraps a save with a catalog.
func New(s *save.Save, cat *Catalog) *Game {
	return &Game{s: s, cat: cat}
}

// Catalog exposes the underlying item catalog (for label edits).
func (g *Game) Catalog() *Catalog { return g.cat }

// Save exposes the underlying save (for the generic database view).
func (g *Game) Save() *save.Save { return g.s }

// UserID returns the (single) account id stored in the save, cached.
func (g *Game) UserID() (int64, error) {
	if g.uidSet {
		return g.userID, nil
	}
	var id int64
	err := g.s.DB().QueryRow(`SELECT USER_DBID FROM tb_user LIMIT 1`).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("domain: reading USER_DBID: %w", err)
	}
	g.userID, g.uidSet = id, true
	return id, nil
}

// Currency is a resolved currency row.
type Currency struct {
	Item
	Amount int64 `json:"amount"`
}

// Currencies lists the account currencies, resolved through the catalog.
func (g *Game) Currencies() ([]Currency, error) {
	uid, err := g.UserID()
	if err != nil {
		return nil, err
	}
	rows, err := g.s.DB().Query(
		`SELECT ITEM_CID, AMOUNT FROM tb_currency WHERE USER_DBID=? ORDER BY ITEM_CID`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Currency
	for rows.Next() {
		var cid, amount int64
		if err := rows.Scan(&cid, &amount); err != nil {
			return nil, err
		}
		out = append(out, Currency{Item: g.cat.LookupCtx(cid, "currency"), Amount: amount})
	}
	return out, rows.Err()
}

// SetCurrency sets the amount of a currency. Requires the row to exist.
func (g *Game) SetCurrency(cid, amount int64) error {
	uid, err := g.UserID()
	if err != nil {
		return err
	}
	return exactlyOne(g.s.Exec(
		`UPDATE tb_currency SET AMOUNT=? WHERE USER_DBID=? AND ITEM_CID=?`, amount, uid, cid))
}

// Kinds of consumable rows, used to route edits to the right table/key.
const (
	KindStackable = "stackable" // tb_stackable_item, keyed by ITEM_CID
	KindCook      = "cook"      // tb_cook_item, keyed by ITEM_DBID (per-instance)
)

// Stack is a resolved consumable/stackable row.
type Stack struct {
	Item
	Kind  string `json:"kind"`
	ID    int64  `json:"id"` // ITEM_CID for stackable, ITEM_DBID for cook
	Count int64  `json:"count"`
}

// Consumables lists stackable items and (non-deleted) cooked food, resolved and
// tagged by category so the UI can group them.
func (g *Game) Consumables() ([]Stack, error) {
	uid, err := g.UserID()
	if err != nil {
		return nil, err
	}
	var out []Stack

	rows, err := g.s.DB().Query(
		`SELECT ITEM_CID, STACK_CNT FROM tb_stackable_item WHERE USER_DBID=? ORDER BY ITEM_CID`, uid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var cid, cnt int64
		if err := rows.Scan(&cid, &cnt); err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, Stack{Item: g.cat.Lookup(cid), Kind: KindStackable, ID: cid, Count: cnt})
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	crows, err := g.s.DB().Query(
		`SELECT ITEM_DBID, ITEM_CID, STACK_CNT FROM tb_cook_item WHERE USER_DBID=? AND DELETED_DATE=0 ORDER BY ITEM_CID`, uid)
	if err != nil {
		return nil, err
	}
	defer crows.Close()
	for crows.Next() {
		var dbid, cid, cnt int64
		if err := crows.Scan(&dbid, &cid, &cnt); err != nil {
			return nil, err
		}
		out = append(out, Stack{Item: g.cat.LookupCtx(cid, "food"), Kind: KindCook, ID: dbid, Count: cnt})
	}
	return out, crows.Err()
}

// SetStack sets the quantity of a consumable, routing by kind.
func (g *Game) SetStack(kind string, id, count int64) error {
	uid, err := g.UserID()
	if err != nil {
		return err
	}
	switch kind {
	case KindStackable:
		return exactlyOne(g.s.Exec(
			`UPDATE tb_stackable_item SET STACK_CNT=? WHERE USER_DBID=? AND ITEM_CID=?`, count, uid, id))
	case KindCook:
		return exactlyOne(g.s.Exec(
			`UPDATE tb_cook_item SET STACK_CNT=? WHERE USER_DBID=? AND ITEM_DBID=?`, count, uid, id))
	default:
		return fmt.Errorf("domain: unknown consumable kind %q", kind)
	}
}

// Character is a resolved, read-only character row.
type Character struct {
	Item
	Level        int64 `json:"level"`
	Exp          int64 `json:"exp"`
	Ascend       int64 `json:"ascend"`
	HP           int64 `json:"hp"`
	Transcend    int64 `json:"transcend"`
	SoldierGrade int64 `json:"soldierGrade"`
}

// Characters lists the owned characters (read-only).
func (g *Game) Characters() ([]Character, error) {
	uid, err := g.UserID()
	if err != nil {
		return nil, err
	}
	rows, err := g.s.DB().Query(
		`SELECT CHARACTER_CID, LEVEL, EXP, ASCEND, HP, TRANSCEND, SOLDIER_GRADE
		   FROM tb_character WHERE USER_DBID=? ORDER BY CHARACTER_CID`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Character
	for rows.Next() {
		var c Character
		var cid int64
		if err := rows.Scan(&cid, &c.Level, &c.Exp, &c.Ascend, &c.HP, &c.Transcend, &c.SoldierGrade); err != nil {
			return nil, err
		}
		c.Item = g.cat.LookupCtx(cid, "character")
		out = append(out, c)
	}
	return out, rows.Err()
}

// characterLevels maps character CID to level, for annotating team slots.
func (g *Game) characterLevels(uid int64) (map[int64]int64, error) {
	rows, err := g.s.DB().Query(`SELECT CHARACTER_CID, LEVEL FROM tb_character WHERE USER_DBID=?`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[int64]int64{}
	for rows.Next() {
		var cid, lvl int64
		if err := rows.Scan(&cid, &lvl); err != nil {
			return nil, err
		}
		m[cid] = lvl
	}
	return m, rows.Err()
}

// TeamSlot is one character slot of a team page (read-only).
type TeamSlot struct {
	Item
	Level int64 `json:"level"`
	Empty bool  `json:"empty"`
}

// TeamPage is a saved team composition (read-only).
type TeamPage struct {
	PageID int64      `json:"pageId"`
	Slots  []TeamSlot `json:"slots"`
}

// Teams lists the saved team pages with their three character slots (read-only).
func (g *Game) Teams() ([]TeamPage, error) {
	uid, err := g.UserID()
	if err != nil {
		return nil, err
	}
	levels, err := g.characterLevels(uid)
	if err != nil {
		return nil, err
	}
	rows, err := g.s.DB().Query(
		`SELECT PAGE_ID, SLOT1_CHARACTER_CID, SLOT2_CHARACTER_CID, SLOT3_CHARACTER_CID
		   FROM tb_team WHERE USER_DBID=? ORDER BY PAGE_ID`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TeamPage
	for rows.Next() {
		var page TeamPage
		var s [3]int64
		if err := rows.Scan(&page.PageID, &s[0], &s[1], &s[2]); err != nil {
			return nil, err
		}
		for _, cid := range s {
			if cid == 0 {
				page.Slots = append(page.Slots, TeamSlot{Empty: true})
				continue
			}
			page.Slots = append(page.Slots, TeamSlot{Item: g.cat.LookupCtx(cid, "character"), Level: levels[cid]})
		}
		out = append(out, page)
	}
	return out, rows.Err()
}

// Equipment is a resolved equipment row. Scalar fields are editable; stat CIDs
// are references shown read-only (their names are not yet decoded).
type Equipment struct {
	Item
	DBID         int64   `json:"dbid"`
	EnchantLevel int64   `json:"enchantLevel"`
	Exp          int64   `json:"exp"`
	IsLock       bool    `json:"isLock"`
	GemDBID      int64   `json:"gemDbid"`
	MainStatCID  int64   `json:"mainStatCid"`
	SubStatCIDs  []int64 `json:"subStatCids"`
}

// Equipments lists active (non-deleted) equipment.
func (g *Game) Equipments() ([]Equipment, error) {
	uid, err := g.UserID()
	if err != nil {
		return nil, err
	}
	rows, err := g.s.DB().Query(
		`SELECT ITEM_DBID, ITEM_CID, ENCHANT_LEVEL, EXP, IS_LOCK, GEM_DBID,
		        MAIN_STAT_CID, SUB_STAT_CID1, SUB_STAT_CID2, SUB_STAT_CID3, SUB_STAT_CID4, SUB_STAT_CID5
		   FROM tb_equipment WHERE USER_DBID=? AND DELETED_DATE=0 ORDER BY ITEM_CID`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Equipment
	for rows.Next() {
		var (
			e         Equipment
			cid, lock int64
			sub       [5]int64
		)
		if err := rows.Scan(&e.DBID, &cid, &e.EnchantLevel, &e.Exp, &lock, &e.GemDBID,
			&e.MainStatCID, &sub[0], &sub[1], &sub[2], &sub[3], &sub[4]); err != nil {
			return nil, err
		}
		e.Item = g.cat.LookupCtx(cid, "gear")
		e.IsLock = lock != 0
		for _, s := range sub {
			if s != 0 {
				e.SubStatCIDs = append(e.SubStatCIDs, s)
			}
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// SetEnchant sets an equipment's enchant level.
func (g *Game) SetEnchant(dbid, level int64) error {
	return g.setEquip(dbid, "ENCHANT_LEVEL", level)
}

// SetEquipExp sets an equipment's experience.
func (g *Game) SetEquipExp(dbid, exp int64) error {
	return g.setEquip(dbid, "EXP", exp)
}

// SetEquipLock locks or unlocks an equipment.
func (g *Game) SetEquipLock(dbid int64, locked bool) error {
	return g.setEquip(dbid, "IS_LOCK", boolToInt(locked))
}

func (g *Game) setEquip(dbid int64, column string, value int64) error {
	uid, err := g.UserID()
	if err != nil {
		return err
	}
	// column is a fixed internal constant, never user input.
	return exactlyOne(g.s.Exec(
		"UPDATE tb_equipment SET "+column+"=? WHERE USER_DBID=? AND ITEM_DBID=?", value, uid, dbid))
}

// Gem is a resolved gem row. Only the lock is editable.
type Gem struct {
	Item
	DBID        int64 `json:"dbid"`
	StatInfoCID int64 `json:"statInfoCid"`
	IsLock      bool  `json:"isLock"`
}

// Gems lists active (non-deleted) gems. May be empty.
func (g *Game) Gems() ([]Gem, error) {
	uid, err := g.UserID()
	if err != nil {
		return nil, err
	}
	rows, err := g.s.DB().Query(
		`SELECT ITEM_DBID, ITEM_CID, STAT_INFO_CID, IS_LOCK
		   FROM tb_gem WHERE USER_DBID=? AND DELETED_DATE=0 ORDER BY ITEM_CID`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Gem{}
	for rows.Next() {
		var gm Gem
		var cid, lock int64
		if err := rows.Scan(&gm.DBID, &cid, &gm.StatInfoCID, &lock); err != nil {
			return nil, err
		}
		gm.Item = g.cat.Lookup(cid)
		gm.IsLock = lock != 0
		out = append(out, gm)
	}
	return out, rows.Err()
}

// SetGemLock locks or unlocks a gem.
func (g *Game) SetGemLock(dbid int64, locked bool) error {
	uid, err := g.UserID()
	if err != nil {
		return err
	}
	return exactlyOne(g.s.Exec(
		`UPDATE tb_gem SET IS_LOCK=? WHERE USER_DBID=? AND ITEM_DBID=?`, boolToInt(locked), uid, dbid))
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func exactlyOne(res sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("domain: update affected %d rows, expected 1", n)
	}
	return nil
}
