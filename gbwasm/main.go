package main

//+build js,wasm

import (
	"flag"
	"log"
	"time"

	"fmt"

	"syscall/js"

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
	c := make(chan struct{}, 0)

	cb := js.NewCallback(func(args []js.Value) {
		data := args[0].Get("detail").String()
		bytes := []byte(data)
		log.Print(bytes)
		startGB(bytes)
		//move := js.Global().Get("document").Call("getElementById", "myText").Get("value").String()
		//fmt.Println(move)
	})
	js.Global().Get("document").Call("addEventListener", "load-file", cb)

	js.Global().Get("document").Call("testMe", "wat")

	flag.Parse()
	fmt.Println(fmt.Sprintf(logo, version))
	<-c
}

func startGB(data []byte) {
	opts := []gb.GameboyOption{gb.WithCGBEnabled()}
	gameboy, err := gb.NewGameboyFromBytes(data, opts...)
	if err != nil {
		log.Fatal(err)
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
