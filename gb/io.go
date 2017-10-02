package gb

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/Humpheh/goboy/bits"
	"image/color"
	"log"
)

var PixelScale float64 = 3

// Interface the screen and input bindings.
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

// Get an Pixelsgl IOBinding
func NewPixelsIOBinding(gameboy *Gameboy, disableVsync bool) PixelsIOBinding {
	monitor := PixelsIOBinding{Gameboy: gameboy}
	monitor.Init(disableVsync)
	return monitor
}

// PixelsIOBinding binds screen output and input using the pixels library.
type PixelsIOBinding struct {
	IOBinding
	Gameboy *Gameboy
	Window  *pixelgl.Window
	picture *pixel.PictureData
	Frames  int
}

// Initalise the Pixels bindings.
func (mon *PixelsIOBinding) Init(disableVsync bool) {
	cfg := pixelgl.WindowConfig{
		Title: "GoBoy",
		Bounds: pixel.R(
			0, 0,
			float64(160*PixelScale), float64(144*PixelScale),
		),
		VSync: !disableVsync,
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

// Update the window camera to center the output.
func (mon *PixelsIOBinding) UpdateCamera() {
	center := pixel.Vec{X: 80 * PixelScale, Y: 72 * PixelScale}
	if mon.Window.Monitor() != nil {
		width, _ := mon.Window.Monitor().Size()
		center.X += width / 2 - center.X
	}
	cam := pixel.IM.Scaled(pixel.ZV, 1).Moved(center.Sub(pixel.ZV))
	mon.Window.SetMatrix(cam)
}

// Returns a bool of if the game should still be running. When
// the window is closed this will be false so the game stops.
func (mon *PixelsIOBinding) IsRunning() bool {
	return !mon.Window.Closed()
}

// Render the pixels on the screen.
func (mon *PixelsIOBinding) RenderScreen() {
	mon.Frames++
	for y := 0; y < 144; y++ {
		for x := 0; x < 160; x++ {
			col := mon.Gameboy.PreparedData[x][y]
			rgb := color.RGBA{R: col[0], G: col[1], B: col[2], A: 0xFF}
			mon.picture.Pix[(143-y)*160+x] = rgb
		}
	}

	r, g, b := GetPaletteColour(3)
	bg := color.RGBA{R: r, G: g, B: b, A: 0xFF}
	mon.Window.Clear(bg)

	spr := pixel.NewSprite(pixel.Picture(mon.picture), pixel.R(0, 0, 160, 144))
	spr.Draw(mon.Window, pixel.IM.Scaled(pixel.ZV, PixelScale))
	mon.Window.Update()
}

// Set the title of the game window.
func (mon *PixelsIOBinding) SetTitle(fps int) {
	title := fmt.Sprintf("GoBoy - %s", mon.Gameboy.Memory.Cart.Name)
	if fps != 0 {
		title += fmt.Sprintf(" (FPS: %2v)", fps)
	}
	log.Println(title)
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
	// Change GB colour palette
	pixelgl.KeyEqual: func(mon *PixelsIOBinding) {
		current_palette = (current_palette + 1) % byte(len(palettes))
	},

	// GPU debugging
	pixelgl.KeyQ: func(mon *PixelsIOBinding) {
		mon.Gameboy.Debug.HideBackground = !mon.Gameboy.Debug.HideBackground
	},
	pixelgl.KeyW: func(mon *PixelsIOBinding) {
		mon.Gameboy.Debug.HideSprites = !mon.Gameboy.Debug.HideSprites
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
	mon.UpdateCamera()
}

// Check the input and process it.
func (mon *PixelsIOBinding) ProcessInput() {
	for key, offset := range keyMap {
		if mon.Window.JustPressed(key) {
			mon.Gameboy.InputMask = bits.Reset(mon.Gameboy.InputMask, offset)
			mon.Gameboy.RequestInterrupt(4) // Joypad interrupt
		}
		if mon.Window.JustReleased(key) {
			mon.Gameboy.InputMask = bits.Set(mon.Gameboy.InputMask, offset)
		}
	}
	// Extra keys not related to emulation
	for key, f := range extraKeyMap {
		if mon.Window.JustPressed(key) {
			f(mon)
		}
	}
}
