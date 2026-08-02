package texture

import (
	"encoding/binary"
	"testing"
)

// solidRedBlock is a DXT5 block decoding to solid opaque red across all 16
// texels: alpha0=255 with all-zero indices → 255; colour0=colour1=565 red with
// all-zero indices → red.
func solidRedBlock() []byte {
	b := make([]byte, 16)
	b[0], b[1] = 0xFF, 0x00                       // alpha endpoints; indices (b[2:8]) = 0 → alpha 255
	binary.LittleEndian.PutUint16(b[8:], 0xF800)  // colour0 = red (565)
	binary.LittleEndian.PutUint16(b[10:], 0xF800) // colour1 = red
	// colour indices (b[12:16]) = 0 → colour0
	return b
}

func TestDecodeDXT5Solid(t *testing.T) {
	img := decodeDXT5(solidRedBlock(), 4, 4)
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			o := img.PixOffset(x, y)
			r, g, b, a := img.Pix[o], img.Pix[o+1], img.Pix[o+2], img.Pix[o+3]
			if r != 255 || g != 0 || b != 0 || a != 255 {
				t.Fatalf("pixel (%d,%d) = %d,%d,%d,%d want 255,0,0,255", x, y, r, g, b, a)
			}
		}
	}
}

func TestDecodeUExpSolid(t *testing.T) {
	// Craft a minimal cooked platform-data tail: SizeX, SizeY, NumSlices, the
	// PF_DXT5 FString, then the mip0 block and the 4-byte package tag.
	var buf []byte
	put32 := func(v int32) {
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, uint32(v))
		buf = append(buf, b...)
	}
	buf = append(buf, make([]byte, 8)...) // some leading export bytes
	put32(4)                              // SizeX
	put32(4)                              // SizeY
	put32(1)                              // NumSlices
	put32(8)                              // FString length ("PF_DXT5\0")
	buf = append(buf, []byte("PF_DXT5\x00")...)
	put32(0)                              // FirstMipToSerialize
	put32(1)                              // NumMips
	put32(0)                              // one mip-header int32
	buf = append(buf, solidRedBlock()...) // mip0 (16 bytes for 4x4 DXT5)
	put32(4)                              // trailing mip SizeX (the locator marker)
	put32(4)                              // trailing mip SizeY
	put32(1)                              // trailing mip SizeZ
	buf = append(buf, 0xC1, 0x83, 0x2A, 0x9E)

	img, err := DecodeUExp(buf)
	if err != nil {
		t.Fatalf("DecodeUExp: %v", err)
	}
	if b := img.Bounds(); b.Dx() != 4 || b.Dy() != 4 {
		t.Fatalf("dims = %dx%d want 4x4", b.Dx(), b.Dy())
	}
	o := img.PixOffset(0, 0)
	if img.Pix[o] != 255 || img.Pix[o+3] != 255 {
		t.Fatalf("top-left = %v want red/opaque", img.Pix[o:o+4])
	}
}
