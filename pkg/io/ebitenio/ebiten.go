package ebitenio

import (
	"fmt"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/Humpheh/goboy/pkg/gb"
)

const defaultScale = 3

type ebitenIOBinding struct {
	running    bool
	imageMutex sync.Mutex
	image      *ebiten.Image

	step                     func(binding gb.IOBinding)
	previouslyPressedButtons []ebiten.Key
}

func New() gb.IOBinding {
	ebiten.SetWindowSize(160*defaultScale, 144*defaultScale)
	ebiten.SetWindowTitle("GoBoy")
	ebiten.SetTPS(ebiten.SyncWithFPS)

	return &ebitenIOBinding{
		image:   ebiten.NewImage(160, 144),
		running: false,
	}
}

func (g *ebitenIOBinding) Update() error {
	g.step(g)
	//g.SetTitle(fmt.Sprintf("GoBoy (FPS: %0.2f)", ebiten.ActualFPS()))
	return nil
}

func (e *ebitenIOBinding) Draw(screen *ebiten.Image) {
	e.imageMutex.Lock()
	defer e.imageMutex.Unlock()

	opts := ebiten.DrawImageOptions{}
	screen.DrawImage(e.image, &opts)
}

func (e *ebitenIOBinding) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 160, 144
}

func (e *ebitenIOBinding) Start(step func(binding gb.IOBinding)) {
	e.step = step
	e.running = true
	if err := ebiten.RunGame(e); err != nil {
		e.running = false
		panic(err)
	}
}

func (e *ebitenIOBinding) SetVSync(enabled bool) {
	ebiten.SetVsyncEnabled(enabled)
}

func (e *ebitenIOBinding) IsRunning() bool {
	return e.running
}

func (e *ebitenIOBinding) Render(screen *[160][144][3]uint8) {
	e.imageMutex.Lock()
	defer e.imageMutex.Unlock()
	//for y := 0; y < gb.ScreenHeight; y++ {
	//	for x := 0; x < gb.ScreenWidth; x++ {
	//		col := screen[x][y]
	//		rgb := color.RGBA{R: col[0], G: col[1], B: col[2], A: 0xFF}
	//		e.image.Set(x, y, rgb)
	//	}
	//}
	pix := make([]byte, 160*144*4)
	for y := 0; y < 144; y++ {
		for x := 0; x < 160; x++ {
			pix[(y*160+x)*4] = screen[x][y][0]
			pix[(y*160+x)*4+1] = screen[x][y][1]
			pix[(y*160+x)*4+2] = screen[x][y][2]
			pix[(y*160+x)*4+3] = 0xff
		}
	}
	e.image.WritePixels(pix)
}

func (e *ebitenIOBinding) SetTitle(title string) {
	ebiten.SetWindowTitle(title + fmt.Sprintf(" (e: %v)", ebiten.CurrentFPS()))
}

var keyMap = map[ebiten.Key]gb.Button{
	ebiten.KeyZ:         gb.ButtonA,
	ebiten.KeyX:         gb.ButtonB,
	ebiten.KeyBackspace: gb.ButtonSelect,
	ebiten.KeyEnter:     gb.ButtonStart,
	ebiten.KeyRight:     gb.ButtonRight,
	ebiten.KeyLeft:      gb.ButtonLeft,
	ebiten.KeyUp:        gb.ButtonUp,
	ebiten.KeyDown:      gb.ButtonDown,

	ebiten.KeyEscape: gb.ButtonPause,
	ebiten.KeyEqual:  gb.ButtonChangePallete,
	ebiten.KeyQ:      gb.ButtonToggleBackground,
	ebiten.KeyW:      gb.ButtonToggleSprites,
	ebiten.KeyE:      gb.ButttonToggleOutputOpCode,
	ebiten.KeyD:      gb.ButtonPrintBGMap,
	ebiten.Key7:      gb.ButtonToggleSoundChannel1,
	ebiten.Key8:      gb.ButtonToggleSoundChannel2,
	ebiten.Key9:      gb.ButtonToggleSoundChannel3,
	ebiten.Key0:      gb.ButtonToggleSoundChannel4,
}

func (e *ebitenIOBinding) keyWasPressed(key ebiten.Key) bool {
	for _, k := range e.previouslyPressedButtons {
		if k == key {
			return true
		}
	}
	return false
}

func (e *ebitenIOBinding) ButtonInput() gb.ButtonInput {
	// TODO: fullscreen?

	var buttonInput gb.ButtonInput
	var pressedButtons []ebiten.Key
	for handledKey, button := range keyMap {
		wasPressed := e.keyWasPressed(handledKey)
		isPressed := ebiten.IsKeyPressed(handledKey)
		if isPressed {
			pressedButtons = append(pressedButtons, handledKey)
		}
		if !wasPressed && isPressed {
			buttonInput.Pressed = append(buttonInput.Pressed, button)
		} else if wasPressed && !isPressed {
			buttonInput.Released = append(buttonInput.Released, button)
		}
	}
	e.previouslyPressedButtons = pressedButtons
	return buttonInput
}
