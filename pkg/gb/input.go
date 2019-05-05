package gb

import "github.com/Humpheh/goboy/pkg/bits"

// Button represents the button on a GameBoy.
type Button byte

const (
	ButtonA      Button = 0
	ButtonB             = 1
	ButtonSelect        = 2
	ButtonStart         = 3
	ButtonRight         = 4
	ButtonLeft          = 5
	ButtonUp            = 6
	ButtonDown          = 7
)

// PressButton notifies the GameBoy that a button has just been pressed
// and requests a joypad interrupt.
func (gb *Gameboy) PressButton(button Button) {
	gb.inputMask = bits.Reset(gb.inputMask, byte(button))
	gb.requestInterrupt(4) // Request the joypad interrupt
}

// ReleaseButton notifies the GameBoy that a button has just been released.
func (gb *Gameboy) ReleaseButton(button Button) {
	gb.inputMask = bits.Set(gb.inputMask, byte(button))
}
