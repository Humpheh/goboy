package main

import (
	"github.com/humpheh/gob"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/humpheh/gob/bits"
	"log"
)

func main() {
	sdl.Init(sdl.INIT_EVERYTHING)

	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int(gob.GPU_PIXEL_SIZE * 160), int(gob.GPU_PIXEL_SIZE * 144), sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()
	defer sdl.Quit()

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}

	cpu := gob.CPU{}
	mem := gob.Memory{}
	gb := gob.Gameboy{}

	gb.CPU = &cpu
	gb.Memory = &mem

	mem.GB = &gb
	mem.LoadCart("/Users/humphreyshotton/pygb/_roms/drmario.gb")//02-interrupts.gb")//cpu_instrs.gb")//10-bit ops.gb")//

	gb.Init()

	cycles := 0
	for true {
		cycles += gb.Update()
		gb.RenderScreen(surface)
		window.UpdateSurface()

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