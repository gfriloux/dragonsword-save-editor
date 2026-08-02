package pak

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gfriloux/dragonsword-save-editor/internal/oodle"
)

// gameProvider opens the real game paks, or skips when DSA_GAME_DIR is unset.
func gameProvider(t *testing.T) *Provider {
	t.Helper()
	dir := os.Getenv("DSA_GAME_DIR")
	if dir == "" {
		t.Skip("set DSA_GAME_DIR=/path/to/DragonSword game folder to run pak tests")
	}
	pv, err := OpenDir(filepath.Join(dir, "DS", "Content", "Paks"))
	if err != nil {
		t.Fatalf("OpenDir: %v", err)
	}
	return pv
}

func TestOpenAndListPaks(t *testing.T) {
	pv := gameProvider(t)
	files := pv.Files()
	if len(files) < 100000 {
		t.Fatalf("expected a large file list, got %d", len(files))
	}
	// Mount points look right.
	if !strings.Contains(files[len(files)/2], "DS/Content/") {
		t.Fatalf("unexpected path sample: %q", files[len(files)/2])
	}
}

func TestReadStoredUAsset(t *testing.T) {
	pv := gameProvider(t)
	// The Gold icon .uasset is stored (uncompressed).
	e := pv.Find("Icon_Item/Common/Icon_Item_Common_Gold.uasset")
	if e == nil {
		t.Skip("Gold icon not found (game version differs?)")
	}
	if !e.Stored() {
		t.Fatalf(".uasset expected stored, got method %q", e.Method)
	}
	b, err := e.ReadStored()
	if err != nil {
		t.Fatalf("ReadStored: %v", err)
	}
	if int64(len(b)) != e.UncompressedSize {
		t.Fatalf("size mismatch: got %d want %d", len(b), e.UncompressedSize)
	}
	// Cooked UE package tag PACKAGE_FILE_TAG = 0x9E2A83C1.
	if binary.LittleEndian.Uint32(b) != 0x9E2A83C1 {
		t.Fatalf("not a UE package (tag %#x)", binary.LittleEndian.Uint32(b))
	}
}

func TestOodleUexpRejectsStoredRead(t *testing.T) {
	pv := gameProvider(t)
	e := pv.Find("Icon_Item/Common/Icon_Item_Common_Gold.uexp")
	if e == nil {
		t.Skip("Gold icon .uexp not found")
	}
	if e.Stored() {
		t.Fatalf(".uexp expected Oodle-compressed, got stored")
	}
	if _, err := e.ReadStored(); err == nil {
		t.Fatalf("ReadStored should reject a compressed entry")
	}
	if len(e.Blocks) == 0 {
		t.Fatalf("expected compressed block ranges")
	}
}

func TestCompressedBlocksCoverEntry(t *testing.T) {
	pv := gameProvider(t)
	e := pv.Find("Icon_Item/Common/Icon_Item_Common_Gold.uexp")
	if e == nil {
		t.Skip("Gold icon .uexp not found")
	}
	blocks, err := e.CompressedBlocks()
	if err != nil {
		t.Fatalf("CompressedBlocks: %v", err)
	}
	var sum int
	for _, b := range blocks {
		if len(b.Data) == 0 || b.RawSize == 0 {
			t.Fatalf("empty block: %+v", b)
		}
		sum += b.RawSize
	}
	if int64(sum) != e.UncompressedSize {
		t.Fatalf("block raw sizes sum to %d, want %d", sum, e.UncompressedSize)
	}
}

// TestReadOodleIcon is the end-to-end pure-Go proof: the pak reader + the
// embedded ooz.wasm (via internal/oodle) decompress the real Gold icon .uexp to
// exactly the bytes the ooz C reference produced (checked by SHA-256).
func TestReadOodleIcon(t *testing.T) {
	pv := gameProvider(t)
	e := pv.Find("Icon_Item/Common/Icon_Item_Common_Gold.uexp")
	if e == nil {
		t.Skip("Gold icon .uexp not found")
	}
	dec, err := oodle.New()
	if err != nil {
		t.Fatalf("oodle.New: %v", err)
	}
	defer dec.Close()

	b, err := e.Read(dec)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if int64(len(b)) != e.UncompressedSize {
		t.Fatalf("got %d bytes, want %d", len(b), e.UncompressedSize)
	}
	const golden = "5d84cb8ea5cbd397d99a320a14f50321a3969a301d2628292256d5215462779d"
	if got := hex.EncodeToString(sha256Sum(b)); got != golden {
		t.Fatalf("sha256 = %s, want %s", got, golden)
	}
}

func sha256Sum(b []byte) []byte { h := sha256.Sum256(b); return h[:] }
