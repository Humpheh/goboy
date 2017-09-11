package gob

import (
	"github.com/veandco/go-sdl2/sdl"
)

const SDL_PIXEL_SIZE int32 = 5

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
	Gameboy *Gameboy
	window  *sdl.Window
	surface *sdl.Surface
}

// Initialise the display
func (mon *SDLMonitor) Init() {
	sdl.Init(sdl.INIT_EVERYTHING)
	window, err := sdl.CreateWindow(
		"test",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		int(SDL_PIXEL_SIZE * 160),
		int(SDL_PIXEL_SIZE * 144),
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		panic(err)
	}
	mon.window = window

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	mon.surface = surface
}

// Combine an array of an RGB array a single number
func (mon *SDLMonitor) toColour(cols [3]int) uint32 {
	return uint32(cols[2]) << 16 | uint32(cols[1]) << 8 | uint32(cols[0])
}

// Draw a frame of the game to the screen
func (mon *SDLMonitor) RenderScreen() {
	for x := 0; x < 160; x++ {
		for y := 0; y < 144; y++ {
			rect := sdl.Rect{
				X: int32(x) * SDL_PIXEL_SIZE,
				Y: int32(y) * SDL_PIXEL_SIZE,
				W: SDL_PIXEL_SIZE,
				H: SDL_PIXEL_SIZE,
			}
			mon.surface.FillRect(&rect, mon.toColour(mon.Gameboy.ScreenData[x][y]))
		}
	}
	mon.window.UpdateSurface()
}

// Destroy the monitor instance
func (mon *SDLMonitor) Destroy() {
	mon.window.Destroy()
	sdl.Quit()
}