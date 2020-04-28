package gbio

import (
	"github.com/Humpheh/goboy/pkg/gb"
)

// IOBinding provides an interface for display and input bindings.
type IOBinding interface {
	// RenderScreen renders a frame of the game.
	Render(screen *[160][144][3]uint8)
	// ProcessInput processes input.
	ProcessInput(gameboy *gb.Gameboy)
	// SetTitle sets the title of the window.
	SetTitle(title string)
	// IsRunning returns if the monitor is still running.
	IsRunning() bool
}
