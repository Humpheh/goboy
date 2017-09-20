package gb

import (
	"github.com/humpheh/goboy/bits"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"time"
	"math"
	"fmt"
)

var squarelimits = map[byte]float64{
	0: -0.25, // 12.5% ( _-------_-------_------- )
	1: -0.5,  // 25%   ( __------__------__------ )
	2: 0,     // 50%   ( ____----____----____---- ) (normal)
	3: 0.5,   // 75%   ( ______--______--______-- )
}

type Envelope struct {
	Time      float64
	StepLen   float64
	Steps     byte
	StepsInit byte
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
			channel.SetAmp(amp)
		}
	}
}

type Sound struct {
	GB           *Gameboy
	Channel1     *Channel
	Channel1Time float64
	Channel1Env  *Envelope
	Channel2     *Channel
	Channel2Time float64
	Channel2Env  *Envelope
	Channel3     *Channel
	Channel3Time float64
	Channel4     *Channel
	Channel4Time float64
	Channel4Env  *Envelope
}

func (s *Sound) Init(gb *Gameboy) {
	sample_rate := beep.SampleRate(48000)
	speaker.Init(sample_rate, sample_rate.N(time.Second/30))

	s.Channel1 = GetChannel(Square)
	s.Channel2 = GetChannel(Square)
	s.Channel3 = GetChannel(Square)
	s.Channel4 = GetChannel(Noise)

	mix := beep.Mix(
		s.Channel1.Stream(float64(sample_rate)),
		s.Channel2.Stream(float64(sample_rate)),
		s.Channel3.Stream(float64(sample_rate)),
		s.Channel4.Stream(float64(sample_rate)),
	)
	speaker.Play(mix)
	s.GB = gb
}

func (s *Sound) Tick(clocks int) {
	/*
	29. FF26 (NR 52)
 Name - NR 52 (Value at reset: $F1-GB, $F0-SGB)
 Contents - Sound on/off (R/W)
 Bit 7 - All sound on/off
 0: stop all sound circuits
 1: operate all sound circuits
 Bit 3 - Sound 4 ON flag
 Bit 2 - Sound 3 ON flag
 Bit 1 - Sound 2 ON flag
 Bit 0 - Sound 1 ON flag
 Bits 0 - 3 of this register are meant to
 be status bits to be read. Writing to
 these bits does NOT enable/disable
 sound.
FF24 - NR50 - Channel control / ON-OFF / Volume (R/W)
The volume bits specify the "Master Volume" for Left/Right sound output.
  Bit 7   - Output Vin to SO2 terminal (1=Enable)
  Bit 6-4 - SO2 output level (volume)  (0-7)
  Bit 3   - Output Vin to SO1 terminal (1=Enable)
  Bit 2-0 - SO1 output level (volume)  (0-7)
The Vin signal is received from the game cartridge bus, allowing external
hardware in the cartridge to supply a fifth sound channel, additionally to the
gameboys internal four channels. As far as I know this feature isn't used by
any existing games.
	 */
	//control := s.GB.Memory.Data[0xFF26]
	//output := s.GB.Memory.Data[0xFF25]
	//volume := s.GB.Memory.Data[0xFF24]
	//
	//sound_on := bits.Test(control, 7)
	//
	//_ = volume
	////vol1 := volume & 0x7
	////vol2 := (volume >> 4) & 0x7
	//vol := snd.DefaultAmpFac// - (snd.DefaultAmpFac / 7) * float64(vol1)
	//
	//chan1out := bits.Test(output, 0) || bits.Test(output, 4)
	//if sound_on && chan1out {
	//	s.Channel1.On()
	//
	//} else {
	//	s.Channel1.Off()
	//	s.Channel1Time = 0
	//}

	secs := float64(clocks) / CLOCK_SPEED

	if s.Channel1Time > 0 {
		s.Channel1Time -= secs
		if s.Channel1Env != nil {
			s.Channel1Env.Update(secs, s.Channel1)
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
}

func (s *Sound) ShouldPlay(channel byte) bool {
	FF25 := s.GB.Memory.Data[0xFF25]
	FF26 := s.GB.Memory.Data[0xFF26]

	// Individual sound control
	return bits.Test(FF25, channel-1) || bits.Test(FF25, channel+3) ||
	// All sound on/off
		bits.Test(FF26, 7)
}

func (s *Sound) StartChannel1(NR14 byte) {

	/*
	NR11 - Channel 1 Sound length/Wave pattern duty (R/W)
  Bit 7-6 - Wave Pattern Duty (Read/Write)
  Bit 5-0 - Sound length data (Write Only) (t1: 0-63)
Wave Duty:
  00: 12.5% ( _-------_-------_------- )
  01: 25%   ( __------__------__------ )
  10: 50%   ( ____----____----____---- ) (normal)
  11: 75%   ( ______--______--______-- )
Sound Length = (64-t1)*(1/256) seconds
The Length value is used only if Bit 6 in NR14 is set.
	 */

	/*
FF12 - NR12 - Channel 1 Volume Envelope (R/W)
 Bit 7-4 - Initial Volume of envelope (0-0Fh) (0=No Sound)
 Bit 3   - Envelope Direction (0=Decrease, 1=Increase)
 Bit 2-0 - Number of envelope sweep (n: 0-7)
		   (If zero, stop envelope operation.)
Length of 1 step = n*(1/64) seconds
*/

	NR11 := s.GB.Memory.Data[0xFF11]
	NR12 := s.GB.Memory.Data[0xFF12]
	NR13 := s.GB.Memory.Data[0xFF13]

	if bits.Test(NR14, 6) {
		// Counter
		s.Channel1Time = (64 - float64(NR11&0x1F)) * (1 / 256)
	} else {
		// Consecutive
		s.Channel1Time = 100000
	}
	pattern := (NR11 >> 6) & 0x3
	s.Channel1.FuncMod = squarelimits[pattern]

	freq_val := uint16(NR13) | (uint16(NR14&0x7) << 8)
	freq := 131072 / (2048 - float64(freq_val))
	s.Channel1.Freq = freq

	// Envelope
	env_volume := (NR12 >> 4) & 0xF
	//env_increase := bits.Test(NR12, 3)
	env_sweep := NR12 & 0x7

	s.Channel1Env = &Envelope{
		StepLen:   float64(env_sweep) / 64,
		Steps:     env_volume,
		StepsInit: env_volume,
	}

	s.Toggle(1, s.ShouldPlay(1))
	s.Channel1.Amp = 1
}

func (s *Sound) StartChannel2(NR24 byte) {

	/*
	NR11 - Channel 1 Sound length/Wave pattern duty (R/W)
  Bit 7-6 - Wave Pattern Duty (Read/Write)
  Bit 5-0 - Sound length data (Write Only) (t1: 0-63)
Wave Duty:
  00: 12.5% ( _-------_-------_------- )
  01: 25%   ( __------__------__------ )
  10: 50%   ( ____----____----____---- ) (normal)
  11: 75%   ( ______--______--______-- )
Sound Length = (64-t1)*(1/256) seconds
The Length value is used only if Bit 6 in NR14 is set.

	FF17 - NR22 - Channel 2 Volume Envelope (R/W)
  Bit 7-4 - Initial Volume of envelope (0-0Fh) (0=No Sound)
  Bit 3   - Envelope Direction (0=Decrease, 1=Increase)
  Bit 2-0 - Number of envelope sweep (n: 0-7)
            (If zero, stop envelope operation.)
Length of 1 step = n*(1/64) seconds

	 */
	NR22 := s.GB.Memory.Data[0xFF17]
	NR23 := s.GB.Memory.Data[0xFF18]

	if bits.Test(NR24, 6) {
		// Counter
		NR21 := s.GB.Memory.Data[0xFF16]
		s.Channel2Time = (64 - float64(NR21&0x1F)) * (1 / 256)
	} else {
		// Consecutive
		s.Channel2Time = 100000
	}
	//pattern := (channel >> 6) & 0x3

	freq_val := uint16(NR23) | (uint16(NR24&0x7) << 8)
	freq := 131072 / (2048 - float64(freq_val))
	s.Channel2.Freq = freq

	// Envelope
	env_volume := (NR22 >> 4) & 0xF
	//env_increase := bits.Test(NR12, 3)
	env_sweep := NR22 & 0x7

	s.Channel2Env = &Envelope{
		StepLen:   float64(env_sweep) / 64,
		Steps:     env_volume,
		StepsInit: env_volume,
	}

	s.Toggle(2, s.ShouldPlay(2))
	s.Channel2.Amp = 1
}

func (s *Sound) StartChannel3(NR34 byte) {

	//NR32 := s.GB.Memory.Data[0xFF1C] // TODO: Volume
	NR33 := s.GB.Memory.Data[0xFF1D]

	if bits.Test(NR34, 6) {
		// Counter
		NR21 := s.GB.Memory.Data[0xFF1B]
		s.Channel3Time = (64 - float64(NR21)) * (1 / 256)
	} else {
		// Consecutive
		s.Channel3Time = 100000
	}
	//pattern := (channel >> 6) & 0x3

	freq_val := uint16(NR33) | (uint16(NR34&0x7) << 8)
	freq := 65536 / (2048 - float64(freq_val))
	s.Channel3.Freq = freq

	s.Toggle(3, s.ShouldPlay(3))
	s.Channel3.Amp = 1
}

func (s *Sound) StartChannel4(NR44 byte) {
	/*

FF21 - NR42 - Channel 4 Volume Envelope (R/W)
  Bit 7-4 - Initial Volume of envelope (0-0Fh) (0=No Sound)
  Bit 3   - Envelope Direction (0=Decrease, 1=Increase)
  Bit 2-0 - Number of envelope sweep (n: 0-7)
            (If zero, stop envelope operation.)
Length of 1 step = n*(1/64) seconds

FF22 - NR43 - Channel 4 Polynomial Counter (R/W)
The amplitude is randomly switched between high and low at the given
frequency. A higher frequency will make the noise to appear 'softer'.
When Bit 3 is set, the output will become more regular, and some frequencies
will sound more like Tone than Noise.
  Bit 7-4 - Shift Clock Frequency (s)
  Bit 3   - Counter Step/Width (0=15 bits, 1=7 bits)
  Bit 2-0 - Dividing Ratio of Frequencies (r)
Frequency = 524288 Hz / r / 2^(s+1)     ;For r=0 assume r=0.5 instead

FF23 - NR44 - Channel 4 Counter/consecutive; Inital (R/W)
  Bit 7   - Initial (1=Restart Sound)     (Write Only)
  Bit 6   - Counter/consecutive selection (Read/Write)
            (1=Stop output when length in NR41 expires)
	 */
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
	//env_increase := bits.Test(NR12, 3)
	env_sweep := NR42 & 0x7

	s.Channel4Env = &Envelope{
		StepLen:   float64(env_sweep) / 64,
		Steps:     env_volume,
		StepsInit: env_volume,
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
	if on {
		fmt.Print("chn ", channel)
		c.On()
		s.GB.Memory.Data[0xFF26] = bits.Set(s.GB.Memory.Data[0xFF26], channel - 1)
	} else {
		c.Off()
		s.GB.Memory.Data[0xFF26] = bits.Reset(s.GB.Memory.Data[0xFF26], channel - 1)
	}
}

func (s *Sound) SetVolume(value byte) {
	/*
	The volume bits specify the "Master Volume" for Left/Right sound output.
	Bit 7   - Output Vin to SO2 terminal (1=Enable)
	Bit 6-4 - SO2 output level (volume)  (0-7)
	Bit 3   - Output Vin to SO1 terminal (1=Enable)
	Bit 2-0 - SO1 output level (volume)  (0-7)
	*/
	so1vol := float64(value & 0x7) / 7
	so2vol := float64((value >> 4) & 0x7) / 7

	s.Channel1.SetVolume(so1vol, so2vol)
	s.Channel2.SetVolume(so1vol, so2vol)
	s.Channel3.SetVolume(so1vol, so2vol)
	s.Channel4.SetVolume(so1vol, so2vol)
}