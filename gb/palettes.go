package gb

import (
	"fmt"

	"github.com/Humpheh/goboy/bits"
)

const (
	PaletteGreyscale = byte(iota)
	PaletteOriginal
	PaletteRGB
)

var currentPalette = PaletteRGB
var palettes = map[byte][4][3]uint8{
	// Greyscale paletter
	PaletteGreyscale: {
		{0xFF, 0xFF, 0xFF},
		{0xCC, 0xCC, 0xCC},
		{0x77, 0x77, 0x77},
		{0x00, 0x00, 0x00},
	},
	// Palette using the colours as it would have been on the GameBoy
	PaletteOriginal: {
		{0x9B, 0xBC, 0x0F},
		{0x8B, 0xAC, 0x0F},
		{0x30, 0x62, 0x30},
		{0x0F, 0x38, 0x0F},
	},
	// Palette used by default in the BGB emulator
	PaletteRGB: {
		{0xE0, 0xF8, 0xD0},
		{0x88, 0xC0, 0x70},
		{0x34, 0x68, 0x56},
		{0x08, 0x18, 0x20},
	},
}

// Get the colour based on the colour index and the currently
// selected palette.
func GetPaletteColour(index byte) (uint8, uint8, uint8) {
	col := palettes[currentPalette][index]
	return col[0], col[1], col[2]
}

// Get a new colour palette.
func NewPalette() *CGBPalette {
	pal := make([]byte, 0x40)
	for i := range pal {
		pal[i] = 0xFF
	}

	return &CGBPalette{
		palette: pal,
	}
}

// Palette information tracking the palette colour info.
type CGBPalette struct {
	// Palette colour information.
	palette []byte
	// Current index the palette is referencing.
	index byte
	// If to auto increment on write.
	inc bool
}

// Update the index the palette is indexing and set
// auto increment if bit 7 is set.
func (pal *CGBPalette) updateIndex(value byte) {
	pal.index = value & 0x3F
	pal.inc = bits.Test(value, 7)
}

// Read the palette information stored at the current
// index.
func (pal *CGBPalette) read() byte {
	return pal.palette[pal.index]
}

// Read the current index.
func (pal *CGBPalette) readIndex() byte {
	return pal.index
}

// Write a value to the palette at the current index.
func (pal *CGBPalette) write(value byte) {
	pal.palette[pal.index] = value
	if pal.inc {
		pal.index = (pal.index + 1) & 0x3F
	}
}

// Get the rgb colour for a palette at a colour number.
func (pal *CGBPalette) get(palette byte, num byte) (uint8, uint8, uint8) {
	idx := (palette * 8) + (num * 2)
	colour := uint16(pal.palette[idx]) | uint16(pal.palette[idx+1])<<8
	r := uint8(colour & 0x1F)
	g := uint8((colour >> 5) & 0x1F)
	b := uint8((colour >> 10) & 0x1F)
	return colMap[r], colMap[g], colMap[b]
}

// String output of the palette values.
func (pal *CGBPalette) String() string {
	out := ""
	for i := 0; i < len(pal.palette); i += 2 {
		out += fmt.Sprintf("%02x%02x ", pal.palette[i+1], pal.palette[i])
		if (i+2)%8 == 0 {
			out += "\n"
		}
	}
	return out
}

// Mapping of the 5 bit colour value to a 8 bit value.
var colMap = map[uint8]uint8{
	0x0:  0x0,
	0x1:  0x8,
	0x2:  0x10,
	0x3:  0x18,
	0x4:  0x20,
	0x5:  0x29,
	0x6:  0x31,
	0x7:  0x39,
	0x8:  0x41,
	0x9:  0x4a,
	0xa:  0x52,
	0xb:  0x5a,
	0xc:  0x62,
	0xd:  0x6a,
	0xe:  0x73,
	0xf:  0x7b,
	0x10: 0x83,
	0x11: 0x8b,
	0x12: 0x94,
	0x13: 0x9c,
	0x14: 0xa4,
	0x15: 0xac,
	0x16: 0xb4,
	0x17: 0xbd,
	0x18: 0xc5,
	0x19: 0xcd,
	0x1a: 0xd5,
	0x1b: 0xde,
	0x1c: 0xe6,
	0x1d: 0xee,
	0x1e: 0xf6,
	0x1f: 0xff,
}
