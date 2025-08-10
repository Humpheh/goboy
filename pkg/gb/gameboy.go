package gb

import (
	"fmt"

	"github.com/Humpheh/goboy/pkg/apu"
	"github.com/Humpheh/goboy/pkg/cart"
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

	// Core components of the Gameboy.
	memory *Memory
	cpu    *CPU
	sound  *apu.APU

	Debug  DebugFlags
	paused bool

	timerCounter int

	// Matrix of pixel data which is used while the screen is rendering. When a
	// frame has been completed, this data is copied into the PreparedData matrix.
	screenData [ScreenWidth][ScreenHeight][3]uint8
	bgPriority [ScreenWidth][ScreenHeight]bool

	// Track colour of tiles in scanline for priority management.
	tileScanline    [ScreenWidth]uint8
	scanlineCounter int
	screenCleared   bool

	// PreparedData is a matrix of screen pixel data for a single frame which has
	// been fully rendered.
	PreparedData [ScreenWidth][ScreenHeight][3]uint8

	interruptsEnabling bool
	interruptsOn       bool
	halted             bool

	cbInst [0x100]func()

	// Mask of the currently pressed buttons.
	inputMask byte

	// Flag if the game is running in cgb mode. For this to be true the game
	// rom must support cgb mode and the option be true.
	cgbMode       bool
	bgPalette     *cgbPalette
	spritePalette *cgbPalette

	currentSpeed byte
	prepareSpeed bool

	thisCpuTicks int

	keyHandlers map[Button]func()
}

// Update update the state of the gameboy by a single frame.
func (gb *Gameboy) Update() int {
	if gb.paused {
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
		cycles += cyclesOp
		gb.updateGraphics(cyclesOp)
		gb.updateTimers(cyclesOp)
		cycles += gb.doInterrupts()

		gb.sound.Buffer(cyclesOp, gb.getSpeed())
	}
	return cycles
}

// togglePaused switches the paused state of the execution.
func (gb *Gameboy) togglePaused() {
	gb.paused = !gb.paused
}

// ToggleSoundChannel toggles a sound channel for debugging.
func (gb *Gameboy) ToggleSoundChannel(channel int) {
	gb.sound.ToggleSoundChannel(channel)
}

func (gb *Gameboy) SoundString() {
	gb.sound.LogSoundState()
}

// BGMapString returns a string of the values in the background map.
func (gb *Gameboy) BGMapString() string {
	out := ""
	for y := uint16(0); y < 0x20; y++ {
		out += fmt.Sprintf("%2x: ", y)
		for x := uint16(0); x < 0x20; x++ {
			out += fmt.Sprintf("%2x ", gb.memory.Read(0x9800+(y*0x20)+x))
		}
		out += "\n"
	}
	return out
}

func (gb *Gameboy) printBGMap() {
	fmt.Printf("BG Map:\n%s", gb.BGMapString())
}

// Get the current CPU speed multiplier (either 1 or 2).
func (gb *Gameboy) getSpeed() int {
	return int(gb.currentSpeed + 1)
}

// Check if the speed needs to be switched for CGB mode.
func (gb *Gameboy) checkSpeedSwitch() {
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
			tima := gb.memory.HighRAM[0x05] /* TIMA */
			if tima == 0xFF {
				gb.memory.HighRAM[TIMA-0xFF00] = gb.memory.HighRAM[0x06] /* TMA */
				gb.requestInterrupt(2)
			} else {
				gb.memory.HighRAM[TIMA-0xFF00] = tima + 1
			}
		}
	}
}

func (gb *Gameboy) isClockEnabled() bool {
	return bitTest(gb.memory.HighRAM[0x07] /* TAC */, 2)
}

func (gb *Gameboy) getClockFreq() byte {
	return gb.memory.HighRAM[0x07] /* TAC */ & 0x3
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
	gb.cpu.Divider += cycles
	if gb.cpu.Divider >= 255 {
		gb.cpu.Divider -= 255
		gb.memory.HighRAM[DIV-0xFF00]++
	}
}

// Request the Gameboy to perform an interrupt.
func (gb *Gameboy) requestInterrupt(interrupt byte) {
	req := gb.memory.HighRAM[0x0F] | 0xE0
	req = bitSet(req, interrupt)
	gb.memory.Write(0xFF0F, req)
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

	req := gb.memory.HighRAM[0x0F] | 0xE0
	enabled := gb.memory.HighRAM[0xFF]

	if req > 0 {
		var i byte
		for i = 0; i < 5; i++ {
			if bitTest(req, i) && bitTest(enabled, i) {
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

	req := gb.memory.ReadHighRam(0xFF0F)
	req = bitReset(req, interrupt)
	gb.memory.Write(0xFF0F, req)

	gb.pushStack(gb.cpu.PC)
	gb.cpu.PC = interruptAddresses[interrupt]
}

// Push a 16 bit value onto the stack and decrement SP.
func (gb *Gameboy) pushStack(address uint16) {
	sp := gb.cpu.SP.HiLo()
	gb.memory.Write(sp-1, byte(uint16(address&0xFF00)>>8))
	gb.memory.Write(sp-2, byte(address&0xFF))
	gb.cpu.SP.Set(gb.cpu.SP.HiLo() - 2)
}

// Pop the next 16 bit value off the stack and increment SP.
func (gb *Gameboy) popStack() uint16 {
	sp := gb.cpu.SP.HiLo()
	byte1 := uint16(gb.memory.Read(sp))
	byte2 := uint16(gb.memory.Read(sp+1)) << 8
	gb.cpu.SP.Set(gb.cpu.SP.HiLo() + 2)
	return byte1 | byte2
}

func (gb *Gameboy) joypadValue(current byte) byte {
	var in byte = 0xF
	if bitTest(current, 4) {
		in = gb.inputMask & 0xF
	} else if bitTest(current, 5) {
		in = (gb.inputMask >> 4) & 0xF
	}
	return current | 0xc0 | in
}

// GetLoadedCart returns the currently loaded cartridge, or nil if no cartridge is loaded.
func (gb *Gameboy) GetLoadedCart() *cart.Cart {
	if gb.memory == nil || gb.memory.Cart == nil {
		return nil
	}
	return gb.memory.Cart
}

// IsCartLoaded returns if there is a game loaded in the gameboy.
func (gb *Gameboy) IsCartLoaded() bool {
	return gb.memory != nil && gb.memory.Cart != nil
}

// IsCGB returns if we are using CGB features.
func (gb *Gameboy) IsCGB() bool {
	return gb.cgbMode
}

// Initialise the Gameboy using a path to a rom.
func (gb *Gameboy) init(romFile string) error {
	gb.setup()

	// Load the ROM file
	hasCGB, err := gb.memory.LoadCart(romFile)
	if err != nil {
		return fmt.Errorf("failed to open rom file: %s", err)
	}
	fmt.Printf("Loaded ROM: %s\n", gb.memory.Cart.GetName())
	gb.cgbMode = gb.options.cgbMode && hasCGB
	return nil
}

func (gb *Gameboy) initKeyHandlers() {
	gb.keyHandlers = map[Button]func(){
		ButtonPause:               gb.togglePaused,
		ButtonChangePallete:       changePalette,
		ButtonToggleBackground:    gb.Debug.toggleBackGround,
		ButtonToggleSprites:       gb.Debug.toggleSprites,
		ButtonToggleOutputOpCode:  gb.Debug.toggleOutputOpCode,
		ButtonPrintBGMap:          gb.printBGMap,
		ButtonToggleSoundChannel1: func() { gb.ToggleSoundChannel(1) },
		ButtonToggleSoundChannel2: func() { gb.ToggleSoundChannel(2) },
		ButtonToggleSoundChannel3: func() { gb.ToggleSoundChannel(3) },
		ButtonToggleSoundChannel4: func() { gb.ToggleSoundChannel(4) },
	}
}

// Setup and instantiate the GameBoys components.
func (gb *Gameboy) setup() {
	// Initialise the CPU
	gb.cpu = &CPU{}
	gb.cpu.Init(gb.options.cgbMode)

	// Initialise the memory
	gb.memory = &Memory{}
	gb.memory.Init(gb)

	gb.sound = &apu.APU{}
	gb.sound.Init(gb.options.sound)

	gb.Debug = DebugFlags{}
	gb.scanlineCounter = 456
	gb.inputMask = 0xFF

	gb.cbInst = gb.cbInstructions()

	gb.spritePalette = NewPalette()
	gb.bgPalette = NewPalette()

	gb.initKeyHandlers()
}

// New returns a new Gameboy instance.
func New(romFile string, opts ...GameboyOption) (*Gameboy, error) {
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
