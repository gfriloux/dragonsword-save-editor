package texture

import (
	"encoding/binary"
	"image"
)

// decodeDXT5 decodes a DXT5/BC3 block-compressed image (w and h are pixel
// dimensions, expected multiples of 4) to a straight-alpha NRGBA image.
func decodeDXT5(src []byte, w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	bw, bh := (w+3)/4, (h+3)/4
	var alpha [8]uint8
	var col [4][3]uint8

	blk := 0
	for by := 0; by < bh; by++ {
		for bx := 0; bx < bw; bx++ {
			b := src[blk*16 : blk*16+16]
			blk++

			// ── Alpha block (BC4) ──
			a0, a1 := b[0], b[1]
			alpha[0], alpha[1] = a0, a1
			if a0 > a1 {
				for i := 2; i < 8; i++ {
					alpha[i] = uint8((uint16(8-i)*uint16(a0) + uint16(i-1)*uint16(a1)) / 7)
				}
			} else {
				for i := 2; i < 6; i++ {
					alpha[i] = uint8((uint16(6-i)*uint16(a0) + uint16(i-1)*uint16(a1)) / 5)
				}
				alpha[6], alpha[7] = 0, 255
			}
			aBits := uint64(b[2]) | uint64(b[3])<<8 | uint64(b[4])<<16 |
				uint64(b[5])<<24 | uint64(b[6])<<32 | uint64(b[7])<<40

			// ── Color block (BC1, always 4-colour in BC3) ──
			c0 := binary.LittleEndian.Uint16(b[8:])
			c1 := binary.LittleEndian.Uint16(b[10:])
			col[0] = rgb565(c0)
			col[1] = rgb565(c1)
			for i := 0; i < 3; i++ {
				col[2][i] = uint8((2*uint16(col[0][i]) + uint16(col[1][i])) / 3)
				col[3][i] = uint8((uint16(col[0][i]) + 2*uint16(col[1][i])) / 3)
			}
			cBits := binary.LittleEndian.Uint32(b[12:])

			for r := 0; r < 4; r++ {
				for c := 0; c < 4; c++ {
					px, py := bx*4+c, by*4+r
					if px >= w || py >= h {
						continue
					}
					t := r*4 + c
					ci := (cBits >> (2 * t)) & 3
					ai := (aBits >> (3 * t)) & 7
					o := img.PixOffset(px, py)
					img.Pix[o+0] = col[ci][0]
					img.Pix[o+1] = col[ci][1]
					img.Pix[o+2] = col[ci][2]
					img.Pix[o+3] = alpha[ai]
				}
			}
		}
	}
	return img
}

// rgb565 expands a 5:6:5 colour to 8:8:8.
func rgb565(v uint16) [3]uint8 {
	r := uint8((v >> 11) & 0x1F)
	g := uint8((v >> 5) & 0x3F)
	b := uint8(v & 0x1F)
	return [3]uint8{
		r<<3 | r>>2,
		g<<2 | g>>4,
		b<<3 | b>>2,
	}
}
