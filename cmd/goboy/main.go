package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/Humpheh/goboy/pkg/gb"
	"github.com/Humpheh/goboy/pkg/pixelbinding"
)

// The version of GoBoy
var version = "develop"

var (
	mute    = flag.Bool("mute", false, "mute sound output")
	dmgMode = flag.Bool("dmg", false, "set to force dmg mode")

	cpuprofile  = flag.String("cpuprofile", "", "write cpu profile to file (debugging)")
	vsyncOff    = flag.Bool("disableVsync", false, "set to disable vsync (debugging)")
	stepThrough = flag.Bool("stepthrough", false, "step through opcodes (debugging)")
	unlocked    = flag.Bool("unlocked", false, "if to unlock the cpu speed (debugging)")
)

func main() {
	flag.Parse()
	pixelbinding.Run(start)
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

	if *unlocked {
		*mute = true
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

func startGBLoop(gameboy *gb.Gameboy, monitor gb.IOBinding) {
	frameTime := time.Second / gb.FramesSecond
	if *unlocked {
		frameTime = 1
	}

	ticker := time.NewTicker(frameTime)
	start := time.Now()
	frames := 0

	var cartName string
	if gameboy.IsCartLoaded() {
		cartName = gameboy.GetLoadedCart().GetName()
	}

	for range ticker.C {
		if !monitor.IsRunning() {
			return
		}

		frames++

		buttons := monitor.ProcessButtonInput()
		gameboy.ProcessInput(buttons)

		_ = gameboy.Update()
		monitor.Render(&gameboy.PreparedData)

		since := time.Since(start)
		if since > time.Second {
			start = time.Now()

			title := fmt.Sprintf("GoBoy - %s (FPS: %2v)", cartName, frames)
			monitor.SetTitle(title)
			frames = 0
		}
	}
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
