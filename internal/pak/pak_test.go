package pak

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
