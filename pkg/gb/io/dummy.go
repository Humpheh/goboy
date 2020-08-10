package io

import "github.com/Humpheh/goboy/pkg/gb"

type Dummy struct{}

func (dummy Dummy) Render(frame *[160][144][3]uint8) {
	_ = frame
}

func (dummy Dummy) ButtonInput() gb.ButtonInput {
	return gb.ButtonInput{}
}

func (dummy Dummy) SetTitle(title string) {
	_ = title
}

func (dummy Dummy) IsRunning() bool {
	return true
}
