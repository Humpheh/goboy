package gb

import (
	"github.com/humpheh/goboy/bits"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"math"
	"time"
)

var squarelimits = map[byte]float64{
	0: -0.25, // 12.5% ( _-------_-------_------- )
	1: -0.5,  // 25%   ( __------__------__------ )
	2: 0,     // 50%   ( ____----____----____---- ) (normal)
	3: 0.5,   // 75%   ( ______--______--______-- )
}

type Envelope struct {
	Time       float64
	StepLen    float64
	Steps      byte
	StepsInit  byte
	Increasing bool
}

func (env *Envelope) Update(secs float64, channel *Channel) {
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

type Sweep struct {
	Time float64
	StepLen byte
	Steps byte
	Step byte
	Increase bool
}

var sweeptime = map[byte]float64{
	1: 7.8  / 1000,
	2: 15.6 / 1000,
	3: 23.4 / 1000,
	4: 31.3 / 1000,
	5: 39.1 / 1000,
	6: 46.9 / 1000,
	7: 54.7 / 1000,
}

func (swp *Sweep) Update(secs float64, channel *Channel) {

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

	if swp.Step < swp.Steps && swp.StepLen != 0 {
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

type Sound struct {
	GB           *Gameboy
	Channel1     *Channel
	Channel1TimeVal float64
	Channel1Time float64
	Channel1Env  *Envelope
	Channel1Sweep *Sweep
	Channel2     *Channel
	Channel2Time float64
	Channel2TimeVal float64
	Channel2Env  *Envelope
	Channel2Sweep *Sweep
	Channel3     *Channel
	Channel3Time float64
	Channel4     *Channel
	Channel4Time float64
	Channel4Env  *Envelope

	WaveformRam [32]int8
	Time float64
}

func (s *Sound) Init(gb *Gameboy) {
	sample_rate := beep.SampleRate(41040)
	speaker.Init(sample_rate, sample_rate.N(time.Second/30))

	s.Time = 0
	s.Channel1 = GetChannel(Square, s.Time)
	s.Channel2 = GetChannel(Square, s.Time)
	s.Channel3 = GetChannel(MakeWaveform(&s.WaveformRam), s.Time)
	s.Channel4 = GetChannel(Noise, s.Time)

	mix := beep.Mix(
		s.Channel1.Stream(float64(sample_rate)),
		s.Channel2.Stream(float64(sample_rate)),
		s.Channel3.Stream(float64(sample_rate)),
		s.Channel4.Stream(float64(sample_rate)),
	)
	speaker.Play(mix)
	s.GB = gb
}

var sound_mask = []byte{
	0xFF, 0xC0, 0xFF, 0x00, 0x40, /*0xFF15->*/
	0x00, 0xC0, 0xFF, 0x00, 0x40,
	0x80, 0x00, 0x60, 0x00, 0x40, /*<-0xFF1E*/
	0x00, 0x3F, 0xFF, 0xFF, 0x40,
	0xFF, 0xFF, 0x80,
}

func (s *Sound) makeSweep(value byte) *Sweep {
	sweep_time := (value >> 4) & 0x7
	sweep_increase := !bits.Test(value, 3)
	sweep_shift := value & 0x7

	return &Sweep{
		StepLen: sweep_time,
		Steps: sweep_shift,
		Step: 0,
		Increase: sweep_increase,
	}
}

func (s *Sound) Write(address uint16, value byte) {
	s.GB.Memory.Data[address] = value

	switch address {
	case 0xFF10:
		s.Channel1Sweep = s.makeSweep(value)

	case 0xFF11:
		NR14 := s.GB.Memory.Read(0xFF14)
		if bits.Test(NR14, 6) {
			// Counter
			s.Channel1Time = (64 - float64(value&0x1F)) * (1 / 256)
		} else {
			// Consecutive
			s.Channel1Time = 100000
		}
		s.Channel1TimeVal = s.Channel1Time
		pattern := (value >> 6) & 0x3
		s.Channel1.FuncMod = squarelimits[pattern]

	case 0xFF12:
		// Envelope
		env_volume := (value >> 4) & 0xF
		env_increase := bits.Test(value, 3)
		env_sweep := value & 0x7

		s.Channel1Env = &Envelope{
			StepLen:    float64(env_sweep) / 64,
			Steps:      env_volume,
			StepsInit:  env_volume,
			Increasing: env_increase,
		}
		if env_volume == 0 {
			s.Toggle(1, false)
		}

	case 0xFF13:
		NR14 := s.GB.Memory.Read(0xFF14)
		s.UpdateChan1Freq(value, NR14)

	case 0xFF14:
		NR13 := s.GB.Memory.Read(0xFF13)
		s.UpdateChan1Freq(NR13, value)
		if bits.Test(value, 7) {
			s.Toggle(1, s.ShouldPlay(1))
			s.Channel1.Amp = 1
			s.Channel1Time = s.Channel1TimeVal
			if s.Channel1Env != nil {
				s.Channel1Env.Steps = s.Channel1Env.StepsInit
			}
			if s.Channel1Sweep != nil {
				s.Channel1Sweep.Step = s.Channel1Sweep.Steps
			}
		}

	case 0xFF16:
		NR24 := s.GB.Memory.Read(0xFF19)
		if bits.Test(NR24, 6) {
			// Counter
			s.Channel2Time = (64 - float64(value&0x1F)) * (1 / 256)
		} else {
			// Consecutive
			s.Channel2Time = 100000
		}
		s.Channel2TimeVal = s.Channel2Time
		pattern := (value >> 6) & 0x3
		s.Channel2.FuncMod = squarelimits[pattern]

	case 0xFF17:
		// Envelope
		env_volume := (value >> 4) & 0xF
		env_increase := bits.Test(value, 3)
		env_sweep := value & 0x7

		s.Channel2Env = &Envelope{
			StepLen:    float64(env_sweep) / 64,
			Steps:      env_volume,
			StepsInit:  env_volume,
			Increasing: env_increase,
		}
		if env_volume == 0 {
			s.Toggle(2, false)
		}

	case 0xFF18:
		NR24 := s.GB.Memory.Read(0xFF19)
		s.UpdateChan2Freq(value, NR24)

	case 0xFF19:
		NR23 := s.GB.Memory.Read(0xFF18)
		s.UpdateChan2Freq(NR23, value)
		if bits.Test(value, 7) {
			s.Toggle(2, s.ShouldPlay(2))
			s.Channel2.Amp = 1
			s.Channel2Time = s.Channel2TimeVal
			if s.Channel2Env != nil {
				s.Channel2Env.Steps = s.Channel2Env.StepsInit
			}
		}

	case 0xFF1A:
		if bits.Test(value, 7) {
			s.Channel3.On()
		} else {
			s.Channel3.Off()
		}

	case 0xFF1B:
		NR34 := s.GB.Memory.Read(0xFF1E)
		if bits.Test(NR34, 6) {
			// Counter
			s.Channel3Time = (64 - float64(value)) * (1 / 256)
		} else {
			// Consecutive
			s.Channel3Time = 100000
		}

	case 0xFF1C:
		s.ToggleCh3Volume(value)

	case 0xFF1D:
		NR34 := s.GB.Memory.Read(0xFF1E)
		s.UpdateChan3Freq(value, NR34)

	case 0xFF1E:
		NR33 := s.GB.Memory.Read(0xFF1D)
		s.UpdateChan3Freq(NR33, value)
		if bits.Test(value, 7) {
			s.Toggle(3, s.ShouldPlay(3))
			s.Channel3.Amp = 1
		}

		// TODO: Channel 4

	case 0xFF23:
		if bits.Test(value, 7) {
			s.StartChannel4(value)
		}

	case 0xFF26:
		s.UpdateOutput(value)

		// END TODO

	case 0xFF24:
		s.SetVolume(value)

	case 0xFF25:
		s.UpdateOutput(value)
	}

	s.GB.Memory.Data[address] = value & sound_mask[address - 0xFF10]
}

func (s *Sound) Tick(clocks int) {
	secs := float64(clocks) / CLOCK_SPEED
	s.Time += secs

	if s.Channel1Time > 0 {
		s.Channel1Time -= secs
		if s.Channel1Env != nil {
			s.Channel1Env.Update(secs, s.Channel1)
		}
		if s.Channel1Sweep != nil {
			s.Channel1Sweep.Update(secs, s.Channel1)
		}
	} else {
		s.Toggle(1, false)
	}

	if s.Channel2Time > 0 {
		s.Channel2Time -= secs
		if s.Channel2Env != nil {
			s.Channel2Env.Update(secs, s.Channel2)
		}
	} else {
		s.Toggle(2, false)
	}

	if s.Channel3Time > 0 {
		s.Channel3Time -= secs
	} else {
		s.Toggle(3, false)
	}

	if s.Channel4Time > 0 {
		s.Channel4Time -= secs
		if s.Channel4Env != nil {
			s.Channel4Env.Update(secs, s.Channel4)
		}
	} else {
		s.Toggle(4, false)
	}

	s.Channel1.DebugMute = s.GB.Debug.MuteChannel1
	s.Channel2.DebugMute = s.GB.Debug.MuteChannel2
	s.Channel3.DebugMute = s.GB.Debug.MuteChannel3
	s.Channel4.DebugMute = s.GB.Debug.MuteChannel4

	s.Channel1.Buffer(684)
	s.Channel2.Buffer(684)
	s.Channel3.Buffer(684)
	s.Channel4.Buffer(684)
}

func (s *Sound) UpdateOutput(value byte) {
	if !s.ShouldPlay(1) {
		s.Channel1.Off()
	}
	if !s.ShouldPlay(2) {
		s.Channel2.Off()
	}
	if !s.ShouldPlay(3) {
		s.Channel3.Off()
	}
	if !s.ShouldPlay(4) {
		s.Channel4.Off()
	}
}

func (s *Sound) ShouldPlay(channel byte) bool {
	FF25 := s.GB.Memory.Data[0xFF25]
	FF26 := s.GB.Memory.Data[0xFF26]

	// Individual sound control
	return bits.Test(FF25, channel-1) && bits.Test(FF25, channel+3) &&
	// All sound on/off
		bits.Test(FF26, 7)
}

func (s *Sound) UpdateChan1Freq(NR13 byte, NR14 byte) {
	freq_val := uint16(NR13) | (uint16(NR14&0x7) << 8)
	freq := 131072 / (2048 - float64(freq_val))
	s.Channel1.Freq = freq
}

func (s *Sound) UpdateChan2Freq(NR23 byte, NR24 byte) {
	freq_val := uint16(NR23) | (uint16(NR24&0x7) << 8)
	freq := 131072 / (2048 - float64(freq_val))
	s.Channel2.Freq = freq
}

func (s *Sound) UpdateChan3Freq(NR33 byte, NR34 byte) {
	freq_val := uint16(NR33) | (uint16(NR34&0x7) << 8)
	freq := 65536 / (2048 - float64(freq_val))
	s.Channel3.Freq = freq
}

func (s *Sound) StartChannel4(NR44 byte) {
	NR42 := s.GB.Memory.Data[0xFF21]
	NR43 := s.GB.Memory.Data[0xFF22]

	if bits.Test(NR44, 6) {
		// Counter
		NR41 := s.GB.Memory.Data[0xFF20]
		s.Channel4Time = (64 - float64(NR41 & 0x3F)) * (1 / 256)
	} else {
		// Consecutive
		s.Channel4Time = 100000
	}
	//pattern := (channel >> 6) & 0x3

	freq_shift_clock := float64((NR43 >> 4) & 0xF)
	freq_divier := float64(NR43 & 0x7)
	if freq_divier == 0 {
		freq_divier = 0.5
	}
	freq := 524288 / freq_divier / math.Pow(2, freq_shift_clock + 1)
	// TODO: Bit 3 NR43 modifier
	s.Channel4.Freq = freq

	// Envelope
	env_volume := (NR42 >> 4) & 0xF
	env_increase := bits.Test(NR42, 3)
	env_sweep := NR42 & 0x7

	s.Channel4Env = &Envelope{
		StepLen:   float64(env_sweep) / 64,
		Steps:     env_volume,
		StepsInit: env_volume,
		Increasing: env_increase,
	}

	s.Toggle(4, s.ShouldPlay(4))
	s.Channel4.Amp = 1
}

func (s *Sound) Toggle(channel byte, on bool) {
	c := s.Channel1
	switch channel {
	case 2: c = s.Channel2
	case 3: c = s.Channel3
	case 4: c = s.Channel4
	}
	if on && s.ShouldPlay(channel) {
		//fmt.Print("chn ", channel)
		c.On()

		s.GB.Memory.Data[0xFF26] = bits.Set(s.GB.Memory.Data[0xFF26], channel - 1)
	} else {
		c.Off()
		s.GB.Memory.Data[0xFF26] = bits.Reset(s.GB.Memory.Data[0xFF26], channel - 1)
	}
}

func (s *Sound) SetVolume(value byte) {
	so1vol := float64(value & 0x7) / 7
	so2vol := float64((value >> 4) & 0x7) / 7

	s.Channel1.SetVolume(so1vol, so2vol)
	s.Channel2.SetVolume(so1vol, so2vol)
	s.Channel3.SetVolume(so1vol, so2vol)
	s.Channel4.SetVolume(so1vol, so2vol)
}

func (s *Sound) ToggleCh3(value byte) {
	if bits.Test(value, 7) {
		s.Channel3.On()
	} else {
		s.Channel3.Off()
	}
}

var ch3vols = map[byte]float64{
	0: 0, 1: 1, 2: 0.5, 3: 0.25,
}

func (s *Sound) ToggleCh3Volume(value byte) {
	/*
	0: Mute (No sound)
    1: 100% Volume (Produce Wave Pattern RAM Data as it is)
    2:  50% Volume (Produce Wave Pattern RAM data shifted once to the right)
    3:  25% Volume (Produce Wave Pattern RAM data shifted twice to the right)
	*/
	// TODO: What does that mean/
	vol := value >> 5 & 0x3
	s.Channel3.Volume = ch3vols[vol]
}