// Package speaker implements playback of beep.Streamer values through physical speakers.
package speaker

import (
	"sync"

	"github.com/faiface/beep"
	"github.com/hajimehoshi/oto"
	"github.com/pkg/errors"
)

var (
	mu       sync.Mutex
	mixer    beep.Mixer
	samples  [][2]float64
	buf      []byte
	player   *oto.Player
	underrun func()
	done     chan struct{}
)

// Init initializes audio playback through speaker. Must be called before using this package.
//
// The bufferSize argument specifies the number of samples of the speaker's buffer. Bigger
// bufferSize means lower CPU usage and more reliable playback. Lower bufferSize means better
// responsiveness and less delay.
func Init(sampleRate beep.SampleRate, bufferSize int) error {
	mu.Lock()
	defer mu.Unlock()

	if player != nil {
		done <- struct{}{}
		player.Close()
	}

	mixer = beep.Mixer{}

	numBytes := bufferSize * 4
	samples = make([][2]float64, bufferSize)
	buf = make([]byte, numBytes)

	var err error
	player, err = oto.NewPlayer(int(sampleRate), 2, 2, numBytes)
	if err != nil {
		return errors.Wrap(err, "failed to initialize speaker")
	}

	if underrun != nil {
		player.SetUnderrunCallback(underrun)
	}

	done = make(chan struct{})

	go func() {
		for {
			select {
			default:
				update()
			case <-done:
				return
			}
		}
	}()

	return nil
}

// UnderrunCallback sets a function which will be called when an underrun occurs. This is useful for
// debugging and optimization purposes.
//
// Underrun happens when program doesn't keep up with the audio playback and doesn't supply audio
// data quickly enough. To fix an underrun, you either need to optimize your audio processing code,
// or increase the buffer size.
//
// Underrun detection currently works on Linux.
func UnderrunCallback(f func()) {
	mu.Lock()
	underrun = f
	if player != nil {
		player.SetUnderrunCallback(underrun)
	}
	mu.Unlock()
}

// Lock locks the speaker. While locked, speaker won't pull new data from the playing Stramers. Lock
// if you want to modify any currently playing Streamers to avoid race conditions.
//
// Always lock speaker for as little time as possible, to avoid playback glitches.
func Lock() {
	mu.Lock()
}

// Unlock unlocks the speaker. Call after modifying any currently playing Streamer.
func Unlock() {
	mu.Unlock()
}

// Play starts playing all provided Streamers through the speaker.
func Play(s ...beep.Streamer) {
	mu.Lock()
	mixer.Play(s...)
	mu.Unlock()
}

// update pulls new data from the playing Streamers and sends it to the speaker. Blocks until the
// data is sent and started playing.
func update() {
	mu.Lock()
	mixer.Stream(samples)
	mu.Unlock()

	for i := range samples {
		for c := range samples[i] {
			val := samples[i][c]
			if val < -1 {
				val = -1
			}
			if val > +1 {
				val = +1
			}
			valInt16 := int16(val * (1<<15 - 1))
			low := byte(valInt16)
			high := byte(valInt16 >> 8)
			buf[i*4+c*2+0] = low
			buf[i*4+c*2+1] = high
		}
	}

	player.Write(buf)
}
