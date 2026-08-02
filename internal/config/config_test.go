package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Missing file → zero config, no error.
	got, err := Load()
	if err != nil {
		t.Fatalf("Load on empty: %v", err)
	}
	if got.GameDir != "" {
		t.Fatalf("expected empty GameDir, got %q", got.GameDir)
	}

	want := Config{GameDir: "/games/DragonSword"}
	if err := want.Store(); err != nil {
		t.Fatalf("Store: %v", err)
	}
	if p := Path(); filepath.Dir(p) != filepath.Join(dir, "dsa-save-editor") {
		t.Fatalf("unexpected config path %q", p)
	}

	got, err = Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.GameDir != want.GameDir {
		t.Fatalf("round-trip: got %q want %q", got.GameDir, want.GameDir)
	}
}

func TestStoreAtomicNoTempLeft(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	if err := (Config{GameDir: "x"}).Store(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(Path() + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("temp file left behind")
	}
}
