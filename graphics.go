package gob

import (
	"github.com/veandco/go-sdl2/sdl"
)

const GPU_PIXEL_SIZE int32 = 5

func toColour(cols [3]int) uint32 {
	return uint32(cols[2]) << 16 | uint32(cols[1]) << 8 | uint32(cols[0])
}

func (gb *Gameboy) RenderScreen(surface *sdl.Surface) {
	for x := 0; x < 160; x++ {
		for y := 0; y < 144; y++ {
			rect := sdl.Rect{int32(x) * GPU_PIXEL_SIZE, int32(y) * GPU_PIXEL_SIZE, GPU_PIXEL_SIZE, GPU_PIXEL_SIZE}
			surface.FillRect(&rect, toColour(gb.ScreenData[x][y]))
		}
	}
}
