package gb

import (
	"github.com/Humpheh/goboy/pkg/bits"
	"github.com/Humpheh/goboy/pkg/gbio"
)

// Button represents the button on a GameBoy.
type Button byte

const (
	// ButtonA is the A button on the GameBoy.
	ButtonA Button = 0
	// ButtonB is the B button on the GameBoy.
	ButtonB = 1
	// ButtonSelect is the select button on the GameBoy.
	ButtonSelect = 2
	// ButtonStart is the start button on the GameBoy.
	ButtonStart = 3
	// ButtonRight is the right dpad direction on the GameBoy.
	ButtonRight = 4
	// ButtonLeft is the left dpad direction on the GameBoy.
	ButtonLeft = 5
	// ButtonUp is the up dpad direction on the GameBoy.
	ButtonUp = 6
	// ButtonDown is the down dpad direction on the GameBoy.
	ButtonDown = 7
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

// Mapping from keys to GB index.
var keyMap = map[string]Button{
	"Z":         ButtonA,
	"X":         ButtonB,
	"Backspace": ButtonSelect,
	"Enter":     ButtonStart,
	"Right":     ButtonRight,
	"Left":      ButtonLeft,
	"Up":        ButtonUp,
	"Down":      ButtonDown,
}

func (gb *Gameboy) ProcessInput(buttons gbio.ButtonInput) {

	if !gb.IsGameLoaded() || gb.paused {
		for _, pressedButton := range buttons.Pressed {
			if pressedButton == "Escape" {
				gb.keyHandlers[pressedButton]()
			}
		}
		return
	}

	for _, pressedButton := range buttons.Pressed {
		if gameboyButton, ok := keyMap[pressedButton]; ok {
			gb.PressButton(gameboyButton)
		}
		if handler, ok := gb.keyHandlers[pressedButton]; ok {
			handler()
		}
	}

	for _, releasedButton := range buttons.Released {
		if gameboyButton, ok := keyMap[releasedButton]; ok {
			gb.ReleaseButton(gameboyButton)
		}
	}
}
