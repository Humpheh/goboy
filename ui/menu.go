package ui

import (
	"github.com/faiface/pixel/text"
	"github.com/faiface/pixel"
	"io/ioutil"
	"log"
	"path"
	"bytes"
	"github.com/faiface/pixel/pixelgl"
	"github.com/golang/freetype/truetype"
	"image/color"
	"github.com/Humpheh/goboy/ui/uifont"
	"os"
	"path/filepath"
)

type ROMOption struct {
	file     string
	name     string
	selected bool
}

type Menu struct {
	txt        *text.Text
	romList    []*ROMOption
	romIndex   int
	background *pixel.PictureData
}

func (menu *Menu) Init() {
	menu.txt = text.New(pixel.ZV, getFont())
	menu.loadROMList()
	menu.updateROMText()

	col := color.RGBA{R: 0, G: 0, B: 0, A: 0xFF / 2}
	pixels := make([]color.RGBA, 144*160)
	for i := range pixels {
		pixels[i] = col
	}
	menu.background = &pixel.PictureData{
		Pix:    pixels,
		Stride: 160,
		Rect:   pixel.R(0, 0, 160, 144),
	}
}

func (menu *Menu) loadROMList() {
	dir := "/Users/humphreyshotton/pygb/_roms/"
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	menu.romList = []*ROMOption{}
	for _, file := range files {
		if file.IsDir() || !menu.isROMFile(file) {
			continue
		}
		menu.romList = append(menu.romList, &ROMOption{
			file: path.Join(dir, file.Name()),
			name: file.Name(),
		})
	}
}

func (menu *Menu) isROMFile(f os.FileInfo) bool {
	switch filepath.Ext(f.Name()) {
	case ".gb", ".gbc", ".rom", ".bin":
		return true
	}
	return false
}

func (menu *Menu) updateROMText() {
	menu.txt.Clear()
	menu.txt.Dot = pixel.V(0, 0)
	var buffer bytes.Buffer
	for i, rom := range menu.romList {
		if i == menu.romIndex {
			buffer.WriteString("~ ")
		} else {
			buffer.WriteString("  ")
		}
		buffer.WriteString(rom.name)
		buffer.WriteString("\n")
	}
	menu.txt.WriteString(buffer.String())
}

func (menu *Menu) GetROMLocation() string {
	if len(menu.romList) == 0 {
		log.Print("no ROMs")
		return ""
	}
	rom := menu.romList[menu.romIndex]
	return rom.file
}

func (menu *Menu) Render(window *pixelgl.Window) {
	scale := float64(3)
	spr := pixel.NewSprite(pixel.Picture(menu.background), pixel.R(0, 0, 160, 144))
	spr.Draw(window, pixel.IM.Scaled(pixel.ZV, scale))

	m := pixel.IM.Moved(pixel.V(-78, 144/2-menu.txt.LineHeight)).Scaled(pixel.ZV, scale)
	menu.txt.Draw(window, m)
}

var romKeyMap = map[pixelgl.Button]func(*Menu) string{
	pixelgl.KeyUp: func(menu *Menu) string {
		menu.romIndex--
		if menu.romIndex < 0 {
			menu.romIndex = len(menu.romList) - 1
		}
		menu.updateROMText()
		return ""
	},
	pixelgl.KeyDown: func(menu *Menu) string {
		menu.romIndex++
		if menu.romIndex >= len(menu.romList) {
			menu.romIndex = 0
		}
		menu.updateROMText()
		return ""
	},
	pixelgl.KeyZ: func(menu *Menu) string {
		return menu.GetROMLocation()
	},
}

func (menu *Menu) ProcessInput(window *pixelgl.Window) string {
	for key, f := range romKeyMap {
		if window.JustPressed(key) {
			str := f(menu)
			if str != "" {
				return str
			}
		}
	}
	return ""
}

func getFont() *text.Atlas {
	ttfFromBytes, err := truetype.Parse(uifont.GBFONT)
	if err != nil {
		panic(err)
	}

	return text.NewAtlas(truetype.NewFace(ttfFromBytes, &truetype.Options{
		Size:              8,
		GlyphCacheEntries: 1,
	}), text.ASCII)
}
