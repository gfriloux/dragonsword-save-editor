# Paks (why item data isn't read from the game)

The game ships all content in Unreal Engine `.pak` archives under
`DS/Content/Paks/` (~50 chunks, ~19 GB). Item/character **names** live there in
DataTables (the exe references `DItemData`, `DCookRecipeData`, `BaseItem.Name`…) plus
StringTables / `.locres` localization. So in principle the names/icons could be read
straight from the game.

In practice they can't, easily: **the paks are custom-protected.**

## What we found

- Every pak footer has `bEncrypted = 1` with an **all-zero encryption-key GUID** — the
  index is AES-256-**ECB** encrypted with the game's single global key.
- The pak footer reports **version 101**, which is bogus for stock Unreal (real pak
  versions are ≤ 12). The exe also carries the string
  `unsupported content encryption algorithm`. So the format is deliberately
  modified/obfuscated: stock tools (repak, FModel/CUE4Parse, UnrealPak) won't read it.
- The global AES key is **not** a plaintext string in the exe, and it is **not** the
  SQLCipher key. A dynamic memory scan of the running game — treating every 16-byte
  aligned window as an AES-256-ECB key and testing it against a pak index (which should
  start with a mount-point `FString`) — found **no** valid key, in both writable and
  read-only regions. That means the standard assumption (AES-256-ECB + standard index
  layout + a resident 32-byte key) does not hold.

## Conclusion

Extracting names/icons from the paks would require reversing the game's *own* pak
reader in the exe to learn the true algorithm and index layout — a deep, uncertain
effort — and then generating a `.usmap` to parse the (unversioned) DataTables, and
resolving localization. It's a wall.

So this project takes the pragmatic path: it gets names and icons from the community
datamine at [th.gl](https://dragonswordawakening.th.gl), keyed by the same
[CIDs](content-ids.md), via `cmd/gen-catalog`.

(By contrast, the **save** database uses standard SQLCipher and a recoverable constant
key — see [Encryption](encryption.md) — which is why editing saves *is* tractable.)
