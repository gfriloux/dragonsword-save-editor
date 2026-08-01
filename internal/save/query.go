package save

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

// Column describes a table column.
type Column struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	NotNull      bool   `json:"notNull"`
	DefaultValue string `json:"default"`
	PrimaryKey   bool   `json:"primaryKey"`
}

// Tables returns the user tables in the save, sorted alphabetically.
func (s *Save) Tables() ([]string, error) {
	rows, err := s.db.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// Columns returns column metadata for a table.
func (s *Save) Columns(table string) ([]Column, error) {
	if !validIdent(table) {
		return nil, fmt.Errorf("invalid table name %q", table)
	}
	rows, err := s.db.Query(fmt.Sprintf("PRAGMA table_info(%s)", quoteIdent(table)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Column
	for rows.Next() {
		var (
			cid     int
			name    string
			ctype   string
			notnull int
			dflt    sql.NullString
			pk      int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		out = append(out, Column{
			Name:         name,
			Type:         ctype,
			NotNull:      notnull != 0,
			DefaultValue: dflt.String,
			PrimaryKey:   pk != 0,
		})
	}
	return out, rows.Err()
}

// RowCount returns the number of rows in a table.
func (s *Save) RowCount(table string) (int, error) {
	if !validIdent(table) {
		return 0, fmt.Errorf("invalid table name %q", table)
	}
	var n int
	err := s.db.QueryRow(fmt.Sprintf("SELECT count(*) FROM %s", quoteIdent(table))).Scan(&n)
	return n, err
}

// Rows returns up to limit rows (offset for paging) as generic values. Values are
// normalised to string/float64/int64/nil so they serialise cleanly to JSON.
func (s *Save) Rows(table string, limit, offset int) (cols []string, data [][]any, err error) {
	if !validIdent(table) {
		return nil, nil, fmt.Errorf("invalid table name %q", table)
	}
	if limit <= 0 {
		limit = 200
	}
	q := fmt.Sprintf("SELECT * FROM %s LIMIT ? OFFSET ?", quoteIdent(table))
	rows, err := s.db.Query(q, limit, offset)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	cols, err = rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, nil, err
		}
		for i, v := range vals {
			if b, ok := v.([]byte); ok {
				vals[i] = string(b)
			}
		}
		data = append(data, vals)
	}
	return cols, data, rows.Err()
}

// UpdateCell sets column col to value on the row identified by the primary-key
// map pk. It requires the update to affect exactly one row.
func (s *Save) UpdateCell(table string, pk map[string]any, col string, value any) error {
	if !validIdent(table) || !validIdent(col) {
		return fmt.Errorf("invalid identifier")
	}
	if len(pk) == 0 {
		return fmt.Errorf("no primary key given")
	}
	keys := make([]string, 0, len(pk))
	for k := range pk {
		if !validIdent(k) {
			return fmt.Errorf("invalid key column %q", k)
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	where := make([]string, len(keys))
	args := []any{value}
	for i, k := range keys {
		where[i] = quoteIdent(k) + " = ?"
	}
	for _, k := range keys {
		args = append(args, pk[k])
	}
	q := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s",
		quoteIdent(table), quoteIdent(col), strings.Join(where, " AND "))
	res, err := s.db.Exec(q, args...)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("update affected %d rows, expected 1", n)
	}
	return nil
}

// Exec runs an arbitrary statement (used by higher-level helpers and power users).
func (s *Save) Exec(query string, args ...any) (sql.Result, error) {
	return s.db.Exec(query, args...)
}

func validIdent(s string) bool {
	if s == "" || len(s) > 128 {
		return false
	}
	for _, r := range s {
		if !(r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
