package gb

import (
	"github.com/faiface/beep"
	"math"
	"math/rand"
)

const volume = 0.1
const two_pi = 2 * math.Pi

func GetChannel(gen func(float64, float64) float64, start float64) *Channel {
	return &Channel{
		on: false,
		Func: gen,
		t: start,
		buffer: [][2]float64{{0, 0}},
		Volume: 1,
	}
}

type Channel struct {
	Freq float64
	Amp  float64
	Func func(t float64, mod float64) float64
	FuncMod float64
	DebugMute bool
	Volume float64
	on   bool
	so1vol float64
	so2vol float64
	t float64

	buffer [][2]float64
}

func (chn *Channel) Stream(sr float64) beep.StreamerFunc {
	counter := 0
	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		buflen := len(chn.buffer) - 1
		for i := range samples {
			index := counter
			if counter < buflen {
				counter++
			} else {
				counter -= 684
				if counter < 0 {
					counter = 0
				}
			}
			samples[i] = chn.buffer[index]
		}
		return len(samples), true
	})
}

func (chn *Channel) Buffer(samples int) {
	step := chn.Freq * two_pi / float64(41040)

	for i := 0; i < samples; i++ {
		chn.t += step

		var val float64 = 0
		if chn.on && !chn.DebugMute {
			val = chn.Func(chn.t, chn.FuncMod) * chn.Amp * volume * chn.Volume
		}

		chn.buffer = append(chn.buffer, [2]float64{val * chn.so1vol, val * chn.so2vol})
	}
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
	return float64(rand.Intn(3) - 1)
}

func MakeWaveform(data *[32]int8) func(float64, float64) float64 {
	return func(t float64, _ float64) float64 {
		idx := int(math.Floor(t / two_pi * 32)) % 32
		data := int16(int8(data[idx] << 4) >> 4)
		return float64(data) / 7
	}
}