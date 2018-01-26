# Beep

A small package that brings sound to any Go program (and real-time audio processing and other casual stuff)

```
$ go get github.com/faiface/beep
```

## A (very short) Tour of Beep

Let's get started! Open an audio file (let's ignore errors for now, never do that in production).

```go
f, _ := os.Open("song.wav")
```

Decode the file into a [Streamer](https://godoc.org/github.com/faiface/beep#Streamer).

```go
// import "github.com/faiface/beep/wav"
s, format, _ := wav.Decode(f)
```

Streamers are super important in Beep. Streamer is anything that can Stream audio (lazily) and
perhaps do other interesting things along the way. The streamer returned from `wav.Decode` streams
audio from the file.

Now, let's play the streamer. First, we need to initialize the speaker.

```go
// import "github.com/faiface/beep/speaker"
speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
```

We set the sample rate of the speaker to the sample rate of the file and we set the buffer size to
1/10s. Buffer size determines the stability and responsiveness of the playback. With smaller buffer,
you get more responsiveness and less latency. With bigger buffer, you get more stability and
reliability.

Finally, let's play!

```go
speaker.Play(s)
```

The streamer now starts playing, but in order to hear anything, we need to prevent our program from
exiting.

```go
select {} // for now
```

Now, this is kind of a hack. Let's fix it. To do that, we'll use `beep.Seq` function, which takes
some streamers and streams them one by one and we'll use `beep.Callback` function, which creates a
streamer, that does not stream any audio, but instead calls our own function. So, here's what we do.
First, we create a channel, which will signal the end of the playback.

```go
done := make(chan struct{})
```

Now, we'll change `speaker.Play(s)` into this.

```go
speaker.Play(beep.Seq(s, beep.Callback(func() {
        close(done)
})))
```

And finally, we'll replace the hacky `select {}` with a receive from the channel.

```go
<-done
```

And that's it!

Take a look at the [documentation](https://godoc.org/github.com/faiface/beep) for other interesting
things you can do with Beep, such as mixing, looping, audio effects, and other useful stuff.