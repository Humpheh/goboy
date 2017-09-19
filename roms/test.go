package main

import (
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep"
	//"math"
	//"log"
	"time"
	"math"
	"log"
)

func main() {
	sample_rate := beep.SampleRate(48000)
	speaker.Init(sample_rate, sample_rate.N(time.Second/30))

	//noise := beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
	//	for i := range samples {
	//		samples[i][0] = rand.Float64()*2 - 1
	//		samples[i][1] = rand.Float64()*2 - 1
	//	}
	//	return len(samples), true
	//})

	const twopi = 2 * math.Pi

	freq := float64(440)

	//waves_per_second := 44100 / freq


	t := float64(0)
	wave := beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		step := freq * twopi / float64(sample_rate)

		for i := range samples {
			t += step
			val := float64(1)
			if math.Sin(t) <= 0 {
				val = -1
			}

			samples[i][0] = val
			samples[i][1] = val
		}
		log.Print(len(samples))

		return len(samples), true
	})


	speaker.Play(wave)
	select {}
}
