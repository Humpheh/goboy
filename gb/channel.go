package gb

import (
	"math"
	"math/rand"
	"sync"

	"github.com/faiface/beep"
)

const volume = 0.1
const twoPi = 2 * math.Pi

// Create a new sound channel using a sampling function.
func NewChannel(gen func(float64, float64) float64, start float64) *Channel {
	return &Channel{
		on:     false,
		Func:   gen,
		t:      start,
		buffer: [][2]float64{{0, 0}},
		Volume: 1,
	}
}

// A sound channel
type Channel struct {
	Freq      float64
	Amp       float64
	Func      func(t float64, mod float64) float64
	FuncMod   float64
	DebugMute bool
	Volume    float64
	on        bool
	so1vol    float64
	so2vol    float64
	t         float64

	buffer     [][2]float64
	counter    int
	bufferlock sync.Mutex
}

// Get a StreamerFunc for streaming the sound output. Uses the
// buffer in the sound channel which can be extended using the
// Buffer function.
func (chn *Channel) Stream(sr float64) beep.StreamerFunc {
	chn.counter = 0
	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		chn.bufferlock.Lock()
		buflen := len(chn.buffer) - 1
		for i := range samples {
			index := chn.counter
			if chn.counter < buflen {
				chn.counter++
			} else {
				chn.counter -= 684
				if chn.counter < 0 {
					chn.counter = 0
				}
			}
			samples[i] = chn.buffer[index]
		}
		chn.bufferlock.Unlock()
		return len(samples), true
	})
}

// Buffer a number of samples for streaming.
func (chn *Channel) Buffer(samples int) {
	chn.bufferlock.Lock()
	// Remove end of buffer if its getting long
	if len(chn.buffer) > samples*120 {
		chn.buffer = chn.buffer[samples*60:]
		chn.counter -= samples * 60
	}

	step := chn.Freq * twoPi / float64(41040)

	for i := 0; i < samples; i++ {
		chn.t += step

		var val float64 = 0
		if chn.on && !chn.DebugMute {
			val = chn.Func(chn.t, chn.FuncMod) * chn.Amp * volume * chn.Volume
		}

		chn.buffer = append(chn.buffer, [2]float64{val * chn.so1vol, val * chn.so2vol})
	}
	chn.bufferlock.Unlock()
}

// Set the amplitude of the sound channel.
func (chn *Channel) SetAmp(amp float64) {
	chn.Amp = amp
}

// Enable the sound channel.
func (chn *Channel) On() {
	chn.on = true
}

// Disable the sound channel.
func (chn *Channel) Off() {
	chn.on = false
}

// Set the volume of the sound channel for the two output terminals.
func (chn *Channel) SetVolume(so1vol float64, so2vol float64) {
	chn.so1vol = so1vol
	chn.so2vol = so2vol
}

func Square(t float64, mod float64) float64 {
	if math.Sin(t) <= mod {
		return -1
	}
	return 1
}

func Noise(t float64, _ float64) float64 {
	return float64(rand.Intn(3) - 1)
}

func MakeWaveform(data *[32]int8) func(float64, float64) float64 {
	return func(t float64, _ float64) float64 {
		idx := int(math.Floor(t/twoPi*32)) % 32
		data := int16(int8(data[idx]<<4) >> 4)
		return float64(data) / 7
	}
}
