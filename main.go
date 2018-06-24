package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"fmt"

	"github.com/Humpheh/goboy/gb"
	"github.com/Humpheh/goboy/gb/gbio/iopixel"
	"github.com/faiface/pixel/pixelgl"
)

// The version of GoBoy
const version = "v0.1"

const logo = `
    ______      ____
   / ____/___  / __ )____  __  __
  / / __/ __ \/ __  / __ \/ / / /
 / /_/ / /_/ / /_/ / /_/ / /_/ /
 \____/\____/_____/\____/\__, /
                 %6s /____/
`

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	rom        = flag.String("rom", "", "location of rom file (required)")
	sound      = flag.Bool("sound", false, "set to enable sound emulation (experimental)")
	vsyncOff   = flag.Bool("disableVsync", false, "set to disable vsync")
	cgbMode    = flag.Bool("cgb", false, "set to enable cgb mode")
)

func main() {
	pixelgl.Run(start)
}

func start() {
	flag.Parse()
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
	fmt.Printf(" %-5v: %v\n", "ROM", *rom)
	fmt.Printf(" %-5v: %v\n", "Sound", *sound)
	fmt.Printf(" %-5v: %v\n\n", "CGB", *cgbMode)

	// Initialise the GameBoy with the flag options
	var opts []gb.GameboyOption
	if *cgbMode {
		opts = append(opts, gb.WithCGBEnabled())
	}
	if *sound {
		opts = append(opts, gb.WithSound())
	}
	gameboy, err := gb.NewGameboy(*rom, opts...)
	if err != nil {
		log.Fatal(err)
	}

	monitor := iopixel.NewPixelsIOBinding(gameboy, *vsyncOff)

	perframe := time.Second / gb.FramesSecond
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
			monitor.SetTitle(monitor.Frames)
			monitor.Frames = 0
		}
	}
}
