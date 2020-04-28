package gbio

// IOBinding provides an interface for display and input bindings.
type IOBinding interface {
	// RenderScreen renders a frame of the game.
	RenderScreen()
	// ProcessInput processes input.
	ProcessInput()
	// SetTitle sets the title of the window.
	SetTitle(fps int)
	// IsRunning returns if the monitor is still running.
	IsRunning() bool
}
