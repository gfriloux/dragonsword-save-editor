# Encryption

The save database (`<accountId>_Slot<N>.db`) is encrypted with **SQLCipher v4**.

## Parameters

The game uses SQLCipher's `cipher_compatibility = 4` defaults:

| Parameter          | Value                              |
| ------------------ | ---------------------------------- |
| Cipher             | AES-256-CBC                        |
| Page size          | 4096 bytes                         |
| KDF                | PBKDF2-HMAC-SHA512, 256000 iterations |
| Per-page integrity | HMAC-SHA512 (16-byte IV + 64-byte MAC per page) |
| Plaintext header   | 0 bytes                            |

The first 16 bytes of the file are the random salt; the file size is always a multiple
of the 4096-byte page size.

## The key

The passphrase is a **hardcoded constant, identical for every save and every player**:

```
13314374259236352028
```

### How it's derived

The client embeds the 67-character string

```
x'1f86483d109b44e81d31181f14b2bba014cf8a174b4fa7f3fe666f075c2ae6c0'
```

and hashes its **UTF-16 code units** with **FNV-1a-64** (offset basis
`0xcbf29ce484222325`, prime `0x100000001b3`), then renders the 64-bit result as a
decimal string via `%llu`:

```
FNV1a64_utf16("x'1f86483d…c2ae6c0'") = 0xb8c628d8a063301c = 13314374259236352028
```

The game passes this through its custom `unreal-fs` SQLite backend as
`PRAGMA key = '13314374259236352028';`. (It looks like a raw SQLCipher key literal, but
the game hashes the literal *text* rather than using it directly — so the effective key
is that single fixed number.)

## Opening it with standard tools

With the [`sqlcipher`](https://www.zetetic.net/sqlcipher/) CLI:

```sql
PRAGMA key = '13314374259236352028';
PRAGMA cipher_compatibility = 4;
-- now query normally, e.g.:
SELECT name FROM sqlite_master WHERE type='table';
```

To export a plaintext copy:

```sql
PRAGMA key = '13314374259236352028';
PRAGMA cipher_compatibility = 4;
ATTACH DATABASE 'plain.db' AS plaintext KEY '';
SELECT sqlcipher_export('plaintext');
DETACH DATABASE plaintext;
```

In-place edits (`UPDATE`/`INSERT`) re-encrypt the touched pages with the same salt and
key, so the game reads an edited save unchanged. **Edit only while the game is closed** —
it keeps the database open and would overwrite external changes.

A pure-Go implementation of this exact format lives in this repo under
`internal/sqlcipher` (no CGO), if you want to read/write saves without the SQLCipher CLI.
