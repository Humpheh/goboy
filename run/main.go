package main

import (
	"github.com/humpheh/gob"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/humpheh/gob/bits"
	"log"
	"os"
)

func exitError(msg string) {
	print(msg, "\n")
	print("Usage: gob romfile\n")
	os.Exit(0)
}

func main() {
	if len(os.Args) == 1 {
		exitError("")
	}
	rom_file := os.Args[1]
	if rom_file == "" {
		exitError("")
	}

	cpu := gob.CPU{}
	mem := gob.Memory{}
	gb := gob.Gameboy{}

	gb.CPU = &cpu
	gb.Memory = &mem

	mem.GB = &gb
	err := mem.LoadCart(rom_file)
	if err != nil {
		exitError("Could not load rom file")
	}

	monitor := gob.GetSDLMonitor(&gb)
	defer monitor.Destroy()

	gb.Init()

	cycles := 0
	for true {
		cycles += gb.Update()
		monitor.RenderScreen()

		evt := sdl.PollEvent()
		if evt != nil {
			switch kevnt := evt.(type) {
			case *sdl.KeyDownEvent:
				val := getKeyIndex(kevnt.Keysym.Sym)
				if val != 255 {
					gb.InputMask = bits.Reset(gb.InputMask, val)
				}
				log.Print("kd")

			case *sdl.KeyUpEvent:
				val := getKeyIndex(kevnt.Keysym.Sym)
				if val != 255 {
					gb.InputMask = bits.Set(gb.InputMask, val)
				}
				log.Print("ku")
			}
		}

		//time.Sleep(time.Microsecond)
	}
}

func getKeyIndex(key sdl.Keycode) byte {
	switch key {
	case sdl.K_z:
		// pressed a
		return 0
	case sdl.K_x:
		// pressed b
		return 1
	case sdl.K_BACKSPACE:
		// pressed select
		return 2
	case sdl.K_EQUALS:
		// pressed start
		return 3
	case sdl.K_RIGHT:
		// pressed right
		return 4
	case sdl.K_LEFT:
		// pressed left
		return 5
	case sdl.K_UP:
		// pressed up
		return 6
	case sdl.K_DOWN:
		// pressed down
		return 7
	}
	// TODO: Returning 255 is naff
	return 255
}