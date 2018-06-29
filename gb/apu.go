package gb

import (
	"math"

	"github.com/Humpheh/goboy/bits"
	"github.com/Humpheh/goboy/gb/sound"
)

var squarelimits = map[byte]float64{
	0: -0.25, // 12.5% ( _-------_-------_------- )
	1: -0.5,  // 25%   ( __------__------__------ )
	2: 0,     // 50%   ( ____----____----____---- ) (normal)
	3: 0.5,   // 75%   ( ______--______--______-- )
}

type envelopeSound struct {
	Time       float64
	StepLen    float64
	Steps      byte
	StepsInit  byte
	Increasing bool
}

func (env *envelopeSound) Update(secs float64, channel *sound.Channel) {
	if env.Steps > 0 {
		env.Time += secs
		if env.Time > env.StepLen {
			env.Time -= env.StepLen
			env.Steps--
			amp := float64(env.Steps) / float64(env.StepsInit)
			if env.Steps == 0 {
				amp = 0
			}
			if env.Increasing {
				amp = 1 - amp
			}
			channel.SetAmp(amp)
		}
	}
}

func (env *envelopeSound) Reset() {
	env.Steps = env.StepsInit
	env.Time = 0
}

type sweepSound struct {
	Time     float64
	StepLen  byte
	Steps    byte
	Step     byte
	Increase bool
}

var sweeptime = map[byte]float64{
	1: 7.8 / 1000,
	2: 15.6 / 1000,
	3: 23.4 / 1000,
	4: 31.3 / 1000,
	5: 39.1 / 1000,
	6: 46.9 / 1000,
	7: 54.7 / 1000,
}

func (swp *sweepSound) Update(secs float64, channel *sound.Channel) {
	/*
		FF10 - NR10 - Channel 1 Sweep register (R/W)
		  Bit 6-4 - Sweep Time
		  Bit 3   - Sweep Increase/Decrease
					 0: Addition    (frequency increases)
					 1: Subtraction (frequency decreases)
		  Bit 2-0 - Number of sweep shift (n: 0-7)
		Sweep Time:
		  000: sweep off - no freq change
		  001: 7.8 ms  (1/128Hz)
		  010: 15.6 ms (2/128Hz)
		  011: 23.4 ms (3/128Hz)
		  100: 31.3 ms (4/128Hz)
		  101: 39.1 ms (5/128Hz)
		  110: 46.9 ms (6/128Hz)
		  111: 54.7 ms (7/128Hz)

		The change of frequency (NR13,NR14) at each shift is calculated by the
		following formula where X(0) is initial freq & X(t-1) is last freq:
		  X(t) = X(t-1) +/- X(t-1)/2^n
	*/
	if swp.Step < swp.Steps {
		t := sweeptime[swp.StepLen]
		swp.Time += secs
		if swp.Time > t {
			swp.Time -= t
			swp.Step += 1

			if swp.Increase {
				channel.Freq += channel.Freq / math.Pow(2, float64(swp.Step))
			} else {
				channel.Freq -= channel.Freq / math.Pow(2, float64(swp.Step))
			}
		}
	}
}

// Struct containing sound data
type Sound struct {
	gb *Gameboy

	// Channel 1 variables
	channel1        *sound.Channel
	channel1TimeVal float64
	channel1Time    float64
	channel1Env     *envelopeSound
	channel1Sweep   *sweepSound

	// Channel 2 variables
	channel2        *sound.Channel
	channel2Time    float64
	channel2TimeVal float64
	channel2Env     *envelopeSound
	channel2Sweep   *sweepSound

	// Channel 3 variables
	channel3     *sound.Channel
	channel3Time float64

	// Channel 4 variables
	channel4     *sound.Channel
	channel4Time float64
	channel4Env  *envelopeSound

	waveformRam [32]int8
	Time        float64

	mixer *sound.Mixer
}

// Initialise the sound emulation for a gameboy.
func (s *Sound) Init(gb *Gameboy) {
	//sampleRate := beep.SampleRate(41040)
	//speaker.Init(sampleRate, sampleRate.N(time.Second/30))

	s.Time = 0

	mixer := sound.NewMixer()

	// Create the channels with their sounds
	s.channel1 = mixer.NewChannel(sound.Square)
	s.channel2 = mixer.NewChannel(sound.Square)
	s.channel3 = mixer.NewChannel(sound.MakeWaveform(&s.waveformRam))
	s.channel4 = mixer.NewChannel(sound.Noise)

	s.mixer = mixer
	//if gb.options.sound {
	//	speaker.Play(mix)
	//}
	s.gb = gb
}

var soundMask = []byte{
	0xFF, 0xC0, 0xFF, 0x00, 0x40, /*0xFF15->*/
	0x00, 0xC0, 0xFF, 0x00, 0x40,
	0x80, 0x00, 0x60, 0x00, 0x40, /*<-0xFF1E*/
	0x00, 0x3F, 0xFF, 0xFF, 0x40,
	0xFF, 0xFF, 0x80,
}

func (s *Sound) makeSweep(value byte) *sweepSound {
	sweepTime := (value >> 4) & 0x7
	sweepIncrease := !bits.Test(value, 3)
	sweepShift := value & 0x7

	if sweepTime == 0 {
		return nil
	}
	return &sweepSound{
		StepLen:  sweepTime,
		Steps:    sweepShift,
		Step:     0,
		Increase: sweepIncrease,
	}
}

func (s *Sound) Write(address uint16, value byte) {
	s.gb.Memory.Data[address] = value

	switch address {
	case 0xFF10:
		s.channel1Sweep = s.makeSweep(value)

	case 0xFF11:
		NR14 := s.gb.Memory.Read(0xFF14)
		if bits.Test(NR14, 6) {
			// Counter
			s.channel1Time = (64 - float64(value&0x1F)) * (1 / 256)
		} else {
			// Consecutive
			s.channel1Time = 100000
		}
		s.channel1TimeVal = s.channel1Time
		pattern := (value >> 6) & 0x3
		s.channel1.Mod = squarelimits[pattern]

	case 0xFF12:
		// Envelope
		envVolume := (value >> 4) & 0xF
		envIncrease := bits.Test(value, 3)
		envSweep := value & 0x7

		if envVolume == 0 {
			s.Toggle(1, false)
		}

		if envSweep == 0 {
			s.channel1Env = nil
		} else {
			s.channel1Env = &envelopeSound{
				StepLen:    float64(envSweep) / 64,
				Steps:      envVolume,
				StepsInit:  envVolume,
				Increasing: envIncrease,
			}
		}

	case 0xFF13:
		NR14 := s.gb.Memory.Read(0xFF14)
		s.UpdateChan1Freq(value, NR14)

	case 0xFF14:
		NR13 := s.gb.Memory.Read(0xFF13)
		s.UpdateChan1Freq(NR13, value)
		if bits.Test(value, 7) {
			s.Toggle(1, s.ShouldPlay(1))
			s.channel1.Amp = 1
			s.channel1Time = s.channel1TimeVal
			if s.channel1Env != nil {
				s.channel1Env.Reset()
			}
			if s.channel1Sweep != nil {
				s.channel1Sweep.Step = 0
			}
		}

	case 0xFF16:
		NR24 := s.gb.Memory.Read(0xFF19)
		if bits.Test(NR24, 6) {
			// Counter
			s.channel2Time = (64 - float64(value&0x1F)) * (1 / 256)
		} else {
			// Consecutive
			s.channel2Time = 100000
		}
		s.channel2TimeVal = s.channel2Time
		pattern := (value >> 6) & 0x3
		s.channel2.Mod = squarelimits[pattern]

	case 0xFF17:
		// Envelope
		envVolume := (value >> 4) & 0xF
		envIncrease := bits.Test(value, 3)
		envSweep := value & 0x7

		if envVolume == 0 {
			s.Toggle(2, false)
		}

		if envSweep == 0 {
			s.channel2Env = nil
		} else {
			s.channel2Env = &envelopeSound{
				StepLen:    float64(envSweep) / 64,
				Steps:      envVolume,
				StepsInit:  envVolume,
				Increasing: envIncrease,
			}
		}

	case 0xFF18:
		NR24 := s.gb.Memory.Read(0xFF19)
		s.UpdateChan2Freq(value, NR24)

	case 0xFF19:
		NR23 := s.gb.Memory.Read(0xFF18)
		s.UpdateChan2Freq(NR23, value)
		if bits.Test(value, 7) {
			s.Toggle(2, s.ShouldPlay(2))
			s.channel2.Amp = 1
			s.channel2Time = s.channel2TimeVal
			if s.channel2Env != nil {
				s.channel2Env.Reset()
			}
		}

	case 0xFF1A:
		if bits.Test(value, 7) {
			s.channel3.On()
		} else {
			s.channel3.Off()
		}

	case 0xFF1B:
		NR34 := s.gb.Memory.Read(0xFF1E)
		if bits.Test(NR34, 6) {
			// Counter
			s.channel3Time = (64 - float64(value)) * (1 / 256)
		} else {
			// Consecutive
			s.channel3Time = 100000
		}

	case 0xFF1C:
		s.ToggleCh3Volume(value)

	case 0xFF1D:
		NR34 := s.gb.Memory.Read(0xFF1E)
		s.UpdateChan3Freq(value, NR34)

	case 0xFF1E:
		NR33 := s.gb.Memory.Read(0xFF1D)
		s.UpdateChan3Freq(NR33, value)
		if bits.Test(value, 7) {
			s.Toggle(3, s.ShouldPlay(3))
			s.channel3.Amp = 1
		}

		// TODO: Channel 4
	case 0xFF20:
		NR44 := s.gb.Memory.Read(0xFF23)
		if bits.Test(NR44, 6) {
			// Counter
			s.channel4Time = (64 - float64(value&0x3F)) * (1 / 256)
		} else {
			// Consecutive
			s.channel4Time = 100000
		}

	case 0xFF21:
		// Envelope
		envVolume := (value >> 4) & 0xF
		envIncrease := bits.Test(value, 3)
		envSweep := value & 0x7

		if envSweep == 0 {
			s.channel4Env = nil
		} else {
			s.channel4Env = &envelopeSound{
				StepLen:    float64(envSweep) / 64,
				Steps:      envVolume,
				StepsInit:  envVolume,
				Increasing: envIncrease,
			}
		}

	case 0xFF22:
		freqShiftClock := float64((value >> 4) & 0xF)
		freqDivier := float64(value & 0x7)
		if freqDivier == 0 {
			freqDivier = 0.5
		}
		freq := 524288 / freqDivier / math.Pow(2, freqShiftClock+1)
		// TODO: Bit 3 NR43 modifier
		s.channel4.Freq = freq

	case 0xFF23:
		if bits.Test(value, 7) {
			envVolume := (s.gb.Memory.Read(0xFF21) >> 4) & 0xF
			s.Toggle(4, envVolume != 0 && s.ShouldPlay(4))
		}

	case 0xFF26:
		s.updateOutput(value)

		// END TODO

	case 0xFF24:
		s.SetVolume(value)

	case 0xFF25:
		s.updateOutput(value)
	}

	//s.GB.Memory.Data[address] = value & soundMask[address-0xFF10]
}

// Update the sound emulation by a number of clock cycles.
func (s *Sound) Tick(clocks int) {
	secs := float64(clocks) / ClockSpeed
	s.Time += secs

	if s.channel1Time > 0 {
		s.channel1Time -= secs
		if s.channel1Env != nil {
			s.channel1Env.Update(secs, s.channel1)
		}
		if s.channel1Sweep != nil {
			s.channel1Sweep.Update(secs, s.channel1)
		}
	} else {
		s.Toggle(1, false)
	}

	if s.channel2Time > 0 {
		s.channel2Time -= secs
		if s.channel2Env != nil {
			s.channel2Env.Update(secs, s.channel2)
		}
	} else {
		s.Toggle(2, false)
	}

	if s.channel3Time > 0 {
		s.channel3Time -= secs
	} else {
		s.Toggle(3, false)
	}

	if s.channel4Time > 0 {
		s.channel4Time -= secs
		if s.channel4Env != nil {
			s.channel4Env.Update(secs, s.channel4)
		}
	} else {
		s.Toggle(4, false)
	}

	s.channel1.DebugMute = s.gb.Debug.MuteChannel1
	s.channel2.DebugMute = s.gb.Debug.MuteChannel2
	s.channel3.DebugMute = s.gb.Debug.MuteChannel3
	s.channel4.DebugMute = s.gb.Debug.MuteChannel4

	go s.mixer.Buffer(int(735 * SpeedDivider))
}

// Update the on/off status of the output channels.
// TODO: what is the meaning of value?
func (s *Sound) updateOutput(value byte) {
	if !s.ShouldPlay(1) {
		s.channel1.Off()
	}
	if !s.ShouldPlay(2) {
		s.channel2.Off()
	}
	if !s.ShouldPlay(3) {
		s.channel3.Off()
	}
	if !s.ShouldPlay(4) {
		s.channel4.Off()
	}
}

// Determine if a channel should be playing.
func (s *Sound) ShouldPlay(channel byte) bool {
	FF25 := s.gb.Memory.Data[0xFF25]
	FF26 := s.gb.Memory.Data[0xFF26]

	// Individual sound control
	return bits.Test(FF25, channel-1) && bits.Test(FF25, channel+3) &&
		// All sound on/off
		bits.Test(FF26, 7)
}

// Update the frequency of channel 1
func (s *Sound) UpdateChan1Freq(NR13 byte, NR14 byte) {
	freqVal := uint16(NR13) | (uint16(NR14&0x7) << 8)
	freq := 131072 / (2048 - float64(freqVal))
	s.channel1.Freq = freq
}

// Update the frequency of channel 2
func (s *Sound) UpdateChan2Freq(NR23 byte, NR24 byte) {
	freqVal := uint16(NR23) | (uint16(NR24&0x7) << 8)
	freq := 131072 / (2048 - float64(freqVal))
	s.channel2.Freq = freq
}

// Update the frequency of channel 3
func (s *Sound) UpdateChan3Freq(NR33 byte, NR34 byte) {
	freqVal := uint16(NR33) | (uint16(NR34&0x7) << 8)
	freq := 65536 / (2048 - float64(freqVal))
	s.channel3.Freq = freq
}

// Toggle a channel on or off.
func (s *Sound) Toggle(channel byte, on bool) {
	var c *sound.Channel
	switch channel {
	case 1:
		c = s.channel1
	case 2:
		c = s.channel2
	case 3:
		c = s.channel3
	case 4:
		c = s.channel4
	}
	if on && s.ShouldPlay(channel) {
		c.On()
		s.gb.Memory.Data[0xFF26] = bits.Set(s.gb.Memory.Data[0xFF26], channel-1)
	} else {
		c.Off()
		s.gb.Memory.Data[0xFF26] = bits.Reset(s.gb.Memory.Data[0xFF26], channel-1)
	}
}

// Set the volume of each channel.
func (s *Sound) SetVolume(value byte) {
	so1vol := float64(value&0x7) / 7
	so2vol := float64((value>>4)&0x7) / 7

	s.channel1.SetVolume(so1vol, so2vol)
	s.channel2.SetVolume(so1vol, so2vol)
	s.channel3.SetVolume(so1vol, so2vol)
	s.channel4.SetVolume(so1vol, so2vol)
}

// Toggle channel 3 on and off from a byte by looking at its
// 8th bit.
func (s *Sound) ToggleCh3(value byte) {
	if bits.Test(value, 7) {
		s.channel3.On()
	} else {
		s.channel3.Off()
	}
}

var ch3vols = map[byte]float64{
	0: 0, 1: 1, 2: 0.5, 3: 0.25,
}

// Toggle the volume of channel 3 - TODO?
func (s *Sound) ToggleCh3Volume(value byte) {
	/*
		0: Mute (No sound)
		1: 100% Volume (Produce Wave Pattern RAM Data as it is)
		2:  50% Volume (Produce Wave Pattern RAM data shifted once to the right)
		3:  25% Volume (Produce Wave Pattern RAM data shifted twice to the right)
	*/
	// TODO: What does that mean?
	vol := value >> 5 & 0x3
	s.channel3.Volume = ch3vols[vol]
}
