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
