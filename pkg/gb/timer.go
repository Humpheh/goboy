package gb

const (
	// DIV is the divider register which is incremented periodically by
	// the Gameboy.
	DIV = 0xFF04
	// TIMA is the timer counter register which is incremented by a clock
	// frequency specified in the TAC register.
	TIMA = 0xFF05
	// TMA is the timer modulo register. When the TIMA overflows, this data
	// will be loaded into the TIMA register.
	TMA = 0xFF06
	// TAC is the timer control register. Writing to this register will
	// start and stop the timer, and select the clock speed for the timer.
	TAC = 0xFF07

	// TODO: move more hardware registers up here.
)

// timerBits is the bit of the internal counter that clocks TIMA for each
// TAC frequency setting.
var timerBits = [4]uint16{9, 3, 5, 7}

// timer holds the state of the divider and timer.
type timer struct {
	// Internal counter incremented every clock. DIV is its upper byte.
	counter uint16
	// Set for the M-cycle after TIMA overflows, during which it reads 0 and
	// a write to TIMA cancels the reload.
	overflowed bool
	// Set for the M-cycle in which TIMA is reloaded from TMA, during which
	// writes to TIMA are ignored and writes to TMA are also copied to TIMA.
	reloading bool
}

// updateTimers advances the divider and timer by a number of clocks, one
// M-cycle at a time.
func (gb *Gameboy) updateTimers(cycles int) {
	for ; cycles > 0; cycles -= 4 {
		gb.timer.reloading = false
		if gb.timer.overflowed {
			gb.timer.overflowed = false
			gb.timer.reloading = true
			gb.memory.HighRAM[TIMA-0xFF00] = gb.memory.HighRAM[TMA-0xFF00]
			gb.requestInterrupt(2)
		}
		gb.setTimerCounter(gb.timer.counter + 4)
	}
}

// setTimerCounter sets the internal counter and DIV, incrementing TIMA if the
// change makes the timer signal fall.
func (gb *Gameboy) setTimerCounter(value uint16) {
	before := gb.timerSignal()
	gb.timer.counter = value
	gb.memory.HighRAM[DIV-0xFF00] = byte(value >> 8)
	if before && !gb.timerSignal() {
		gb.incrementTIMA()
	}
}

// writeTAC sets the timer control register, incrementing TIMA if the change
// makes the timer signal fall.
func (gb *Gameboy) writeTAC(value byte) {
	before := gb.timerSignal()
	gb.memory.HighRAM[TAC-0xFF00] = value | 0xF8
	if before && !gb.timerSignal() {
		gb.incrementTIMA()
	}
}

// writeTIMA sets TIMA unless it is being reloaded. A write in the M-cycle
// after an overflow cancels the reload and its interrupt.
func (gb *Gameboy) writeTIMA(value byte) {
	if gb.timer.reloading {
		return
	}
	gb.timer.overflowed = false
	gb.memory.HighRAM[TIMA-0xFF00] = value
}

// writeTMA sets TMA, which also goes into TIMA if it is being reloaded.
func (gb *Gameboy) writeTMA(value byte) {
	gb.memory.HighRAM[TMA-0xFF00] = value
	if gb.timer.reloading {
		gb.memory.HighRAM[TIMA-0xFF00] = value
	}
}

// timerSignal returns the signal TIMA is clocked by, which is the selected
// bit of the internal counter while the timer is enabled. TIMA increments
// when it falls.
func (gb *Gameboy) timerSignal() bool {
	tac := gb.memory.HighRAM[TAC-0xFF00]
	return tac&0x04 != 0 && gb.timer.counter>>timerBits[tac&0x03]&1 != 0
}

// incrementTIMA increments TIMA. On overflow it reads 0 until it is reloaded
// from TMA in the next M-cycle.
func (gb *Gameboy) incrementTIMA() {
	gb.memory.HighRAM[TIMA-0xFF00]++
	if gb.memory.HighRAM[TIMA-0xFF00] == 0 {
		gb.timer.overflowed = true
	}
}
