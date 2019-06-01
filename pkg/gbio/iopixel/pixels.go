package iopixel

import (
	"fmt"
	"image/color"
	"log"

	"math"

	"github.com/Humpheh/goboy/pkg/gb"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

// PixelScale is the multiplier on the pixels on display
var PixelScale float64 = 3

// NewPixelsIOBinding returns a new Pixelsgl IOBinding
func NewPixelsIOBinding(disableVsync bool) *PixelsIOBinding {
	monitor := PixelsIOBinding{}
	monitor.Init(disableVsync)
	return &monitor
}

// PixelsIOBinding binds screen output and input using the pixels library.
type PixelsIOBinding struct {
	Gameboy *gb.Gameboy
	Window  *pixelgl.Window
	picture *pixel.PictureData
}

// Init initialises the Pixels bindings.
func (mon *PixelsIOBinding) Init(disableVsync bool) {
	cfg := pixelgl.WindowConfig{
		Title: "GoBoy",
		Bounds: pixel.R(
			0, 0,
			float64(gb.ScreenWidth*PixelScale), float64(gb.ScreenHeight*PixelScale),
		),
		VSync:     !disableVsync,
		Resizable: true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}

	// Hack so that pixelgl renders on Darwin
	win.SetPos(win.GetPos().Add(pixel.V(0, 1)))

	mon.Window = win
	mon.UpdateCamera()

	mon.picture = &pixel.PictureData{
		Pix:    make([]color.RGBA, gb.ScreenWidth*gb.ScreenHeight),
		Stride: gb.ScreenWidth,
		Rect:   pixel.R(0, 0, gb.ScreenWidth, gb.ScreenHeight),
	}
}

// UpdateCamera updates the window camera to center the output.
func (mon *PixelsIOBinding) UpdateCamera() {
	xScale := mon.Window.Bounds().W() / 160
	yScale := mon.Window.Bounds().H() / 144
	scale := math.Min(yScale, xScale)

	shift := mon.Window.Bounds().Size().Scaled(0.5).Sub(pixel.ZV)
	cam := pixel.IM.Scaled(pixel.ZV, scale).Moved(shift)
	mon.Window.SetMatrix(cam)
}

// IsRunning returns if the game should still be running. When
// the window is closed this will be false so the game stops.
func (mon *PixelsIOBinding) IsRunning() bool {
	return !mon.Window.Closed()
}

// RenderScreen renders the pixels on the screen.
func (mon *PixelsIOBinding) RenderScreen() {
	for y := 0; y < gb.ScreenHeight; y++ {
		for x := 0; x < gb.ScreenWidth; x++ {
			col := mon.Gameboy.PreparedData[x][y]
			rgb := color.RGBA{R: col[0], G: col[1], B: col[2], A: 0xFF}
			mon.picture.Pix[(gb.ScreenHeight-1-y)*gb.ScreenWidth+x] = rgb
		}
	}

	r, g, b := gb.GetPaletteColour(3)
	bg := color.RGBA{R: r, G: g, B: b, A: 0xFF}
	mon.Window.Clear(bg)

	spr := pixel.NewSprite(pixel.Picture(mon.picture), pixel.R(0, 0, gb.ScreenWidth, gb.ScreenHeight))
	spr.Draw(mon.Window, pixel.IM)

	mon.UpdateCamera()
	mon.Window.Update()
}

// Destroy implements IOBinding.Destroy.
func (mon *PixelsIOBinding) Destroy() {
	mon.Window.Destroy()
}

// SetTitle sets the title of the game window.
func (mon *PixelsIOBinding) SetTitle(fps int) {
	title := "GoBoy"
	if mon.Gameboy.IsGameLoaded() {
		title += fmt.Sprintf(" - %s", mon.Gameboy.Memory.Cart.GetName())
		if fps != 0 {
			title += fmt.Sprintf(" (FPS: %2v)", fps)
		}
	}
	mon.Window.SetTitle(title)
}

// Mapping from keys to GB index.
var keyMap = map[pixelgl.Button]gb.Button{
	pixelgl.KeyZ:         gb.ButtonA,
	pixelgl.KeyX:         gb.ButtonB,
	pixelgl.KeyBackspace: gb.ButtonSelect,
	pixelgl.KeyEnter:     gb.ButtonStart,
	pixelgl.KeyRight:     gb.ButtonRight,
	pixelgl.KeyLeft:      gb.ButtonLeft,
	pixelgl.KeyUp:        gb.ButtonUp,
	pixelgl.KeyDown:      gb.ButtonDown,
}

// Extra key bindings to functions.
var extraKeyMap = map[pixelgl.Button]func(*PixelsIOBinding){
	// Pause execution
	pixelgl.KeyEscape: func(mon *PixelsIOBinding) {
		// Toggle the paused state
		mon.Gameboy.SetPaused(!mon.Gameboy.IsPaused())
	},

	// Change GB colour palette
	pixelgl.KeyEqual: func(mon *PixelsIOBinding) {
		gb.CurrentPalette = (gb.CurrentPalette + 1) % byte(len(gb.Palettes))
	},

	// GPU debugging
	pixelgl.KeyQ: func(mon *PixelsIOBinding) {
		mon.Gameboy.Debug.HideBackground = !mon.Gameboy.Debug.HideBackground
	},
	pixelgl.KeyW: func(mon *PixelsIOBinding) {
		mon.Gameboy.Debug.HideSprites = !mon.Gameboy.Debug.HideSprites
	},
	pixelgl.KeyD: func(mon *PixelsIOBinding) {
		fmt.Println("BG Map:")
		fmt.Println(mon.Gameboy.BGMapString())
	},

	// CPU debugging
	pixelgl.KeyE: func(mon *PixelsIOBinding) {
		mon.Gameboy.Debug.OutputOpcodes = !mon.Gameboy.Debug.OutputOpcodes
	},

	// Audio channel debugging
	pixelgl.Key7: func(mon *PixelsIOBinding) {
		mon.Gameboy.ToggleSoundChannel(1)
	},
	pixelgl.Key8: func(mon *PixelsIOBinding) {
		mon.Gameboy.ToggleSoundChannel(2)
	},
	pixelgl.Key9: func(mon *PixelsIOBinding) {
		mon.Gameboy.ToggleSoundChannel(3)
	},
	pixelgl.Key0: func(mon *PixelsIOBinding) {
		mon.Gameboy.ToggleSoundChannel(4)
	},

	// Fullscreen toggle
	pixelgl.KeyF: func(mon *PixelsIOBinding) {
		mon.toggleFullscreen()
	},
}

// Toggle the fullscreen window on the main monitor.
func (mon *PixelsIOBinding) toggleFullscreen() {
	if mon.Window.Monitor() == nil {
		monitor := pixelgl.PrimaryMonitor()
		_, height := monitor.Size()
		mon.Window.SetMonitor(monitor)
		PixelScale = height / 144
	} else {
		mon.Window.SetMonitor(nil)
		PixelScale = 3
	}
}

// ProcessInput checks the input and process it.
func (mon *PixelsIOBinding) ProcessInput() {
	if mon.Gameboy.IsGameLoaded() && !mon.Gameboy.IsPaused() {
		mon.processGBInput()
	}
	// Extra keys not related to emulation
	for key, f := range extraKeyMap {
		if mon.Window.JustPressed(key) {
			f(mon)
		}
	}
}

// Check the input and process it.
func (mon *PixelsIOBinding) processGBInput() {
	for key, button := range keyMap {
		if mon.Window.JustPressed(key) {
			mon.Gameboy.PressButton(button)
		}
		if mon.Window.JustReleased(key) {
			mon.Gameboy.ReleaseButton(button)
		}
	}
}
