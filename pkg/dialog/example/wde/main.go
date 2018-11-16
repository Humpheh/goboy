package main

import (
	"fmt"
	"github.com/skelterjohn/go.wde"
	"github.com/sqweek/dialog"
	"image"
	"image/color"
	"image/draw"

	_ "github.com/skelterjohn/go.wde/init"
)

var loadR, saveR image.Rectangle

func events(events <-chan interface{}) {
	for ei := range events {
		switch e := ei.(type) {
		case wde.MouseUpEvent:
			switch e.Which {
			case wde.LeftButton:
				var f string
				var err error
				/* launching dialogs within the event loop like this has
				** a serious problem in practice. On Linux/Windows the
				** dialog is not modal, which means events on the main
				** window still get queued up.
				**
				** But the dialog functions are synchronous, so we don't
				** process any of these events until the dialog closes. End
				** result is if the user clicks multiple times they end up
				** facing multiple dialogs, one after the other.
				**
				** For this reason, it is recommended to launch dialogs
				** in a separate goroutine, and arrange modality via some
				** other mechanism (if desired). */
				if e.Where.In(loadR) {
					f, err = dialog.File().Title("LOL").Filter("Image", "png").Filter("Audio", "mp3").Filter("All files", "*").Load()
				} else {
					f, err = dialog.File().Title("Hilarious").Save()
				}
				fmt.Println(f)
				fmt.Println("Error:", err)
			}
		case wde.KeyTypedEvent:
			switch {
			case e.Glyph == "a":
				fmt.Println(dialog.Message("Is this sentence false?").YesNo())
			case e.Glyph == "b":
				fmt.Println(dialog.Message("R U OK?").Title("Just checking").YesNo())
			case e.Glyph == "c":
				dialog.Message("Operation failed").Error()
			}
		case wde.CloseEvent:
			wde.Stop()
			return
		}
	}
}

func main() {
	go func() {
		w, _ := wde.NewWindow(300, 300)
		loadR = image.Rect(0, 0, 300, 150)
		saveR = image.Rect(0, 150, 300, 300)
		w.Show()
		draw.Draw(w.Screen(), loadR, &image.Uniform{color.RGBA{0,0xff,0,0xff}}, image.ZP, draw.Src)
		draw.Draw(w.Screen(), saveR, &image.Uniform{color.RGBA{0xff,0,0,0xff}}, image.ZP, draw.Src)
		w.FlushImage()
		go events(w.EventChan())
	}()
	wde.Run()
}
