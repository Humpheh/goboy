package gb

import "sort"

const (
	// ScreenWidth is the number of pixels width on the GameBoy LCD panel.
	ScreenWidth = 160

	// ScreenHeight is the number of pixels height on the GameBoy LCD panel.
	ScreenHeight = 144

	// LCDC is the main LCD Control register.
	LCDC = 0xFF40
)

// PPU timings, in dots from the start of a scanline.
const (
	dotsPerLine = 456
	// Dot at which mode 2 shows in STAT on lines 0-143. The mode 2 STAT
	// interrupt is raised at the start of the line, before this.
	mode2StartDot = 1
	// Dot at which mode 3 starts on lines 0-143.
	mode3StartDot = mode2StartDot + 80
	// Length of mode 3 before it is extended by scrolling, the window and sprites.
	mode3BaseLength = 172
	// The STAT mode 0 interrupt is raised this many dots before mode 0 shows in STAT.
	mode0IrqLead = 1
	// Dot at which LY=LYC is compared after LY changes.
	lycCompareDot = 4
	// Dot on the first line after the LCD is turned on that the PPU starts from.
	lcdOnStartDot = 4
)

// ppu holds the PPU timing and rendering state.
type ppu struct {
	// Matrix of pixel data which is used while the screen is rendering. When a
	// frame has been completed, this data is copied into the PreparedData matrix.
	screenData [ScreenWidth][ScreenHeight][3]uint8
	bgPriority [ScreenWidth][ScreenHeight]bool

	// Track colour of tiles in scanline for priority management.
	tileScanline  [ScreenWidth]uint8
	screenCleared bool

	// Scanline being processed, which only differs from LY on line 153.
	line int
	// Dot within the current scanline.
	dot int
	// Dot of the next event on the current scanline.
	eventDot int
	// Dot on the current line at which mode 3 ends.
	mode3End int
	// Set for the first line after the LCD is turned on.
	lcdFirstLine bool
	// LY value being compared with LYC, or -1 if no comparison is being made.
	lycCompareLY int
	// STAT interrupt conditions which don't match the mode shown in STAT.
	statMode0Irq bool
	statMode2Irq bool
	// Level of the STAT interrupt line. An interrupt is requested when it rises.
	statIrqLine bool
}

// Init sets the default PPU values.
func (ppu *ppu) Init() {
	// The boot ROM hands over on line 153, after LY has started reading 0
	ppu.line = 153
	ppu.dot = 400
	ppu.eventDot = dotsPerLine
}

// Update the state of the graphics by advancing the PPU one dot at a time.
func (gb *Gameboy) updateGraphics(cycles int) {
	if !gb.isLCDEnabled() {
		return
	}
	for dots := cycles / gb.getSpeed(); dots > 0; dots-- {
		gb.ppu.dot++
		if gb.ppu.dot == gb.ppu.eventDot {
			gb.ppuEvent()
		}
	}
	gb.updateStat()
}

// ppuEvent handles the PPU reaching ppuEventDot on the current line, and then
// schedules the next event.
func (gb *Gameboy) ppuEvent() {
	if gb.ppu.dot == dotsPerLine {
		gb.startLine()
	}

	dot := gb.ppu.dot
	switch {
	case gb.ppu.line < ScreenHeight:
		if dot == lycCompareDot {
			gb.ppu.lycCompareLY = int(gb.memory.HighRAM[0x44])
		}
		if dot == mode2StartDot && !gb.ppu.lcdFirstLine {
			gb.setMode(2)
		}
		if dot == mode3StartDot {
			gb.ppu.statMode2Irq = false
			gb.setMode(3)
			gb.ppu.mode3End = mode3StartDot + gb.mode3Length()
			// Draw the whole scanline at the start of mode 3. In the real
			// GameBoy this is done throughout mode 3 by reading OAM and VRAM.
			gb.drawScanline(byte(gb.ppu.line))
		}
		if dot == gb.ppu.mode3End-mode0IrqLead {
			gb.ppu.statMode0Irq = true
		}
		if dot == gb.ppu.mode3End {
			gb.setMode(0)
			gb.memory.doHDMATransfer()
		}

	case dot == lycCompareDot:
		gb.ppu.lycCompareLY = int(gb.memory.HighRAM[0x44])
		gb.ppu.statMode2Irq = false
		if gb.ppu.line == 153 {
			// LY reads 0 for the rest of line 153, but is compared with LYC
			// as 153 for one more M-cycle
			gb.memory.HighRAM[0x44] = 0
		}

	case dot == lycCompareDot+4:
		gb.ppu.lycCompareLY = int(gb.memory.HighRAM[0x44])
	}

	gb.updateStat()
	gb.ppu.eventDot = gb.nextPPUEventDot()
}

// startLine moves the PPU on to the start of the next scanline.
func (gb *Gameboy) startLine() {
	gb.ppu.dot = 0
	gb.ppu.lcdFirstLine = false
	gb.ppu.statMode0Irq = false
	gb.ppu.line++
	if gb.ppu.line == 154 {
		gb.ppu.line = 0
		gb.PreparedData = gb.ppu.screenData
		gb.ppu.screenData = [ScreenWidth][ScreenHeight][3]uint8{}
		gb.ppu.bgPriority = [ScreenWidth][ScreenHeight]bool{}
	}

	// LY=LYC isn't compared for a few dots after LY changes
	if gb.memory.HighRAM[0x44] != byte(gb.ppu.line) {
		gb.memory.HighRAM[0x44] = byte(gb.ppu.line)
		gb.ppu.lycCompareLY = -1
	}

	switch {
	case gb.ppu.line < ScreenHeight:
		gb.ppu.statMode2Irq = true
	case gb.ppu.line == ScreenHeight:
		gb.setMode(1)
		gb.requestInterrupt(0)
		// On DMG the mode 2 STAT interrupt is also raised when VBlank starts
		gb.ppu.statMode2Irq = !gb.IsCGB()
	}
}

// nextPPUEventDot returns the dot of the next PPU event on the current line.
func (gb *Gameboy) nextPPUEventDot() int {
	next := dotsPerLine
	events := [...]int{mode2StartDot, lycCompareDot, lycCompareDot + 4, mode3StartDot, gb.ppu.mode3End - mode0IrqLead, gb.ppu.mode3End}
	for _, event := range events {
		if event > gb.ppu.dot && event < next {
			next = event
		}
	}
	return next
}

// mode3Length returns the number of dots mode 3 takes on the current line,
// which is extended by fine scrolling, the window and each sprite drawn.
func (gb *Gameboy) mode3Length() int {
	control := gb.memory.HighRAM[0x40]
	scx := int(gb.memory.HighRAM[0x43] & 7)
	length := mode3BaseLength + scx

	// The window being drawn pauses the pixel fetcher
	if bitTest(control, 5) && int(gb.memory.HighRAM[0x4A]) <= gb.ppu.line && gb.memory.HighRAM[0x4B] <= 166 {
		length += 6
	}

	if !bitTest(control, 1) {
		return length
	}

	// Find the sprites on this line, as selected by the OAM scan
	height := 8
	if bitTest(control, 2) {
		height = 16
	}
	var spriteXs []int
	for i := 0; i < 0xA0 && len(spriteXs) < 10; i += 4 {
		y := int(gb.memory.OAM[i]) - 16
		if gb.ppu.line >= y && gb.ppu.line < y+height {
			spriteXs = append(spriteXs, int(gb.memory.OAM[i+1]))
		}
	}
	sort.Ints(spriteXs)

	// Each sprite costs 6 dots, plus up to 5 more for the first sprite in each
	// background tile depending on how far into the tile it starts
	var tileFetched [22]bool
	for _, x := range spriteXs {
		if x >= 168 {
			continue
		}
		pixel := x + scx
		if x == 0 {
			pixel = 0
		}
		length += 6
		if tile := pixel / 8; !tileFetched[tile] {
			tileFetched[tile] = true
			if extra := 5 - pixel%8; extra > 0 {
				length += extra
			}
		}
	}
	return length
}

// setMode sets the PPU mode shown in STAT.
func (gb *Gameboy) setMode(mode byte) {
	gb.memory.HighRAM[0x41] = gb.memory.HighRAM[0x41]&^0x03 | mode
}

// updateStat updates the LY=LYC flag in STAT and requests a STAT interrupt on
// the rising edge of the STAT interrupt line.
func (gb *Gameboy) updateStat() {
	stat := gb.memory.HighRAM[0x41]
	if gb.ppu.lycCompareLY == int(gb.memory.HighRAM[0x45]) {
		stat = bitSet(stat, 2)
	} else {
		stat = bitReset(stat, 2)
	}
	gb.memory.HighRAM[0x41] = stat

	mode := stat & 0x03
	line := (gb.ppu.statMode0Irq && bitTest(stat, 3)) ||
		(mode == 1 && bitTest(stat, 4)) ||
		(gb.ppu.statMode2Irq && bitTest(stat, 5)) ||
		(bitTest(stat, 2) && bitTest(stat, 6))
	if line && !gb.ppu.statIrqLine {
		gb.requestInterrupt(1)
	}
	gb.ppu.statIrqLine = line
}

// lcdOn starts the PPU from the first line when the LCD is turned on. The
// first line skips mode 2.
func (gb *Gameboy) lcdOn() {
	gb.ppu.line = 0
	gb.ppu.dot = lcdOnStartDot
	gb.ppu.lcdFirstLine = true
	gb.ppu.lycCompareLY = 0
	gb.ppu.screenCleared = false
	gb.ppu.eventDot = gb.nextPPUEventDot()
	gb.updateStat()
}

// lcdOff stops the PPU when the LCD is turned off. The LY=LYC flag keeps the
// value it had.
func (gb *Gameboy) lcdOff() {
	gb.ppu.line = 0
	gb.ppu.dot = 0
	gb.memory.HighRAM[0x44] = 0
	gb.setMode(0)
	gb.ppu.statMode0Irq = false
	gb.ppu.statMode2Irq = false
	gb.clearScreen()
}

// ppuLocked returns if the PPU is blocking the cpu from accessing the address,
// which happens for OAM during modes 2 and 3 and for VRAM during mode 3.
func (gb *Gameboy) ppuLocked(address uint16) bool {
	if address < 0x8000 || address >= 0xFEA0 || (address >= 0xA000 && address < 0xFE00) {
		return false
	}
	mode := gb.memory.HighRAM[0x41] & 0x03
	if address < 0xA000 {
		return mode == 3
	}
	// OAM is locked from the start of the line, slightly before mode 2 shows
	return mode >= 2 || (gb.ppu.statMode2Irq && gb.ppu.line < ScreenHeight)
}

// Checks if the LCD is enabled by examining 0xFF40.
func (gb *Gameboy) isLCDEnabled() bool {
	return bitTest(gb.memory.HighRAM[0x40], 7)
}

// Draw a single scanline to the graphics output.
func (gb *Gameboy) drawScanline(scanline byte) {
	control := gb.memory.ReadHighRam(LCDC)

	// LCDC bit 0 clears tiles on DMG but controls priority on CGB.
	if (gb.IsCGB() || bitTest(control, 0)) && !gb.Debug.HideBackground {
		gb.renderTiles(control, scanline)
	}

	if bitTest(control, 1) && !gb.Debug.HideSprites {
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

	if bitTest(lcdControl, 5) {
		// Is current scanline we're drawing within windows Y position?
		if windowY <= gb.memory.ReadHighRam(0xFF44) {
			usingWindow = true
		}
	}

	// Test if we're using unsigned bytes
	if bitTest(lcdControl, 4) {
		tileData = 0x8000
		unsigned = true
	}

	// Work out where to look in background memory.
	var testBit byte = 3
	if usingWindow {
		testBit = 6
	}
	backgroundMemory = uint16(0x9800)
	if bitTest(lcdControl, testBit) {
		backgroundMemory = 0x9C00
	}
	return
}

// Render a scanline of the tile map to the graphics output based
// on the state of the lcdControl register.
func (gb *Gameboy) renderTiles(lcdControl byte, scanline byte) {
	scrollY := gb.memory.ReadHighRam(0xFF42)
	scrollX := gb.memory.ReadHighRam(0xFF43)
	windowY := gb.memory.ReadHighRam(0xFF4A)
	windowX := gb.memory.ReadHighRam(0xFF4B) - 7

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
	var palette = gb.memory.ReadHighRam(0xFF47)

	// start drawing the 160 horizontal pixels for this scanline
	gb.ppu.tileScanline = [ScreenWidth]uint8{}
	for pixel := range byte(ScreenWidth) {
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
			tileNum := int16(gb.memory.VRAM[tileAddress-0x8000])
			tileLocation = tileLocation + uint16(tileNum*16)
		} else {
			tileNum := int16(int8(gb.memory.VRAM[tileAddress-0x8000]))
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
		tileAttr := gb.memory.VRAM[tileAddress-0x6000]
		if gb.IsCGB() && bitTest(tileAttr, 3) {
			bankOffset = 0x6000
		}
		priority := bitTest(tileAttr, 7)

		var line byte
		if gb.IsCGB() && bitTest(tileAttr, 6) {
			// Vertical flip
			line = ((7 - yPos) % 8) * 2
		} else {
			line = (yPos % 8) * 2
		}
		// Get the tile data from memory
		data1 := gb.memory.VRAM[tileLocation+uint16(line)-bankOffset]
		data2 := gb.memory.VRAM[tileLocation+uint16(line)+1-bankOffset]

		if gb.IsCGB() && bitTest(tileAttr, 5) {
			// Horizontal flip
			xPos = 7 - xPos
		}
		colourBit := byte(int8((xPos%8)-7) * -1)
		colourNum := (bitGet(data2, colourBit) << 1) | bitGet(data1, colourBit)
		gb.setTilePixel(pixel, scanline, tileAttr, colourNum, palette, priority)
	}
}

func (gb *Gameboy) setTilePixel(x, y, tileAttr, colourNum, palette byte, priority bool) {
	// Set the pixel
	if gb.IsCGB() {
		cgbPalette := tileAttr & 0x7
		red, green, blue := gb.bgPalette.get(cgbPalette, colourNum)
		gb.setPixel(x, y, red, green, blue, true)
		gb.ppu.bgPriority[x][y] = priority
	} else {
		red, green, blue := gb.getColour(colourNum, palette)
		gb.setPixel(x, y, red, green, blue, true)
	}

	// Store for the current scanline so sprite priority can be managed
	gb.ppu.tileScanline[x] = colourNum
}

// Get the RGB colour value for a colour num at an address using the current palette.
func (gb *Gameboy) getColour(colourNum byte, palette byte) (uint8, uint8, uint8) {
	hi := colourNum<<1 | 1
	lo := colourNum << 1
	col := (bitGet(palette, hi) << 1) | bitGet(palette, lo)
	return GetPaletteColour(col)
}

const spritePriorityOffset = 100

// Render the sprites to the screen on the current scanline using the lcdControl register.
func (gb *Gameboy) renderSprites(lcdControl byte, scanline int32) {
	var ySize int32 = 8
	if bitTest(lcdControl, 2) {
		ySize = 16
	}

	// Load the two palettes which sprites can be drawn in
	var palette1 = gb.memory.ReadHighRam(0xFF48)
	var palette2 = gb.memory.ReadHighRam(0xFF49)

	var minx [ScreenWidth]int32
	var lineSprites = 0
	for sprite := uint16(0); sprite < 40; sprite++ {
		// Load sprite data from memory.
		index := sprite * 4

		// If this is true the scanline is out of the area we care about
		yPos := int32(gb.memory.Read(uint16(0xFE00+index))) - 16
		if scanline < yPos || scanline >= (yPos+ySize) {
			continue
		}

		// Only 10 sprites are allowed to be displayed on each line
		if lineSprites >= 10 {
			break
		}
		lineSprites++

		xPos := int32(gb.memory.Read(uint16(0xFE00+index+1))) - 8
		tileLocation := gb.memory.Read(uint16(0xFE00 + index + 2))
		attributes := gb.memory.Read(uint16(0xFE00 + index + 3))

		yFlip := bitTest(attributes, 6)
		xFlip := bitTest(attributes, 5)
		priority := !bitTest(attributes, 7)

		// Bank the sprite data in is (CGB only)
		var bank uint16 = 0
		if gb.IsCGB() && bitTest(attributes, 3) {
			bank = 1
		}

		// Set the line to draw based on if the sprite is flipped on the y
		line := scanline - yPos
		if yFlip {
			line = ySize - line - 1
		}

		// Load the data containing the sprite data for this line
		dataAddress := (uint16(tileLocation) * 16) + uint16(line*2) + (bank * 0x2000)
		data1 := gb.memory.VRAM[dataAddress]
		data2 := gb.memory.VRAM[dataAddress+1]

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
			colourNum := (bitGet(data2, colourBit) << 1) | bitGet(data1, colourBit)

			// Colour 0 is transparent for sprites
			if colourNum == 0 {
				continue
			}

			if gb.IsCGB() {
				cgbPalette := attributes & 0x7
				red, green, blue := gb.spritePalette.get(cgbPalette, colourNum)
				gb.setPixel(byte(pixel), byte(scanline), red, green, blue, priority)
			} else {
				// Determine the colour palette to use
				var palette = palette1
				if bitTest(attributes, 4) {
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
	if (priority && !gb.ppu.bgPriority[x][y]) || gb.ppu.tileScanline[x] == 0 {
		gb.ppu.screenData[x][y][0] = r
		gb.ppu.screenData[x][y][1] = g
		gb.ppu.screenData[x][y][2] = b
	}
}

// Clear the screen by setting every pixel to white.
func (gb *Gameboy) clearScreen() {
	// Check if we have cleared the screen already
	if gb.ppu.screenCleared {
		return
	}

	// Set every pixel to white
	for x := 0; x < len(gb.ppu.screenData); x++ {
		for y := 0; y < len(gb.ppu.screenData[x]); y++ {
			gb.ppu.screenData[x][y][0] = 255
			gb.ppu.screenData[x][y][1] = 255
			gb.ppu.screenData[x][y][2] = 255
		}
	}

	// Push the cleared data right now
	gb.PreparedData = gb.ppu.screenData
	gb.ppu.screenCleared = true
}
