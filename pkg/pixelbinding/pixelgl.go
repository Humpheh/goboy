package pixelbinding

import (
	"image/color"
	"log"
	"math"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"

	"github.com/Humpheh/goboy/pkg/gb"
)

// DefaultPixelScale is the default multiplier on the pixels on display
const DefaultPixelScale float64 = 3

// pixelsIOBinding binds screen output and input using the pixels library.
type pixelsIOBinding struct {
	window  *opengl.Window
	picture *pixel.PictureData
}

// Run runs the Gameboy using pixel IOBinding
func Run(start func(self gb.IOBinding)) {
	opengl.Run(func() {
		window, err := opengl.NewWindow(opengl.WindowConfig{
			Title: "GoBoy",
			Bounds: pixel.R(
				0, 0,
				gb.ScreenWidth*DefaultPixelScale, gb.ScreenHeight*DefaultPixelScale,
			),
			Resizable: true,
		})
		if err != nil {
			log.Fatalf("failed to create window: %v", err)
		}

		picture := &pixel.PictureData{
			Pix:    make([]color.RGBA, gb.ScreenWidth*gb.ScreenHeight),
			Stride: gb.ScreenWidth,
			Rect:   pixel.R(0, 0, gb.ScreenWidth, gb.ScreenHeight),
		}

		monitor := pixelsIOBinding{
			window:  window,
			picture: picture,
		}

		monitor.updateCamera()

		// Start the game loop with the monitor
		start(&monitor)
	})
}

func (mon *pixelsIOBinding) SetEnableVSync(enable bool) {
	mon.window.SetVSync(enable)
}

// updateCamera updates the window camera to center the output.
func (mon *pixelsIOBinding) updateCamera() {
	xScale := mon.window.Bounds().W() / gb.ScreenWidth
	yScale := mon.window.Bounds().H() / gb.ScreenHeight
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

// toggleFullscreen toggles the fullscreen window on the main monitor.
func (mon *pixelsIOBinding) toggleFullscreen() {
	if mon.window.Monitor() == nil {
		monitor := opengl.PrimaryMonitor()
		mon.window.SetMonitor(monitor)
	} else {
		mon.window.SetMonitor(nil)
	}
}

var keyMap = map[pixel.Button]gb.Button{
	pixel.KeyZ:         gb.ButtonA,
	pixel.KeyX:         gb.ButtonB,
	pixel.KeyBackspace: gb.ButtonSelect,
	pixel.KeyEnter:     gb.ButtonStart,
	pixel.KeyRight:     gb.ButtonRight,
	pixel.KeyLeft:      gb.ButtonLeft,
	pixel.KeyUp:        gb.ButtonUp,
	pixel.KeyDown:      gb.ButtonDown,

	pixel.KeyEscape: gb.ButtonPause,
	pixel.KeyEqual:  gb.ButtonChangePallete,
	pixel.KeyQ:      gb.ButtonToggleBackground,
	pixel.KeyW:      gb.ButtonToggleSprites,
	pixel.KeyE:      gb.ButtonToggleOutputOpCode,
	pixel.KeyD:      gb.ButtonPrintBGMap,
	pixel.Key7:      gb.ButtonToggleSoundChannel1,
	pixel.Key8:      gb.ButtonToggleSoundChannel2,
	pixel.Key9:      gb.ButtonToggleSoundChannel3,
	pixel.Key0:      gb.ButtonToggleSoundChannel4,
}

// ProcessButtonInput checks the input and process it.
func (mon *pixelsIOBinding) ProcessButtonInput() gb.ButtonInput {
	if mon.window.JustPressed(pixel.KeyF) {
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
