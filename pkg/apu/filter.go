package apu

// highPass is a DC-blocking filter, like the capacitor on the GameBoy's
// audio output. Without it silence sits at the bottom of the 8-bit range,
// so every gap in the audio stream (which oto plays as the midpoint) pops.
type highPass struct {
	lastIn, lastOut float64
}

// highPassCharge is the filter's pole; 0.995 puts the cutoff at about 35Hz
// at 44.1kHz.
const highPassCharge = 0.995

// filter removes the DC offset from in and returns it as an unsigned 8-bit
// sample centred on 128.
func (h *highPass) filter(in float64) byte {
	out := in - h.lastIn + highPassCharge*h.lastOut
	h.lastIn, h.lastOut = in, out
	return byte(max(0, min(255, 128+out)))
}
