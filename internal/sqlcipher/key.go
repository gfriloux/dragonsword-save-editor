package sqlcipher

import "strconv"

// keySeed is the constant string DragonSword Awakening embeds in its client and
// feeds to an FNV-1a-64 hash to produce the SQLCipher passphrase. It looks like a
// raw SQLCipher key literal, but the game hashes the literal *text* rather than
// using it directly, so the effective passphrase is a single fixed number shared
// by every save file.
const keySeed = "x'1f86483d109b44e81d31181f14b2bba014cf8a174b4fa7f3fe666f075c2ae6c0'"

// fnv1a64UTF16 hashes the UTF-16 code units of s, matching the game's routine.
func fnv1a64UTF16(s string) uint64 {
	const (
		offset uint64 = 0xcbf29ce484222325
		prime  uint64 = 0x100000001b3
	)
	h := offset
	for _, r := range s {
		// The seed is pure ASCII, so each rune is a single UTF-16 code unit.
		h = (h ^ uint64(r)) * prime
	}
	return h
}

// Passphrase is the SQLCipher key for every DragonSword Awakening save. It equals
// DefaultPassphrase() and is provided as a constant for convenience.
const Passphrase = "13314374259236352028"

// DefaultPassphrase recomputes the passphrase from the embedded seed. It always
// returns Passphrase; keeping the derivation in code documents where the key
// comes from and guards against typos.
func DefaultPassphrase() string {
	return strconv.FormatUint(fnv1a64UTF16(keySeed), 10)
}
