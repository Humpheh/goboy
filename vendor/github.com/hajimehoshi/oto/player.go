// Copyright 2017 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package oto offers io.Writer to play sound on multiple platforms.
package oto

import (
	"time"
)

// Player is a PCM (pulse-code modulation) audio player. It implements io.Writer, use Write method
// to play samples.
type Player struct {
	player         *player
	sampleRate     int
	channelNum     int
	bytesPerSample int
	bufferSize     int
}

// NewPlayer creates a new, ready-to-use Player.
//
// The sampleRate argument specifies the number of samples that should be played during one second.
// Usual numbers are 44100 or 48000.
//
// The channelNum argument specifies the number of channels. One channel is mono playback. Two
// channels are stereo playback. No other values are supported.
//
// The bytesPerSample argument specifies the number of bytes per sample per channel. The usual value
// is 2. Only values 1 and 2 are supported.
//
// The bufferSizeInBytes argument specifies the size of the buffer of the Player. This means, how
// many bytes can Player remember before actually playing them. Bigger buffer can reduce the number
// of Write calls, thus reducing CPU time. Smaller buffer enables more precise timing. The longest
// delay between when samples were written and when they started playing is equal to the size of the
// buffer.
func NewPlayer(sampleRate, channelNum, bytesPerSample, bufferSizeInBytes int) (*Player, error) {
	p, err := newPlayer(sampleRate, channelNum, bytesPerSample, bufferSizeInBytes)
	if err != nil {
		return nil, err
	}
	return &Player{
		player:         p,
		sampleRate:     sampleRate,
		channelNum:     channelNum,
		bytesPerSample: bytesPerSample,
		bufferSize:     bufferSizeInBytes,
	}, nil
}

func (p *Player) bytesPerSec() int {
	return p.sampleRate * p.channelNum * p.bytesPerSample
}

// SetUnderrunCallback sets a function which will be called whenever an underrun occurs. This is
// mostly for debugging and optimization purposes.
//
// Underrun occurs when not enough samples is written to the player in a certain amount of time and
// thus there's nothing to play. This usually happens when there's too much audio data processing,
// or the audio data processing code gets stuck for a while, or the player's buffer is too small.
//
// Example:
//
//   player.SetUnderrunCallback(func() {
//       log.Println("UNDERRUN, YOUR CODE IS SLOW")
//   })
//
// Supported platforms: Linux.
func (p *Player) SetUnderrunCallback(f func()) {
	p.player.SetUnderrunCallback(f)
}

// Write writes PCM samples to the Player.
//
// The format is as follows:
//   [data]      = [sample 1] [sample 2] [sample 3] ...
//   [sample *]  = [channel 1] ...
//   [channel *] = [byte 1] [byte 2] ...
// Byte ordering is little endian.
//
// The data is first put into the Player's buffer. Once the buffer is full, Player starts playing
// the data and empties the buffer.
//
// If the supplied data doesn't fit into the Player's buffer, Write block until a sufficient amount
// of data has been played (or at least started playing) and the remaining unplayed data fits into
// the buffer.
//
// Note, that the Player won't start playing anything until the buffer is full.
func (p *Player) Write(data []byte) (int, error) {
	written := 0
	for len(data) > 0 {
		n, err := p.player.TryWrite(data)
		written += n
		if err != nil {
			return written, err
		}
		data = data[n:]
		// When not all data is written, the underlying buffer is full.
		// Mitigate the busy loop by sleeping (#10).
		if len(data) > 0 {
			t := time.Second * time.Duration(p.bufferSize) / time.Duration(p.bytesPerSec()) / 8
			time.Sleep(t)
		}
	}
	return written, nil
}

// Close closes the Player and frees any resources associated with it. The Player is no longer
// usable after calling Close.
func (p *Player) Close() error {
	return p.player.Close()
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
