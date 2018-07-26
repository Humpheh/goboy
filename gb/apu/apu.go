package apu

import (
	"math"

	"time"

	"log"

	"math/rand"

	"github.com/hajimehoshi/oto"
)

const sampleRate = 44100
const clock = 4194304
const twoPi = 2 * math.Pi
const cycleTime = float64(41040) / float64(4194304)
const perSample = 1 / float64(sampleRate)

type APU struct {
	memory [52]byte

	chn1, chn2, chn3, chn4 *Channel

	lVol, rVol float64

	// TODO: waveform RAM
	waveformRam []byte
}

// Init the sound emulation for a gameboy.
func (a *APU) Init() {
	a.waveformRam = make([]byte, 0x20)

	// Create the channels with their sounds
	a.chn1 = NewChannel()
	a.chn2 = NewChannel()
	a.chn3 = NewChannel()
	a.chn4 = NewChannel()

	player, err := oto.NewPlayer(sampleRate, 1, 1, sampleRate/30)
	if err != nil {
		log.Fatalf("Failed to start audio: %v", err)
	}
	go func() {
		for range time.Tick(time.Second / 60) {
			buffer := make([]byte, sampleRate/60)
			vol := (a.lVol + a.rVol) / 10
			for i := range buffer {
				val := (a.chn1.Sample() + a.chn2.Sample() + a.chn3.Sample() + a.chn4.Sample()) / 4
				buffer[i] = byte(float64(val) * vol)
			}

			player.Write(buffer)
		}
	}()
}

//var soundMask = []byte{
//	/* 0xFF10 */ 0xFF, 0xC0, 0xFF, 0x00, 0x40,
//	/* 0xFF15 */ 0x00, 0xC0, 0xFF, 0x00, 0x40,
//	/* 0xFF1A */ 0x80, 0x00, 0x60, 0x00, 0x40,
//	/* 0xFF20 */ 0x00, 0x3F, 0xFF, 0xFF, 0x40,
//	/* 0xFF24 */ 0xFF, 0xFF, 0x80,
//}

var sound3Volume = map[byte]float64{0: 0, 1: 1, 2: 0.5, 3: 0.25}

func (a *APU) Write(address uint16, value byte) {
	a.memory[address-0xFF00] = value // & soundMask[address-0xFF10]

	switch address {
	// Channel 1
	case 0xFF14:
		if value&0x80 == 0x80 {
			a.start1()
		}
		fallthrough //TODO: this is broken, needs to do below logic
	case 0xFF13:
		frequencyValue := uint16(a.memory[0x14]&0x7)<<8 | uint16(a.memory[0x13])
		a.chn1.frequency = 131072 / (2048 - float64(frequencyValue))
	case 0xFF11:
		pattern := (a.memory[0x11] & 0xC0) >> 6
		a.chn1.generator = Square(squareLimits[pattern])

	// Channel 2
	case 0xFF19:
		if value&0x80 == 0x80 {
			a.start2()
		}
		fallthrough //TODO: this is broken, needs to do below logic
	case 0xFF18:
		frequencyValue := uint16(a.memory[0x19]&0x7)<<8 | uint16(a.memory[0x18])
		a.chn2.frequency = 131072 / (2048 - float64(frequencyValue))
	case 0xFF16:
		pattern := (a.memory[0x16] & 0xC0) >> 6
		a.chn2.generator = Square(squareLimits[pattern])

	// Channel 3
	case 0xFF1A:
		// TODO: simplify
		soundOn := a.memory[0x1A]&0x80 == 0x80
		if soundOn {
			a.chn3.envelopeStepsInit = 1
		} else {
			a.chn3.envelopeStepsInit = 0
		}
	case 0xFF1E:
		if value&0x80 == 0x80 {
			a.start3()
		}
		fallthrough //TODO: this is broken, needs to do below logic
	case 0xFF1F:
		frequencyValue := uint16(a.memory[0x1E]&0x7)<<8 | uint16(a.memory[0x1D])
		a.chn3.frequency = 65536 / (2048 - float64(frequencyValue))
	case 0xFF1C:
		// Output level
		value := (a.memory[0x1C] & 0x60) >> 5
		a.chn3.amplitude = sound3Volume[value]

	// Channel 4
	case 0xFF22:
		shiftClock := float64((value & 0xF0) >> 4)
		// TODO: counter step width
		divRatio := float64(value & 0x7)
		if divRatio == 0 {
			divRatio = 0.5
		}
		a.chn4.frequency = 524288 / divRatio / math.Pow(2, shiftClock+1)
	case 0xFF23:
		if value&0x80 == 0x80 {
			a.chn4.generator = Noise()
			a.start4()
		}

	case 0xFF24:
		// Volume control
		a.lVol = float64((a.memory[0x24]&0x70)>>4) / 7
		a.rVol = float64(a.memory[0x24]&0x7) / 7

	case 0xFF25:
		// Channel control
		// Right output for each channel
		output1r := a.memory[0x25]&0x1 == 0x1
		output2r := a.memory[0x25]&0x2 == 0x2
		output3r := a.memory[0x25]&0x4 == 0x4
		output4r := a.memory[0x25]&0x8 == 0x8

		// Left output for each channel
		output1l := a.memory[0x25]&0x10 == 0x10
		output2l := a.memory[0x25]&0x20 == 0x20
		output3l := a.memory[0x25]&0x40 == 0x40
		output4l := a.memory[0x25]&0x80 == 0x80

		a.chn1.on = output1r || output1l
		a.chn2.on = output2r || output2l
		a.chn3.on = output3r || output3l
		a.chn4.on = output4r || output4l
	}
	// TODO: if writing to FF26 bit 7 destroy all contents (also cannot access)
}

func (a *APU) WriteWaveform(address uint16, value byte) {
	soundIndex := (address - 0xFF30) * 2
	a.waveformRam[soundIndex] = byte((value>>4)&0xF) * 0x11
	a.waveformRam[soundIndex+1] = byte(value&0xF) * 0x11
}

func (a *APU) start1() {
	selection := (a.memory[0x14] & 0x40) >> 6 // 1 = stop when length in NR11 expires
	length := a.memory[0x11] & 0x3F

	// Envelope settings
	envVolume, envDirection, envSweep := a.extractEnvelope(a.memory[0x12])

	// Sweep
	sweepTime := (a.memory[0x10] & 0x70) >> 4
	sweepDirection := a.memory[0x10] >> 3 // 1 = decrease
	sweepNumber := a.memory[0x10] & 0x7

	duration := -1
	if selection == 1 {
		duration = int(float64(length)*(1/64)) * sampleRate
	}

	a.chn1.Reset(duration)
	a.chn1.envelopeSteps = int(envVolume)
	a.chn1.envelopeStepsInit = int(envVolume)
	a.chn1.envelopeSamples = int(envSweep) * sampleRate / 64
	a.chn1.envelopeIncreasing = envDirection == 1

	a.chn1.sweepStepLen = sweepTime
	a.chn1.sweepSteps = sweepNumber
	a.chn1.sweepIncrease = sweepDirection == 0
}

func (a *APU) start2() {
	selection := (a.memory[0x19] & 0x40) >> 6 // 1 = stop when length in NR24 expires
	length := a.memory[0x16] & 0x3F

	// Envelope settings
	envVolume, envDirection, envSweep := a.extractEnvelope(a.memory[0x17])

	duration := -1
	if selection == 1 {
		duration = int(float64(length)*(1/64)) * sampleRate
	}

	a.chn2.Reset(duration)
	a.chn2.envelopeSteps = int(envVolume)
	a.chn2.envelopeStepsInit = int(envVolume)
	a.chn2.envelopeSamples = int(envSweep) * sampleRate / 64
	a.chn2.envelopeIncreasing = envDirection == 1
}

func (a *APU) start3() {
	selection := (a.memory[0x1E] & 0x40) >> 6 // 1 = stop when length in NR31 expires
	length := a.memory[0x1B]

	duration := -1
	if selection == 1 {
		duration = int((256-float64(length))*(1/256)) * sampleRate
	}
	// TODO: remove and use wave ram
	a.chn3.generator = Waveform(a.waveformRam)
	a.chn3.Reset(duration)
}

func (a *APU) start4() {
	selection := (a.memory[0x23] & 0x40) >> 6 // 1 = stop when length in NR44 expires
	length := a.memory[0x20] & 0x3F

	// Envelope settings
	envVolume, envDirection, envSweep := a.extractEnvelope(a.memory[0x21])

	duration := -1
	if selection == 1 {
		duration = int(float64(61-length)*(1/256)) * sampleRate
	}

	a.chn4.Reset(duration)
	a.chn4.envelopeSteps = int(envVolume)
	a.chn4.envelopeStepsInit = int(envVolume)
	a.chn4.envelopeSamples = int(envSweep) * sampleRate / 64
	a.chn4.envelopeIncreasing = envDirection == 1
}

func (a *APU) extractEnvelope(val byte) (volume, direction, sweep byte) {
	volume = (val & 0xF0) >> 4
	direction = (val & 0x8) >> 3 // 1 or 0
	sweep = val & 0x7
	return
}

var squareLimits = map[byte]float64{
	0: -0.25, // 12.5% ( _-------_-------_------- )
	1: -0.5,  // 25%   ( __------__------__------ )
	2: 0,     // 50%   ( ____----____----____---- ) (normal)
	3: 0.5,   // 75%   ( ______--______--______-- )
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

func Square(mod float64) WaveGenerator {
	return func(t float64) byte {
		if math.Sin(t) <= mod {
			return 255
		}
		return 0
	}
}

func Waveform(ram []byte) WaveGenerator {
	return func(t float64) byte {
		idx := int(math.Floor(t/twoPi*32)) % 0x20
		return ram[idx]
	}
}

func Noise() WaveGenerator {
	var last byte = 0
	low := true
	return func(t float64) byte {
		sin := math.Sin(t)
		if (sin < 0 && low) || (sin > 0 && !low) {
			last = byte(rand.Intn(2)) * 0xFF
			low = !low
		}
		return last
	}
}

// WaveGenerator is a function which can be used for generating waveform
// samples for different channels.
type WaveGenerator func(t float64) byte

// NewChannel returns a new sound channel using a sampling function.
func NewChannel() *Channel {
	return &Channel{}
}

// Channel represents one of four gameboy sound channels.
type Channel struct {
	frequency float64
	generator WaveGenerator
	time      float64
	amplitude float64

	// Duration in samples
	duration int

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

	on bool
}

// Stream returns a StreamerFunc for streaming the sound output. Uses
// the buffer in the sound channel which can be extended using the
// Buffer function.
func (chn *Channel) Sample() (output uint16) {
	step := chn.frequency * twoPi / float64(sampleRate)
	chn.time += step
	if chn.shouldPlay() && chn.on {
		// Take the sample value from the generator
		output = uint16(float64(chn.generator(chn.time)) * chn.amplitude)
		if chn.duration > 0 {
			chn.duration--
		}
	}
	chn.updateEnvelope()
	chn.updateSweep()
	return output
}

func (chn *Channel) Reset(duration int) {
	chn.amplitude = 1
	chn.envelopeTime = 0
	chn.sweepTime = 0
	chn.sweepStep = 0
	chn.duration = duration
}

func (chn *Channel) shouldPlay() bool {
	return (chn.duration == -1 || chn.duration > 0) &&
		chn.generator != nil && chn.envelopeStepsInit > 0
}

func (chn *Channel) updateEnvelope() {
	if chn.envelopeSamples > 0 {
		chn.envelopeTime += 1
		if chn.envelopeSteps > 0 && chn.envelopeTime >= chn.envelopeSamples {
			//log.Print(chn.envelopeSteps, chn.envelopeSamples)
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
