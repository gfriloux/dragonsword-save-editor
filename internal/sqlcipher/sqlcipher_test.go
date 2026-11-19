package sqlcipher

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"
)

func TestDerivedPassphrase(t *testing.T) {
	if got := DefaultPassphrase(); got != Passphrase {
		t.Fatalf("derived passphrase = %q, want %q", got, Passphrase)
	}
}

// TestKeyDerivationVector checks deriveKeys against an independently computed
// test vector (Python hashlib.pbkdf2_hmac, not derived from this code) for a
// fixed passphrase and all-zero salt. Unlike a pure encrypt/decrypt round
// trip, this catches a bug that breaks deriveKeys symmetrically on both the
// encrypt and decrypt side (which a self-consistent round trip would miss).
func TestKeyDerivationVector(t *testing.T) {
	salt := make([]byte, SaltSize) // all zero
	wantEncKey := mustHex(t, "0d1bb8c88399bb1eb9c4a64948dcf0d5f5d05ac6b6cc0283ce04379bbb0d4a92")
	wantHMACKey := mustHex(t, "15e65464e0e684e5065465ae40cb85e0a1d5f5446f9a496c02f613bbf9791aaf")

	encKey, hmacKey, err := deriveKeys(Passphrase, salt)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encKey, wantEncKey) {
		t.Fatalf("encKey = %x, want %x", encKey, wantEncKey)
	}
	if !bytes.Equal(hmacKey, wantHMACKey) {
		t.Fatalf("hmacKey = %x, want %x", hmacKey, wantHMACKey)
	}
}

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestEncryptDecryptRoundTrip exercises the whole page-level cipher pipeline
// (PBKDF2 key derivation, AES-256-CBC, per-page HMAC-SHA512, salt placement and
// the reserved-bytes layout) against a synthetic two-page plaintext, so the
// core round-trip is verified by `go test ./...` with no external save file
// or DSA_SAVE env var required. TestDecryptRealSave above stays as an opt-in
// extra check against a real save for maintainers.
func TestEncryptDecryptRoundTrip(t *testing.T) {
	const nPages = 2
	plain := make([]byte, nPages*PageSize)
	// Page 0 starts with the SQLite magic — Decrypt always restores it there,
	// so seeding it here lets a plain bytes.Equal prove a true round trip.
	copy(plain, sqliteMagic)
	for p := 0; p < nPages; p++ {
		page := plain[p*PageSize : (p+1)*PageSize]
		start := 0
		if p == 0 {
			start = SaltSize
		}
		// Fill the page's content region with non-trivial bytes; leave the
		// trailing Reserve bytes zero since Decrypt always zeroes them.
		if _, err := rand.Read(page[start : PageSize-Reserve]); err != nil {
			t.Fatal(err)
		}
	}

	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		t.Fatal(err)
	}

	enc, err := Encrypt(plain, Passphrase, salt)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if len(enc) != len(plain) {
		t.Fatalf("encrypted size %d != plaintext size %d", len(enc), len(plain))
	}
	if gotSalt, err := Salt(enc); err != nil || !bytes.Equal(gotSalt, salt) {
		t.Fatalf("Salt(enc) = %x, %v; want %x, nil", gotSalt, err, salt)
	}

	got, err := Decrypt(enc, Passphrase)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatal("round-trip plaintext mismatch")
	}

	// A wrong passphrase must fail HMAC verification rather than silently
	// returning garbage.
	if _, err := Decrypt(enc, Passphrase+"x"); err == nil {
		t.Fatal("decrypt with wrong passphrase: expected HMAC mismatch error, got nil")
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
