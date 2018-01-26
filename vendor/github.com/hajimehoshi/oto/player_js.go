// Copyright 2015 Hajime Hoshi
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

// +build js

package oto

import (
	"errors"
	"strings"

	"github.com/gopherjs/gopherjs/js"
)

type player struct {
	sampleRate     int
	channelNum     int
	bytesPerSample int
	nextPos        float64
	tmp            []byte
	bufferSize     int
	context        *js.Object
	lastTime       float64
	lastAudioTime  float64
}

func isIOSSafari() bool {
	ua := js.Global.Get("navigator").Get("userAgent").String()
	if !strings.Contains(ua, "iPhone") {
		return false
	}
	return true
}

func isAndroidChrome() bool {
	ua := js.Global.Get("navigator").Get("userAgent").String()
	if !strings.Contains(ua, "Android") {
		return false
	}
	if !strings.Contains(ua, "Chrome") {
		return false
	}
	return true
}

const audioBufferSamples = 3200

func newPlayer(sampleRate, channelNum, bytesPerSample, bufferSize int) (*player, error) {
	class := js.Global.Get("AudioContext")
	if class == js.Undefined {
		class = js.Global.Get("webkitAudioContext")
	}
	if class == js.Undefined {
		return nil, errors.New("oto: audio couldn't be initialized")
	}
	p := &player{
		sampleRate:     sampleRate,
		channelNum:     channelNum,
		bytesPerSample: bytesPerSample,
		context:        class.New(),
		bufferSize:     max(bufferSize, audioBufferSamples*channelNum*bytesPerSample),
	}
	// iOS Safari and Android Chrome requires touch event to use AudioContext.
	if isIOSSafari() || isAndroidChrome() {
		var f *js.Object
		f = js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
			// Resuming is necessary as of Chrome 55+ in some cases like different
			// domain page in an iframe.
			p.context.Call("resume")
			p.context.Call("createBufferSource").Call("start", 0)
			js.Global.Get("document").Call("removeEventListener", "touchend", f)
			return nil
		})
		js.Global.Get("document").Call("addEventListener", "touchend", f)
	}
	return p, nil
}

func toLR(data []byte) ([]float32, []float32) {
	const max = 1 << 15

	l := make([]float32, len(data)/4)
	r := make([]float32, len(data)/4)
	for i := 0; i < len(data)/4; i++ {
		l[i] = float32(int16(data[4*i])|int16(data[4*i+1])<<8) / max
		r[i] = float32(int16(data[4*i+2])|int16(data[4*i+3])<<8) / max
	}
	return l, r
}

func (p *player) SetUnderrunCallback(f func()) {
	//TODO
}

func nowInSeconds() float64 {
	return js.Global.Get("performance").Call("now").Float() / 1000.0
}

func (p *player) TryWrite(data []byte) (int, error) {
	n := min(len(data), max(0, p.bufferSize-len(p.tmp)))
	p.tmp = append(p.tmp, data[:n]...)

	c := p.context.Get("currentTime").Float()
	now := nowInSeconds()

	if p.lastTime != 0 && p.lastAudioTime != 0 && p.lastAudioTime >= c && p.lastTime != now {
		// Unfortunately, currentTime might not be precise enough on some devices
		// (e.g. Android Chrome). Adjust the audio time with OS clock.
		c = p.lastAudioTime + now - p.lastTime
	}

	p.lastAudioTime = c
	p.lastTime = now

	if p.nextPos < c {
		p.nextPos = c
	}

	// It's too early to enqueue a buffer.
	// Highly likely, there are two playing buffers now.
	if c+float64(p.bufferSize/p.bytesPerSample/p.channelNum)/float64(p.sampleRate) < p.nextPos {
		return n, nil
	}

	le := audioBufferSamples * p.bytesPerSample * p.channelNum
	if len(p.tmp) < le {
		return n, nil
	}

	buf := p.context.Call("createBuffer", p.channelNum, audioBufferSamples, p.sampleRate)
	l, r := toLR(p.tmp[:le])
	if buf.Get("copyToChannel") != js.Undefined {
		buf.Call("copyToChannel", l, 0, 0)
		buf.Call("copyToChannel", r, 1, 0)
	} else {
		// copyToChannel is not defined on Safari 11
		outL := buf.Call("getChannelData", 0).Interface().([]float32)
		outR := buf.Call("getChannelData", 1).Interface().([]float32)
		copy(outL, l)
		copy(outR, r)
	}

	s := p.context.Call("createBufferSource")
	s.Set("buffer", buf)
	s.Call("connect", p.context.Get("destination"))
	s.Call("start", p.nextPos)
	p.nextPos += buf.Get("duration").Float()

	p.tmp = p.tmp[le:]
	return n, nil
}

func (p *player) Close() error {
	return nil
}
