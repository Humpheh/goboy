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

// +build darwin freebsd
// +build !js
// +build !android

package oto

// #cgo darwin  LDFLAGS: -framework OpenAL
// #cgo freebsd LDFLAGS: -lopenal
//
// #ifdef __APPLE__
// #include <OpenAL/al.h>
// #include <OpenAL/alc.h>
// #else
// #include <AL/al.h>
// #include <AL/alc.h>
// #endif
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

// As x/mobile/exp/audio/al is broken on macOS (https://github.com/golang/go/issues/15075),
// and that doesn't support FreeBSD, use OpenAL directly here.

type player struct {
	// alContext represents a pointer to ALCcontext. The type is uintptr since the value
	// can be 0x18 on macOS, which is invalid as a pointer value, and this might cause
	// GC errors.
	alContext    alContext
	alDevice     alDevice
	alDeviceName string
	alSource     C.ALuint
	sampleRate   int
	isClosed     bool
	alFormat     C.ALenum

	bufs       []C.ALuint
	tmp        []byte
	bufferSize int
}

type alContext uintptr

func (a alContext) cALCcontext() *C.ALCcontext {
	return (*C.struct_ALCcontext_struct)(unsafe.Pointer(a))
}

type alDevice uintptr

func (a alDevice) getError() error {
	c := C.alcGetError(a.cALCDevice())
	switch c {
	case C.ALC_NO_ERROR:
		return nil
	case C.ALC_INVALID_DEVICE:
		return errors.New("OpenAL error: invalid device")
	case C.ALC_INVALID_CONTEXT:
		return errors.New("OpenAL error: invalid context")
	case C.ALC_INVALID_ENUM:
		return errors.New("OpenAL error: invalid enum")
	case C.ALC_INVALID_VALUE:
		return errors.New("OpenAL error: invalid value")
	case C.ALC_OUT_OF_MEMORY:
		return errors.New("OpenAL error: out of memory")
	default:
		return fmt.Errorf("OpenAL error: code %d", c)
	}
}

func (a alDevice) cALCDevice() *C.ALCdevice {
	return (*C.struct_ALCdevice_struct)(unsafe.Pointer(a))
}

func alFormat(channelNum, bytesPerSample int) C.ALenum {
	switch {
	case channelNum == 1 && bytesPerSample == 1:
		return C.AL_FORMAT_MONO8
	case channelNum == 1 && bytesPerSample == 2:
		return C.AL_FORMAT_MONO16
	case channelNum == 2 && bytesPerSample == 1:
		return C.AL_FORMAT_STEREO8
	case channelNum == 2 && bytesPerSample == 2:
		return C.AL_FORMAT_STEREO16
	}
	panic(fmt.Sprintf("oto: invalid channel num (%d) or bytes per sample (%d)", channelNum, bytesPerSample))
}

const numBufs = 2

func newPlayer(sampleRate, channelNum, bytesPerSample, bufferSizeInBytes int) (*player, error) {
	name := C.alGetString(C.ALC_DEFAULT_DEVICE_SPECIFIER)
	d := alDevice(unsafe.Pointer(C.alcOpenDevice((*C.ALCchar)(name))))
	if d == 0 {
		return nil, fmt.Errorf("oto: alcOpenDevice must not return null")
	}
	c := alContext(unsafe.Pointer(C.alcCreateContext((*C.struct_ALCdevice_struct)(unsafe.Pointer(d)), nil)))
	if c == 0 {
		return nil, fmt.Errorf("oto: alcCreateContext must not return null")
	}

	// Don't check getError until making the current context is done.
	// Linux might fail this check even though it succeeds (hajimehoshi/ebiten#204).
	C.alcMakeContextCurrent(c.cALCcontext())
	if err := d.getError(); err != nil {
		return nil, fmt.Errorf("oto: Activate: %v", err)
	}

	s := C.ALuint(0)
	C.alGenSources(1, &s)
	if err := d.getError(); err != nil {
		return nil, fmt.Errorf("oto: NewSource: %v", err)
	}

	p := &player{
		alContext:    c,
		alDevice:     d,
		alSource:     s,
		alDeviceName: C.GoString((*C.char)(name)),
		sampleRate:   sampleRate,
		alFormat:     alFormat(channelNum, bytesPerSample),
		bufs:         make([]C.ALuint, numBufs),
		bufferSize:   bufferSizeInBytes,
	}
	runtime.SetFinalizer(p, (*player).Close)
	C.alGenBuffers(C.ALsizei(numBufs), &p.bufs[0])
	C.alSourcePlay(p.alSource)

	if err := d.getError(); err != nil {
		return nil, fmt.Errorf("oto: Play: %v", err)
	}

	return p, nil
}

func (p *player) SetUnderrunCallback(f func()) {
	//TODO
}

func (p *player) TryWrite(data []byte) (int, error) {
	if err := p.alDevice.getError(); err != nil {
		return 0, fmt.Errorf("oto: starting Write: %v", err)
	}
	n := min(len(data), max(0, p.bufferSize-len(p.tmp)))
	p.tmp = append(p.tmp, data[:n]...)
	if len(p.tmp) < p.bufferSize {
		return n, nil
	}

	pn := C.ALint(0)
	C.alGetSourcei(p.alSource, C.AL_BUFFERS_PROCESSED, &pn)

	if pn > 0 {
		bufs := make([]C.ALuint, pn)
		C.alSourceUnqueueBuffers(p.alSource, C.ALsizei(len(bufs)), &bufs[0])
		if err := p.alDevice.getError(); err != nil {
			return 0, fmt.Errorf("oto: UnqueueBuffers: %v", err)
		}
		p.bufs = append(p.bufs, bufs...)
	}

	if len(p.bufs) == 0 {
		return n, nil
	}

	buf := p.bufs[0]
	p.bufs = p.bufs[1:]
	C.alBufferData(buf, p.alFormat, unsafe.Pointer(&p.tmp[0]), C.ALsizei(p.bufferSize), C.ALsizei(p.sampleRate))
	C.alSourceQueueBuffers(p.alSource, 1, &buf)
	if err := p.alDevice.getError(); err != nil {
		return 0, fmt.Errorf("oto: QueueBuffer: %v", err)
	}

	state := C.ALint(0)
	C.alGetSourcei(p.alSource, C.AL_SOURCE_STATE, &state)
	if state == C.AL_STOPPED || state == C.AL_INITIAL {
		C.alSourceRewind(p.alSource)
		C.alSourcePlay(p.alSource)
		if err := p.alDevice.getError(); err != nil {
			return 0, fmt.Errorf("oto: Rewind or Play: %v", err)
		}
	}

	p.tmp = nil
	return n, nil
}

func (p *player) Close() error {
	if err := p.alDevice.getError(); err != nil {
		return fmt.Errorf("oto: starting Close: %v", err)
	}
	if p.isClosed {
		return nil
	}

	n := C.ALint(0)
	C.alGetSourcei(p.alSource, C.AL_BUFFERS_QUEUED, &n)
	if 0 < n {
		bs := make([]C.ALuint, n)
		C.alSourceUnqueueBuffers(p.alSource, C.ALsizei(len(bs)), &bs[0])
		p.bufs = append(p.bufs, bs...)
	}

	C.alSourceStop(p.alSource)
	C.alDeleteSources(1, &p.alSource)
	if len(p.bufs) != 0 {
		C.alDeleteBuffers(C.ALsizei(numBufs), &p.bufs[0])
	}
	C.alcDestroyContext(p.alContext.cALCcontext())

	if err := p.alDevice.getError(); err != nil {
		return fmt.Errorf("oto: CloseDevice: %v", err)
	}

	b := C.alcCloseDevice(p.alDevice.cALCDevice())
	if b == C.ALC_FALSE {
		return fmt.Errorf("oto: CloseDevice: %s failed to close", p.alDeviceName)
	}

	p.isClosed = true
	runtime.SetFinalizer(p, nil)
	return nil
}
