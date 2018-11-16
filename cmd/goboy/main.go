package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/Humpheh/goboy/dialog"

	"fmt"

	"github.com/Humpheh/goboy/pkg/gb"
	"github.com/Humpheh/goboy/pkg/gbio"
	"github.com/Humpheh/goboy/pkg/gbio/iopixel"
	"github.com/faiface/pixel/pixelgl"
)

// The version of GoBoy
var version = "develop"

const logo = `
    ______      ____
   / ____/___  / __ )____  __  __
  / / __/ __ \/ __  / __ \/ / / /
 / /_/ / /_/ / /_/ / /_/ / /_/ /
 \____/\____/_____/\____/\__, /
%23s /____/
`

var (
	mute    = flag.Bool("mute", false, "mute sound output")
	dmgMode = flag.Bool("dmg", false, "set to force dmg mode")

	cpuprofile  = flag.String("cpuprofile", "", "write cpu profile to file (debugging)")
	vsyncOff    = flag.Bool("disableVsync", false, "set to disable vsync (debugging)")
	saveState   = flag.String("load", "", "location of save state to load (experimental)")
	stepThrough = flag.Bool("stepthrough", false, "step through opcodes (debugging)")
	unlocked    = flag.Bool("unlocked", false, "if to unlock the cpu speed (debugging)")
)

func main() {
	flag.Parse()
	rom := flag.Arg(0)

	// Check if rom was passed in as first argument. If not, we
	// should prompt the user to select a rom file
	if rom == "" {
		var err error
		rom, err = dialog.File().
			Filter("GameBoy ROM", "zip", "gb", "gbc", "bin").
			Title("Load GameBoy ROM File").Load()
		if err != nil {
			os.Exit(1)
		}
	}

	pixelgl.Run(func() {
		start(rom)
	})
}

func start(rom string) {
	// Check if to run the CPU profile
	if *cpuprofile != "" {
		log.Print("start profile")
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatalf("Failed to create CPU profile: %v", err)
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			log.Fatalf("Failed to start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}

	fmt.Println(fmt.Sprintf(logo, version))
	fmt.Printf(" %-5v: %v\n", "ROM", rom)
	fmt.Printf(" %-5v: %v\n", "Sound", !*mute)
	fmt.Printf(" %-5v: %v\n\n", "CGB", !*dmgMode)

	var opts []gb.GameboyOption
	if !*dmgMode {
		opts = append(opts, gb.WithCGBEnabled())
	}
	if !*mute {
		opts = append(opts, gb.WithSound())
	}

	var gameboy *gb.Gameboy
	var err error
	if rom != "" {
		// Initialise the GameBoy with the flag options
		gameboy, err = gb.NewGameboy(rom, opts...)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// Load the gameboy from a save state
		gameboy, err = gb.NewGameboyFromGob(*saveState)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *stepThrough {
		gameboy.Debug.OutputOpcodes = true
	}

	monitor := iopixel.NewPixelsIOBinding(gameboy, *vsyncOff || *unlocked)
	startGB(gameboy, monitor)
}

func startGB(gameboy *gb.Gameboy, monitor gbio.IOBinding) {
	var div time.Duration = 1.0
	if *unlocked {
		div = 1000.0
	}

	perframe := time.Second / gb.FramesSecond / div
	ticker := time.NewTicker(perframe)
	start := time.Now()

	frames := 0
	for range ticker.C {
		if !monitor.IsRunning() {
			return
		}

		frames++
		monitor.ProcessInput()

		if gameboy.IsGameLoaded() {
			_ = gameboy.Update()
		}
		monitor.RenderScreen()

		since := time.Since(start)
		if since > time.Second {
			start = time.Now()
			monitor.SetTitle(frames)
			frames = 0
		}
	}
}
