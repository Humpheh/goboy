package gbio

import (
	"github.com/faiface/pixel/pixelgl"
)

type ButtonInput struct {
	Pressed, Released []pixelgl.Button
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
