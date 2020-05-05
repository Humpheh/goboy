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

	ButtonPause               = 8
	ButtonChangePallete       = 9
	ButtonToggleBackground    = 10
	ButtonToggleSprites       = 11
	ButttonToggleOutputOpCode = 12
	ButtonPrintBGMap          = 13
	ButtonToggleSoundChannel1 = 14
	ButtonToggleSoundChannel2 = 15
	ButtonToggleSoundChannel3 = 16
	ButtonToggleSoundChannel4 = 17
)

// IsGameBoyInput checks whether a button value represents a physical button on a gameboy
func (button Button) IsGameBoyButton() bool {
	return button <= ButtonDown
}

type ButtonInput struct {
	// Pressed and Released are inputs on this frame
	Pressed, Released []Button
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

	if gb.paused || !gb.IsGameLoaded() {
		return
	}

	gb.inputMask = bits.Reset(gb.inputMask, byte(button))
	gb.requestInterrupt(4) // Request the joypad interrupt
}

// releaseButton notifies the GameBoy that a button has just been released.
func (gb *Gameboy) releaseButton(button Button) {
	if gb.paused || !gb.IsGameLoaded() {
		return
	}

	gb.inputMask = bits.Set(gb.inputMask, byte(button))
}

func (gb *Gameboy) ProcessInput(buttons ButtonInput) {

	for _, button := range buttons.Pressed {
		if button.IsGameBoyButton() {
			gb.pressButton(button)
		} else if handler, ok := gb.keyHandlers[button]; ok {
			handler()
		}
	}

	for _, button := range buttons.Released {
		if button.IsGameBoyButton() {
			gb.releaseButton(button)
		}
	}
}
