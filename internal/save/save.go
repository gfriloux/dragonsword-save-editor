// Package save opens, edits and writes back DragonSword Awakening save files.
//
// A save is decrypted into a temporary plaintext SQLite database that is edited
// with the pure-Go modernc.org/sqlite driver, then re-encrypted (reusing the
// original salt) when persisted. A timestamped backup of the original file is
// created the first time a save is written.
package save

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gfriloux/dragonswordawakening/internal/sqlcipher"
	_ "modernc.org/sqlite"
)

// Save is an open save file being edited.
type Save struct {
	db     *sql.DB
	path   string // original encrypted file
	salt   []byte
	tmp    string // temporary plaintext file
	backup bool   // whether a backup has already been made
}

// Open decrypts path and returns an editable Save. The passphrase defaults to the
// game's constant key when empty.
func Open(path, passphrase string) (*Save, error) {
	if passphrase == "" {
		passphrase = sqlcipher.Passphrase
	}
	enc, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	salt, err := sqlcipher.Salt(enc)
	if err != nil {
		return nil, err
	}
	plain, err := sqlcipher.Decrypt(enc, passphrase)
	if err != nil {
		return nil, err
	}
	tmp, err := os.CreateTemp("", "dsa-save-*.db")
	if err != nil {
		return nil, err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(plain); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return nil, err
	}
	tmp.Close()

	// Rollback journal keeps the whole DB in the single file after each commit,
	// so we can read it back for re-encryption without closing the connection.
	db, err := sql.Open("sqlite", "file:"+tmpPath+"?_pragma=journal_mode(DELETE)&_pragma=foreign_keys(0)")
	if err != nil {
		os.Remove(tmpPath)
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		os.Remove(tmpPath)
		return nil, err
	}
	return &Save{db: db, path: path, salt: salt, tmp: tmpPath}, nil
}

// DB exposes the underlying plaintext database for queries and edits.
func (s *Save) DB() *sql.DB { return s.db }

// Path returns the original encrypted file path.
func (s *Save) Path() string { return s.path }

// Save re-encrypts the current database state and writes it back to the original
// path, creating a one-time timestamped backup of the original file first.
func (s *Save) Save() error {
	if err := s.checkpoint(); err != nil {
		return err
	}
	plain, err := os.ReadFile(s.tmp)
	if err != nil {
		return err
	}
	enc, err := sqlcipher.Encrypt(plain, sqlcipher.Passphrase, s.salt)
	if err != nil {
		return err
	}
	if !s.backup {
		if err := s.makeBackup(); err != nil {
			return err
		}
		s.backup = true
	}
	return atomicWrite(s.path, enc)
}

// checkpoint forces SQLite to flush committed pages to the temp file.
func (s *Save) checkpoint() error {
	// A no-op transaction commit ensures the main db file is up to date under the
	// rollback journal; PRAGMA wal_checkpoint is harmless if WAL were ever used.
	if _, err := s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE);"); err != nil {
		// ignore: not in WAL mode
	}
	return nil
}

func (s *Save) makeBackup() error {
	ts := time.Now().Format("20060102-150405")
	bak := fmt.Sprintf("%s.%s.bak", s.path, ts)
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return os.WriteFile(bak, data, 0o644)
}

// BackupPath returns where the next backup would be written (informational).
func (s *Save) BackupPath() string {
	return s.path + ".<timestamp>.bak"
}

// Close releases the database and removes the temporary plaintext file.
func (s *Save) Close() error {
	var err error
	if s.db != nil {
		err = s.db.Close()
		s.db = nil
	}
	if s.tmp != "" {
		os.Remove(s.tmp)
		os.Remove(s.tmp + "-journal")
		os.Remove(s.tmp + "-wal")
		os.Remove(s.tmp + "-shm")
		s.tmp = ""
	}
	return err
}

func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".dsa-write-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}
