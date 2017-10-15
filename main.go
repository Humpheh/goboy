package main

import (
	"flag"
	"github.com/Humpheh/goboy/gb"
	"github.com/faiface/pixel/pixelgl"
	"log"
	"os"
	"runtime/pprof"
	"time"
)

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

	// Initalise the GameBoy.
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

	cycles := 0
	frames := 0
	for range ticker.C {
		if !monitor.IsRunning() {
			return
		}

		frames++
		monitor.ProcessInput()

		if gameboy.IsGameLoaded() {
			cycles += gameboy.Update()
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
