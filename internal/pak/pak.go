// Package pak reads DragonSword Awakening's custom UE5 pak archives in pure Go
// (no CGO). The container is a standard UE5 pak whose footer version is
// obfuscated to 101 and whose encrypted index carries an extra XOR layer,
// reverse-engineered from CUE4Parse's GameTypes/DragonSword profile.
//
// The public AES key and the deobfuscation are documented inline. Actual asset
// pixel data (.uexp) is Oodle-compressed and decoded via the sibling
// internal/oodle package; stored (uncompressed) records are read directly here.
package pak

import (
	"crypto/aes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// pakMagic is the standard UE pak footer magic (present despite the bogus
// version 101).
const pakMagic = 0x5A6F12E1

// footerLen is the fixed footer size for these paks: EncryptionKeyGuid(16) +
// bEncrypted(1) + magic(4) + version(4) + indexOffset(8) + indexSize(8) +
// indexHash(20) + 5×32 compression-method names.
const footerLen = 221

// aesKey is the game's public pak key (static; the all-zero key GUID selects it).
var aesKey = [32]byte{
	0x26, 0x34, 0x79, 0xC4, 0x42, 0xD4, 0x5B, 0x7E, 0xED, 0xE7, 0xB3, 0xA3, 0x6B, 0xBB, 0x3C, 0x3B,
	0x39, 0xEF, 0x91, 0x78, 0xA2, 0xF8, 0x2A, 0xB6, 0x94, 0xFB, 0x41, 0x0A, 0xB1, 0x5E, 0x01, 0xAD,
}

// Block is a compressed block's byte range within the pak file.
type Block struct{ Start, End int64 }

// Entry is one file inside a pak.
type Entry struct {
	Path             string
	pak              *Pak
	Offset           int64
	UncompressedSize int64
	CompressedSize   int64
	Method           string // "None" (stored) or a compression method name (e.g. "Oodle")
	Encrypted        bool
	BlockSize        uint32
	Blocks           []Block
}

// Stored reports whether the entry is uncompressed.
func (e *Entry) Stored() bool { return e.Method == "None" || e.Method == "" }

// Pak is a single opened .pak archive.
type Pak struct {
	path    string
	methods []string // index 0 = "None", 1.. = footer names
	entries map[string]*Entry
}

// decryptIndex AES-256-ECB decrypts an index region, then removes the custom XOR
// layer (whole buffer XOR decrypted byte [2]).
func decryptIndex(data []byte) []byte {
	blk, _ := aes.NewCipher(aesKey[:])
	out := make([]byte, len(data))
	for i := 0; i+16 <= len(data); i += 16 {
		blk.Decrypt(out[i:i+16], data[i:i+16])
	}
	if len(out) >= 16 {
		k := out[2]
		for i := range out {
			out[i] ^= k
		}
	}
	return out
}

// ecbDecrypt AES-256-ECB decrypts file data in place (no XOR layer — that is
// index-only). len(data) must be a multiple of 16.
func ecbDecrypt(data []byte) {
	blk, _ := aes.NewCipher(aesKey[:])
	for i := 0; i+16 <= len(data); i += 16 {
		blk.Decrypt(data[i:i+16], data[i:i+16])
	}
}

func align16(n int64) int64 { return (n + 15) &^ 15 }

type cursor struct {
	b []byte
	p int
}

func (c *cursor) i32() int32    { v := int32(binary.LittleEndian.Uint32(c.b[c.p:])); c.p += 4; return v }
func (c *cursor) i64() int64    { v := int64(binary.LittleEndian.Uint64(c.b[c.p:])); c.p += 8; return v }
func (c *cursor) skip(n int)    { c.p += n }
func (c *cursor) boolean() bool { return c.i32() != 0 } // UE index bool = int32

// str reads an FString and removes its per-string XOR (bytes XOR their own last
// byte), dropping the trailing null.
func (c *cursor) str() string {
	n := c.i32()
	if n == 0 {
		return ""
	}
	if n > 0 {
		s := append([]byte(nil), c.b[c.p:c.p+int(n)]...)
		c.p += int(n)
		k := s[len(s)-1]
		for i := range s {
			s[i] ^= k
		}
		return string(s[:len(s)-1])
	}
	m := int(-n)
	raw := append([]byte(nil), c.b[c.p:c.p+m*2]...)
	c.p += m * 2
	k := raw[len(raw)-1]
	for i := range raw {
		raw[i] ^= k
	}
	out := make([]byte, 0, m)
	for i := 0; i < m-1; i++ {
		out = append(out, raw[i*2])
	}
	return string(out)
}

// openPak parses a single .pak's footer and index.
func openPak(path string) (*Pak, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	sz := fi.Size()

	foot := make([]byte, footerLen)
	if _, err := f.ReadAt(foot, sz-footerLen); err != nil {
		return nil, err
	}
	if binary.LittleEndian.Uint32(foot[17:]) != pakMagic {
		return nil, fmt.Errorf("pak %s: bad footer magic", filepath.Base(path))
	}
	idxOff := int64(binary.LittleEndian.Uint64(foot[25:]))
	idxSize := int64(binary.LittleEndian.Uint64(foot[33:]))

	p := &Pak{path: path, methods: []string{"None"}, entries: map[string]*Entry{}}
	for i := 0; i < 5; i++ {
		name := foot[61+i*32 : 61+i*32+32]
		if z := strings.IndexByte(string(name), 0); z >= 0 {
			name = name[:z]
		}
		if n := strings.TrimSpace(string(name)); n != "" {
			p.methods = append(p.methods, n)
		}
	}

	raw := make([]byte, idxSize)
	if _, err := f.ReadAt(raw, idxOff); err != nil {
		return nil, err
	}
	pi := &cursor{b: decryptIndex(raw)}
	mount := pi.str()
	if !strings.Contains(mount, "/") {
		return nil, fmt.Errorf("pak %s: bad mount point %q", filepath.Base(path), mount)
	}
	_ = pi.i32() // file count (recomputed below)
	pi.skip(8)   // PathHashSeed
	if !pi.boolean() {
		return nil, fmt.Errorf("pak %s: no path hash index", filepath.Base(path))
	}
	pi.skip(36)
	if !pi.boolean() {
		return nil, fmt.Errorf("pak %s: no directory index", filepath.Base(path))
	}
	dirOff := pi.i64()
	dirSize := pi.i64()
	pi.skip(20)
	encSize := pi.i32()
	encoded := append([]byte(nil), pi.b[pi.p:pi.p+int(encSize)]...)
	if encSize > 0 {
		k := encoded[5] // encoded entries XOR layer
		for i := range encoded {
			encoded[i] ^= k
		}
	}

	dir := make([]byte, dirSize)
	if _, err := f.ReadAt(dir, dirOff); err != nil {
		return nil, err
	}
	di := &cursor{b: decryptIndex(dir)}
	dirCount := di.i32()
	for i := int32(0); i < dirCount; i++ {
		d := di.str()
		fc := di.i32()
		for j := int32(0); j < fc; j++ {
			name := di.str()
			off := di.i32()
			if off < 0 {
				continue // non-encoded entries are unused by this game's paks
			}
			e := decodeEntry(p, encoded, int(off))
			e.Path = mount + strings.TrimPrefix(d, "/") + name
			p.entries[e.Path] = e
		}
	}
	return p, nil
}

// decodeEntry decodes an encoded FPakEntry (UE FPakFile::DecodePakEntry) and
// computes its compressed block ranges.
func decodeEntry(p *Pak, enc []byte, off int) *Entry {
	a := &cursor{b: enc, p: off}
	bitfield := uint32(a.i32())
	cmi := (bitfield >> 23) & 0x3F
	blocksCount := (bitfield >> 6) & 0xFFFF
	e := &Entry{pak: p, Encrypted: bitfield&(1<<22) != 0}
	if int(cmi) < len(p.methods) {
		e.Method = p.methods[cmi]
	}

	var blockSize uint32
	if (bitfield & 0x3F) == 0x3F {
		blockSize = uint32(a.i32())
	} else {
		blockSize = (bitfield & 0x3F) << 11
	}

	off32 := bitfield&(1<<31) != 0
	usz32 := bitfield&(1<<30) != 0
	csz32 := bitfield&(1<<29) != 0
	if off32 {
		e.Offset = int64(uint32(a.i32()))
	} else {
		e.Offset = a.i64()
	}
	if usz32 {
		e.UncompressedSize = int64(uint32(a.i32()))
	} else {
		e.UncompressedSize = a.i64()
	}
	if cmi != 0 {
		if csz32 {
			e.CompressedSize = int64(uint32(a.i32()))
		} else {
			e.CompressedSize = a.i64()
		}
	} else {
		e.CompressedSize = e.UncompressedSize
	}
	if blocksCount == 1 {
		e.BlockSize = uint32(e.UncompressedSize)
	} else {
		e.BlockSize = blockSize
	}

	// StructSize: the on-disk FPakEntry header prepended to the payload.
	structSize := int64(8*3 + 4*2 + 1 + 20)
	if cmi != 0 {
		structSize += int64(4 + int(blocksCount)*2*8)
	}
	start := e.Offset + structSize
	if blocksCount == 1 && !e.Encrypted {
		e.Blocks = []Block{{start, start + e.CompressedSize}}
	} else if blocksCount > 0 {
		align := int64(1)
		if e.Encrypted {
			align = 16
		}
		for i := uint32(0); i < blocksCount; i++ {
			length := int64(uint32(a.i32()))
			e.Blocks = append(e.Blocks, Block{start, start + length})
			start += (length + align - 1) / align * align
		}
	}
	return e
}

// Provider is a read-only view over a directory of paks. Later paks in glob
// order win on duplicate paths (matching UE mount ordering closely enough for
// this single-source game).
type Provider struct {
	paks  []*Pak
	byPth map[string]*Entry
}

// OpenDir opens every *.pak under dir.
func OpenDir(dir string) (*Provider, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.pak"))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	pv := &Provider{byPth: map[string]*Entry{}}
	for _, m := range matches {
		p, err := openPak(m)
		if err != nil {
			return nil, err
		}
		pv.paks = append(pv.paks, p)
		for path, e := range p.entries {
			pv.byPth[path] = e
		}
	}
	if len(pv.paks) == 0 {
		return nil, fmt.Errorf("no .pak files in %s", dir)
	}
	return pv, nil
}

// Files returns every mounted virtual path, sorted.
func (pv *Provider) Files() []string {
	out := make([]string, 0, len(pv.byPth))
	for p := range pv.byPth {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// Find looks up an entry by exact virtual path, or by a path suffix if no exact
// match is found (paths carry a "../../../DS/Content/" mount prefix).
func (pv *Provider) Find(path string) *Entry {
	if e, ok := pv.byPth[path]; ok {
		return e
	}
	for p, e := range pv.byPth {
		if strings.HasSuffix(p, path) {
			return e
		}
	}
	return nil
}

// rawBlocks reads and concatenates an entry's on-disk block bytes (still
// compressed if the entry is Oodle).
func (e *Entry) rawBlocks() ([]byte, error) {
	f, err := os.Open(e.pak.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if len(e.Blocks) == 0 {
		// Stored, single unblocked record: payload right after the header.
		structSize := int64(8*3 + 4*2 + 1 + 20)
		rd := e.UncompressedSize
		if e.Encrypted {
			rd = align16(rd)
		}
		buf := make([]byte, rd)
		if _, err := f.ReadAt(buf, e.Offset+structSize); err != nil {
			return nil, err
		}
		if e.Encrypted {
			ecbDecrypt(buf)
		}
		return buf[:e.UncompressedSize], nil
	}
	var out []byte
	for _, b := range e.Blocks {
		length := b.End - b.Start
		rd := length
		if e.Encrypted {
			rd = align16(length)
		}
		buf := make([]byte, rd)
		if _, err := f.ReadAt(buf, b.Start); err != nil {
			return nil, err
		}
		if e.Encrypted {
			ecbDecrypt(buf)
		}
		out = append(out, buf[:length]...)
	}
	return out, nil
}

// ReadStored returns the bytes of an uncompressed entry (decrypting it if the
// record is AES-encrypted). It errors on compressed (Oodle) entries, which need
// internal/oodle (wired in a later step).
func (e *Entry) ReadStored() ([]byte, error) {
	if !e.Stored() {
		return nil, fmt.Errorf("%s is %s-compressed (not stored)", e.Path, e.Method)
	}
	return e.rawBlocks()
}
