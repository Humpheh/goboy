package io

import (
	"bytes"
	"encoding/json"
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

type Input struct {
	Type    string `json:"type"`
	Key     string `json:"key"`
	Pressed bool   `json:"pressed"`
}

type WebServer struct {
	sync.RWMutex
	frame bytes.Buffer
	input gb.ButtonInput
}

func NewWebServer() *WebServer {
	server := &WebServer{}

	log.Printf("Starting webserver, go to http://localhost:8080/ to play!")

	go func() {
		if err := http.ListenAndServe(":8080", server); err != nil {
			log.Fatalf("Webserver crashed: %s", err.Error())
		}
	}()
	return server
}

func (server *WebServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	handlers := map[string]func(http.ResponseWriter, *http.Request){
		"/":          serveFile("web/index.html"),
		"/ws":        server.serveWebSocket,
		"/index.js":  serveFile("web/index.js"),
		"/jquery.js": serveFile("web/jquery.js"),
	}

	log.Printf("%6s %s", r.Method, r.URL)

	if handler, ok := handlers[r.URL.String()]; ok {
		handler(w, r)
		return
	}

	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("Not found"))
}

func serveFile(filename string) func(http.ResponseWriter, *http.Request) {
	bytes, err := ioutil.ReadFile(filename)

	return func(w http.ResponseWriter, r *http.Request) {
		if err != nil {
			log.Printf("serve file error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Internal server error"))
		}

		_, _ = w.Write(bytes)
	}
}

func (server *WebServer) Render(rgbMatrix *[gb.ScreenWidth][gb.ScreenHeight][3]uint8) {

	copy := *rgbMatrix

	go func() {
		pngFrame := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{gb.ScreenWidth, gb.ScreenHeight}})

		for x := 0; x < gb.ScreenWidth; x++ {
			for y := 0; y < gb.ScreenHeight; y++ {
				pixel := copy[x][y]
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
	}()
}

func (server *WebServer) ButtonInput() gb.ButtonInput {
	var input gb.ButtonInput

	server.RLock()
	// create a deep copy
	// TODO simplify this
	input.Pressed = append(input.Pressed, server.input.Pressed...)
	input.Released = append(input.Released, server.input.Released...)

	// input will be sent to emulator, so reset our input state
	server.input = gb.ButtonInput{}
	server.RUnlock()

	return input
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

	go server.readWebSocket(ws)

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
			switch err.(type) {
			case *websocket.CloseError:
				break
			default:
				log.Printf("error writing to ws: (type %T) %s", err, err.Error())
			}
			return
		}
	}
}

var webKeyMap = map[string]gb.Button{
	"z":          gb.ButtonA,
	"x":          gb.ButtonB,
	"Backspace":  gb.ButtonSelect,
	"Enter":      gb.ButtonStart,
	"ArrowRight": gb.ButtonRight,
	"ArrowLeft":  gb.ButtonLeft,
	"ArrowUp":    gb.ButtonUp,
	"ArrowDown":  gb.ButtonDown,
}

func (server *WebServer) readWebSocket(ws *websocket.Conn) {
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Printf("error reading from ws: (type %T) %s", err, err.Error())
			return
		}

		var inputMessage Input
		if err := json.Unmarshal(message, &inputMessage); err != nil {
			log.Printf("failed to parse json: %s", err.Error())
			continue
		}

		if inputMessage.Type != "input" {
			log.Printf("unhandled message type: %s", inputMessage.Type)
			continue
		}

		key, ok := webKeyMap[inputMessage.Key]
		if !ok {
			continue
		}

		server.Lock()
		if inputMessage.Pressed {
			server.input.Pressed = append(server.input.Pressed, key)
		} else {
			server.input.Released = append(server.input.Released, key)
		}
		server.Unlock()
	}
}
