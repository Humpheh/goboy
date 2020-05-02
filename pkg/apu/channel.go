package apu

import "math"

// NewChannel returns a new sound channel using a sampling function.
func NewChannel() *Channel {
	return &Channel{}
}

// Channel represents one of four Gameboy sound channels.
type Channel struct {
	frequency float64
	generator WaveGenerator
	time      float64
	amplitude float64

	// Duration in samples
	duration int
	length   int

	envelopeVolume     int
	envelopeTime       int
	envelopeSteps      int
	envelopeStepsInit  int
	envelopeSamples    int
	envelopeIncreasing bool

	sweepTime     float64
	sweepStepLen  byte
	sweepSteps    byte
	sweepStep     byte
	sweepIncrease bool

	onL bool
	onR bool
	// Debug flag to turn off sound output
	debugOff bool
}

// Sample returns a single sample for streaming the sound output. Each sample
// will increase the internal timer based on the global sample rate.
func (chn *Channel) Sample() (outputL, outputR uint16) {
	var output uint16
	step := chn.frequency * twoPi / float64(sampleRate)
	chn.time += step
	if chn.shouldPlay() {
		// Take the sample value from the generator
		if !chn.debugOff {
			output = uint16(float64(chn.generator(chn.time)) * chn.amplitude)
		}
		if chn.duration > 0 {
			chn.duration--
		}
	}
	chn.updateEnvelope()
	chn.updateSweep()
	if chn.onL {
		outputL = output
	}
	if chn.onR {
		outputR = output
	}
	return
}

// Reset the channel to some default variables for the sweep, amplitude,
// envelope and duration.
func (chn *Channel) Reset(duration int) {
	chn.amplitude = 1
	chn.envelopeTime = 0
	chn.sweepTime = 0
	chn.sweepStep = 0
	chn.duration = duration
}

// Returns if the channel should be playing or not.
func (chn *Channel) shouldPlay() bool {
	return (chn.duration == -1 || chn.duration > 0) &&
		chn.generator != nil && chn.envelopeStepsInit > 0
}

// Update the state of the channels envelope.
func (chn *Channel) updateEnvelope() {
	if chn.envelopeSamples > 0 {
		chn.envelopeTime += 1
		if chn.envelopeSteps > 0 && chn.envelopeTime >= chn.envelopeSamples {
			chn.envelopeTime -= chn.envelopeSamples
			chn.envelopeSteps--
			if chn.envelopeSteps == 0 {
				chn.amplitude = 0
			} else if chn.envelopeIncreasing {
				chn.amplitude = 1 - float64(chn.envelopeSteps)/float64(chn.envelopeStepsInit)
			} else {
				chn.amplitude = float64(chn.envelopeSteps) / float64(chn.envelopeStepsInit)
			}
		}
	}
}

var sweepTimes = map[byte]float64{
	1: 7.8 / 1000,
	2: 15.6 / 1000,
	3: 23.4 / 1000,
	4: 31.3 / 1000,
	5: 39.1 / 1000,
	6: 46.9 / 1000,
	7: 54.7 / 1000,
}

// Update the state of the channels sweep.
func (chn *Channel) updateSweep() {
	if chn.sweepStep < chn.sweepSteps {
		t := sweepTimes[chn.sweepStepLen]
		chn.sweepTime += perSample
		if chn.sweepTime > t {
			chn.sweepTime -= t
			chn.sweepStep += 1

			if chn.sweepIncrease {
				chn.frequency += chn.frequency / math.Pow(2, float64(chn.sweepStep))
			} else {
				chn.frequency -= chn.frequency / math.Pow(2, float64(chn.sweepStep))
			}
		}
	}
}
