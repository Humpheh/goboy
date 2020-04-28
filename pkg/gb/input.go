package gb

import (
	"fmt"

	"github.com/Humpheh/goboy/pkg/bits"
	"github.com/Humpheh/goboy/pkg/gbio"
	"github.com/faiface/pixel/pixelgl"
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

func (gb *Gameboy) ProcessInput(buttons gbio.ButtonInput) {

	// Mapping from keys to GB index.
	var keyMap = map[pixelgl.Button]Button{
		pixelgl.KeyZ:         ButtonA,
		pixelgl.KeyX:         ButtonB,
		pixelgl.KeyBackspace: ButtonSelect,
		pixelgl.KeyEnter:     ButtonStart,
		pixelgl.KeyRight:     ButtonRight,
		pixelgl.KeyLeft:      ButtonLeft,
		pixelgl.KeyUp:        ButtonUp,
		pixelgl.KeyDown:      ButtonDown,
	}

	// Extra key bindings to functions.
	var extraKeyMap = map[pixelgl.Button]func(*Gameboy){
		// Pause execution
		pixelgl.KeyEscape: func(gameboy *Gameboy) {
			// Toggle the paused state
			gameboy.SetPaused(!gameboy.IsPaused())
		},

		// Change GB colour palette
		pixelgl.KeyEqual: func(gameboy *Gameboy) {
			CurrentPalette = (CurrentPalette + 1) % byte(len(Palettes))
		},

		// GPU debugging
		pixelgl.KeyQ: func(gameboy *Gameboy) {
			gameboy.Debug.HideBackground = !gameboy.Debug.HideBackground
		},
		pixelgl.KeyW: func(gameboy *Gameboy) {
			gameboy.Debug.HideSprites = !gameboy.Debug.HideSprites
		},
		pixelgl.KeyD: func(gameboy *Gameboy) {
			fmt.Println("BG Map:")
			fmt.Println(gameboy.BGMapString())
		},

		// CPU debugging
		pixelgl.KeyE: func(gameboy *Gameboy) {
			gameboy.Debug.OutputOpcodes = !gameboy.Debug.OutputOpcodes
		},

		// Audio channel debugging
		pixelgl.Key7: func(gameboy *Gameboy) {
			gameboy.ToggleSoundChannel(1)
		},
		pixelgl.Key8: func(gameboy *Gameboy) {
			gameboy.ToggleSoundChannel(2)
		},
		pixelgl.Key9: func(gameboy *Gameboy) {
			gameboy.ToggleSoundChannel(3)
		},
		pixelgl.Key0: func(gameboy *Gameboy) {
			gameboy.ToggleSoundChannel(4)
		},
	}

	if gb.IsGameLoaded() && !gb.IsPaused() {
		for _, pressedButton := range buttons.Pressed {
			if gameboyButton, ok := keyMap[pressedButton]; ok {
				gb.PressButton(gameboyButton)
			}
			if handler, ok := extraKeyMap[pressedButton]; ok {
				handler(gb)
			}
		}

		for _, releasedButton := range buttons.Released {
			if gameboyButton, ok := keyMap[releasedButton]; ok {
				gb.ReleaseButton(gameboyButton)
			}
		}
	}
}
