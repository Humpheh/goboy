package gb

import (
	"github.com/Humpheh/goboy/bits"
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
			gb.screenData = [160][144][3]uint8{}
			gb.Memory.HighRAM[0x44] = 0
		}

		currentLine := gb.Memory.Read(0xFF44)
		gb.scanlineCounter += 456 * gb.getSpeed()

		if currentLine == 144 {
			gb.requestInterrupt(0)
		}
	}
}

// Set the status of the LCD based on the current state of memory.
func (gb *Gameboy) setLCDStatus() {
	status := gb.Memory.Read(0xFF41)

	if !gb.isLCDEnabled() {
		// TODO: Set screen to white in this instance
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

	currentLine := gb.Memory.Read(0xFF44)
	currentMode := status & 0x3

	var mode byte
	requestInterrupt := false

	if currentLine >= 144 {
		mode = 1
		status = bits.Set(status, 0)
		status = bits.Reset(status, 1)
		requestInterrupt = bits.Test(status, 4)
	} else {
		mode2Bounds := 456 - 80
		mode3Bounds := mode2Bounds - 172

		if gb.scanlineCounter >= mode2Bounds {
			if currentLine != gb.lastDrawnScanline {
				// Draw the scanline at the start of the v-blank period
				gb.drawScanline(gb.lastDrawnScanline)
				gb.lastDrawnScanline = currentLine
			}
			mode = 2
			status = bits.Reset(status, 0)
			status = bits.Set(status, 1)
			requestInterrupt = bits.Test(status, 5)
		} else if gb.scanlineCounter >= mode3Bounds {
			mode = 3
			status = bits.Set(status, 0)
			status = bits.Set(status, 1)
		} else {
			mode = 0
			status = bits.Reset(status, 0)
			status = bits.Reset(status, 1)
			requestInterrupt = bits.Test(status, 3)
			if mode != currentMode {
				gb.Memory.hbHDMATransfer()
			}
		}
	}

	if requestInterrupt && mode != currentMode {
		gb.requestInterrupt(1)
	}

	// Check is LYC == LY (coincidence flag)
	if gb.Memory.Read(0xFF44) == gb.Memory.Read(0xFF45) {
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
	return bits.Test(gb.Memory.Read(0xFF40), 7)
}

// Draw a single scanline to the graphics output.
func (gb *Gameboy) drawScanline(scanline byte) {
	control := gb.Memory.Read(0xFF40)
	if bits.Test(control, 0) && !gb.Debug.HideBackground {
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
		if windowY <= gb.Memory.Read(0xFF44) {
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
	scrollY := gb.Memory.Read(0xFF42)
	scrollX := gb.Memory.Read(0xFF43)
	windowY := gb.Memory.Read(0xFF4A)
	windowX := gb.Memory.Read(0xFF4B) - 7

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
		//	 Bit 0-2  Background Palette number  (BGP0-7)
		//	 Bit 5    Horizontal Flip            (0=Normal, 1=Mirror horizontally)
		//	 Bit 6    Vertical Flip              (0=Normal, 1=Mirror vertically)
		//	 Bit 7    BG-to-OAM Priority         (0=Use OAM priority bit, 1=BG Priority
		//
		tileAttr := gb.Memory.VRAM[tileAddress-0x6000]
		if gb.IsCGB() && bits.Test(tileAttr, 3) {
			bankOffset = 0x6000
		}

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
		gb.setTilePixel(pixel, scanline, tileAttr, colourNum)
	}
}

func (gb *Gameboy) setTilePixel(x, y, tileAttr, colourNum byte) {
	// Set the pixel
	if gb.IsCGB() {
		cgbPalette := tileAttr & 0x7 //
		red, green, blue := gb.BGPalette.get(cgbPalette, colourNum)
		gb.setPixel(x, y, red, green, blue, true)
	} else {
		red, green, blue := gb.getColour(colourNum, 0xFF47)
		gb.setPixel(x, y, red, green, blue, true)
	}

	// Store for the current scanline so sprite priority can be managed
	gb.tileScanline[x] = colourNum
}

// Get the RGB colour value for a colour num at an address using the current palette.
func (gb *Gameboy) getColour(colourNum byte, address uint16) (uint8, uint8, uint8) {
	var hi, lo byte
	switch colourNum {
	case 0:
		hi, lo = 1, 0
	case 1:
		hi, lo = 3, 2
	case 2:
		hi, lo = 5, 4
	case 3:
		hi, lo = 7, 6
	}

	palette := gb.Memory.Read(address)
	col := (bits.Val(palette, hi) << 1) | bits.Val(palette, lo)

	return GetPaletteColour(col)
}

// Render the sprites to the screen on the current scanline using the lcdControl register.
func (gb *Gameboy) renderSprites(lcdControl byte, scanline int32) {
	var ySize int32 = 8
	if bits.Test(lcdControl, 2) {
		ySize = 16
	}

	for sprite := 0; sprite < 40; sprite++ {
		// Load sprite data from memory.
		index := sprite * 4
		yPos := int32(gb.Memory.Read(uint16(0xFE00+index))) - 16
		xPos := int32(gb.Memory.Read(uint16(0xFE00+index+1))) - 8
		tileLocation := gb.Memory.Read(uint16(0xFE00 + index + 2))
		attributes := gb.Memory.Read(uint16(0xFE00 + index + 3))

		yFlip := bits.Test(attributes, 6)
		xFlip := bits.Test(attributes, 5)
		priority := !bits.Test(attributes, 7)

		// Bank the sprite data in is (CGB only)
		var bank uint16
		if gb.IsCGB() {
			bank = uint16((attributes & 0x8) >> 3)
		}

		// If this is true the scanline is out of the area we care about
		if scanline < yPos || scanline >= (yPos+ySize) {
			continue
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
			pixel := int16(xPos) + int16(7-tilePixel)

			// Set the pixel if it is in bounds
			if pixel >= 0 && pixel < 160 {
				if gb.IsCGB() {
					cgbPalette := attributes & 0x7
					red, green, blue := gb.SpritePalette.get(cgbPalette, colourNum)
					gb.setPixel(byte(pixel), byte(scanline), red, green, blue, priority)
				} else {
					// Determine the colour palette to use
					var colourAddress uint16 = 0xFF48
					if bits.Test(attributes, 4) {
						colourAddress = 0xFF49
					}
					red, green, blue := gb.getColour(colourNum, colourAddress)
					gb.setPixel(byte(pixel), byte(scanline), red, green, blue, priority)
				}
			}
		}
	}
}

// Set a pixel in the graphics screen data.
func (gb *Gameboy) setPixel(x byte, y byte, r uint8, g uint8, b uint8, priority bool) {
	// If priority is false then sprite pixel is only set if tile colour is 0
	if priority || gb.tileScanline[x] == 0 {
		gb.screenData[x][y][0] = r
		gb.screenData[x][y][1] = g
		gb.screenData[x][y][2] = b
	}
}
