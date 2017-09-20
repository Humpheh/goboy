package gb

import (
	"github.com/faiface/beep"
	"math"
	"math/rand"
)

const volume = 0.1
const two_pi = 2 * math.Pi

func GetChannel(gen func(float64, float64) float64) *Channel {
	return &Channel{
		Amp:  0,
		Func: gen,
	}
}

type Channel struct {
	Freq float64
	Amp  float64
	Func func(t float64, mod float64) float64
	FuncMod float64
	DebugMute bool
	on   bool
	so1vol float64
	so2vol float64
}

func (chn *Channel) Stream(sr float64) beep.StreamerFunc {
	var t float64 = 0
	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		step := chn.Freq * two_pi / float64(sr)

		for i := range samples {
			t += step

			var val float64 = 0
			if chn.on && !chn.DebugMute {
				val = chn.Func(t, chn.FuncMod) * chn.Amp * volume
			}

			samples[i][0] = val * chn.so1vol
			samples[i][1] = val * chn.so2vol
		}
		return len(samples), true
	})
}

func (chn *Channel) SetAmp(amp float64) {
	chn.Amp = amp
}

func (chn *Channel) On() {
	chn.on = true
}

func (chn *Channel) Off() {
	chn.on = false
}

func (chn *Channel) SetVolume(so1vol float64, so2vol float64) {
	chn.so1vol = so1vol
	chn.so2vol = so2vol
}

func Square(t float64, mod float64) float64 {
	val := float64(1)
	if math.Sin(t) <= mod {
		val = -1
	}
	return val
}

func Noise(t float64, _ float64) float64 {
	return rand.Float64()*2 - 1
}
