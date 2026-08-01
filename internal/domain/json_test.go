package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

// 64-bit ITEM_DBIDs exceed 2^53 and must cross the wire as JSON strings, or the
// browser's number parsing rounds them and edits target the wrong row.
func TestBigIDsMarshalAsStrings(t *testing.T) {
	const big = 73605708194892760 // real cook/equipment ITEM_DBID magnitude
	b, _ := json.Marshal(Stack{ID: big})
	if !strings.Contains(string(b), `"id":"73605708194892760"`) {
		t.Fatalf("Stack.ID not a JSON string: %s", b)
	}
	b, _ = json.Marshal(Equipment{DBID: big})
	if !strings.Contains(string(b), `"dbid":"73605708194892760"`) {
		t.Fatalf("Equipment.DBID not a JSON string: %s", b)
	}
	b, _ = json.Marshal(Gem{DBID: big})
	if !strings.Contains(string(b), `"dbid":"73605708194892760"`) {
		t.Fatalf("Gem.DBID not a JSON string: %s", b)
	}
}
