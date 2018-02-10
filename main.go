package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"fmt"

	"github.com/Humpheh/goboy/gb"
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

	// Initialise the GameBoy.
	gameboy := gb.Gameboy{
		EnableSound: *sound,
	}

	// Get the name of the ROM cartridge.
	romFile := *rom
	if romFile == "" {
		// If no rom, use menu
		gameboy.EnableSound = true
	} else {
		err := gameboy.Init(romFile, *cgbMode)
		if err != nil {
			flag.PrintDefaults()
			log.Fatal(err)
		}
	}

	monitor := gb.NewPixelsIOBinding(&gameboy, *vsyncOff, *cgbMode)

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
