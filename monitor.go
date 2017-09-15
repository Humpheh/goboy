// TODO: Rename monitor so something better
package gob

import (
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel"
	"image/color"
	"github.com/humpheh/gob/bits"
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
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	mon.Window = win
	cam := pixel.IM.Scaled(pixel.ZV, 1).Moved(win.Bounds().Center().Sub(pixel.ZV))
	win.SetMatrix(cam)

	mon.picture = &pixel.PictureData{
		Pix:    make([]color.RGBA, 144 * 160),
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
	mon.Window.Clear(color.White)
	spr.Draw(mon.Window, pixel.IM.Scaled(pixel.ZV, PIXEL_SCALE))
	mon.Window.Update()
}

func (mon *PixelsMonitor) Destroy() {
}

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
}