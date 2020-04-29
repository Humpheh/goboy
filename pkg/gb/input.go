package gb

import (
	"github.com/Humpheh/goboy/pkg/bits"
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

type ButtonInput struct {
	// Pressed and Released are gameboy button inputs on this frame
	Pressed, Released []Button

	// KeysPressed list all keyboard keys pressed last frame
	// This is used to call handlers that list/modify emulator config
	KeysPressed []string
}

// IOBinding provides an interface for display and input bindings.
type IOBinding interface {
	// RenderScreen renders a frame of the game.
	Render(screen *[160][144][3]uint8)
	// ButtonInput returns which buttons were pressed and released
	ButtonInput() ButtonInput
	// SetTitle sets the title of the window.
	SetTitle(title string)
	// IsRunning returns if the monitor is still running.
	IsRunning() bool
}

// pressButton notifies the GameBoy that a button has just been pressed
// and requests a joypad interrupt.
func (gb *Gameboy) pressButton(button Button) {
	gb.inputMask = bits.Reset(gb.inputMask, byte(button))
	gb.requestInterrupt(4) // Request the joypad interrupt
}

// releaseButton notifies the GameBoy that a button has just been released.
func (gb *Gameboy) releaseButton(button Button) {
	gb.inputMask = bits.Set(gb.inputMask, byte(button))
}

func (gb *Gameboy) ProcessInput(buttons ButtonInput) {

	if !gb.IsGameLoaded() || gb.paused {
		for _, pressedButton := range buttons.KeysPressed {
			if pressedButton == "Escape" {
				gb.keyHandlers[pressedButton]()
			}
		}
		return
	}

	for _, button := range buttons.Pressed {
		gb.pressButton(button)
	}

	for _, button := range buttons.Released {
		gb.releaseButton(button)
	}

	for _, key := range buttons.KeysPressed {
		if handler, ok := gb.keyHandlers[key]; ok {
			handler()
		}
	}
}
