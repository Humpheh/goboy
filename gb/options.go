package gb

// GameboyOption is an option for the Gameboy execution.
type GameboyOption func(o *gameboyOptions)

type gameboyOptions struct {
	sound   bool
	cgbMode bool
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
