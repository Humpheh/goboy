package io

import (
	"net/http"

	"github.com/Humpheh/goboy/pkg/gb"
)

type WebServer struct{}

func NewWebServer() *WebServer {

	server := &WebServer{}
	http.HandleFunc("/", serveFrontend)

	go http.ListenAndServe(":8080", nil)

	return server
}

func serveFrontend(w http.ResponseWriter, r *http.Request) {
	body := []byte(`<html><head></head><body><p style="text-align:center;"><img src='https://http.cat/500' width='900' height='810' style='margin-left: auto; margin-right: auto;' /></p></body></html>`)
	w.Write(body)
}

func (server *WebServer) Render(frame *[160][144][3]uint8) {
	_ = frame
}

func (server *WebServer) ButtonInput() gb.ButtonInput {
	return gb.ButtonInput{}
}

func (server *WebServer) SetTitle(title string) {
	_ = title
}

func (server *WebServer) IsRunning() bool {
	return true
}
