// TODO: Rename monitor so something better
package gb

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/humpheh/goboy/bits"
	"image/color"
	"log"
)

const PIXEL_SCALE float64 = 3

// Interface for a screen which displays the game
type Monitor interface {
	// Initialise the monitor
	Init()
	// Render a frame of the game
	RenderScreen()
	// Destroy the monitor instance
	Destroy()
	// Process input
	ProcessInput()
}

// Get an Pixelsgl monitor
func GetPixelsMonitor(gameboy *Gameboy) PixelsMonitor {
	monitor := PixelsMonitor{Gameboy: gameboy}
	monitor.Init()
	return monitor
}

type PixelsMonitor struct {
	Monitor
	Gameboy *Gameboy
	Window  *pixelgl.Window
	picture *pixel.PictureData
	Frames  int
}

func (mon *PixelsMonitor) Init() {
	mon.initDisplay()
}

func (mon *PixelsMonitor) initDisplay() {
	cfg := pixelgl.WindowConfig{
		Title: "GoBoy",
		Bounds: pixel.R(
			0, 0,
			float64(160*PIXEL_SCALE), float64(144*PIXEL_SCALE),
		),
		VSync: true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	mon.Window = win
	cam := pixel.IM.Scaled(pixel.ZV, 1).Moved(win.Bounds().Center().Sub(pixel.ZV))
	win.SetMatrix(cam)

	mon.picture = &pixel.PictureData{
		Pix:    make([]color.RGBA, 144*160),
		Stride: 160,
		Rect:   pixel.R(0, 0, 160, 144),
	}
}

func (mon *PixelsMonitor) RenderScreen() {
	mon.Frames++
	for y := 0; y < 144; y++ {
		for x := 0; x < 160; x++ {
			col := mon.Gameboy.PreparedData[x][y]
			rgb := color.RGBA{R: col[0], G: col[1], B: col[2], A: 0xFF}
			mon.picture.Pix[(143-y)*160+x] = rgb
		}
	}

	spr := pixel.NewSprite(pixel.Picture(mon.picture), pixel.R(0, 0, 160, 144))
	r, g, b := GetPaletteColour(0)
	bg := color.RGBA{R: r, G: g, B: b, A: 0xFF}
	mon.Window.Clear(bg)
	spr.Draw(mon.Window, pixel.IM.Scaled(pixel.ZV, PIXEL_SCALE))
	mon.Window.Update()
}

func (mon *PixelsMonitor) SetTitle(fps int) {
	title := fmt.Sprintf("GoBoy - %s", mon.Gameboy.Memory.Cart.Name)
	if fps != 0 {
		title += fmt.Sprintf(" (FPS: %2v)", fps)
	}
	log.Println(title)
	mon.Window.SetTitle(title)
}

// Mapping from keys to GB index.
var key_map = map[pixelgl.Button]byte{
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
var extra_map = map[pixelgl.Button]func(*PixelsMonitor){
	// Change GB colour palette
	pixelgl.KeyEqual: func(mon *PixelsMonitor) {
		current_palette = (current_palette + 1) % byte(len(palettes))
	},

	// GPU debugging
	pixelgl.KeyQ: func(mon *PixelsMonitor) {
		mon.Gameboy.Debug.HideBackground = !mon.Gameboy.Debug.HideBackground
	},
	pixelgl.KeyW: func(mon *PixelsMonitor) {
		mon.Gameboy.Debug.HideSprites = !mon.Gameboy.Debug.HideSprites
	},

	// CPU debugging
	pixelgl.KeyE: func(mon *PixelsMonitor) {
		mon.Gameboy.Debug.OutputOpcodes = !mon.Gameboy.Debug.OutputOpcodes
	},

	// Audio channel debugging
	pixelgl.Key7: func(mon *PixelsMonitor) {
		mon.Gameboy.Debug.MuteChannel1 = !mon.Gameboy.Debug.MuteChannel1
		log.Print("Channel 1 mute =", mon.Gameboy.Debug.MuteChannel1)
	},
	pixelgl.Key8: func(mon *PixelsMonitor) {
		mon.Gameboy.Debug.MuteChannel2 = !mon.Gameboy.Debug.MuteChannel2
		log.Print("Channel 2 mute =", mon.Gameboy.Debug.MuteChannel2)
	},
	pixelgl.Key9: func(mon *PixelsMonitor) {
		mon.Gameboy.Debug.MuteChannel3 = !mon.Gameboy.Debug.MuteChannel3
		log.Print("Channel 3 mute =", mon.Gameboy.Debug.MuteChannel3)
	},
	pixelgl.Key0: func(mon *PixelsMonitor) {
		mon.Gameboy.Debug.MuteChannel4 = !mon.Gameboy.Debug.MuteChannel4
		log.Print("Channel 4 mute =", mon.Gameboy.Debug.MuteChannel4)
	},
}

func (mon *PixelsMonitor) ProcessInput() {
	for key, offset := range key_map {
		if mon.Window.JustPressed(key) {
			mon.Gameboy.InputMask = bits.Reset(mon.Gameboy.InputMask, offset)
			mon.Gameboy.RequestInterrupt(4) // Joypad interrupt
		}
		if mon.Window.JustReleased(key) {
			mon.Gameboy.InputMask = bits.Set(mon.Gameboy.InputMask, offset)
		}
	}
	// Extra keys not related to emulation
	for key, f := range extra_map {
		if mon.Window.JustPressed(key) {
			f(mon)
		}
	}
}
