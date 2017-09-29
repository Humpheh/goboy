package main

import (
	"flag"
	"github.com/faiface/pixel/pixelgl"
	"github.com/humpheh/goboy/gb"
	"log"
	"os"
	"runtime/pprof"
	"time"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var rom = flag.String("rom", "", "location of rom file")

func main() {
	pixelgl.Run(start)
}

func start() {
	flag.Parse()
	if *cpuprofile != "" {
		log.Print("start profile")
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	rom_file := *rom
	if rom_file == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	gameboy := gb.Gameboy{}
	err := gameboy.Init(rom_file)
	if err != nil {
		log.Fatal(err)
	}

	monitor := gb.GetPixelsMonitor(&gameboy)

	perframe := time.Second / 60
	ticker := time.NewTicker(perframe)
	start := time.Now()

	cycles := 0
	frames := 0
	for range ticker.C {
		frames++
		monitor.ProcessInput()

		cycles += gameboy.Update()
		monitor.RenderScreen()

		since := time.Since(start)
		if since > time.Second {
			start = time.Now()
			monitor.SetTitle(monitor.Frames)
			monitor.Frames = 0
		}
	}
}
