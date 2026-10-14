package gb

import (
	"fmt"
	"strings"

	"github.com/Humpheh/goboy/pkg/apu"
	"github.com/Humpheh/goboy/pkg/cart"
)

const (
	// ClockSpeed is the number of cycles the GameBoy cpu performs each second.
	ClockSpeed = 4194304
	// FramesSecond is the target number of frames for each frame of GameBoy output.
	FramesSecond = 60
	// CyclesFrame is the number of cpu cycles in each frame.
	CyclesFrame = ClockSpeed / FramesSecond
)

// Gameboy is the master struct which contains all of the sub components
// for running the Gameboy emulator.
type Gameboy struct {
	options gameboyOptions

	// Core components of the Gameboy.
	memory memory
	cpu    cpu
	sound  apu.APU
	ppu    ppu
	timer  timer

	Debug  DebugFlags
	paused bool

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
	// Cycles ticked so far while executing the current opcode.
	tickedCpuTicks int

	keyHandlers map[Button]func()
}

// Update update the state of the gameboy by a single frame.
func (gb *Gameboy) Update() int {
	if gb.paused {
		return 0
	}

	cycles := 0
	for cycles < CyclesFrame*gb.getSpeed() {
		if !gb.halted {
			if gb.Debug.OutputOpcodes {
				LogOpcode(gb, false)
			}
			cycles += gb.ExecuteNextOpcode()
		} else {
			// TODO: This is incorrect
			cycles += gb.tick(4)
		}
		cycles += gb.doInterrupts()
	}
	return cycles
}

// tick advances the graphics, timers and sound by the given number of cycles
// and returns the cycles.
func (gb *Gameboy) tick(cycles int) int {
	gb.tickedCpuTicks += cycles
	gb.memory.stepDMACycles(cycles)
	gb.updateGraphics(cycles)
	gb.updateTimers(cycles)
	gb.sound.Buffer(cycles, gb.getSpeed())
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
	var out strings.Builder
	for y := range uint16(0x20) {
		out.WriteString(fmt.Sprintf("%2x: ", y))
		for x := range uint16(0x20) {
			out.WriteString(fmt.Sprintf("%2x ", gb.memory.Read(0x9800+(y*0x20)+x)))
		}
		out.WriteString("\n")
	}
	return out.String()
}

func (gb *Gameboy) printBGMap() {
	fmt.Printf("BG Map:\n%s", gb.BGMapString())
}

// Get the current cpu speed multiplier (either 1 or 2).
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
		for i = range 5 {
			if bitTest(req, i) && bitTest(enabled, i) {
				return gb.serviceInterrupt(i)
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
// enabled and will jump to the interrupt address, returning the number of
// cycles taken.
func (gb *Gameboy) serviceInterrupt(interrupt byte) int {
	// If was halted without interrupts, do not jump or reset IF
	if !gb.interruptsOn && gb.halted {
		gb.halted = false
		return 0
	}
	gb.interruptsOn = false
	gb.halted = false

	req := gb.memory.ReadHighRam(0xFF0F)
	req = bitReset(req, interrupt)
	gb.memory.Write(0xFF0F, req)

	// Dispatch takes 5 cycles: two internal, two to push PC and one to set PC
	gb.tick(4)
	gb.pushStack(gb.cpu.PC)
	gb.cpu.PC = interruptAddresses[interrupt]
	gb.tick(4)
	return 20
}

// Push a 16 bit value onto the stack and decrement SP. Every push is preceded
// by an internal cycle, so this ticks the hardware for that and the two writes.
func (gb *Gameboy) pushStack(address uint16) {
	gb.tick(4)
	sp := gb.cpu.SP.HiLo()
	gb.write(sp-1, byte(uint16(address&0xFF00)>>8))
	gb.write(sp-2, byte(address&0xFF))
	gb.cpu.SP.Set(gb.cpu.SP.HiLo() - 2)
}

// Pop the next 16 bit value off the stack and increment SP.
func (gb *Gameboy) popStack() uint16 {
	sp := gb.cpu.SP.HiLo()
	byte1 := uint16(gb.read(sp))
	byte2 := uint16(gb.read(sp+1)) << 8
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
	return gb.memory.Cart
}

// IsCartLoaded returns if there is a game loaded in the gameboy.
func (gb *Gameboy) IsCartLoaded() bool {
	return gb.memory.Cart != nil
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

// configureDebugKeys enables or disables debug functionality, which
// controls whether the non-GameBoy buttons have any effect.
func (gb *Gameboy) configureDebugKeys(enabled bool) {
	gb.Debug.Enabled = enabled
	if enabled {
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
	} else {
		gb.keyHandlers = map[Button]func(){}
	}
}

// Setup and instantiate the GameBoys components.
func (gb *Gameboy) setup() {
	// Initialise components
	gb.cpu.Init(gb.options.cgbMode)
	gb.memory.Init(gb)
	gb.sound.Init(gb.options.sound)
	gb.ppu.Init()
	gb.timer.counter = uint16(gb.memory.HighRAM[DIV-0xFF00]) << 8

	gb.inputMask = 0xFF

	gb.cbInst = gb.cbInstructions()

	gb.spritePalette = NewPalette()
	gb.bgPalette = NewPalette()

	gb.configureDebugKeys(gb.options.debugEnabled)
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
