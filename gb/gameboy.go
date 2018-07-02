package gb

import (
	"bufio"
	"fmt"
	"log"

	"encoding/gob"
	"os"

	"github.com/Humpheh/goboy/bits"
)

const (
	ClockSpeed   = 4194304
	FramesSecond = 60
	CyclesFrame  = ClockSpeed / FramesSecond

	TIMA = 0xFF05
	TMA  = 0xFF06
	TAC  = 0xFF07
)

// Gameboy is the master struct which contains all of the sub components
// for running the Gameboy emulator.
type Gameboy struct {
	options gameboyOptions

	Memory          *Memory
	CPU             *CPU
	Sound           *Sound
	TimerCounter    int
	ScanlineCounter int

	ScreenData   [160][144][3]uint8
	PreparedData [160][144][3]uint8
	// Track colour of tiles in scanline
	TileScanline [160]uint8

	InterruptsEnabling bool
	InterruptsOn       bool
	Halted             bool

	cbInst    map[byte]func()
	InputMask byte

	Debug           DebugFlags
	ExecutionPaused bool

	// Flag if the game is running in cgb mode. For this to be true the game
	// rom must support cgb mode and the option be true.
	CGBMode       bool
	BGPalette     *cgbPalette
	SpritePalette *cgbPalette

	currentSpeed byte
	prepareSpeed bool

	thisCpuTicks int
	debugScanner *bufio.Scanner
}

// Update should be called 60 times/second. This will update the state
// of the gameboy by a single frame.
func (gb *Gameboy) Update() int {
	if gb.ExecutionPaused {
		return 0
	}

	cycles := 0
	for cycles < CyclesFrame*gb.getSpeed() {
		cyclesOp := 4
		if !gb.Halted {
			if gb.Debug.OutputOpcodes {
				logOpcode(gb)
				fmt.Println(cpuStateString(gb.CPU, ""))
			}
			cyclesOp = gb.ExecuteNextOpcode()
			cycles += cyclesOp
		} else {
			// TODO: This is incorrect
			cycles += 4
		}
		if gb.IsCGB() {
			gb.checkSpeedSwitch()
		}
		gb.updateTimers(cyclesOp)
		gb.updateGraphics(cyclesOp)
		gb.doInterrupts()
	}
	//gb.Sound.Tick(cycles)

	return cycles
}

// Get the current CPU speed multiplier (either 1 or 2).
func (gb *Gameboy) getSpeed() int {
	return int(gb.currentSpeed + 1)
}

// Check if the speed needs to be switched for CGB mode.
func (gb *Gameboy) checkSpeedSwitch() {
	// TODO: This should actually happen after a STOP after asking to switch
	if gb.prepareSpeed {
		// Switch speed
		gb.prepareSpeed = false
		if gb.currentSpeed == 0 {
			gb.currentSpeed = 1
		} else {
			gb.currentSpeed = 0
		}
		log.Print("new speed", gb.currentSpeed)
		gb.Halted = false
	}
}

func (gb *Gameboy) updateTimers(cycles int) {
	gb.dividerRegister(cycles)

	if gb.isClockEnabled() {
		gb.TimerCounter -= cycles

		if gb.TimerCounter <= 0 {
			gb.TimerCounter += gb.getClockFreqCount()
			tima := gb.Memory.Read(TIMA)
			if tima == 255 {
				gb.Memory.Write(TIMA, gb.Memory.Read(TMA))
				gb.requestInterrupt(2)
			} else {
				gb.Memory.Write(TIMA, tima+1)
			}
		}
	}
}

func (gb *Gameboy) isClockEnabled() bool {
	return bits.Test(gb.Memory.Read(TAC), 2)
}

func (gb *Gameboy) getClockFreq() byte {
	return gb.Memory.Read(TAC) & 0x3
}

func (gb *Gameboy) getClockFreqCount() int {
	switch gb.getClockFreq() {
	case 0:
		return 1024
	case 1:
		return 16
	case 2:
		return 64
	default:
		return 256
	}
}

func (gb *Gameboy) setClockFreq() {
	gb.TimerCounter = gb.getClockFreqCount()
}

func (gb *Gameboy) dividerRegister(cycles int) {
	gb.CPU.Divider += cycles
	if gb.CPU.Divider >= 255 {
		gb.CPU.Divider -= 255
		gb.Memory.Data[0xFF04]++
	}
}

// RequestJoypadInterrupt triggers a joypad interrupt. To be called by the io
// binding implementation.
func (gb *Gameboy) RequestJoypadInterrupt() {
	gb.requestInterrupt(4) // Joypad interrupt
}

// Request the Gameboy to perform an interrupt.
func (gb *Gameboy) requestInterrupt(interrupt byte) {
	req := gb.Memory.Read(0xFF0F)
	req = bits.Set(req, interrupt)
	gb.Memory.Write(0xFF0F, req)
}

func (gb *Gameboy) doInterrupts() {
	if gb.InterruptsEnabling {
		gb.InterruptsOn = true
		gb.InterruptsEnabling = false
		return
	}
	if !gb.InterruptsOn && !gb.Halted {
		return
	}

	req := gb.Memory.Read(0xFF0F)
	enabled := gb.Memory.Read(0xFFFF)

	if req > 0 {
		var i byte
		for i = 0; i < 5; i++ {
			if bits.Test(req, i) && bits.Test(enabled, i) {
				gb.serviceInterrupt(i)
			}
		}
	}
}

// Address that should be jumped to by interrupt.
var interruptAddresses = map[byte]uint16{
	0: 0x40, // V-Blank
	1: 0x48, // LCDC Status
	2: 0x50, // Timer Overflow
	3: 0x58, // Serial Transfer
	4: 0x60, // Hi-Lo P10-P13
}

// Called if an interrupt has been raised. Will check if interrupts are
// enabled and will jump to the interrupt address.
func (gb *Gameboy) serviceInterrupt(interrupt byte) {
	// If was halted without interrupts, do not jump or reset IF
	if !gb.InterruptsOn && gb.Halted {
		gb.Halted = false
		return
	}
	gb.InterruptsOn = false
	gb.Halted = false

	req := gb.Memory.Read(0xFF0F)
	req = bits.Reset(req, interrupt)
	gb.Memory.Write(0xFF0F, req)

	gb.pushStack(gb.CPU.PC)
	gb.CPU.PC = interruptAddresses[interrupt]
}

// Push a 16 bit value onto the stack and decrement SP.
func (gb *Gameboy) pushStack(address uint16) {
	sp := gb.CPU.SP.HiLo()
	gb.Memory.Write(sp-1, byte(uint16(address&0xFF00)>>8))
	gb.Memory.Write(sp-2, byte(address&0xFF))
	gb.CPU.SP.Set(gb.CPU.SP.HiLo() - 2)
}

// Pop the next 16 bit value off the stack and increment SP.
func (gb *Gameboy) popStack() uint16 {
	sp := gb.CPU.SP.HiLo()
	byte1 := uint16(gb.Memory.Read(sp))
	byte2 := uint16(gb.Memory.Read(sp+1)) << 8
	gb.CPU.SP.Set(gb.CPU.SP.HiLo() + 2)
	return byte1 | byte2
}

// Update the state of the graphics.
func (gb *Gameboy) updateGraphics(cycles int) {
	gb.setLCDStatus()

	if !gb.isLCDEnabled() {
		return
	}
	gb.ScanlineCounter -= cycles

	if gb.ScanlineCounter <= 0 {
		gb.Memory.Data[0xFF44]++
		if gb.Memory.Data[0xFF44] > 153 {
			gb.PreparedData = gb.ScreenData
			gb.ScreenData = [160][144][3]uint8{}
			gb.Memory.Data[0xFF44] = 0
		}

		currentLine := gb.Memory.Read(0xFF44)
		gb.ScanlineCounter += 456 * gb.getSpeed()

		if currentLine < 144 {
			gb.drawScanline(currentLine)
		}
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
		gb.ScanlineCounter = 456
		gb.Memory.Data[0xFF44] = 0
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

	var mode byte = 0
	requestInterrupt := false

	if currentLine >= 144 {
		mode = 1
		status = bits.Set(status, 0)
		status = bits.Reset(status, 1)
		requestInterrupt = bits.Test(status, 4)
	} else {
		mode2Bounds := 456 - 80
		mode3Bounds := mode2Bounds - 172

		if gb.ScanlineCounter >= mode2Bounds {
			mode = 2
			status = bits.Reset(status, 0)
			status = bits.Set(status, 1)
			requestInterrupt = bits.Test(status, 5)
		} else if gb.ScanlineCounter >= mode3Bounds {
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

// Render a scanline of the tile map to the graphics output based
// on the state of the lcdControl register.
func (gb *Gameboy) renderTiles(lcdControl byte, scanline byte) {
	unsigned := false
	tileData := uint16(0x8800)

	scrollY := gb.Memory.Read(0xFF42)
	scrollX := gb.Memory.Read(0xFF43)
	windowY := gb.Memory.Read(0xFF4A)
	windowX := gb.Memory.Read(0xFF4B) - 7

	usingWindow := false

	if bits.Test(lcdControl, 5) {
		// is current scanline we're drawing within windows Y position?
		if windowY <= gb.Memory.Read(0xFF44) {
			usingWindow = true
		}
	}

	// Test if we're using unsigned bytes
	if bits.Test(lcdControl, 4) {
		tileData = 0x8000
		unsigned = true
	}

	var testBit byte = 3
	if usingWindow {
		testBit = 6
	}

	backgroundMemory := uint16(0x9800)
	if bits.Test(lcdControl, testBit) {
		backgroundMemory = 0x9C00
	}

	// yPos is used to calc which of 32 v-lines the current scanline is drawing
	var yPos byte = 0
	if !usingWindow {
		yPos = scrollY + scanline
	} else {
		yPos = scanline - windowY
	}

	// which of the 8 vertical pixels of the current tile is the scanline on?
	var tileRow = uint16(yPos/8) * 32

	// start drawing the 160 horizontal pixels for this scanline
	gb.TileScanline = [160]uint8{}
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
		/*
			Bit 0-2  Background Palette number  (BGP0-7)
			Bit 5    Horizontal Flip            (0=Normal, 1=Mirror horizontally)
			Bit 6    Vertical Flip              (0=Normal, 1=Mirror vertically)
			Bit 7    BG-to-OAM Priority         (0=Use OAM priority bit, 1=BG Priority
		*/
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

		// Set the pixel
		if gb.IsCGB() {
			cgbPalette := tileAttr & 0x7
			red, green, blue := gb.BGPalette.get(cgbPalette, colourNum)
			gb.setPixel(pixel, scanline, red, green, blue, true)
		} else {
			red, green, blue := gb.getColour(colourNum, 0xFF47)
			gb.setPixel(pixel, scanline, red, green, blue, true)
		}

		// Store for the current scanline so sprite priority can be managed
		gb.TileScanline[pixel] = colourNum
	}
}

// Get the RGB colour value for a colour num at an address using the current palette.
func (gb *Gameboy) getColour(colourNum byte, address uint16) (uint8, uint8, uint8) {
	var hi, lo byte = 0, 0
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
		// Load sprite data from memory. Note: for speed purposes
		// we are accessing the Data array directly instead of using
		// the read() method.
		index := sprite * 4
		yPos := int32(gb.Memory.Read(uint16(0xFE00+index))) - 16
		xPos := gb.Memory.Read(uint16(0xFE00+index+1)) - 8
		tileLocation := gb.Memory.Read(uint16(0xFE00 + index + 2))
		attributes := gb.Memory.Read(uint16(0xFE00 + index + 3))

		yFlip := bits.Test(attributes, 6)
		xFlip := bits.Test(attributes, 5)
		priority := !bits.Test(attributes, 7)

		// Bank the sprite data in is (CGB only)
		var bank uint16 = 0
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
	if priority || gb.TileScanline[x] == 0 {
		gb.ScreenData[x][y][0] = r
		gb.ScreenData[x][y][1] = g
		gb.ScreenData[x][y][2] = b
	}
}

func (gb *Gameboy) joypadValue(current byte) byte {
	var in byte = 0xF
	if bits.Test(current, 4) {
		in = gb.InputMask & 0xF
	} else if bits.Test(current, 5) {
		in = (gb.InputMask >> 4) & 0xF
	}
	return current | 0xc0 | in
}

// IsGameLoaded returns if there is a game loaded in the gameboy or not.
func (gb *Gameboy) IsGameLoaded() bool {
	return gb.Memory != nil && gb.Memory.Cart != nil
}

// IsCGB returns if we are using CGB features
func (gb *Gameboy) IsCGB() bool {
	return gb.CGBMode
}

// Initialise the Gameboy using a path to a rom.
func (gb *Gameboy) init(cart *Cartridge) {
	gb.ExecutionPaused = false

	// Initialise the CPU
	gb.CPU = &CPU{}
	gb.CPU.Init(gb.options.cgbMode)

	// Initialise the memory
	gb.Memory = &Memory{}
	gb.Memory.Init(gb)

	//gb.Sound = &Sound{}
	//gb.Sound.Init(gb)

	// Load the ROM file
	gb.Memory.LoadCart(cart)
	gb.CGBMode = gb.options.cgbMode && cart.hasCGB

	gb.Debug = DebugFlags{}
	gb.ScanlineCounter = 456
	gb.TimerCounter = 1024
	gb.InputMask = 0xFF

	gb.cbInst = gb.cbInstructions()

	gb.SpritePalette = NewPalette()
	gb.BGPalette = NewPalette()
}

func (gb *Gameboy) Gob() error {
	f, err := os.Create("test.gob")
	if err != nil {
		return err
	}
	defer f.Close()
	enc := gob.NewEncoder(f)
	err = enc.Encode(gb)
	if err != nil {
		return err
	}
	log.Print("Gob'd")
	return nil
}

// NewGameboy returns a new Gameboy instance.
func NewGameboy(romFile string, opts ...GameboyOption) (*Gameboy, error) {
	// Build the gameboy
	gameboy := Gameboy{}
	for _, opt := range opts {
		opt(&gameboy.options)
	}
	cart := &Cartridge{}
	err := cart.LoadFromFile(romFile)
	if err != nil {
		return nil, err
	}
	gameboy.init(cart)
	return &gameboy, nil
}

func NewGameboyFromBytes(bytes []byte, opts ...GameboyOption) (*Gameboy, error) {
	// Build the gameboy
	gameboy := Gameboy{}
	for _, opt := range opts {
		opt(&gameboy.options)
	}
	cart := &Cartridge{}
	err := cart.Load("test", bytes)
	if err != nil {
		return nil, err
	}
	gameboy.init(cart)
	if err != nil {
		return nil, err
	}
	return &gameboy, nil
}

func NewGameboyFromGob(gobFile string, opts ...GameboyOption) (*Gameboy, error) {
	f, err := os.Open(gobFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := gob.NewDecoder(f)
	gameboy := Gameboy{}
	for _, opt := range opts {
		opt(&gameboy.options)
	}
	err = dec.Decode(&gameboy)
	if err != nil {
		return nil, err
	}
	gameboy.Memory.gb = &gameboy
	//gameboy.Sound.Init(&gameboy)
	gameboy.cbInst = gameboy.cbInstructions()
	return &gameboy, nil
}
