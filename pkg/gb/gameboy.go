package gb

import (
	"fmt"
	"log"

	"encoding/gob"
	"os"

	"github.com/Humpheh/goboy/pkg/apu"
	"github.com/Humpheh/goboy/pkg/bits"
)

const (
	// ClockSpeed is the number of cycles the GameBoy CPU performs each second.
	ClockSpeed = 4194304
	// FramesSecond is the target number of frames for each frame of GameBoy output.
	FramesSecond = 60
	// CyclesFrame is the number of CPU cycles in each frame.
	CyclesFrame = ClockSpeed / FramesSecond
)

// Gameboy is the master struct which contains all of the sub components
// for running the Gameboy emulator.
type Gameboy struct {
	options gameboyOptions

	Memory *Memory
	CPU    *CPU
	Sound  *apu.APU

	Debug           DebugFlags
	ExecutionPaused bool

	timerCounter int

	// Matrix of pixel data which is used while the screen is rendering. When a
	// frame has been completed, this data is copied into the PreparedData matrix.
	screenData [160][144][3]uint8
	// Track colour of tiles in scanline for priority management.
	tileScanline      [160]uint8
	scanlineCounter   int
	lastDrawnScanline byte

	// PreparedData is a matrix of screen pixel data for a single frame which has
	// been fully rendered.
	PreparedData [160][144][3]uint8

	interruptsEnabling bool
	interruptsOn       bool
	halted             bool

	cbInst    [0x100]func()
	InputMask byte

	mainInst [0x100]func()

	// Flag if the game is running in cgb mode. For this to be true the game
	// rom must support cgb mode and the option be true.
	cgbMode       bool
	BGPalette     *cgbPalette
	SpritePalette *cgbPalette

	currentSpeed byte
	prepareSpeed bool

	thisCpuTicks int
}

// Update update the state of the gameboy by a single frame.
func (gb *Gameboy) Update() int {
	if gb.ExecutionPaused {
		return 0
	}

	cycles := 0
	for cycles < CyclesFrame*gb.getSpeed() {
		cyclesOp := 4
		if !gb.halted {
			if gb.Debug.OutputOpcodes {
				LogOpcode(gb, false)
			}
			cyclesOp = gb.ExecuteNextOpcode()
		} else {
			// TODO: This is incorrect
		}
		if gb.IsCGB() {
			gb.checkSpeedSwitch()
		}
		cycles += cyclesOp
		gb.updateGraphics(cyclesOp)
		gb.updateTimers(cyclesOp)
		cycles += gb.doInterrupts()
	}
	return cycles
}

// ToggleSoundChannel toggles a sound channel for debugging.
func (gb *Gameboy) ToggleSoundChannel(channel int) {
	gb.Sound.ToggleSoundChannel(channel)
}

// BGMapString returns a string of the values in the background map.
func (gb *Gameboy) BGMapString() string {
	out := ""
	for y := uint16(0); y < 0x20; y++ {
		out += fmt.Sprintf("%2x: ", y)
		for x := uint16(0); x < 0x20; x++ {
			out += fmt.Sprintf("%2x ", gb.Memory.Read(0x9800+(y*0x20)+x))
		}
		out += "\n"
	}
	return out
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
		gb.halted = false
	}
}

func (gb *Gameboy) updateTimers(cycles int) {
	gb.dividerRegister(cycles)
	if gb.isClockEnabled() {
		gb.timerCounter += cycles

		freq := gb.getClockFreqCount()
		for gb.timerCounter >= freq {
			gb.timerCounter -= freq
			tima := gb.Memory.Read(TIMA)
			if tima == 0xFF {
				gb.Memory.HighRAM[TIMA-0xFF00] = gb.Memory.Read(TMA)
				gb.requestInterrupt(2)
			} else {
				gb.Memory.HighRAM[TIMA-0xFF00] = tima + 1
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
	gb.timerCounter = 0
}

func (gb *Gameboy) dividerRegister(cycles int) {
	gb.CPU.Divider += cycles
	if gb.CPU.Divider >= 255 {
		gb.CPU.Divider -= 255
		gb.Memory.HighRAM[DIV-0xFF00]++
	}
}

// RequestJoypadInterrupt triggers a joypad interrupt. To be called by the io
// binding implementation.
func (gb *Gameboy) RequestJoypadInterrupt() {
	gb.requestInterrupt(4) // Joypad interrupt
}

// Request the Gameboy to perform an interrupt.
func (gb *Gameboy) requestInterrupt(interrupt byte) {
	req := gb.Memory.ReadHighRam(0xFF0F)
	req = bits.Set(req, interrupt)
	gb.Memory.Write(0xFF0F, req)
}

func (gb *Gameboy) doInterrupts() (cycles int) {
	if gb.interruptsEnabling {
		gb.interruptsOn = true
		gb.interruptsEnabling = false
		return 0
	}
	if !gb.interruptsOn && !gb.halted {
		return 0
	}

	req := gb.Memory.ReadHighRam(0xFF0F)
	enabled := gb.Memory.ReadHighRam(0xFFFF)

	if req > 0 {
		var i byte
		for i = 0; i < 5; i++ {
			if bits.Test(req, i) && bits.Test(enabled, i) {
				gb.serviceInterrupt(i)
				return 20
			}
		}
	}
	return 0
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
	if !gb.interruptsOn && gb.halted {
		gb.halted = false
		return
	}
	gb.interruptsOn = false
	gb.halted = false

	req := gb.Memory.ReadHighRam(0xFF0F)
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
	return gb.cgbMode
}

// Initialise the Gameboy using a path to a rom.
func (gb *Gameboy) init(romFile string) error {
	gb.setup()

	// Load the ROM file
	hasCGB, err := gb.Memory.LoadCart(romFile)
	if err != nil {
		return fmt.Errorf("failed to open rom file: %s", err)
	}
	gb.cgbMode = gb.options.cgbMode && hasCGB
	return nil
}

// Setup and instantitate the gameboys components.
func (gb *Gameboy) setup() error {
	gb.ExecutionPaused = false

	// Initialise the CPU
	gb.CPU = &CPU{}
	gb.CPU.Init(gb.options.cgbMode)

	// Initialise the memory
	gb.Memory = &Memory{}
	gb.Memory.Init(gb)

	gb.Sound = &apu.APU{}
	gb.Sound.Init(gb.options.sound)

	gb.Debug = DebugFlags{}
	gb.scanlineCounter = 456
	gb.InputMask = 0xFF

	gb.cbInst = gb.cbInstructions()

	gb.mainInst = gb.mainInstructions()

	gb.SpritePalette = NewPalette()
	gb.BGPalette = NewPalette()

	return nil
}

// Gob writes a gob'd version of the Gameboy instance to a file (experimental).
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
	err := gameboy.init(romFile)
	if err != nil {
		return nil, err
	}
	return &gameboy, nil
}

// NewGameboyFromGob returns a new Gameboy instance from an existing gob output
// of a Gameboy (experimental).
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
	gameboy.Sound.Init(gameboy.options.sound)
	gameboy.cbInst = gameboy.cbInstructions()
	return &gameboy, nil
}
