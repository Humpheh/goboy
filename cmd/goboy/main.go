package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/Humpheh/goboy/pkg/ebitenbinding"
	"github.com/Humpheh/goboy/pkg/gb"
)

// The version of GoBoy
var version = "develop"

var (
	mute    = flag.Bool("mute", false, "mute sound output")
	dmgMode = flag.Bool("dmg", false, "set to force dmg mode")

	debugEnabled = flag.Bool("debug", false, "enable debug functionality on startup")
	cpuprofile   = flag.String("cpuprofile", "", "write cpu profile to file (debugging)")
	vsyncOff     = flag.Bool("disableVsync", false, "set to disable vsync (debugging)")
	stepThrough  = flag.Bool("stepthrough", false, "step through opcodes (debugging)")
	unlocked     = flag.Bool("unlocked", false, "if to unlock the cpu speed (debugging)")
)

func main() {
	flag.Parse()
	ebitenbinding.Run(start)
}

func start(binding gb.IOBinding) {
	rom := flag.Arg(0)
	if rom == "" {
		log.Fatal("No ROM file specified. Please provide a ROM file as an argument.")
	}

	// If the CPU profile flag is set, then setup the profiling
	if *cpuprofile != "" {
		startCPUProfiling()
		defer pprof.StopCPUProfile()
	}

	// Print the logo and the run settings to the console
	fmt.Printf("[\u001B[1;32mGoBoy\u001B[0m] %v :: apu=%v cgb=%v\n", version, !*mute, !*dmgMode)

	var opts []gb.GameboyOption
	if !*dmgMode {
		opts = append(opts, gb.WithCGBEnabled())
	}
	if !*mute {
		opts = append(opts, gb.WithSound())
	}
	if *debugEnabled {
		opts = append(opts, gb.WithDebugEnabled())
	}

	// Initialise the GameBoy with the flag options
	gameboy, err := gb.New(rom, opts...)
	if err != nil {
		log.Fatal(err)
	}
	if *stepThrough {
		gameboy.Debug.OutputOpcodes = true
	}

	// Create the monitor for pixels
	enableVSync := !(*vsyncOff || *unlocked)
	binding.SetEnableVSync(enableVSync)
	startGBLoop(gameboy, binding)
}

const (
	// fastForwardSpeed is the emulation speed multiplier while fast forward is active.
	fastForwardSpeed = 2

	// maxFrameLag is how far the loop may fall behind schedule before it
	// stops trying to catch up, so a stall isn't followed by a burst of
	// frames.
	maxFrameLag = 100 * time.Millisecond
)

func startGBLoop(gameboy *gb.Gameboy, monitor gb.IOBinding) {
	frameTime := time.Second / gb.FramesSecond
	fastForward := false

	start := time.Now()
	nextFrame := start
	frames := 0

	var cartName string
	if gameboy.IsCartLoaded() {
		cartName = gameboy.GetLoadedCart().GetName()
	}

	for monitor.IsRunning() {
		frames++

		buttons := monitor.ProcessButtonInput()
		fastForward = isFastForwarding(fastForward, buttons)
		gameboy.ProcessInput(buttons)

		_ = gameboy.Update()
		monitor.Render(&gameboy.PreparedData)

		since := time.Since(start)
		if since > time.Second {
			start = time.Now()

			extra := ""
			if fastForward {
				extra = fmt.Sprintf(" (%xx)", fastForwardSpeed)
			}
			title := fmt.Sprintf("GoBoy - %s (FPS: %2v)%v", cartName, frames, extra)
			monitor.SetTitle(title)
			frames = 0
		}

		if *unlocked {
			continue
		}

		// Wait until the next frame is due
		interval := frameTime
		if fastForward {
			interval /= fastForwardSpeed
		}
		nextFrame = nextFrame.Add(interval)
		if wait := time.Until(nextFrame); wait > 0 {
			time.Sleep(wait)
		} else if wait < -maxFrameLag {
			nextFrame = time.Now()
		}
	}
}

// isFastForwarding returns whether fast forward is held after applying the
// button changes in buttons to the current state.
func isFastForwarding(current bool, buttons gb.ButtonInput) bool {
	if buttons.IsPressed(gb.ButtonFastForward) {
		current = true
	}
	if buttons.IsReleased(gb.ButtonFastForward) {
		current = false
	}
	return current
}

// Start the CPU profile to a the file passed in from the flag.
func startCPUProfiling() {
	log.Print("Starting CPU profile...")
	f, err := os.Create(*cpuprofile)
	if err != nil {
		log.Fatalf("Failed to create CPU profile: %v", err)
	}
	err = pprof.StartCPUProfile(f)
	if err != nil {
		log.Fatalf("Failed to start CPU profile: %v", err)
	}
}
