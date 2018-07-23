package apu

import (
	"math"
	"sync"

	"time"

	"log"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

const sampleRate = 41040
const clock = 4194304
const twoPi = 2 * math.Pi
const cycleTime = float64(41040) / float64(4194304)

type APU struct {
	memory [52]byte

	sampleTime   float64
	currentCycle int64
	lastCycle    int64

	buffer [][2]float64

	chn1 *Channel

	// TODO: waveform RAM
	WaveformRam []int8
}

// Init the sound emulation for a gameboy.
func (a *APU) Init() {
	sampleRate := beep.SampleRate(41040)
	speaker.Init(sampleRate, sampleRate.N(time.Second/30))

	a.WaveformRam = make([]int8, 0x10*2)

	// Create the channels with their sounds
	a.chn1 = NewChannel()
	mix := beep.Mix(
		a.chn1.Stream(float64(sampleRate)),
	)
	speaker.Play(mix)
}

//var soundMask = []byte{
//	/* 0xFF10 */ 0xFF, 0xC0, 0xFF, 0x00, 0x40,
//	/* 0xFF15 */ 0x00, 0xC0, 0xFF, 0x00, 0x40,
//	/* 0xFF1A */ 0x80, 0x00, 0x60, 0x00, 0x40,
//	/* 0xFF20 */ 0x00, 0x3F, 0xFF, 0xFF, 0x40,
//	/* 0xFF24 */ 0xFF, 0xFF, 0x80,
//}

func (a *APU) Write(address uint16, value byte) {
	a.memory[address-0xFF00] = value // & soundMask[address-0xFF10]

	switch address {
	case 0xFF14:
		if value&0x80 == 0x80 {
			a.start1()
		}
	}
	// TODO: if writing to FF26 bit 7 destroy all contents (also cannot access)
}

func (a *APU) Update(cycles int) {
	//lVolume := a.memory[0x24] & 0x70
	//rVolume := a.memory[0x24] & 0x7
	//
	//// Right output for each channel
	output1r := a.memory[0x25]&0x1 == 0x1
	//output2r := a.memory[0x25]&0x2 == 0x2
	//output3r := a.memory[0x25]&0x4 == 0x4
	//output4r := a.memory[0x25]&0x8 == 0x8
	//
	//// Left output for each channel
	output1l := a.memory[0x25]&0x10 == 0x10
	//output2l := a.memory[0x25]&0x20 == 0x20
	//output3l := a.memory[0x25]&0x40 == 0x40
	//output4l := a.memory[0x25]&0x80 == 0x80

	a.currentCycle += int64(cycles)
	diff := a.currentCycle - a.lastCycle
	samples := math.Floor(float64(diff) / clock * sampleRate)
	if samples > 0 {
		a.chn1.Buffer(int(samples), !output1r && !output1l)
		a.lastCycle = a.currentCycle
	}
}

func (a *APU) start1() {
	selection := (a.memory[0x14] & 0x40) >> 6 // 1 = stop when length in NR11 expires
	frequencyValue := uint16(a.memory[0x14]&0x7)<<8 | uint16(a.memory[0x13])

	pattern := (a.memory[0x11] & 0xC0) >> 6
	length := a.memory[0x11] & 0x3F

	// Envelope settings
	envVolume := (a.memory[0x12] & 0xF0) >> 4
	envDirection := (a.memory[0x12] & 0x8) >> 3 // 1 or 0
	envSweep := a.memory[0x12] & 0x7

	// Sweep
	sweepTime := (a.memory[0x10] & 0x70) >> 4
	sweepDirection := a.memory[0x10] >> 3 // 1 = decrease
	sweepNumber := a.memory[0x10] & 0x7

	duration := -1
	if selection == 1 {
		duration = int(float64(length)*(1/64)) * sampleRate
	}

	a.chn1.frequency = 131062 / (2048 - float64(frequencyValue))
	a.chn1.generator = Square(squareLimits[pattern])
	a.chn1.duration = duration
	a.chn1.amplitude = 1

	a.chn1.envelopeTime = 0
	a.chn1.envelopeSteps = int(envVolume)
	a.chn1.envelopeStepsInit = int(envVolume)
	a.chn1.envelopeStepLength = float64(envSweep) / 64
	a.chn1.envelopeIncreasing = envDirection == 1

	a.chn1.sweepTime = 0
	a.chn1.sweepStepLen = sweepTime
	a.chn1.sweepSteps = sweepNumber
	a.chn1.sweepStep = 0
	a.chn1.sweepIncrease = sweepDirection == 0

	log.Print(duration)
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
	return func(t float64) float64 {
		if math.Sin(t) <= mod {
			return -1
		}
		return 1
	}
}

// WaveGenerator is a function which can be used for generating waveform
// samples for different channels.
type WaveGenerator func(t float64) float64

// NewChannel returns a new sound channel using a sampling function.
func NewChannel() *Channel {
	return &Channel{
		buffer: [][2]float64{{0, 0}},
	}
}

// Channel represents one of four gameboy sound channels.
type Channel struct {
	frequency float64
	generator WaveGenerator
	time      float64
	amplitude float64

	// Duration in samples
	duration int

	buffer     [][2]float64
	counter    int
	bufferLock sync.Mutex

	envelopeTime       float64
	envelopeSteps      int
	envelopeStepsInit  int
	envelopeStepLength float64
	envelopeIncreasing bool

	sweepTime     float64
	sweepStepLen  byte
	sweepSteps    byte
	sweepStep     byte
	sweepIncrease bool
}

// Stream returns a StreamerFunc for streaming the sound output. Uses
// the buffer in the sound channel which can be extended using the
// Buffer function.
func (chn *Channel) Stream(sr float64) beep.StreamerFunc {
	chn.counter = 0
	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		chn.bufferLock.Lock()
		bufferLen := len(chn.buffer) - 1
		for i := range samples {
			index := chn.counter
			if chn.counter < bufferLen {
				chn.counter++
			} else {
				chn.counter -= 684
				if chn.counter < 0 {
					chn.counter = 0
				}
			}
			if index >= 0 && index < len(chn.buffer) {
				val := chn.buffer[index]
				samples[i] = val
			}
		}
		chn.bufferLock.Unlock()
		return len(samples), true
	})
}

// Buffer a number of samples for streaming.
func (chn *Channel) Buffer(samples int, silent bool) {
	chn.bufferLock.Lock()
	// Remove end of buffer if its getting long
	if len(chn.buffer) > samples*120 {
		chn.buffer = chn.buffer[samples*60:]
		chn.counter -= samples * 60
	}

	step := chn.frequency * twoPi / float64(sampleRate)
	perSample := 1 / float64(sampleRate)
	for i := 0; i < samples; i++ {
		chn.time += step
		var val float64
		if (chn.duration == -1 || chn.duration > 0) && chn.generator != nil && !silent {
			val = chn.generator(chn.time) * chn.amplitude
			if chn.duration > 0 {
				chn.duration--
			}
		}
		chn.buffer = append(chn.buffer, [2]float64{val, val})

		// Update the envelope
		if chn.envelopeStepLength > 0 {
			chn.envelopeTime += perSample
			if chn.envelopeSteps > 0 && chn.envelopeTime > chn.envelopeStepLength {
				log.Print(chn.envelopeSteps, chn.envelopeStepLength)
				chn.envelopeTime -= chn.envelopeStepLength
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
	chn.bufferLock.Unlock()
}
