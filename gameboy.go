package gob

import "github.com/humpheh/gob/bits"

const (
	CLOCK_SPEED   = 4194304
	FRAMES_SECOND = 60
	CYCLES_FRAME  = CLOCK_SPEED / FRAMES_SECOND
)

const (
	TIMA = 0xFF05
	TMA  = 0xFF06
	TMC  = 0xFF07
)

type Gameboy struct {
	Memory       *Memory
	CPU 		 *CPU
	TimerCounter int

	InterruptsOn bool
}

// Should be called 60 times/second
func (gb *Gameboy) Update() {
	cycles := 0
	for cycles < CYCLES_FRAME {
		cycles_op := gb.ExecuteOpcode()
		cycles += cycles_op

		gb.UpdateTimers(cycles_op)
		gb.UpdateGraphics(cycles_op)
		gb.DoInterrupts()
	}
	gb.RenderScreen()
}

func (gb *Gameboy) ExecuteOpcode() int {
	return 0
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

func (gb *Gameboy) RequestInterrupt(interrupt int) {
	req := gb.Memory.Read(0xFF0F)
	req = BitSet(req, id)
	gb.Memory.Write(0xFF0F, req)
}

func (gb *Gameboy) UpdateGraphics(cycles int) {

}

func (gb *Gameboy) DoInterrupts() {
	if !gb.InterruptsOn {
		return
	}

	req := gb.Memory.Read(0xFF0F)
	enabled := gb.Memory.Read(0xFFFF)

	if req > 0 {
		for i := 0; i < 5; i++ {
			if TestBit(req, i) {
				if TestBit(enabled, i) {
					gb.ServiceInterrupt(i)
				}
			}
		}
	}
}

func (gb *Gameboy) ServiceInterrupt(interrupt int) {
	gb.InterruptsOn = false

	req := gb.Memory.Read(0xFF0F)
	req = BitReset(req, interrupt)
	gb.Memory.Write(0xFF0F, req)

	gb.PushStack(gb.CPU.PC)

	switch interrupt {
	case 0: gb.CPU.PC = 0x40
	case 1: gb.CPU.PC = 0x48
	case 2: gb.CPU.PC = 0x50
	case 4: gb.CPU.PC = 0x60
	}
}


func (gb *Gameboy) RenderScreen() {

}