package apu

import (
	"math"
	"math/rand"
)

// WaveGenerator is a function which can be used for generating waveform
// samples for different channels.
type WaveGenerator func(t float64) byte

// Square returns a square wave generator with a given mod. This is used
// for channels 1 and 2.
func Square(mod float64) WaveGenerator {
	return func(t float64) byte {
		if math.Sin(t) <= mod {
			return 0xFF
		}
		return 0
	}
}

// Waveform returns a wave generator for some waveform ram. This is used
// by channel 3.
func Waveform(ram func(i int) byte) WaveGenerator {
	return func(t float64) byte {
		idx := int(math.Floor(t/twoPi*32)) % 0x20
		return ram(idx)
	}
}

// Noise returns a wave generator for a noise channel. This is used by
// channel 4.
func Noise() WaveGenerator {
	var last float64
	var val byte
	return func(t float64) byte {
		if t-last > twoPi {
			last = t
			val = byte(rand.Intn(2)) * 0xFF
		}
		return val
	}
}
