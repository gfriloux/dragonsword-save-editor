// Package sqlcipher implements a pure-Go reader/writer for the SQLCipher v4
// database format as used by DragonSword Awakening save files.
//
// It intentionally supports only the exact configuration the game uses
// (cipher_compatibility = 4): AES-256-CBC, 4096-byte pages, PBKDF2-HMAC-SHA512
// key derivation with 256000 iterations, and per-page HMAC-SHA512.
//
// Because it is pure Go (only the standard library), it cross-compiles to any
// GOOS/GOARCH with no CGO.
package sqlcipher

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
)

// Parameters fixed by cipher_compatibility = 4.
const (
	PageSize     = 4096
	SaltSize     = 16
	IVSize       = 16
	HMACSize     = 64 // SHA-512 digest
	KeySize      = 32 // AES-256
	KDFIter      = 256000
	FastKDFIter  = 2
	HMACSaltMask = 0x3a
	// Reserve = IV + HMAC, rounded up to the AES block size (already a multiple).
	Reserve = IVSize + HMACSize // 80
)

// sqliteMagic is the fixed 16-byte header of any SQLite file. SQLCipher stores
// the random salt in its place on page 1, so we restore it on decrypt and drop
// it (replacing it with the salt) on encrypt.
var sqliteMagic = []byte("SQLite format 3\x00")

func deriveKeys(passphrase string, salt []byte) (encKey, hmacKey []byte, err error) {
	encKey, err = pbkdf2.Key(sha512.New, passphrase, salt, KDFIter, KeySize)
	if err != nil {
		return nil, nil, err
	}
	hmacSalt := make([]byte, len(salt))
	for i, b := range salt {
		hmacSalt[i] = b ^ HMACSaltMask
	}
	// The HMAC key is derived from the *encryption key bytes* (not the passphrase).
	hmacKey, err = pbkdf2.Key(sha512.New, string(encKey), hmacSalt, FastKDFIter, KeySize)
	if err != nil {
		return nil, nil, err
	}
	return encKey, hmacKey, nil
}

func pageHMAC(hmacKey, cipherText, iv []byte, pgno uint32) []byte {
	mac := hmac.New(sha512.New, hmacKey)
	mac.Write(cipherText)
	mac.Write(iv)
	var no [4]byte
	binary.LittleEndian.PutUint32(no[:], pgno) // game runs on little-endian x86
	mac.Write(no[:])
	return mac.Sum(nil)
}

// Decrypt turns a SQLCipher v4 database into a plaintext SQLite database.
// The returned bytes start with the standard "SQLite format 3" header and can be
// opened by any SQLite implementation (each page keeps 80 reserved bytes).
func Decrypt(enc []byte, passphrase string) ([]byte, error) {
	if len(enc) == 0 || len(enc)%PageSize != 0 {
		return nil, fmt.Errorf("sqlcipher: size %d is not a multiple of page size %d", len(enc), PageSize)
	}
	salt := enc[:SaltSize]
	encKey, hmacKey, err := deriveKeys(passphrase, salt)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, err
	}

	nPages := len(enc) / PageSize
	out := make([]byte, len(enc))
	reserveStart := PageSize - Reserve

	for p := 0; p < nPages; p++ {
		pgno := uint32(p + 1)
		page := enc[p*PageSize : (p+1)*PageSize]
		encStart := 0
		if p == 0 {
			encStart = SaltSize
		}
		cipherText := page[encStart:reserveStart]
		iv := page[reserveStart : reserveStart+IVSize]
		storedMAC := page[reserveStart+IVSize : reserveStart+IVSize+HMACSize]

		if !hmac.Equal(storedMAC, pageHMAC(hmacKey, cipherText, iv, pgno)) {
			return nil, fmt.Errorf("sqlcipher: HMAC mismatch on page %d (wrong key or corrupt file)", pgno)
		}

		plain := make([]byte, len(cipherText))
		cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, cipherText)

		dst := out[p*PageSize : (p+1)*PageSize]
		if p == 0 {
			copy(dst[:SaltSize], sqliteMagic)
			copy(dst[SaltSize:], plain)
		} else {
			copy(dst, plain)
		}
		// reserved bytes (last 80) are left as zero — unused by SQLite.
	}
	return out, nil
}

// Encrypt turns a plaintext SQLite database back into the SQLCipher v4 format,
// reusing the given 16-byte salt so the game derives the same key. Pass the salt
// obtained from the original file (Salt) to keep it byte-for-byte compatible.
func Encrypt(plain []byte, passphrase string, salt []byte) ([]byte, error) {
	if len(plain) == 0 || len(plain)%PageSize != 0 {
		return nil, fmt.Errorf("sqlcipher: plaintext size %d is not a multiple of page size %d", len(plain), PageSize)
	}
	if len(salt) != SaltSize {
		return nil, fmt.Errorf("sqlcipher: salt must be %d bytes, got %d", SaltSize, len(salt))
	}
	encKey, hmacKey, err := deriveKeys(passphrase, salt)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, err
	}

	nPages := len(plain) / PageSize
	out := make([]byte, len(plain))
	reserveStart := PageSize - Reserve

	for p := 0; p < nPages; p++ {
		pgno := uint32(p + 1)
		page := plain[p*PageSize : (p+1)*PageSize]
		encStart := 0
		if p == 0 {
			encStart = SaltSize
		}
		content := page[encStart:reserveStart]

		iv := make([]byte, IVSize)
		if _, err := rand.Read(iv); err != nil {
			return nil, err
		}
		cipherText := make([]byte, len(content))
		cipher.NewCBCEncrypter(block, iv).CryptBlocks(cipherText, content)
		mac := pageHMAC(hmacKey, cipherText, iv, pgno)

		dst := out[p*PageSize : (p+1)*PageSize]
		if p == 0 {
			copy(dst[:SaltSize], salt)
			copy(dst[SaltSize:], cipherText)
		} else {
			copy(dst, cipherText)
		}
		copy(dst[reserveStart:reserveStart+IVSize], iv)
		copy(dst[reserveStart+IVSize:reserveStart+IVSize+HMACSize], mac)
	}
	return out, nil
}

// Salt returns the 16-byte salt stored at the start of a SQLCipher file.
func Salt(enc []byte) ([]byte, error) {
	if len(enc) < SaltSize {
		return nil, fmt.Errorf("sqlcipher: file too small")
	}
	return bytes.Clone(enc[:SaltSize]), nil
}
