package oodle

import "testing"

// TestNewClose checks the embedded ooz.wasm instantiates and exposes the
// expected exports on the pure-Go runtime (no game data needed).
func TestNewClose(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := d.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

// TestDecompressRejectsGarbage ensures a bogus block fails cleanly rather than
// panicking (the wasm decoder returns a negative size).
func TestDecompressRejectsGarbage(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	if _, err := d.Decompress([]byte{0, 1, 2, 3, 4, 5, 6, 7}, 4096); err == nil {
		t.Fatalf("expected an error on garbage input")
	}
}
