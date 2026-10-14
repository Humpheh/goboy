package gb

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

	ButtonDebugToggle         = 8
	ButtonPause               = 9
	ButtonChangePallete       = 10
	ButtonToggleBackground    = 11
	ButtonToggleSprites       = 12
	ButtonToggleOutputOpCode  = 13
	ButtonPrintBGMap          = 14
	ButtonToggleSoundChannel1 = 15
	ButtonToggleSoundChannel2 = 16
	ButtonToggleSoundChannel3 = 17
	ButtonToggleSoundChannel4 = 18
)

// IsGameBoyButton checks whether a button value represents a physical button on a GameBoy
func (button Button) IsGameBoyButton() bool {
	return button <= ButtonDown
}

type ButtonInput struct {
	// Pressed and Released are inputs on this frame
	Pressed, Released []Button
}

// IOBinding provides an interface for display and input bindings.
type IOBinding interface {
	// SetEnableVSync sets whether vertical sync is enabled.
	SetEnableVSync(enable bool)

	// Render renders a frame of the game.
	Render(screen *[160][144][3]uint8)

	// ProcessButtonInput returns which buttons were pressed and released
	ProcessButtonInput() ButtonInput

	// SetTitle sets the title of the window.
	SetTitle(title string)

	// IsRunning returns if the monitor is still running.
	IsRunning() bool
}

// pressButton notifies the GameBoy that a button has just been pressed
// and requests a joypad interrupt.
func (gb *Gameboy) pressButton(button Button) {
	if gb.paused || !gb.IsCartLoaded() {
		return
	}

	gb.inputMask = bitReset(gb.inputMask, byte(button))
	gb.requestInterrupt(4) // Request the joypad interrupt
}

// releaseButton notifies the GameBoy that a button has just been released.
func (gb *Gameboy) releaseButton(button Button) {
	if gb.paused || !gb.IsCartLoaded() {
		return
	}

	gb.inputMask = bitSet(gb.inputMask, byte(button))
}

func (gb *Gameboy) ProcessInput(buttons ButtonInput) {
	for _, button := range buttons.Pressed {
		if button.IsGameBoyButton() {
			gb.pressButton(button)
		} else if button == ButtonDebugToggle {
			gb.configureDebugKeys(!gb.Debug.Enabled)
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
