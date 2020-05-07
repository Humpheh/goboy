package io

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Humpheh/goboy/pkg/gb"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type Frame struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
}

type WebServer struct {
	sync.RWMutex
	frame bytes.Buffer
}

func NewWebServer() *WebServer {

	server := &WebServer{}
	http.HandleFunc("/", serveFile("web/index.html"))
	http.HandleFunc("/ws", server.serveWebSocket)
	http.HandleFunc("/index.js", serveFile("web/index.js"))
	http.HandleFunc("/jquery.js", serveFile("web/jquery.js"))

	go http.ListenAndServe(":8080", nil)

	return server
}

func serveFile(filename string) func(http.ResponseWriter, *http.Request) {
	bytes, err := ioutil.ReadFile(filename)

	return func(w http.ResponseWriter, r *http.Request) {
		if err != nil {
			log.Printf("serve file error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))
		}

		w.Write(bytes)
	}
}

func (server *WebServer) Render(rgbMatrix *[gb.ScreenWidth][gb.ScreenHeight][3]uint8) {
	pngFrame := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{gb.ScreenWidth, gb.ScreenHeight}})

	for x := 0; x < gb.ScreenWidth; x++ {
		for y := 0; y < gb.ScreenHeight; y++ {
			pixel := rgbMatrix[x][y]
			pngFrame.Set(x, y, color.RGBA{R: pixel[0], G: pixel[1], B: pixel[2], A: 0xff})
		}
	}

	var frame bytes.Buffer

	if err := png.Encode(&frame, pngFrame); err != nil {
		log.Printf("error encoding png frame: %s", err.Error())
		return
	}

	server.Lock()
	server.frame = frame
	server.Unlock()
}

func (server *WebServer) ButtonInput() gb.ButtonInput {
	return gb.ButtonInput{}
}

func (server *WebServer) SetTitle(title string) {
	// TODO
	_ = title
}

func (server *WebServer) IsRunning() bool {
	return true
}

func (server *WebServer) serveWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %s", err)
		return
	}

	ticker := time.NewTicker(time.Second / 60)
	for range ticker.C {

		server.RLock()
		frame := server.frame
		server.RUnlock()

		frameMessage := Frame{
			Type: "frame",
			Data: frame.Bytes(),
		}

		err := ws.WriteJSON(frameMessage)

		if err != nil {
			log.Printf("error sending ws json: %s", err.Error())
		}
	}
}
