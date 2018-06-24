package gbio

// IOBinding provides an interface for display and input bindings.
type IOBinding interface {
	// Initialise the IOBinding
	Init(disableVsync bool)
	// Render a frame of the game
	RenderScreen()
	// Destroy the IOBinding instance
	Destroy()
	// Process input
	ProcessInput()
	// Set the title of the window
	SetTitle(fps int)
}
