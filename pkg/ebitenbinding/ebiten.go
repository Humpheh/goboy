package ebitenbinding

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/Humpheh/goboy/pkg/gb"
)

// DefaultPixelScale is the default multiplier on the pixels on display.
const DefaultPixelScale = 3

// frameBytes is the size of one RGBA frame of the GameBoy screen.
const frameBytes = gb.ScreenWidth * gb.ScreenHeight * 4

// errStopped is returned from Update to end the ebiten game loop once the
// emulator has stopped.
var errStopped = errors.New("emulator stopped")

// ebitenIOBinding binds screen output and input using the ebiten library.
//
// Ebiten owns the main thread and calls Update and Draw, while the emulator
// runs in its own goroutine and calls the gb.IOBinding methods. Frames and
// input are handed between the two under mu.
type ebitenIOBinding struct {
	mu sync.Mutex
	// back is written by Render on the emulator goroutine; front is read by
	// Draw. They are swapped under mu so neither side copies a whole frame
	// while holding the lock.
	back, front []byte
	// newFrame is set when front holds a frame not yet uploaded to the GPU.
	newFrame bool
	// input accumulates key changes seen by Update until the emulator drains
	// them in ProcessButtonInput.
	input gb.ButtonInput

	running atomic.Bool
	stopped atomic.Bool
}

// Run runs the Gameboy using the ebiten IOBinding. It must be called from
// the main goroutine and blocks until the window is closed and start has
// returned.
func Run(start func(self gb.IOBinding)) {
	binding := &ebitenIOBinding{
		back:  make([]byte, frameBytes),
		front: make([]byte, frameBytes),
	}
	binding.running.Store(true)

	ebiten.SetWindowTitle("GoBoy")
	ebiten.SetWindowSize(gb.ScreenWidth*DefaultPixelScale, gb.ScreenHeight*DefaultPixelScale)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	// Draw overwrites the whole screen when a frame arrives, so keeping the
	// previous contents lets unchanged frames skip all GPU work.
	ebiten.SetScreenClearedEveryFrame(false)

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer binding.stopped.Store(true)
		start(binding)
	}()

	err := ebiten.RunGame(binding)
	binding.running.Store(false)
	if err != nil && !errors.Is(err, errStopped) {
		log.Printf("ebiten: %v", err)
	}
	<-done
}

// SetEnableVSync sets whether vertical sync is enabled.
func (b *ebitenIOBinding) SetEnableVSync(enable bool) {
	ebiten.SetVsyncEnabled(enable)
}

// IsRunning returns if the game should still be running. When
// the window is closed this will be false so the game stops.
func (b *ebitenIOBinding) IsRunning() bool {
	return b.running.Load()
}

// Render converts the frame to RGBA and queues it to be drawn on the next
// ebiten Draw call. Frames rendered faster than the display refreshes
// replace any frame that has not been drawn yet.
func (b *ebitenIOBinding) Render(screen *[160][144][3]uint8) {
	pix := b.back
	for x := range gb.ScreenWidth {
		col := &screen[x]
		i := x * 4
		for y := range gb.ScreenHeight {
			p := pix[i : i+4 : i+4]
			p[0], p[1], p[2], p[3] = col[y][0], col[y][1], col[y][2], 0xFF
			i += gb.ScreenWidth * 4
		}
	}

	b.mu.Lock()
	b.back, b.front = b.front, b.back
	b.newFrame = true
	b.mu.Unlock()
}

// SetTitle sets the title of the game window.
func (b *ebitenIOBinding) SetTitle(title string) {
	ebiten.SetWindowTitle(title)
}

// ProcessButtonInput returns the buttons pressed and released since the
// previous call.
func (b *ebitenIOBinding) ProcessButtonInput() gb.ButtonInput {
	b.mu.Lock()
	input := b.input
	b.input = gb.ButtonInput{}
	b.mu.Unlock()
	return input
}

// keyMap maps keyboard keys to GameBoy and emulator buttons.
var keyMap = []struct {
	key    ebiten.Key
	button gb.Button
}{
	{ebiten.KeyZ, gb.ButtonA},
	{ebiten.KeyX, gb.ButtonB},
	{ebiten.KeyBackspace, gb.ButtonSelect},
	{ebiten.KeyEnter, gb.ButtonStart},
	{ebiten.KeyArrowRight, gb.ButtonRight},
	{ebiten.KeyArrowLeft, gb.ButtonLeft},
	{ebiten.KeyArrowUp, gb.ButtonUp},
	{ebiten.KeyArrowDown, gb.ButtonDown},

	{ebiten.KeyTab, gb.ButtonFastForward},
	{ebiten.KeyBackslash, gb.ButtonDebugToggle},
	{ebiten.KeyEscape, gb.ButtonPause},
	{ebiten.KeyEqual, gb.ButtonChangePallete},
	{ebiten.KeyQ, gb.ButtonToggleBackground},
	{ebiten.KeyW, gb.ButtonToggleSprites},
	{ebiten.KeyE, gb.ButtonToggleOutputOpCode},
	{ebiten.KeyD, gb.ButtonPrintBGMap},
	{ebiten.KeyDigit7, gb.ButtonToggleSoundChannel1},
	{ebiten.KeyDigit8, gb.ButtonToggleSoundChannel2},
	{ebiten.KeyDigit9, gb.ButtonToggleSoundChannel3},
	{ebiten.KeyDigit0, gb.ButtonToggleSoundChannel4},
}

// Update implements ebiten.Game. It collects key changes for the emulator
// and ends the game loop once the emulator has stopped.
func (b *ebitenIOBinding) Update() error {
	if b.stopped.Load() {
		return errStopped
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	// Collect key changes for the emulator
	b.mu.Lock()
	for _, m := range keyMap {
		if inpututil.IsKeyJustPressed(m.key) {
			b.input.Pressed = append(b.input.Pressed, m.button)
		}
		if inpututil.IsKeyJustReleased(m.key) {
			b.input.Released = append(b.input.Released, m.button)
		}
	}
	b.mu.Unlock()
	return nil
}

// Draw implements ebiten.Game. The screen is the native GameBoy resolution
// (see Layout), so the frame is uploaded straight into it with no
// intermediate image or draw call, and ebiten scales it to the window.
func (b *ebitenIOBinding) Draw(screen *ebiten.Image) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.newFrame {
		return
	}
	screen.WritePixels(b.front)
	b.newFrame = false
}

// Layout implements ebiten.Game, fixing the logical screen to the GameBoy
// resolution regardless of window size.
func (b *ebitenIOBinding) Layout(_, _ int) (int, int) {
	return gb.ScreenWidth, gb.ScreenHeight
}
