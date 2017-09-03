package main

import (
	"github.com/humpheh/gob"
	"github.com/veandco/go-sdl2/sdl"
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
	mem.LoadCart("/Users/humphreyshotton/pygb/_roms/cpu_instrs.gb")//tetris.gb")//06-ld r,r.gb")//

	gb.Init()

	cycles := 0
	for true {
		cycles += gb.Update()
		gb.RenderScreen(surface)
		window.UpdateSurface()
	}
}
