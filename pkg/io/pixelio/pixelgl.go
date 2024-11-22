package pixelio

import (
	"image/color"
	"log"
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"

	"github.com/Humpheh/goboy/pkg/gb"
)

// defaultPixelScale is the default multiplier on the pixels on display
var defaultPixelScale float64 = 3

// pixelsIOBinding binds screen output and input using the pixels library.
type pixelsIOBinding struct {
	window  *pixelgl.Window
	picture *pixel.PictureData
	vsync bool
}

// New returns a new PixelGL IOBinding
func New(start func(iob gb.IOBinding)) {
	binding := &pixelsIOBinding{}
	pixelgl.Run(func() {
		start(binding)
	})
}

func (mon *pixelsIOBinding) Start() {
	windowConfig := pixelgl.WindowConfig{
		Title: "GoBoy",
		Bounds: pixel.R(
			0, 0,
			gb.ScreenWidth*defaultPixelScale, gb.ScreenHeight*defaultPixelScale,
		),
		VSync:     mon.vsync,
		Resizable: true,
	}

	var err error
	mon.window, err = pixelgl.NewWindow(windowConfig)
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}

	// Hack so that pixelgl renders on Darwin
	mon.window.SetPos(mon.window.GetPos().Add(pixel.V(0, 1)))

	mon.picture = &pixel.PictureData{
		Pix:    make([]color.RGBA, gb.ScreenWidth*gb.ScreenHeight),
		Stride: gb.ScreenWidth,
		Rect:   pixel.R(0, 0, gb.ScreenWidth, gb.ScreenHeight),
	}
	mon.updateCamera()
}

func (mon *pixelsIOBinding) SetVSync(enabled bool) {
	mon.vsync = enabled
}

// updateCamera updates the window camera to center the output.
func (mon *pixelsIOBinding) updateCamera() {
	xScale := mon.window.Bounds().W() / 160
	yScale := mon.window.Bounds().H() / 144
	scale := math.Min(yScale, xScale)

	shift := mon.window.Bounds().Size().Scaled(0.5).Sub(pixel.ZV)
	cam := pixel.IM.Scaled(pixel.ZV, scale).Moved(shift)
	mon.window.SetMatrix(cam)
}

// IsRunning returns if the game should still be running. When
// the window is closed this will be false so the game stops.
func (mon *pixelsIOBinding) IsRunning() bool {
	return !mon.window.Closed()
}

// Render renders the pixels on the screen.
func (mon *pixelsIOBinding) Render(screen *[160][144][3]uint8) {
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
func (mon *pixelsIOBinding) SetTitle(title string) {
	mon.window.SetTitle(title)
}

// Toggle the fullscreen window on the main monitor.
func (mon *pixelsIOBinding) toggleFullscreen() {
	if mon.window.Monitor() == nil {
		monitor := pixelgl.PrimaryMonitor()
		_, height := monitor.Size()
		mon.window.SetMonitor(monitor)
		defaultPixelScale = height / 144
	} else {
		mon.window.SetMonitor(nil)
		defaultPixelScale = 3
	}
}

var keyMap = map[pixelgl.Button]gb.Button{
	pixelgl.KeyZ:         gb.ButtonA,
	pixelgl.KeyX:         gb.ButtonB,
	pixelgl.KeyBackspace: gb.ButtonSelect,
	pixelgl.KeyEnter:     gb.ButtonStart,
	pixelgl.KeyRight:     gb.ButtonRight,
	pixelgl.KeyLeft:      gb.ButtonLeft,
	pixelgl.KeyUp:        gb.ButtonUp,
	pixelgl.KeyDown:      gb.ButtonDown,

	pixelgl.KeyEscape: gb.ButtonPause,
	pixelgl.KeyEqual:  gb.ButtonChangePallete,
	pixelgl.KeyQ:      gb.ButtonToggleBackground,
	pixelgl.KeyW:      gb.ButtonToggleSprites,
	pixelgl.KeyE:      gb.ButttonToggleOutputOpCode,
	pixelgl.KeyD:      gb.ButtonPrintBGMap,
	pixelgl.Key7:      gb.ButtonToggleSoundChannel1,
	pixelgl.Key8:      gb.ButtonToggleSoundChannel2,
	pixelgl.Key9:      gb.ButtonToggleSoundChannel3,
	pixelgl.Key0:      gb.ButtonToggleSoundChannel4,
}

// ProcessInput checks the input and process it.
func (mon *pixelsIOBinding) ButtonInput() gb.ButtonInput {
	if mon.window.JustPressed(pixelgl.KeyF) {
		mon.toggleFullscreen()
	}

	var buttonInput gb.ButtonInput
	for handledKey, button := range keyMap {
		if mon.window.JustPressed(handledKey) {
			buttonInput.Pressed = append(buttonInput.Pressed, button)
		}
		if mon.window.JustReleased(handledKey) {
			buttonInput.Released = append(buttonInput.Released, button)
		}
	}
	return buttonInput
}
