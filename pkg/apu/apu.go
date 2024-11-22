package apu

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/hajimehoshi/oto"
)

const (
	sampleRate = 44100
	twoPi      = 2 * math.Pi
	perSample  = 1 / float64(sampleRate)

	cpuTicksPerSample    = float64(4194304) / sampleRate
	maxFrameBufferLength = 5000
)

// APU is the GameBoy's audio processing unit. Audio comprises four
// channels, each one controlled by a set of registers.
//
// Channels 1 and 2 are both Square channels, channel 3 is a arbitrary
// waveform channel which can be set in RAM, and channel 4 outputs noise.
type APU struct {
	playing bool

	memory      [52]byte
	waveformRam []byte

	player                 *oto.Player
	chn1, chn2, chn3, chn4 *Channel
	tickCounter            float64
	lVol, rVol             float64

	audioBuffer chan [2]byte
}

// Init the sound emulation for a Gameboy.
func (a *APU) Init(sound bool) {
	a.playing = sound
	a.waveformRam = make([]byte, 0x20)
	a.audioBuffer = make(chan [2]byte, maxFrameBufferLength)

	// Sets waveform ram to:
	// 00 FF 00 FF  00 FF 00 FF  00 FF 00 FF  00 FF 00 FF
	for x := 0x0; x < 0x20; x++ {
		if x&2 == 0 {
			a.waveformRam[x] = 0x00
		} else {
			a.waveformRam[x] = 0xFF
		}
	}

	// Create the channels with their sounds
	a.chn1 = NewChannel()
	a.chn2 = NewChannel()
	a.chn3 = NewChannel()
	a.chn4 = NewChannel()

	const bufferSeconds = 120

	if sound {
		ctx, err := oto.NewContext(sampleRate, 2, 1, sampleRate/bufferSeconds)
		if err != nil {
			log.Fatalf("Failed to start audio: %v", err)
		}
		a.player = ctx.NewPlayer()
		a.playSound(bufferSeconds)
	}
}

// Starts a goroutine which plays the sound
func (a *APU) playSound(bufferSeconds int) {
	frameTime := time.Second / time.Duration(bufferSeconds)
	ticker := time.NewTicker(frameTime)
	targetSamples := sampleRate / bufferSeconds
	go func() {
		var reading [2]byte
		var buffer []byte
		for range ticker.C {
			fbLen := len(a.audioBuffer)
			if fbLen >= targetSamples/2 {
				newBuffer := make([]byte, fbLen*2)
				for i := 0; i < fbLen*2; i += 2 {
					reading = <-a.audioBuffer
					newBuffer[i], newBuffer[i+1] = reading[0], reading[1]
				}
				buffer = newBuffer
			}

			_, err := a.player.Write(buffer)
			if err != nil {
				log.Printf("error sampling: %v", err)
			}
		}
	}()
}

func (a *APU) Buffer(cpuTicks int, speed int) {
	if !a.playing {
		return
	}
	a.tickCounter += float64(cpuTicks) / float64(speed)
	if a.tickCounter < cpuTicksPerSample {
		return
	}
	a.tickCounter -= cpuTicksPerSample

	chn1l, chn1r := a.chn1.Sample()
	chn2l, chn2r := a.chn2.Sample()
	chn3l, chn3r := a.chn3.Sample()
	chn4l, chn4r := a.chn4.Sample()

	valL := (chn1l + chn2l + chn3l + chn4l) / 4
	valR := (chn1r + chn2r + chn3r + chn4r) / 4

	a.audioBuffer <- [2]byte{byte(float64(valL) * a.lVol), byte(float64(valR) * a.rVol)}
}

var soundMask = []byte{
	/* 0xFF10 */ 0xFF, 0xC0, 0xFF, 0x00, 0x40,
	/* 0xFF15 */ 0x00, 0xC0, 0xFF, 0x00, 0x40,
	/* 0xFF1A */ 0x80, 0x00, 0x60, 0x00, 0x40,
	/* 0xFF20 */ 0x00, 0x3F, 0xFF, 0xFF, 0x40,
	/* 0xFF24 */ 0xFF, 0xFF, 0x80,
}

var channel3Volume = map[byte]float64{0: 0, 1: 1, 2: 0.5, 3: 0.25}

var squareLimits = map[byte]float64{
	0: -0.25, // 12.5% ( _-------_-------_------- )
	1: -0.5,  // 25%   ( __------__------__------ )
	2: 0,     // 50%   ( ____----____----____---- ) (normal)
	3: 0.5,   // 75%   ( ______--______--______-- )
}

// Read returns a value from the APU.
func (a *APU) Read(address uint16) byte {
	if address >= 0xFF30 {
		return a.waveformRam[address-0xFF30]
	}
	// TODO: we should modify the sound memory as we're sampling
	return a.memory[address-0xFF00] & soundMask[address-0xFF10]
}

// Write a value to the APU registers.
func (a *APU) Write(address uint16, value byte) {
	a.memory[address-0xFF00] = value

	switch address {
	// Channel 1
	case 0xFF10:
		// -PPP NSSS Sweep period, negate, shift
		a.chn1.sweepStepLen = (a.memory[0x10] & 0b111_0000) >> 4
		a.chn1.sweepSteps = a.memory[0x10] & 0b111
		a.chn1.sweepIncrease = a.memory[0x10]&0b1000 == 0 // 1 = decrease
	case 0xFF11:
		// DDLL LLLL Duty, Length load
		duty := (value & 0b1100_0000) >> 6
		a.chn1.generator = Square(squareLimits[duty])
		a.chn1.length = int(value & 0b0011_1111)
	case 0xFF12:
		// VVVV APPP - Starting volume, Envelop add mode, period
		envVolume, envDirection, envSweep := a.extractEnvelope(value)
		a.chn1.envelopeVolume = int(envVolume)
		a.chn1.envelopeSamples = int(envSweep) * sampleRate / 64
		a.chn1.envelopeIncreasing = envDirection == 1
	case 0xFF13:
		// FFFF FFFF Frequency LSB
		frequencyValue := uint16(a.memory[0x14]&0b111)<<8 | uint16(value)
		a.chn1.frequency = 131072 / (2048 - float64(frequencyValue))
	case 0xFF14:
		// TL-- -FFF Trigger, Length Enable, Frequencu MSB
		frequencyValue := uint16(value&0b111)<<8 | uint16(a.memory[0x13])
		a.chn1.frequency = 131072 / (2048 - float64(frequencyValue))
		if value&0b1000_0000 != 0 {
			if a.chn1.length == 0 {
				a.chn1.length = 64
			}
			duration := -1
			if value&0b100_0000 != 0 { // 1 = use length
				duration = int(float64(a.chn1.length)*(1/64)) * sampleRate
			}
			a.chn1.Reset(duration)
			a.chn1.envelopeSteps = a.chn1.envelopeVolume
			a.chn1.envelopeStepsInit = a.chn1.envelopeVolume
			// TODO: Square 1's sweep does several things (see frequency sweep).
		}

	// Channel 2
	case 0xFF15:
		// ---- ---- Not used
	case 0xFF16:
		// DDLL LLLL Duty, Length load (64-L)
		pattern := (value & 0b1100_0000) >> 6
		a.chn2.generator = Square(squareLimits[pattern])
		a.chn2.length = int(value & 0b11_1111)
	case 0xFF17:
		// VVVV APPP Starting volume, Envelope add mode, period
		envVolume, envDirection, envSweep := a.extractEnvelope(value)
		a.chn2.envelopeVolume = int(envVolume)
		a.chn2.envelopeSamples = int(envSweep) * sampleRate / 64
		a.chn2.envelopeIncreasing = envDirection == 1
	case 0xFF18:
		// FFFF FFFF Frequency LSB
		frequencyValue := uint16(a.memory[0x19]&0b111)<<8 | uint16(value)
		a.chn2.frequency = 131072 / (2048 - float64(frequencyValue))
	case 0xFF19:
		// TL-- -FFF Trigger, Length enable, Frequency MSB
		if value&0b1000_0000 != 0 {
			if a.chn2.length == 0 {
				a.chn2.length = 64
			}
			duration := -1
			if value&0b100_0000 != 0 {
				duration = int(float64(a.chn2.length)*(1/64)) * sampleRate
			}
			a.chn2.Reset(duration)
			a.chn2.envelopeSteps = a.chn2.envelopeVolume
			a.chn2.envelopeStepsInit = a.chn2.envelopeVolume
		}
		frequencyValue := uint16(value&0b111)<<8 | uint16(a.memory[0x18])
		a.chn2.frequency = 131072 / (2048 - float64(frequencyValue))

	// Channel 3
	case 0xFF1A:
		// E--- ---- DAC power
		a.chn3.envelopeStepsInit = int((value & 0b1000_0000) >> 7)
	case 0xFF1B:
		// LLLL LLLL Length load
		a.chn3.length = int(value)
	case 0xFF1C:
		// -VV- ---- Volume code
		selection := (value & 0b110_0000) >> 5
		a.chn3.amplitude = channel3Volume[selection]
	case 0xFF1D:
		// FFFF FFFF Frequency LSB
		frequencyValue := uint16(a.memory[0x1E]&0b111)<<8 | uint16(value)
		a.chn3.frequency = 65536 / (2048 - float64(frequencyValue))
	case 0xFF1E:
		// TL-- -FFF Trigger, Length enable, Frequency MSB
		if value&0b1000_0000 != 0 {
			if a.chn3.length == 0 {
				a.chn3.length = 256
			}
			duration := -1
			if value&0b100_0000 != 0 { // 1 = use length
				duration = int((256-float64(a.chn3.length))*(1/256)) * sampleRate
			}
			a.chn3.generator = Waveform(func(i int) byte { return a.waveformRam[i] })
			a.chn3.duration = duration
		}
		frequencyValue := uint16(value&0b111)<<8 | uint16(a.memory[0x1D])
		a.chn3.frequency = 65536 / (2048 - float64(frequencyValue))

	// Channel 4
	case 0xFF1F:
		// ---- ---- Not used
	case 0xFF20:
		// --LL LLLL Length load
		a.chn4.length = int(value & 0b11_1111)
	case 0xFF21:
		// VVVV APPP Starting volume, Envelope add mode, period
		envVolume, envDirection, envSweep := a.extractEnvelope(value)
		a.chn4.envelopeVolume = int(envVolume)
		a.chn4.envelopeSamples = int(envSweep) * sampleRate / 64
		a.chn4.envelopeIncreasing = envDirection == 1
	case 0xFF22:
		// SSSS WDDD Clock shift, Width mode of LFSR, Divisor code
		shiftClock := float64((value & 0b1111_0000) >> 4)
		// TODO: counter step width
		divRatio := float64(value & 0b111)
		if divRatio == 0 {
			divRatio = 0.5
		}
		a.chn4.frequency = 524288 / divRatio / math.Pow(2, shiftClock+1)
	case 0xFF23:
		// TL-- ---- Trigger, Length enable
		if value&0x80 == 0x80 {
			duration := -1
			if value&0b100_0000 != 0 { // 1 = use length
				duration = int(float64(61-a.chn4.length)*(1/256)) * sampleRate
			}
			a.chn4.generator = Noise()
			a.chn4.Reset(duration)
			a.chn4.envelopeSteps = a.chn4.envelopeVolume
			a.chn4.envelopeStepsInit = a.chn4.envelopeVolume
		}

	case 0xFF24:
		// Volume control
		a.lVol = float64((a.memory[0x24]&0x70)>>4) / 7
		a.rVol = float64(a.memory[0x24]&0x7) / 7

	case 0xFF25:
		// Channel control
		a.chn1.onR = value&0x1 != 0
		a.chn2.onR = value&0x2 != 0
		a.chn3.onR = value&0x4 != 0
		a.chn4.onR = value&0x8 != 0
		a.chn1.onL = value&0x10 != 0
		a.chn2.onL = value&0x20 != 0
		a.chn3.onL = value&0x40 != 0
		a.chn4.onL = value&0x80 != 0
	}
	// TODO: if writing to FF26 bit 7 destroy all contents (also cannot access)
}

// WriteWaveform writes a value to the waveform ram.
func (a *APU) WriteWaveform(address uint16, value byte) {
	soundIndex := (address - 0xFF30) * 2
	a.waveformRam[soundIndex] = (value >> 4) & 0xF * 0x11
	a.waveformRam[soundIndex+1] = value & 0xF * 0x11
}

// ToggleSoundChannel toggles a sound channel for debugging.
func (a *APU) ToggleSoundChannel(channel int) {
	switch channel {
	case 1:
		a.chn1.debugOff = !a.chn1.debugOff
	case 2:
		a.chn2.debugOff = !a.chn2.debugOff
	case 3:
		a.chn3.debugOff = !a.chn3.debugOff
	case 4:
		a.chn4.debugOff = !a.chn4.debugOff
	}
	log.Printf("Toggle Channel %v mute", channel)
}

func (a *APU) LogSoundState() {
	fmt.Println("Channel 3")
	fmt.Printf("  0xFF1A E--- ---- = %08b\n", a.memory[0x1A])
	fmt.Printf("  0xFF1B LLLL LLLL = %08b\n", a.memory[0x1B])
	fmt.Printf("  0xFF1C -VV- ---- = %08b\n", a.memory[0x1C])
	fmt.Printf("  0xFF1D FFFF FFFF = %08b\n", a.memory[0x1D])
	fmt.Printf("  0xFF1E TL-- -FFF = %08b\n", a.memory[0x1E])
}

// Extract some envelope variables from a byte.
func (a *APU) extractEnvelope(val byte) (volume, direction, sweep byte) {
	volume = (val & 0xF0) >> 4
	direction = (val & 0x8) >> 3 // 1 or 0
	sweep = val & 0x7
	return
}
