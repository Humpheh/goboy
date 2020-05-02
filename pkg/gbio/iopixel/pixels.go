package iopixel

import (
	"fmt"
	"image/color"
	"log"

	"math"

	"github.com/Humpheh/goboy/pkg/gb"
	"github.com/Humpheh/goboy/pkg/gbio"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

// PixelScale is the multiplier on the pixels on display
var PixelScale float64 = 3

// PixelsIOBinding binds screen output and input using the pixels library.
type PixelsIOBinding struct {
	window  *pixelgl.Window
	picture *pixel.PictureData
}

// NewPixelsIOBinding returns a new Pixelsgl IOBinding
func NewPixelsIOBinding(enableVSync bool, gameboy *gb.Gameboy) *PixelsIOBinding {
	windowConfig := pixelgl.WindowConfig{
		Title: "GoBoy",
		Bounds: pixel.R(
			0, 0,
			float64(gb.ScreenWidth*PixelScale), float64(gb.ScreenHeight*PixelScale),
		),
		VSync:     enableVSync,
		Resizable: true,
	}

	window, err := pixelgl.NewWindow(windowConfig)
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}

	// Hack so that pixelgl renders on Darwin
	window.SetPos(window.GetPos().Add(pixel.V(0, 1)))

	picture := &pixel.PictureData{
		Pix:    make([]color.RGBA, gb.ScreenWidth*gb.ScreenHeight),
		Stride: gb.ScreenWidth,
		Rect:   pixel.R(0, 0, gb.ScreenWidth, gb.ScreenHeight),
	}

	monitor := PixelsIOBinding{
		window:  window,
		picture: picture,
	}

	monitor.updateCamera()

	return &monitor
}

// updateCamera updates the window camera to center the output.
func (mon *PixelsIOBinding) updateCamera() {
	xScale := mon.window.Bounds().W() / 160
	yScale := mon.window.Bounds().H() / 144
	scale := math.Min(yScale, xScale)

	shift := mon.window.Bounds().Size().Scaled(0.5).Sub(pixel.ZV)
	cam := pixel.IM.Scaled(pixel.ZV, scale).Moved(shift)
	mon.window.SetMatrix(cam)
}

// IsRunning returns if the game should still be running. When
// the window is closed this will be false so the game stops.
func (mon *PixelsIOBinding) IsRunning() bool {
	return !mon.window.Closed()
}

// Render renders the pixels on the screen.
func (mon *PixelsIOBinding) Render(screen *[160][144][3]uint8) {
	for y := 0; y < gb.ScreenHeight; y++ {
		for x := 0; x < gb.ScreenWidth; x++ {
			col := screen[x][y]
			rgb := color.RGBA{R: col[0], G: col[1], B: col[2], A: 0xFF}
			mon.picture.Pix[(gb.ScreenHeight-1-y)*gb.ScreenWidth+x] = rgb
		}
	}

	r, g, b := gb.GetPaletteColour(3)
	bg := color.RGBA{R: r, G: g, B: b, A: 0xFF}
	mon.window.Clear(bg)

	spr := pixel.NewSprite(pixel.Picture(mon.picture), pixel.R(0, 0, gb.ScreenWidth, gb.ScreenHeight))
	spr.Draw(mon.window, pixel.IM)

	mon.updateCamera()
	mon.window.Update()
}

// SetTitle sets the title of the game window.
func (mon *PixelsIOBinding) SetTitle(title string) {
	mon.window.SetTitle(title)
}

// Toggle the fullscreen window on the main monitor.
func (mon *PixelsIOBinding) toggleFullscreen() {
	if mon.window.Monitor() == nil {
		monitor := pixelgl.PrimaryMonitor()
		_, height := monitor.Size()
		mon.window.SetMonitor(monitor)
		PixelScale = height / 144
	} else {
		mon.window.SetMonitor(nil)
		PixelScale = 3
	}
}

// ProcessInput checks the input and process it.
func (mon *PixelsIOBinding) ProcessInput(gameboyArg *gb.Gameboy) {

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
	var extraKeyMap = map[pixelgl.Button]func(*gb.Gameboy){
		// Pause execution
		pixelgl.KeyEscape: func(gameboy *gb.Gameboy) {
			// Toggle the paused state
			gameboy.SetPaused(!gameboy.IsPaused())
		},

		// Change GB colour palette
		pixelgl.KeyEqual: func(gameboy *gb.Gameboy) {
			gb.CurrentPalette = (gb.CurrentPalette + 1) % byte(len(gb.Palettes))
		},

		// GPU debugging
		pixelgl.KeyQ: func(gameboy *gb.Gameboy) {
			gameboy.Debug.HideBackground = !gameboy.Debug.HideBackground
		},
		pixelgl.KeyW: func(gameboy *gb.Gameboy) {
			gameboy.Debug.HideSprites = !gameboy.Debug.HideSprites
		},
		pixelgl.KeyD: func(gameboy *gb.Gameboy) {
			fmt.Println("BG Map:")
			fmt.Println(gameboy.BGMapString())
		},

		// CPU debugging
		pixelgl.KeyE: func(gameboy *gb.Gameboy) {
			gameboy.Debug.OutputOpcodes = !gameboy.Debug.OutputOpcodes
		},

		// Audio channel debugging
		pixelgl.Key7: func(gameboy *gb.Gameboy) {
			gameboy.ToggleSoundChannel(1)
		},
		pixelgl.Key8: func(gameboy *gb.Gameboy) {
			gameboy.ToggleSoundChannel(2)
		},
		pixelgl.Key9: func(gameboy *gb.Gameboy) {
			gameboy.ToggleSoundChannel(3)
		},
		pixelgl.Key0: func(gameboy *gb.Gameboy) {
			gameboy.ToggleSoundChannel(4)
		},
	}

	var ioInput gbio.IOInput

	for input := pixelgl.Button(0); input < pixelgl.KeyLast; input++ {
		if mon.window.JustPressed(input) {
			ioInput.Pressed = append(ioInput.Pressed, input)
		}
		if mon.window.JustReleased(input) {
			ioInput.Released = append(ioInput.Released, input)
		}
	}

	if gameboyArg.IsGameLoaded() && !gameboyArg.IsPaused() {
		for _, pressedButton := range ioInput.Pressed {
			if gameboyButton, ok := keyMap[pressedButton]; ok {
				gameboyArg.PressButton(gameboyButton)
			}
			if handler, ok := extraKeyMap[pressedButton]; ok {
				handler(gameboyArg)
			}
			if pressedButton == pixelgl.KeyF {
				mon.toggleFullscreen()
			}
		}

		for _, releasedButton := range ioInput.Released {
			if gameboyButton, ok := keyMap[releasedButton]; ok {
				gameboyArg.ReleaseButton(gameboyButton)
			}
		}
	}
}
