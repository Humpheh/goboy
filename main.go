package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"fmt"

	"github.com/Humpheh/goboy/gb"
	"github.com/Humpheh/goboy/gb/gbio"
	"github.com/Humpheh/goboy/gb/gbio/iopixel"
	"github.com/faiface/pixel/pixelgl"
)

// The version of GoBoy
const version = "v0.2"

const logo = `
    ______      ____
   / ____/___  / __ )____  __  __
  / / __/ __ \/ __  / __ \/ / / /
 / /_/ / /_/ / /_/ / /_/ / /_/ /
 \____/\____/_____/\____/\__, /
                 %6s /____/
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
	pixelgl.Run(start)
}

func start() {
	flag.Parse()
	rom := flag.Arg(0)

	// Check if to run the CPU profile
	if *cpuprofile != "" {
		log.Print("start profile")
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
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
