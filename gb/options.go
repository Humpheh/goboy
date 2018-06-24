package gb

// GameboyOption is an option for the Gameboy execution.
type GameboyOption func(o *gameboyOptions)

type gameboyOptions struct {
	sound   bool
	cgbMode bool

	// Callback when the serial port is written to
	transferFunction func(byte)
}

type DebugFlags struct {
	HideSprites    bool
	HideBackground bool
	OutputOpcodes  bool

	MuteChannel1 bool
	MuteChannel2 bool
	MuteChannel3 bool
	MuteChannel4 bool
}

// WithCGBEnabled runs the Gameboy with cgb mode enabled.
func WithCGBEnabled() GameboyOption {
	return func(o *gameboyOptions) {
		o.cgbMode = true
	}
}

// WithSound runs the Gameboy with sound output.
func WithSound() GameboyOption {
	return func(o *gameboyOptions) {
		o.sound = true
	}
}

// WithTransferFunction provides a function to callback on when the serial transfer
// address is written to.
func WithTransferFunction(transfer func(byte)) GameboyOption {
	return func(o *gameboyOptions) {
		o.transferFunction = transfer
	}
}
