package gb

import (
	"fmt"

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

func (gb *Gameboy) ProcessInput(buttons gbio.ButtonInput) {

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

	// Extra key bindings to functions.
	var extraKeyMap = map[string]func(*Gameboy){
		// Pause execution
		"Escape": func(gameboy *Gameboy) {
			// Toggle the paused state
			gameboy.SetPaused(!gameboy.IsPaused())
		},

		// Change GB colour palette
		"Equal": func(gameboy *Gameboy) {
			CurrentPalette = (CurrentPalette + 1) % byte(len(Palettes))
		},

		// GPU debugging
		"Q": func(gameboy *Gameboy) {
			gameboy.Debug.HideBackground = !gameboy.Debug.HideBackground
		},
		"W": func(gameboy *Gameboy) {
			gameboy.Debug.HideSprites = !gameboy.Debug.HideSprites
		},
		"D": func(gameboy *Gameboy) {
			fmt.Println("BG Map:")
			fmt.Println(gameboy.BGMapString())
		},

		// CPU debugging
		"E": func(gameboy *Gameboy) {
			gameboy.Debug.OutputOpcodes = !gameboy.Debug.OutputOpcodes
		},

		// Audio channel debugging
		"7": func(gameboy *Gameboy) {
			gameboy.ToggleSoundChannel(1)
		},
		"8": func(gameboy *Gameboy) {
			gameboy.ToggleSoundChannel(2)
		},
		"9": func(gameboy *Gameboy) {
			gameboy.ToggleSoundChannel(3)
		},
		"0": func(gameboy *Gameboy) {
			gameboy.ToggleSoundChannel(4)
		},
	}

	if !gb.IsGameLoaded() || gb.IsPaused() {
		for _, pressedButton := range buttons.Pressed {
			if pressedButton == "Escape" {
				extraKeyMap[pressedButton](gb)
			}
		}
		return
	}

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
