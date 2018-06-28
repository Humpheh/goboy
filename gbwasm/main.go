package main

import (
	"flag"
	"log"
	"time"

	"fmt"

	"github.com/Humpheh/goboy/gb"
	"github.com/Humpheh/goboy/gb/gbio/ionet"
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
	rom     = flag.String("rom", "", "location of rom file (required)")
	sound   = flag.Bool("sound", false, "set to enable sound emulation (experimental)")
	cgbMode = flag.Bool("cgb", false, "set to enable cgb mode")
)

func main() {
	flag.Parse()
	fmt.Println(fmt.Sprintf(logo, version))
	fmt.Printf(" %-5v: %v\n", "ROM", *rom)
	fmt.Printf(" %-5v: %v\n", "Sound", *sound)
	fmt.Printf(" %-5v: %v\n\n", "CGB", *cgbMode)

	var opts []gb.GameboyOption
	if *cgbMode {
		opts = append(opts, gb.WithCGBEnabled())
	}
	if *sound {
		opts = append(opts, gb.WithSound())
	}

	var gameboy *gb.Gameboy
	var err error
	if *rom != "" {
		// Initialise the GameBoy with the flag options
		gameboy, err = gb.NewGameboy(*rom, opts...)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("No rom supplied")
	}

	monitor := ionet.WASMIOBinding{}

	perframe := time.Second / gb.FramesSecond
	ticker := time.NewTicker(perframe)
	start := time.Now()

	frames := 0
	for range ticker.C {
		frames++
		monitor.ProcessInput()

		if gameboy.IsGameLoaded() {
			_ = gameboy.Update()
		}
		monitor.RenderScreen()

		since := time.Since(start)
		if since > time.Second {
			start = time.Now()
		}
	}
}
