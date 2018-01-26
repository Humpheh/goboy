// Copyright 2016 Hajime Hoshi
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

package oto

/*

#include <jni.h>
#include <stdlib.h>

static jclass android_media_AudioFormat;
static jclass android_media_AudioManager;
static jclass android_media_AudioTrack;

static char* initAudioTrack(uintptr_t java_vm, uintptr_t jni_env,
    int sampleRate, int channelNum, int bytesPerSample, jobject* audioTrack, int* bufferSize) {
  // bufferSize has an initial value, and this is updated when bufferSize is
  // smaller than AudioTrack's minimum buffer size.

  JavaVM* vm = (JavaVM*)java_vm;
  JNIEnv* env = (JNIEnv*)jni_env;

  jclass local = (*env)->FindClass(env, "android/media/AudioFormat");
  android_media_AudioFormat = (*env)->NewGlobalRef(env, local);
  (*env)->DeleteLocalRef(env, local);

  local = (*env)->FindClass(env, "android/media/AudioManager");
  android_media_AudioManager = (*env)->NewGlobalRef(env, local);
  (*env)->DeleteLocalRef(env, local);

  local = (*env)->FindClass(env, "android/media/AudioTrack");
  android_media_AudioTrack = (*env)->NewGlobalRef(env, local);
  (*env)->DeleteLocalRef(env, local);

  const jint android_media_AudioManager_STREAM_MUSIC =
      (*env)->GetStaticIntField(
          env, android_media_AudioManager,
          (*env)->GetStaticFieldID(env, android_media_AudioManager, "STREAM_MUSIC", "I"));
  const jint android_media_AudioTrack_MODE_STREAM =
      (*env)->GetStaticIntField(
          env, android_media_AudioTrack,
          (*env)->GetStaticFieldID(env, android_media_AudioTrack, "MODE_STREAM", "I"));
  const jint android_media_AudioFormat_CHANNEL_OUT_MONO =
      (*env)->GetStaticIntField(
          env, android_media_AudioFormat,
          (*env)->GetStaticFieldID(env, android_media_AudioFormat, "CHANNEL_OUT_MONO", "I"));
  const jint android_media_AudioFormat_CHANNEL_OUT_STEREO =
      (*env)->GetStaticIntField(
          env, android_media_AudioFormat,
          (*env)->GetStaticFieldID(env, android_media_AudioFormat, "CHANNEL_OUT_STEREO", "I"));
  const jint android_media_AudioFormat_ENCODING_PCM_8BIT =
      (*env)->GetStaticIntField(
          env, android_media_AudioFormat,
          (*env)->GetStaticFieldID(env, android_media_AudioFormat, "ENCODING_PCM_8BIT", "I"));
  const jint android_media_AudioFormat_ENCODING_PCM_16BIT =
      (*env)->GetStaticIntField(
          env, android_media_AudioFormat,
          (*env)->GetStaticFieldID(env, android_media_AudioFormat, "ENCODING_PCM_16BIT", "I"));

  jint channel = android_media_AudioFormat_CHANNEL_OUT_MONO;
  switch (channelNum) {
  case 1:
    channel = android_media_AudioFormat_CHANNEL_OUT_MONO;
    break;
  case 2:
    channel = android_media_AudioFormat_CHANNEL_OUT_STEREO;
    break;
  default:
    return "invalid channel";
  }

  jint encoding = android_media_AudioFormat_ENCODING_PCM_8BIT;
  switch (bytesPerSample) {
  case 1:
    encoding = android_media_AudioFormat_ENCODING_PCM_8BIT;
    break;
  case 2:
    encoding = android_media_AudioFormat_ENCODING_PCM_16BIT;
    break;
  default:
    return "invalid bytesPerSample";
  }

  int minBufferSize =
      (*env)->CallStaticIntMethod(
          env, android_media_AudioTrack,
          (*env)->GetStaticMethodID(env, android_media_AudioTrack, "getMinBufferSize", "(III)I"),
          sampleRate, channel, encoding);
  if (*bufferSize < minBufferSize) {
    *bufferSize = minBufferSize;
  }

  const jobject tmpAudioTrack =
      (*env)->NewObject(
          env, android_media_AudioTrack,
          (*env)->GetMethodID(env, android_media_AudioTrack, "<init>", "(IIIIII)V"),
          android_media_AudioManager_STREAM_MUSIC,
          sampleRate, channel, encoding, *bufferSize,
          android_media_AudioTrack_MODE_STREAM);
  *audioTrack = (*env)->NewGlobalRef(env, tmpAudioTrack);
  (*env)->DeleteLocalRef(env, tmpAudioTrack);

  (*env)->CallVoidMethod(
      env, *audioTrack,
      (*env)->GetMethodID(env, android_media_AudioTrack, "play", "()V"));

  return NULL;
}

static char* writeToAudioTrack(uintptr_t java_vm, uintptr_t jni_env,
    jobject audioTrack, int bytesPerSample, void* data, int length) {
  JavaVM* vm = (JavaVM*)java_vm;
  JNIEnv* env = (JNIEnv*)jni_env;

  jbyteArray arrInBytes;
  jshortArray arrInShorts;
  switch (bytesPerSample) {
  case 1:
    arrInBytes = (*env)->NewByteArray(env, length);
    (*env)->SetByteArrayRegion(env, arrInBytes, 0, length, data);
    break;
  case 2:
    arrInShorts = (*env)->NewShortArray(env, length);
    (*env)->SetShortArrayRegion(env, arrInShorts, 0, length, data);
    break;
  }

  jint result;
  static jmethodID write1 = NULL;
  static jmethodID write2 = NULL;
  if (!write1) {
    write1 = (*env)->GetMethodID(env, android_media_AudioTrack, "write", "([BII)I");
  }
  if (!write2) {
    write2 = (*env)->GetMethodID(env, android_media_AudioTrack, "write", "([SII)I");
  }
  switch (bytesPerSample) {
  case 1:
    result = (*env)->CallIntMethod(env, audioTrack, write1, arrInBytes, 0, length);
    (*env)->DeleteLocalRef(env, arrInBytes);
    break;
  case 2:
    result = (*env)->CallIntMethod(env, audioTrack, write2, arrInShorts, 0, length);
    (*env)->DeleteLocalRef(env, arrInShorts);
    break;
  }

  switch (result) {
  case -3: // ERROR_INVALID_OPERATION
    return "invalid operation";
  case -2: // ERROR_BAD_VALUE
    return "bad value";
  case -1: // ERROR
    return "error";
  }
  if (result < 0) {
    return "unknown error";
  }
  return NULL;
}

static char* releaseAudioTrack(uintptr_t java_vm, uintptr_t jni_env,
    jobject audioTrack) {
  JavaVM* vm = (JavaVM*)java_vm;
  JNIEnv* env = (JNIEnv*)jni_env;

  (*env)->CallVoidMethod(
      env, audioTrack,
      (*env)->GetMethodID(env, android_media_AudioTrack, "release", "()V"));
  return NULL;
}

*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"

	"github.com/hajimehoshi/oto/internal/jni"
)

type player struct {
	sampleRate     int
	channelNum     int
	bytesPerSample int
	audioTrack     C.jobject
	chErr          chan error
	chBuffer       chan []byte
	tmp            []byte
	bufferSize     int
}

func newPlayer(sampleRate, channelNum, bytesPerSample, bufferSizeInBytes int) (*player, error) {
	p := &player{
		sampleRate:     sampleRate,
		channelNum:     channelNum,
		bytesPerSample: bytesPerSample,
		chErr:          make(chan error),
		chBuffer:       make(chan []byte),
	}
	runtime.SetFinalizer(p, (*player).Close)

	if err := jni.RunOnJVM(func(vm, env, ctx uintptr) error {
		audioTrack := C.jobject(nil)
		bufferSize := C.int(bufferSizeInBytes)
		if msg := C.initAudioTrack(C.uintptr_t(vm), C.uintptr_t(env),
			C.int(sampleRate), C.int(channelNum), C.int(bytesPerSample),
			&audioTrack, &bufferSize); msg != nil {
			return errors.New("oto: initAutioTrack failed: " + C.GoString(msg))
		}
		p.audioTrack = audioTrack
		p.bufferSize = int(bufferSize) // bufferSize can be updated at initAudioTrack.
		return nil
	}); err != nil {
		return nil, err
	}

	go p.loop()
	return p, nil
}

func (p *player) loop() {
	for bufInBytes := range p.chBuffer {
		var bufInShorts []int16
		if p.bytesPerSample == 2 {
			bufInShorts = make([]int16, len(bufInBytes)/2)
			for i := 0; i < len(bufInShorts); i++ {
				bufInShorts[i] = int16(bufInBytes[2*i]) | (int16(bufInBytes[2*i+1]) << 8)
			}
		}
		if err := jni.RunOnJVM(func(vm, env, ctx uintptr) error {
			msg := (*C.char)(nil)
			switch p.bytesPerSample {
			case 1:
				msg = C.writeToAudioTrack(C.uintptr_t(vm), C.uintptr_t(env),
					p.audioTrack, C.int(p.bytesPerSample),
					unsafe.Pointer(&bufInBytes[0]), C.int(len(bufInBytes)))
			case 2:
				msg = C.writeToAudioTrack(C.uintptr_t(vm), C.uintptr_t(env),
					p.audioTrack, C.int(p.bytesPerSample),
					unsafe.Pointer(&bufInShorts[0]), C.int(len(bufInShorts)))
			default:
				panic("not reach")
			}
			if msg != nil {
				return errors.New("oto: loop failed: " + C.GoString(msg))
			}
			return nil
		}); err != nil {
			p.chErr <- err
			return
		}
	}
}

func (p *player) SetUnderrunCallback(f func()) {
	//TODO
}

func (p *player) TryWrite(data []byte) (int, error) {
	n := min(len(data), p.bufferSize-len(p.tmp))
	p.tmp = append(p.tmp, data[:n]...)

	if len(p.tmp) < p.bufferSize {
		return n, nil
	}

	select {
	case p.chBuffer <- p.tmp:
	case err := <-p.chErr:
		return 0, err
	}

	p.tmp = nil
	return n, nil
}

func (p *player) Close() error {
	if p.audioTrack == nil {
		return nil
	}

	runtime.SetFinalizer(p, nil)
	err := jni.RunOnJVM(func(vm, env, ctx uintptr) error {
		if msg := C.releaseAudioTrack(C.uintptr_t(vm), C.uintptr_t(env),
			p.audioTrack); msg != nil {
			return errors.New("oto: release failed: " + C.GoString(msg))
		}
		return nil
	})

	p.audioTrack = nil
	return err
}
