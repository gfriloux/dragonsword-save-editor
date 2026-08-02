// Package config persists the editor's small user settings (chiefly the game
// folder) in the OS config directory, next to the item-label overrides.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config is the persisted user configuration.
type Config struct {
	// GameDir is the remembered game installation folder (the one holding
	// DS/Content/Paks and DS/Saved/SaveGames). Empty until the user picks it.
	GameDir string `json:"gameDir"`
}

// Path returns the config file location (…/dsa-save-editor/config.json), or ""
// if the OS config directory cannot be determined.
func Path() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "dsa-save-editor", "config.json")
}

// Load reads the config file. A missing file is not an error: it returns a zero
// Config so first-run is indistinguishable from "nothing set yet".
func Load() (Config, error) {
	var c Config
	p := Path()
	if p == "" {
		return c, nil
	}
	b, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return c, err
	}
	return c, nil
}

// Store writes the config file, creating its directory if needed. It writes
// atomically (temp file + rename) so a crash never leaves a truncated config.
func (c Config) Store() error {
	p := Path()
	if p == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}
