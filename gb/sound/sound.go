package sound

import (
	"math"

	"math/rand"

	"github.com/hajimehoshi/oto"
)

const twoPi = 2 * math.Pi
const SampleRate float64 = 44100

func NewMixer() *Mixer {
	player, err := oto.NewPlayer(int(SampleRate), 1, 1, 735*10)
	if err != nil {
		panic(err)
	}
	return &Mixer{
		player: player,
	}
}

type Mixer struct {
	channels []*Channel
	player   *oto.Player
}

func (m *Mixer) NewChannel(gen WaveGenerator) *Channel {
	channel := &Channel{
		Gen:    gen,
		Volume: 1,
	}
	m.channels = append(m.channels, channel)
	return channel
}

func (m *Mixer) Buffer(samples int) {
	var buf []byte
	for i := 0; i < samples; i++ {
		total := 0
		for _, channel := range m.channels {
			total += int(channel.Sample())
		}
		buf = append(buf, byte(total/len(m.channels)))
	}
	m.player.Write(buf)
}

type Channel struct {
	Freq   float64
	Amp    float64
	Mod    float64
	Gen    WaveGenerator
	Volume float64

	DebugMute bool
	on        bool

	time float64
}

func (c *Channel) Sample() byte {
	if !c.on || c.DebugMute {
		return 0
	}
	c.time += 1
	tpc := 44100.0 / c.Freq
	cycles := c.time / tpc
	amp := c.Amp * c.Volume * 120 // TODO: 126?
	return byte(c.Gen(twoPi*cycles, c.Mod)*amp) + 126
}

// Set the amplitude of the sound channel.
func (c *Channel) SetAmp(amp float64) {
	c.Amp = amp
}

// Enable the sound channel.
func (c *Channel) On() {
	c.on = true
}

// Disable the sound channel.
func (c *Channel) Off() {
	c.on = false
}

// Set the volume of the sound channel for the two output terminals.
func (c *Channel) SetVolume(so1vol float64, so2vol float64) {
	// TODO: fix this?
	c.Volume = so1vol
}

type WaveGenerator func(rad, mod float64) float64

func Sine(rad, _ float64) float64 {
	return math.Sin(rad)
}

func Square(rad, mod float64) float64 {
	var val float64 = 1
	if math.Sin(rad) <= mod {
		val = -1
	}
	return val
}

func Noise(_, _ float64) float64 {
	// TODO: improve this so it depends on rad
	var val float64 = 1
	if rand.Intn(2) == 0 {
		val = -1
	}
	return val
}

func MakeWaveform(data *[32]int8) func(rad, mod float64) float64 {
	return func(rad, mod float64) float64 {
		idx := int(math.Floor(rad/twoPi*32)) % 32
		data := int16(int8(data[idx]<<4) >> 4)
		return float64(data)/3.5 - 1
	}
}
