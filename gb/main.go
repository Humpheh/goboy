package main

import (
	"github.com/humpheh/gob"
	"log"
	"os"
	"flag"
	"runtime/pprof"
	"github.com/faiface/pixel/pixelgl"
	"time"
)

func exitError(msg string) {
	print(msg, "\n")
	print("Usage: gob romfile\n")
	os.Exit(0)
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var rom = flag.String("rom", "", "location of rom file")

func main() {
	pixelgl.Run(main2)
}

func main2() {
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
		exitError("must supply rom file")
	}

	gb := gob.Gameboy{}
	err := gb.Init(rom_file)
	if err != nil {
		exitError(err.Error())
	}

	monitor := gob.GetPixelsMonitor(&gb)
	//defer monitor.Destroy()

	perframe := time.Second / 60
	ticker := time.NewTicker(perframe)
	start := time.Now()

	cycles := 0
	frames := 0
	for range ticker.C {
		frames++
		monitor.ProcessInput()
		cycles += gb.Update()
		monitor.RenderScreen()

		since := time.Since(start)
		if since > time.Second {
			start = time.Now()
			log.Print("FPS:", monitor.Frames)
			monitor.Frames = 0
		}
	}
}
