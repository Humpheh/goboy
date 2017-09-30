package gb

const (
	PaletteGreyscale = byte(iota)
	PaletteOriginal
	PaletteRGB
)

var current_palette = PaletteRGB
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
	col := palettes[current_palette][index]
	return col[0], col[1], col[2]
}
