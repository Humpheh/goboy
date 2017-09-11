package gob

import (
	"github.com/veandco/go-sdl2/sdl"
)

const PIXEL_SCALE int = 3

// Interface for a screen which displays the game
type Monitor interface {
	// Initialise the monitor
	Init()
	// Render a frame of the game
	RenderScreen()
	// Destroy the monitor instance
	Destroy()
}

// Get an SDL2 monitor
func GetSDLMonitor(gameboy *Gameboy) SDLMonitor {
	monitor := SDLMonitor{Gameboy: gameboy}
	monitor.Init()
	return monitor
}

// And SDL2 monitor instance
type SDLMonitor struct {
	Monitor
	Gameboy  *Gameboy
	window   *sdl.Window
	surface  *sdl.Surface
	renderer *sdl.Renderer
}

// Initialise the display
func (mon *SDLMonitor) Init() {
	sdl.Init(sdl.INIT_EVERYTHING)
	window, renderer, err := sdl.CreateWindowAndRenderer(
		160 * PIXEL_SCALE, 144 * PIXEL_SCALE, 0,
	)
	if err != nil {
		panic(err)
	}
	renderer.SetScale(float32(PIXEL_SCALE), float32(PIXEL_SCALE))
	mon.window = window
	mon.renderer = renderer
}

// Combine an array of an RGB array a single number
func (mon *SDLMonitor) toColour(cols [3]int) uint32 {
	return uint32(cols[2]) << 16 | uint32(cols[1]) << 8 | uint32(cols[0])
}

// Draw a frame of the game to the screen
func (mon *SDLMonitor) RenderScreen() {
	for x := 0; x < 160; x++ {
		for y := 0; y < 144; y++ {
			col := mon.Gameboy.ScreenData[x][y]
			mon.renderer.SetDrawColor(uint8(col[0]), uint8(col[1]), uint8(col[2]), 0xFF)
			mon.renderer.DrawPoint(x, y)
		}
	}
	mon.renderer.Present()
}

// Destroy the monitor instance
func (mon *SDLMonitor) Destroy() {
	mon.window.Destroy()
	sdl.Quit()
}