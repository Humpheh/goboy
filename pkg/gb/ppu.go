package gb

import (
	"github.com/Humpheh/goboy/pkg/bits"
)

const (
	// ScreenWidth is the number of pixels width on the GameBoy LCD panel.
	ScreenWidth = 160

	// ScreenHeight is the number of pixels height on the GameBoy LCD panel.
	ScreenHeight = 144

	// LCDC is the main LCD Control register.
	LCDC = 0xFF40
)

// Update the state of the graphics.
func (gb *Gameboy) updateGraphics(cycles int) {
	gb.setLCDStatus()

	if !gb.isLCDEnabled() {
		return
	}
	gb.scanlineCounter -= cycles

	if gb.scanlineCounter <= 0 {
		gb.Memory.HighRAM[0x44]++
		if gb.Memory.HighRAM[0x44] > 153 {
			gb.PreparedData = gb.screenData
			gb.screenData = [ScreenWidth][ScreenHeight][3]uint8{}
			gb.bgPriority = [ScreenWidth][ScreenHeight]bool{}
			gb.Memory.HighRAM[0x44] = 0
		}

		currentLine := gb.Memory.ReadHighRam(0xFF44)
		gb.scanlineCounter += 456 * gb.getSpeed()

		if currentLine == ScreenHeight {
			gb.requestInterrupt(0)
		}
	}
}

const (
	lcdMode2Bounds = 456 - 80
	lcdMode3Bounds = lcdMode2Bounds - 172
)

// Set the status of the LCD based on the current state of memory.
func (gb *Gameboy) setLCDStatus() {
	status := gb.Memory.ReadHighRam(0xFF41)

	if !gb.isLCDEnabled() {
		// set the screen to white
		gb.clearScreen()

		gb.scanlineCounter = 456
		gb.Memory.HighRAM[0x44] = 0
		status &= 252
		// TODO: Check this is correct
		// We aren't in a mode so reset the values
		status = bits.Reset(status, 0)
		status = bits.Reset(status, 1)
		gb.Memory.Write(0xFF41, status)
		return
	}
	gb.screenCleared = false

	currentLine := gb.Memory.ReadHighRam(0xFF44)
	currentMode := status & 0x3

	var mode byte
	requestInterrupt := false

	switch {
	case currentLine >= 144:
		mode = 1
		status = bits.Set(status, 0)
		status = bits.Reset(status, 1)
		requestInterrupt = bits.Test(status, 4)
	case gb.scanlineCounter >= lcdMode2Bounds:
		mode = 2
		status = bits.Reset(status, 0)
		status = bits.Set(status, 1)
		requestInterrupt = bits.Test(status, 5)
	case gb.scanlineCounter >= lcdMode3Bounds:
		mode = 3
		status = bits.Set(status, 0)
		status = bits.Set(status, 1)
		if mode != currentMode {
			// Draw the scanline when we start mode 3. In the real GameBoy
			// this would be done throughout mode 3 by reading OAM and VRAM
			// to generate the picture.
			gb.drawScanline(currentLine)
		}
	default:
		mode = 0
		status = bits.Reset(status, 0)
		status = bits.Reset(status, 1)
		requestInterrupt = bits.Test(status, 3)
		if mode != currentMode {
			gb.Memory.doHDMATransfer()
		}
	}

	if requestInterrupt && mode != currentMode {
		gb.requestInterrupt(1)
	}

	// Check if LYC == LY (coincidence flag)
	if currentLine == gb.Memory.ReadHighRam(0xFF45) {
		status = bits.Set(status, 2)
		// If enabled request an interrupt for this
		if bits.Test(status, 6) {
			gb.requestInterrupt(1)
		}
	} else {
		status = bits.Reset(status, 2)
	}

	gb.Memory.Write(0xFF41, status)
}

// Checks if the LCD is enabled by examining 0xFF40.
func (gb *Gameboy) isLCDEnabled() bool {
	return bits.Test(gb.Memory.ReadHighRam(LCDC), 7)
}

// Draw a single scanline to the graphics output.
func (gb *Gameboy) drawScanline(scanline byte) {
	control := gb.Memory.ReadHighRam(LCDC)

	// LCDC bit 0 clears tiles on DMG but controls priority on CGB.
	if (gb.IsCGB() || bits.Test(control, 0)) && !gb.Debug.HideBackground {
		gb.renderTiles(control, scanline)
	}

	if bits.Test(control, 1) && !gb.Debug.HideSprites {
		gb.renderSprites(control, int32(scanline))
	}
}

// Get settings to be used when rendering tiles.
func (gb *Gameboy) getTileSettings(lcdControl byte, windowY byte) (
	usingWindow bool,
	unsigned bool,
	tileData uint16,
	backgroundMemory uint16,
) {
	tileData = uint16(0x8800)

	if bits.Test(lcdControl, 5) {
		// Is current scanline we're drawing within windows Y position?
		if windowY <= gb.Memory.ReadHighRam(0xFF44) {
			usingWindow = true
		}
	}

	// Test if we're using unsigned bytes
	if bits.Test(lcdControl, 4) {
		tileData = 0x8000
		unsigned = true
	}

	// Work out where to look in background memory.
	var testBit byte = 3
	if usingWindow {
		testBit = 6
	}
	backgroundMemory = uint16(0x9800)
	if bits.Test(lcdControl, testBit) {
		backgroundMemory = 0x9C00
	}
	return
}

// Render a scanline of the tile map to the graphics output based
// on the state of the lcdControl register.
func (gb *Gameboy) renderTiles(lcdControl byte, scanline byte) {
	scrollY := gb.Memory.ReadHighRam(0xFF42)
	scrollX := gb.Memory.ReadHighRam(0xFF43)
	windowY := gb.Memory.ReadHighRam(0xFF4A)
	windowX := gb.Memory.ReadHighRam(0xFF4B) - 7

	usingWindow, unsigned, tileData, backgroundMemory := gb.getTileSettings(lcdControl, windowY)

	// yPos is used to calc which of 32 v-lines the current scanline is drawing
	var yPos byte
	if !usingWindow {
		yPos = scrollY + scanline
	} else {
		yPos = scanline - windowY
	}

	// which of the 8 vertical pixels of the current tile is the scanline on?
	var tileRow = uint16(yPos/8) * 32

	// Load the palette which will be used to draw the tiles
	var palette = gb.Memory.ReadHighRam(0xFF47)

	// start drawing the 160 horizontal pixels for this scanline
	gb.tileScanline = [160]uint8{}
	for pixel := byte(0); pixel < 160; pixel++ {
		xPos := pixel + scrollX

		// Translate the current x pos to window space if necessary
		if usingWindow && pixel >= windowX {
			xPos = pixel - windowX
		}

		// Which of the 32 horizontal tiles does this x_pox fall within?
		tileCol := uint16(xPos / 8)

		// Get the tile identity number
		tileAddress := backgroundMemory + tileRow + tileCol

		// Deduce where this tile id is in memory
		tileLocation := tileData
		if unsigned {
			tileNum := int16(gb.Memory.VRAM[tileAddress-0x8000])
			tileLocation = tileLocation + uint16(tileNum*16)
		} else {
			tileNum := int16(int8(gb.Memory.VRAM[tileAddress-0x8000]))
			tileLocation = uint16(int32(tileLocation) + int32((tileNum+128)*16))
		}

		bankOffset := uint16(0x8000)

		// Attributes used in CGB mode TODO: check in CGB mode
		//
		//    Bit 0-2  Background Palette number  (BGP0-7)
		//    Bit 3    Tile VRAM Bank number      (0=Bank 0, 1=Bank 1)
		//    Bit 5    Horizontal Flip            (0=Normal, 1=Mirror horizontally)
		//    Bit 6    Vertical Flip              (0=Normal, 1=Mirror vertically)
		//    Bit 7    BG-to-OAM Priority         (0=Use OAM priority bit, 1=BG Priority)
		//
		tileAttr := gb.Memory.VRAM[tileAddress-0x6000]
		if gb.IsCGB() && bits.Test(tileAttr, 3) {
			bankOffset = 0x6000
		}
		priority := bits.Test(tileAttr, 7)

		var line byte
		if gb.IsCGB() && bits.Test(tileAttr, 6) {
			// Vertical flip
			line = ((7 - yPos) % 8) * 2
		} else {
			line = (yPos % 8) * 2
		}
		// Get the tile data from memory
		data1 := gb.Memory.VRAM[tileLocation+uint16(line)-bankOffset]
		data2 := gb.Memory.VRAM[tileLocation+uint16(line)+1-bankOffset]

		if gb.IsCGB() && bits.Test(tileAttr, 5) {
			// Horizontal flip
			xPos = 7 - xPos
		}
		colourBit := byte(int8((xPos%8)-7) * -1)
		colourNum := (bits.Val(data2, colourBit) << 1) | bits.Val(data1, colourBit)
		gb.setTilePixel(pixel, scanline, tileAttr, colourNum, palette, priority)
	}
}

func (gb *Gameboy) setTilePixel(x, y, tileAttr, colourNum, palette byte, priority bool) {
	// Set the pixel
	if gb.IsCGB() {
		cgbPalette := tileAttr & 0x7
		red, green, blue := gb.BGPalette.get(cgbPalette, colourNum)
		gb.setPixel(x, y, red, green, blue, true)
		gb.bgPriority[x][y] = priority
	} else {
		red, green, blue := gb.getColour(colourNum, palette)
		gb.setPixel(x, y, red, green, blue, true)
	}

	// Store for the current scanline so sprite priority can be managed
	gb.tileScanline[x] = colourNum
}

// Get the RGB colour value for a colour num at an address using the current palette.
func (gb *Gameboy) getColour(colourNum byte, palette byte) (uint8, uint8, uint8) {
	hi := colourNum<<1 | 1
	lo := colourNum << 1
	col := (bits.Val(palette, hi) << 1) | bits.Val(palette, lo)
	return GetPaletteColour(col)
}

const spritePriorityOffset = 100

// Render the sprites to the screen on the current scanline using the lcdControl register.
func (gb *Gameboy) renderSprites(lcdControl byte, scanline int32) {
	var ySize int32 = 8
	if bits.Test(lcdControl, 2) {
		ySize = 16
	}

	// Load the two palettes which sprites can be drawn in
	var palette1 = gb.Memory.ReadHighRam(0xFF48)
	var palette2 = gb.Memory.ReadHighRam(0xFF49)

	var minx [ScreenWidth]int32
	var lineSprites = 0
	for sprite := uint16(0); sprite < 40; sprite++ {
		// Load sprite data from memory.
		index := sprite * 4

		// If this is true the scanline is out of the area we care about
		yPos := int32(gb.Memory.Read(uint16(0xFE00+index))) - 16
		if scanline < yPos || scanline >= (yPos+ySize) {
			continue
		}

		// Only 10 sprites are allowed to be displayed on each line
		if lineSprites >= 10 {
			break
		}
		lineSprites++

		xPos := int32(gb.Memory.Read(uint16(0xFE00+index+1))) - 8
		tileLocation := gb.Memory.Read(uint16(0xFE00 + index + 2))
		attributes := gb.Memory.Read(uint16(0xFE00 + index + 3))

		yFlip := bits.Test(attributes, 6)
		xFlip := bits.Test(attributes, 5)
		priority := !bits.Test(attributes, 7)

		// Bank the sprite data in is (CGB only)
		var bank uint16 = 0
		if gb.IsCGB() && bits.Test(attributes, 3) {
			bank = 1
		}

		// Set the line to draw based on if the sprite is flipped on the y
		line := scanline - yPos
		if yFlip {
			line = ySize - line - 1
		}

		// Load the data containing the sprite data for this line
		dataAddress := (uint16(tileLocation) * 16) + uint16(line*2) + (bank * 0x2000)
		data1 := gb.Memory.VRAM[dataAddress]
		data2 := gb.Memory.VRAM[dataAddress+1]

		// Draw the line of the sprite
		for tilePixel := byte(0); tilePixel < 8; tilePixel++ {
			pixel := int16(xPos) + int16(7-tilePixel)
			if pixel < 0 || pixel >= ScreenWidth {
				continue
			}

			// Check if the pixel has priority.
			//  - In DMG this is determined by the sprite with the smallest X coordinate,
			//    then the first sprite in the OAM.
			//  - In CGB this is determined by the first sprite appearing in the OAM.
			// We add a fixed 100 to the xPos so we can use the 0 value as the absence of a sprite.
			if minx[pixel] != 0 && (gb.IsCGB() || minx[pixel] <= xPos+spritePriorityOffset) {
				continue
			}

			colourBit := tilePixel
			if xFlip {
				colourBit = byte(int8(colourBit-7) * -1)
			}

			// Find the colour value by combining the data bits
			colourNum := (bits.Val(data2, colourBit) << 1) | bits.Val(data1, colourBit)

			// Colour 0 is transparent for sprites
			if colourNum == 0 {
				continue
			}

			if gb.IsCGB() {
				cgbPalette := attributes & 0x7
				red, green, blue := gb.SpritePalette.get(cgbPalette, colourNum)
				gb.setPixel(byte(pixel), byte(scanline), red, green, blue, priority)
			} else {
				// Determine the colour palette to use
				var palette = palette1
				if bits.Test(attributes, 4) {
					palette = palette2
				}
				red, green, blue := gb.getColour(colourNum, palette)
				gb.setPixel(byte(pixel), byte(scanline), red, green, blue, priority)
			}

			// Store the xpos of the sprite for this pixel for priority resolution
			minx[pixel] = xPos + spritePriorityOffset
		}
	}
}

// Set a pixel in the graphics screen data.
func (gb *Gameboy) setPixel(x byte, y byte, r uint8, g uint8, b uint8, priority bool) {
	// If priority is false then sprite pixel is only set if tile colour is 0
	if (priority && !gb.bgPriority[x][y]) || gb.tileScanline[x] == 0 {
		gb.screenData[x][y][0] = r
		gb.screenData[x][y][1] = g
		gb.screenData[x][y][2] = b
	}
}

// Clear the screen by setting every pixel to white.
func (gb *Gameboy) clearScreen() {
	// Check if we have cleared the screen already
	if gb.screenCleared {
		return
	}

	// Set every pixel to white
	for x := 0; x < len(gb.screenData); x++ {
		for y := 0; y < len(gb.screenData[x]); y++ {
			gb.screenData[x][y][0] = 255
			gb.screenData[x][y][1] = 255
			gb.screenData[x][y][2] = 255
		}
	}

	// Push the cleared data right now
	gb.PreparedData = gb.screenData
	gb.screenCleared = true
}
