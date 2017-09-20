package main

import (
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep"
	"time"
	"github.com/humpheh/goboy/gb"
)

func main() {
	sample_rate := beep.SampleRate(48100)
	speaker.Init(sample_rate, sample_rate.N(time.Second/30))

	channel := gb.GetChannel(gb.Square)

	channel.Freq = 64
	channel.SetAmp(1)
	channel.On()
	channel.SetVolume(1, 1)

	speaker.Play(channel.Stream(float64(sample_rate)))
	select {}
}
