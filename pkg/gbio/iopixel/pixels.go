package iopixel

import (
	"fmt"
	"image/color"
	"log"

	"math"

	"github.com/Humpheh/goboy/gb"
	"github.com/Humpheh/goboy/pkg/bits"
	"github.com/Humpheh/goboy/pkg/gbio"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

// PixelScale is the multiplier on the pixels on display
var PixelScale float64 = 3

// NewPixelsIOBinding returns a new Pixelsgl IOBinding
func NewPixelsIOBinding(gameboy *gb.Gameboy, disableVsync bool) gbio.IOBinding {
	monitor := PixelsIOBinding{Gameboy: gameboy}
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
			float64(160*PixelScale), float64(144*PixelScale),
		),
		VSync:     !disableVsync,
		Resizable: true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	mon.Window = win
	mon.UpdateCamera()

	mon.picture = &pixel.PictureData{
		Pix:    make([]color.RGBA, 144*160),
		Stride: 160,
		Rect:   pixel.R(0, 0, 160, 144),
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
	for y := 0; y < 144; y++ {
		for x := 0; x < 160; x++ {
			col := mon.Gameboy.PreparedData[x][y]
			rgb := color.RGBA{R: col[0], G: col[1], B: col[2], A: 0xFF}
			mon.picture.Pix[(143-y)*160+x] = rgb
		}
	}

	r, g, b := gb.GetPaletteColour(3)
	bg := color.RGBA{R: r, G: g, B: b, A: 0xFF}
	mon.Window.Clear(bg)

	spr := pixel.NewSprite(pixel.Picture(mon.picture), pixel.R(0, 0, 160, 144))
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
var keyMap = map[pixelgl.Button]byte{
	// A button
	pixelgl.KeyZ: 0,
	// B button
	pixelgl.KeyX: 1,
	// SELECT button
	pixelgl.KeyBackspace: 2,
	// START button
	pixelgl.KeyEnter: 3,
	// RIGHT button
	pixelgl.KeyRight: 4,
	// LEFT button
	pixelgl.KeyLeft: 5,
	// UP button
	pixelgl.KeyUp: 6,
	// DOWN button
	pixelgl.KeyDown: 7,
}

// Extra key bindings to functions.
var extraKeyMap = map[pixelgl.Button]func(*PixelsIOBinding){
	// Pause execution
	pixelgl.KeyEscape: func(mon *PixelsIOBinding) {
		mon.Gameboy.ExecutionPaused = !mon.Gameboy.ExecutionPaused
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
	pixelgl.KeyA: func(mon *PixelsIOBinding) {
		fmt.Println("BG Tile Palette:")
		fmt.Println(mon.Gameboy.BGPalette.String())
	},
	pixelgl.KeyS: func(mon *PixelsIOBinding) {
		fmt.Println("Sprite Palette:")
		fmt.Println(mon.Gameboy.SpritePalette.String())
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
		mon.Gameboy.Debug.MuteChannel1 = !mon.Gameboy.Debug.MuteChannel1
		log.Print("Channel 1 mute =", mon.Gameboy.Debug.MuteChannel1)
	},
	pixelgl.Key8: func(mon *PixelsIOBinding) {
		mon.Gameboy.Debug.MuteChannel2 = !mon.Gameboy.Debug.MuteChannel2
		log.Print("Channel 2 mute =", mon.Gameboy.Debug.MuteChannel2)
	},
	pixelgl.Key9: func(mon *PixelsIOBinding) {
		mon.Gameboy.Debug.MuteChannel3 = !mon.Gameboy.Debug.MuteChannel3
		log.Print("Channel 3 mute =", mon.Gameboy.Debug.MuteChannel3)
	},
	pixelgl.Key0: func(mon *PixelsIOBinding) {
		mon.Gameboy.Debug.MuteChannel4 = !mon.Gameboy.Debug.MuteChannel4
		log.Print("Channel 4 mute =", mon.Gameboy.Debug.MuteChannel4)
	},

	// Fullscreen toggle
	pixelgl.KeyF: func(mon *PixelsIOBinding) {
		mon.toggleFullscreen()
	},
	// Gob
	pixelgl.KeyG: func(mon *PixelsIOBinding) {
		err := mon.Gameboy.Gob()
		log.Print(err)
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
	if mon.Gameboy.IsGameLoaded() && !mon.Gameboy.ExecutionPaused {
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
	for key, offset := range keyMap {
		if mon.Window.JustPressed(key) {
			mon.Gameboy.InputMask = bits.Reset(mon.Gameboy.InputMask, offset)
			mon.Gameboy.RequestJoypadInterrupt() // Joypad interrupt
		}
		if mon.Window.JustReleased(key) {
			mon.Gameboy.InputMask = bits.Set(mon.Gameboy.InputMask, offset)
		}
	}
}
