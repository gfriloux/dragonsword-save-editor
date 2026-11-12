package sqlcipher

import (
	"bytes"
	"encoding/hex"
	"math/rand"
	"os"
	"testing"
)

func TestDerivedPassphrase(t *testing.T) {
	if got := DefaultPassphrase(); got != Passphrase {
		t.Fatalf("derived passphrase = %q, want %q", got, Passphrase)
	}
}

// TestDeriveKeysVector pins deriveKeys' output for a fixed passphrase/salt
// against a hardcoded hex vector. Unlike the round-trip test below, this
// catches a regression that stays internally self-consistent (e.g. a broken
// HMAC-salt mask applied identically on both the encrypt and decrypt path) —
// see https://github.com/gfriloux/dragonsword-save-editor/issues/2.
func TestDeriveKeysVector(t *testing.T) {
	const (
		wantEncKey  = "48357a0ff9622c0400977be64fd90bf25eb125fe39a078feb7aeb533d20a7ffe"
		wantHMACKey = "e66b5f4841e8db0d54d5d680a40255a05843a4a392c32cbb0d7a1aba087d0b4f"
	)
	salt := bytes.Repeat([]byte{0x5a}, SaltSize)
	encKey, hmacKey, err := deriveKeys(Passphrase, salt)
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(encKey); got != wantEncKey {
		t.Fatalf("encKey = %s, want %s", got, wantEncKey)
	}
	if got := hex.EncodeToString(hmacKey); got != wantHMACKey {
		t.Fatalf("hmacKey = %s, want %s", got, wantHMACKey)
	}
}

// syntheticPlaintext builds a deterministic, self-consistent multi-page
// plaintext for the round-trip test: page 0 starts with the standard SQLite
// magic header (Decrypt always writes that back in, so it must be present for
// a byte-exact round trip) and the per-page reserved bytes are left zero
// (Decrypt always zeroes them; Encrypt ignores whatever is there).
func syntheticPlaintext(nPages int) []byte {
	out := make([]byte, nPages*PageSize)
	rng := rand.New(rand.NewSource(42))
	for p := 0; p < nPages; p++ {
		page := out[p*PageSize : (p+1)*PageSize]
		start := 0
		if p == 0 {
			copy(page[:SaltSize], sqliteMagic)
			start = SaltSize
		}
		content := page[start : PageSize-Reserve]
		for i := range content {
			content[i] = byte(rng.Intn(256))
		}
		// page[PageSize-Reserve:] stays zero, matching Decrypt's output.
	}
	return out
}

// TestEncryptDecryptRoundTrip exercises Decrypt/Encrypt on a synthetic
// in-memory fixture (no real save file / DSA_SAVE needed), so it always runs
// under plain `go test ./...`. It catches regressions in the AES-CBC, PBKDF2
// key derivation, per-page HMAC, salt placement, or reserved-byte layout —
// see https://github.com/gfriloux/dragonsword-save-editor/issues/2.
func TestEncryptDecryptRoundTrip(t *testing.T) {
	plain := syntheticPlaintext(3)
	salt := bytes.Repeat([]byte{0x5a}, SaltSize)

	enc, err := Encrypt(plain, Passphrase, salt)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if len(enc) != len(plain) {
		t.Fatalf("encrypted size %d != plaintext size %d", len(enc), len(plain))
	}
	if !bytes.Equal(enc[:SaltSize], salt) {
		t.Fatal("salt not placed at the start of page 1")
	}

	gotSalt, err := Salt(enc)
	if err != nil {
		t.Fatalf("salt: %v", err)
	}
	if !bytes.Equal(gotSalt, salt) {
		t.Fatal("Salt() did not recover the salt written by Encrypt")
	}

	decrypted, err := Decrypt(enc, Passphrase)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plain) {
		t.Fatal("round-trip plaintext mismatch: Decrypt(Encrypt(plain)) != plain")
	}

	// Wrong passphrase must fail HMAC verification, not silently succeed.
	if _, err := Decrypt(enc, Passphrase+"x"); err == nil {
		t.Fatal("decrypt with wrong passphrase unexpectedly succeeded")
	}

	// A single flipped ciphertext byte must be caught by the per-page HMAC.
	tampered := bytes.Clone(enc)
	tampered[SaltSize] ^= 0x01
	if _, err := Decrypt(tampered, Passphrase); err == nil {
		t.Fatal("decrypt of tampered ciphertext unexpectedly succeeded")
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
