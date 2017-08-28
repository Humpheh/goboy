package gob

import (
	"github.com/humpheh/gob/bits"
	"os"
	"bufio"
	"strings"
	"strconv"
)

const (
	CLOCK_SPEED   = 4194304
	FRAMES_SECOND = 60
	CYCLES_FRAME  = CLOCK_SPEED / FRAMES_SECOND
)

const (
	TIMA = 0xFF05
	TMA  = 0xFF06
	TMC  = 0xFF07

	WHITE = 1
	LIGHT_GRAY = 2
	DARK_GRAY = 3
	BLACK = 4
)

type Gameboy struct {
	Memory       *Memory
	CPU 		 *CPU
	TimerCounter int
	ScanlineCounter int

	ScreenData [160][144][3]int

	InterruptsOn bool

	CBInst map[byte]func()

	scanner *bufio.Scanner
	waitscan bool
}

// Should be called 60 times/second
func (gb *Gameboy) Update() int {
	cycles := 0
	for cycles < CYCLES_FRAME {
		cycles_op := gb.ExecuteNextOpcode()
		cycles += cycles_op

		gb.UpdateTimers(cycles_op)
		gb.UpdateGraphics(cycles_op)
		gb.DoInterrupts()
	}
	return cycles
}

func (gb *Gameboy) GetDebugNum() uint16 {
	if !gb.waitscan {
		gb.scanner.Scan()
	}
	line := gb.scanner.Text()

	split := strings.Split(line, " ")
	num, _ := strconv.ParseUint(split[1], 16, 16)

	return uint16(num)
}

func (gb *Gameboy) UpdateTimers(cycles int) {
	gb.dividerRegister(cycles)

	if gb.isClockEnabled() {
		gb.TimerCounter -= cycles

		if gb.TimerCounter <= 0 {
			gb.GetClockFreq()

			if gb.Memory.Read(TIMA) == 255 {
				gb.Memory.Write(TIMA, gb.Memory.Read(TMA))
				gb.RequestInterrupt(2)
			} else {
				gb.Memory.Write(TIMA, gb.Memory.Read(TIMA) + 1)
			}
		}
	}
}

func (gb *Gameboy) isClockEnabled() bool {
	return bits.Test(gb.Memory.Read(TMC), 2)
}

func (gb *Gameboy) GetClockFreq() byte {
	return gb.Memory.Read(TMC) & 0x3
}

func (gb *Gameboy) SetClockFreq() {
	// Set the frequency of the timer
	switch gb.GetClockFreq() {
	case 0: gb.TimerCounter = 1024
	case 1: gb.TimerCounter = 16
	case 2: gb.TimerCounter = 64
	case 3: gb.TimerCounter = 256
	}
}

func (gb *Gameboy) dividerRegister(cycles int) {
	gb.CPU.Divider += cycles
	if gb.CPU.Divider >= 255 {
		gb.CPU.Divider = 0
		gb.Memory.Data[0xFF04]++
	}
}

func (gb *Gameboy) RequestInterrupt(interrupt byte) {
	req := gb.Memory.Read(0xFF0F)
	req = bits.Set(req, interrupt)
	gb.Memory.Write(0xFF0F, req)
}

func (gb *Gameboy) DoInterrupts() {
	if !gb.InterruptsOn {
		return
	}

	req := gb.Memory.Read(0xFF0F)
	enabled := gb.Memory.Read(0xFFFF)

	if req > 0 {
		var i byte
		for i = 0; i < 5; i++ {
			if bits.Test(req, i) {
				if bits.Test(enabled, i) {
					gb.ServiceInterrupt(i)
				}
			}
		}
	}
}

func (gb *Gameboy) ServiceInterrupt(interrupt byte) {
	gb.InterruptsOn = false

	req := gb.Memory.Read(0xFF0F)
	req = bits.Reset(req, interrupt)
	gb.Memory.Write(0xFF0F, req)

	gb.PushStack(gb.CPU.PC)

	switch interrupt {
	case 0: gb.CPU.PC = 0x40
	case 1: gb.CPU.PC = 0x48
	case 2: gb.CPU.PC = 0x50
	case 4: gb.CPU.PC = 0x60
	}
}

func (gb *Gameboy) PushStack(address uint16) {
	sp := gb.CPU.SP.HiLo()
	gb.Memory.Data[sp - 1] = byte(uint16(address & 0xFF00) >> 8)
	gb.Memory.Data[sp - 2] = byte(address & 0xFF)
	gb.CPU.SP.Set(gb.CPU.SP.HiLo() - 2)
}

func (gb *Gameboy) PopStack() uint16 {
	sp := gb.CPU.SP.HiLo()
	// TODO: Check this works!
	byte1 := uint16(gb.Memory.Data[sp])
	byte2 := uint16(gb.Memory.Data[sp + 1]) << 8
	gb.CPU.SP.Set(gb.CPU.SP.HiLo() + 2)
	return byte1 | byte2
}

func (gb *Gameboy) UpdateGraphics(cycles int) {
	gb.setLCDStatus()

	if !gb.IsLCDEnabled() {
		return
	}
	gb.ScanlineCounter -= cycles

	if gb.ScanlineCounter <= 0 {
		gb.Memory.Data[0xFF44]++
		current_line := gb.Memory.Read(0xFF44)

		// TODO: This could be -456?
		gb.ScanlineCounter = 0

		if current_line == 144 {
			gb.RequestInterrupt(0)
		} else if current_line > 153 {
			gb.Memory.Data[0xFF44] = 0
		} else if current_line < 144 {
			gb.DrawScanline()
		}
	}
}

func (gb *Gameboy) setLCDStatus() {
	status := gb.Memory.Read(0xFF41)
	if !gb.IsLCDEnabled() {
		gb.ScanlineCounter = 456
		gb.Memory.Data[0xFF44] = 0
		status &= 252
		status = bits.Set(status, 0)
		gb.Memory.Write(0xFF41, status)
		return
	}

	current_line := gb.Memory.Read(0xFF44)
	current_mode := status & 0x3

	var mode byte = 0
	request_interrupt := false

	if current_line >= 144 {
		mode = 1
		status = bits.Set(status, 0)
		status = bits.Reset(status, 1)
		request_interrupt = bits.Test(status, 4)
	} else {
		mode2_bounds := 456 - 80
		mode3_bounds := mode2_bounds - 172

		if gb.ScanlineCounter >= mode2_bounds {
			mode = 2
			status = bits.Set(status, 1)
			status = bits.Reset(status, 0)
			request_interrupt = bits.Test(status, 5)
		} else if gb.ScanlineCounter >= mode3_bounds {
			mode = 3
			status = bits.Set(status, 1)
			status = bits.Set(status, 0)
		} else {
			mode = 0
			status = bits.Reset(status, 1)
			status = bits.Reset(status, 0)
			request_interrupt = bits.Test(status, 3)
		}
	}

	if request_interrupt && mode != current_mode {
		gb.RequestInterrupt(1)
	}

	if gb.Memory.Read(0xFF44) == gb.Memory.Read(0xFF45) {
		status = bits.Set(status, 2)
		if bits.Test(status, 6) {
			gb.RequestInterrupt(1)
		}
	} else {
		status = bits.Reset(status, 2)
	}

	gb.Memory.Write(0xFF41, status)
}

func (gb *Gameboy) IsLCDEnabled() bool {
	return bits.Test(gb.Memory.Read(0xFF40), 7)
}

func (gb *Gameboy) DrawScanline() {
	control := gb.Memory.Read(0xFF40)
	if bits.Test(control, 0 ) {
		gb.RenderTiles(control)
	}

	if bits.Test(control, 1) {
		gb.RenderSprites(control)
	}
}

func (gb *Gameboy) RenderTiles(lcdControl byte) {
	var tile_data uint16 = 0
	var background_memory uint16 = 0
	unsig := true

	scroll_y := gb.Memory.Read(0xFF42)
	scroll_x := gb.Memory.Read(0xFF43)
	window_y := gb.Memory.Read(0xFF4A)
	window_x := gb.Memory.Read(0xFF4B) - 7

	using_window := false

	if bits.Test(lcdControl, 5) {
		// is current scanline we're drawing within windows Y position?
		if window_y <= gb.Memory.Read(0xFF44) {
			using_window = true
		}
	}

	if bits.Test(lcdControl, 4) {
		tile_data = 0x8000
	} else {
		// This memory region uses signed bytes as tile identifiers
		tile_data = 0x8800
		unsig = false
	}

	var test_bit byte = 3
	if using_window {
		test_bit = 6
	}
	if bits.Test(lcdControl, test_bit) {
		background_memory = 0x9C00
	} else {
		background_memory = 0x9800
	}

	// yPos is used to calc which of 32 v-lines the current scanline is drawing
	var y_pos byte = 0
	if !using_window {
		y_pos = scroll_y + gb.Memory.Read(0xFF44)
	} else {
		y_pos = gb.Memory.Read(0xFF44) - window_y
	}

	// which of the 8 vertical pixels of the current tile is the scanline on?
	var tile_row uint16 = uint16(byte(y_pos / 8) * 32)

	// start drawing the 160 horizontal pixels for this scanline
	var pixel byte
	for pixel = 0; pixel < 160; pixel++ {
		x_pos := pixel + scroll_x

		// translate the current x pos to window space if necessary
		if using_window && pixel >= window_x {
			x_pos = pixel - window_x
		}

		// which of the 32 horizontal tiles does this x_pox fall within?
		tile_col := uint16(x_pos / 8)
		var tile_num int16 // signed

		// get the tile identity number
		tile_address := background_memory + tile_row + tile_col
		// TODO: ensure signing works correctly
		if unsig {
			tile_num = int16(gb.Memory.Read(tile_address))
		} else {
			tile_num = int16(int8(gb.Memory.Read(tile_address)))
		}

		// deduce where this tile id is in memory
		tile_location := tile_data
		// TODO: ensure signing works correctly
		if unsig {
			tile_location = tile_location + uint16(tile_num * 16)
		} else {
			tile_location = uint16(int16(tile_location) + ((tile_num + 128) * 16))
		}

		// find the correct v-line we're on of the tile to get the tile data from in memory (???)
		var line byte = (y_pos % 8) * 2 // each v-line takes up two bytes of memory
		data1 := gb.Memory.Read(tile_location + uint16(line))
		data2 := gb.Memory.Read(tile_location + uint16(line) + 1)

		colour_bit := byte(int8((x_pos % 8) - 7) * -1)

		colour_num := (bits.Val(data2, colour_bit) << 1) | bits.Val(data1, colour_bit)
		col := gb.GetColour(colour_num, 0xFF47)
		red, green, blue := 0, 0, 0

		// setup the RGB values
		switch col {
		case WHITE: red, green, blue = 255, 255, 255
		case LIGHT_GRAY: red, green, blue = 0xCC, 0xCC, 0xCC
		case DARK_GRAY: red, green, blue = 0x77, 0x77, 0x77
		}

		finally := gb.Memory.Read(0xFF44)

		// safety check to make sure about to set is in bounds
		if finally < 0 || finally > 143 || pixel < 0 || pixel > 159 {
			continue
		}

		gb.ScreenData[pixel][finally][0] = red
		gb.ScreenData[pixel][finally][1] = green
		gb.ScreenData[pixel][finally][2] = blue
	}
}

func (gb *Gameboy) GetColour(colour_num byte, address uint16) int {
	res := WHITE
	palette := gb.Memory.Read(address)
	var hi, lo byte = 0, 0

	switch colour_num {
	case 0: hi, lo = 1, 0
	case 1: hi, lo = 3, 2
	case 2: hi, lo = 5, 4
	case 3: hi, lo = 7, 6
	}

	colour := (bits.Val(palette, hi) << 1) | bits.Val(palette, lo)

	switch colour {
	case 0: res = WHITE
	case 1: res = LIGHT_GRAY
	case 2: res = DARK_GRAY
	case 3: res = BLACK
	}

	return res
}

func (gb *Gameboy) RenderSprites(lcdControl byte) {
	use8x16 := false
	if bits.Test(lcdControl, 2) {
		use8x16 = true
	}

	for sprite := 0; sprite < 40; sprite++ {
		index := sprite * 4
		y_pos := gb.Memory.Read(uint16(0xFE00 + index)) - 16
		x_pos := gb.Memory.Read(uint16(0xFE00 + index + 1)) - 8
		tile_location := gb.Memory.Read(uint16(0xFE00 + index + 2))
		attributes := gb.Memory.Read(uint16(0xFE00 + index + 3))

		y_flip := bits.Test(attributes, 6)
		x_flip := bits.Test(attributes, 5)

		scanline := gb.Memory.Read(0xFF44)

		var y_size byte = 8
		if use8x16 {
			y_size = 16
		}

		if scanline >= y_pos && scanline < (y_pos + y_size) {
			line := scanline - y_pos

			if y_flip {
				line = byte(int8(line - y_size) * -1)
			}

			line *= 2
			data_address := 0x8000 + uint16((tile_location * 16) + line)
			data1 := gb.Memory.Read(data_address)
			data2 := gb.Memory.Read(data_address + 1)

			var tile_pixel byte
			for tile_pixel = 7; tile_pixel >= 0; tile_pixel-- {
				colour_bit := tile_pixel

				if x_flip {
					colour_bit = byte(int8(colour_bit - 7) * -1)
				}

				colour_num := (bits.Val(data2, colour_bit) << 1) | bits.Val(data1, colour_bit)

				var colour_address uint16 = 0xFF48
				if bits.Test(attributes, 4) {
					colour_address = 0xFF49
				}
				col := gb.GetColour(colour_num, colour_address)

				// white is transparent for sprites
				if col == WHITE {
					continue
				}

				red, green, blue := 0, 0, 0
				switch col {
				case LIGHT_GRAY: red, green, blue = 0xCC, 0xCC, 0xCC
				case DARK_GRAY: red, green, blue = 0x77, 0x77, 0x77
				}

				var x_pix byte = 7 - tile_pixel
				pixel := x_pos + x_pix

				// safety check to make sure about to set is in bounds
				if scanline < 0 || scanline > 143 || pixel < 0 || pixel > 159 {
					continue
				}

				gb.ScreenData[pixel][scanline][0] = red
				gb.ScreenData[pixel][scanline][1] = green
				gb.ScreenData[pixel][scanline][2] = blue
			}
		}
	}
}

func (gb *Gameboy) Init() {
	// Load debug file
	file, err := os.Open("/Users/humphreyshotton/Documents/gb-inst.log")
	if err != nil {
		panic(err)
	}

	gb.scanner = bufio.NewScanner(file)
	gb.scanner.Split(bufio.ScanLines)
	// ^ TEMP ^

	gb.ScanlineCounter = 0//456

	gb.CBInst = gb.CBInstructions()
	gb.CPU.AF.isAF = true

	gb.CPU.PC = 0x100
	gb.CPU.AF.Set(0x01B0)
	gb.CPU.BC.Set(0x0013)
	gb.CPU.DE.Set(0x00D8)
	gb.CPU.HL.Set(0x014D)
	gb.CPU.SP.Set(0xFFFE)
	gb.Memory.Data[0xFF05] = 0x00
	gb.Memory.Data[0xFF06] = 0x00
	gb.Memory.Data[0xFF07] = 0x00
	gb.Memory.Data[0xFF10] = 0x80
	gb.Memory.Data[0xFF11] = 0xBF
	gb.Memory.Data[0xFF12] = 0xF3
	gb.Memory.Data[0xFF14] = 0xBF
	gb.Memory.Data[0xFF16] = 0x3F
	gb.Memory.Data[0xFF17] = 0x00
	gb.Memory.Data[0xFF19] = 0xBF
	gb.Memory.Data[0xFF1A] = 0x7F
	gb.Memory.Data[0xFF1B] = 0xFF
	gb.Memory.Data[0xFF1C] = 0x9F
	gb.Memory.Data[0xFF1E] = 0xBF
	gb.Memory.Data[0xFF20] = 0xFF
	gb.Memory.Data[0xFF21] = 0x00
	gb.Memory.Data[0xFF22] = 0x00
	gb.Memory.Data[0xFF23] = 0xBF
	gb.Memory.Data[0xFF24] = 0x77
	gb.Memory.Data[0xFF25] = 0xF3
	gb.Memory.Data[0xFF26] = 0xF1
	gb.Memory.Data[0xFF40] = 0x91
	gb.Memory.Data[0xFF41] = 0x85
	gb.Memory.Data[0xFF42] = 0x00
	gb.Memory.Data[0xFF43] = 0x00
	gb.Memory.Data[0xFF45] = 0x00
	gb.Memory.Data[0xFF47] = 0xFC
	gb.Memory.Data[0xFF48] = 0xFF
	gb.Memory.Data[0xFF49] = 0xFF
	gb.Memory.Data[0xFF4A] = 0x00
	gb.Memory.Data[0xFF4B] = 0x00
	gb.Memory.Data[0xFFFF] = 0x00
}