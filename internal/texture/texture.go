// Package texture decodes the game's cooked UTexture2D icons to images in pure
// Go. The icons are single-mip DXT5 (BC3) textures whose pixel data lives at the
// tail of the decompressed .uexp; this package locates it and decodes it.
package texture

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"strings"
)

// DecodeUExp decodes a cooked UTexture2D from its decompressed .uexp bytes.
// It supports the format the game uses for item icons (PF_DXT5, single mip).
func DecodeUExp(uexp []byte) (*image.NRGBA, error) {
	pf := bytes.LastIndex(uexp, []byte("PF_"))
	if pf < 16 {
		return nil, fmt.Errorf("texture: no pixel format found")
	}
	// Cooked FTexturePlatformData: [SizeX i32][SizeY i32][NumSlices i32]
	// [FString len i32]["PF_..\0"]. The format string starts at pf.
	sizeX := int(int32(binary.LittleEndian.Uint32(uexp[pf-16:])))
	sizeY := int(int32(binary.LittleEndian.Uint32(uexp[pf-12:])))
	end := bytes.IndexByte(uexp[pf:], 0)
	if end < 0 {
		return nil, fmt.Errorf("texture: unterminated pixel format")
	}
	format := string(uexp[pf : pf+end])
	if sizeX <= 0 || sizeY <= 0 || sizeX > 1<<14 || sizeY > 1<<14 {
		return nil, fmt.Errorf("texture: bad dimensions %dx%d", sizeX, sizeY)
	}
	if !strings.EqualFold(format, "PF_DXT5") {
		return nil, fmt.Errorf("texture: unsupported pixel format %s", format)
	}

	// Single-mip DXT5: mip0 is w*h bytes (16 bytes per 4x4 block). The mip header
	// varies slightly, but the bulk data is always immediately followed by the
	// mip's own SizeX/SizeY — locate it by that marker rather than a fixed offset.
	mipSize := sizeX * sizeY
	for off := pf + end + 1; off+mipSize+8 <= len(uexp); off++ {
		sx := int(int32(binary.LittleEndian.Uint32(uexp[off+mipSize:])))
		sy := int(int32(binary.LittleEndian.Uint32(uexp[off+mipSize+4:])))
		if sx == sizeX && sy == sizeY {
			return decodeDXT5(uexp[off:off+mipSize], sizeX, sizeY), nil
		}
	}
	return nil, fmt.Errorf("texture: could not locate mip0 (%dx%d)", sizeX, sizeY)
}
