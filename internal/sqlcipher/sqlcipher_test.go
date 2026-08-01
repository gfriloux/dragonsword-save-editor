package sqlcipher

import (
	"bytes"
	"os"
	"testing"
)

func TestDerivedPassphrase(t *testing.T) {
	if got := DefaultPassphrase(); got != Passphrase {
		t.Fatalf("derived passphrase = %q, want %q", got, Passphrase)
	}
}

// TestDecryptRealSave decrypts a real save file when DSA_SAVE points at one.
// It verifies the plaintext is a valid SQLite database and that a full
// encrypt/decrypt round-trip reproduces the same plaintext.
func TestDecryptRealSave(t *testing.T) {
	path := os.Getenv("DSA_SAVE")
	if path == "" {
		t.Skip("set DSA_SAVE to a real .db to run this test")
	}
	enc, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := Decrypt(enc, Passphrase)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.HasPrefix(plain, sqliteMagic) {
		t.Fatalf("plaintext does not start with SQLite magic: %x", plain[:16])
	}
	if plain[20] != Reserve {
		t.Fatalf("reserved-bytes-per-page = %d, want %d", plain[20], Reserve)
	}

	salt, err := Salt(enc)
	if err != nil {
		t.Fatal(err)
	}
	reEnc, err := Encrypt(plain, Passphrase, salt)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	plain2, err := Decrypt(reEnc, Passphrase)
	if err != nil {
		t.Fatalf("re-decrypt: %v", err)
	}
	if !bytes.Equal(plain, plain2) {
		t.Fatal("round-trip plaintext mismatch")
	}
	if len(reEnc) != len(enc) {
		t.Fatalf("re-encrypted size %d != original %d", len(reEnc), len(enc))
	}
	if !bytes.Equal(reEnc[:SaltSize], enc[:SaltSize]) {
		t.Fatal("salt not preserved")
	}
	if out := os.Getenv("DSA_OUT"); out != "" {
		if err := os.WriteFile(out, reEnc, 0o644); err != nil {
			t.Fatal(err)
		}
		if plainOut := os.Getenv("DSA_PLAIN_OUT"); plainOut != "" {
			if err := os.WriteFile(plainOut, plain, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
}
