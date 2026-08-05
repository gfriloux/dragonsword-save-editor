package domain

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

// randID draws a non-zero random 32-bit instance id. Costume and vehicle DBIDs
// observed in real saves fit in a uint32 and look random; 0 is reserved for the
// "none/unset" meaning, so we never return it.
func randID() (int64, error) {
	var b [4]byte
	for {
		if _, err := rand.Read(b[:]); err != nil {
			return 0, fmt.Errorf("domain: draw dbid: %w", err)
		}
		if id := int64(binary.BigEndian.Uint32(b[:])); id != 0 {
			return id, nil
		}
	}
}

// mintIDFrom returns the first drawn id that exists reports as absent. draw and
// exists are seams so the retry loop is testable without RNG or a database.
func mintIDFrom(draw func() (int64, error), exists func(int64) (bool, error)) (int64, error) {
	for tries := 0; tries < 1000; tries++ {
		id, err := draw()
		if err != nil {
			return 0, err
		}
		taken, err := exists(id)
		if err != nil {
			return 0, err
		}
		if !taken {
			return id, nil
		}
	}
	return 0, fmt.Errorf("domain: mint dbid: exhausted attempts")
}

// mintID forges a random id absent according to exists.
func mintID(exists func(int64) (bool, error)) (int64, error) {
	return mintIDFrom(randID, exists)
}

// mintDBID forges a unique instance id for table.pkCol, absent from the save.
// table and pkCol are fixed internal constants, never user input.
func (g *Game) mintDBID(table, pkCol string) (int64, error) {
	return mintID(func(id int64) (bool, error) {
		var n int
		if err := g.s.DB().QueryRow(
			"SELECT COUNT(*) FROM "+table+" WHERE "+pkCol+"=?", id).Scan(&n); err != nil {
			return false, fmt.Errorf("domain: mint dbid %s.%s: %w", table, pkCol, err)
		}
		return n != 0, nil
	})
}
