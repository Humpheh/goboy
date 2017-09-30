package gb

type Memory struct {
	GB   *Gameboy
	Cart *Cartridge
	Data [0x10000]byte
}

// Init the gb2 memory to the post-boot values
func (mem *Memory) Init(gameboy *Gameboy) {
	mem.GB = gameboy

	// Set the default values
	mem.Data[0xFF05] = 0x00
	mem.Data[0xFF06] = 0x00
	mem.Data[0xFF07] = 0x00
	mem.Data[0xFF0F] = 0xE1
	mem.Data[0xFF10] = 0x80
	mem.Data[0xFF11] = 0xBF
	mem.Data[0xFF12] = 0xF3
	mem.Data[0xFF14] = 0xBF
	mem.Data[0xFF16] = 0x3F
	mem.Data[0xFF17] = 0x00
	mem.Data[0xFF19] = 0xBF
	mem.Data[0xFF1A] = 0x7F
	mem.Data[0xFF1B] = 0xFF
	mem.Data[0xFF1C] = 0x9F
	mem.Data[0xFF1E] = 0xBF
	mem.Data[0xFF20] = 0xFF
	mem.Data[0xFF21] = 0x00
	mem.Data[0xFF22] = 0x00
	mem.Data[0xFF23] = 0xBF
	mem.Data[0xFF24] = 0x77
	mem.Data[0xFF25] = 0xF3
	mem.Data[0xFF26] = 0xF1

	// Sets wave ram to
	// 00 FF 00 FF  00 FF 00 FF  00 FF 00 FF  00 FF 00 FF
	for x := 0xFF30; x < 0xFF3F; x++ {
		if x&2 == 0 {
			mem.Data[x] = 0x00
		} else {
			mem.Data[x] = 0xFF
		}
	}
	mem.Data[0xFF40] = 0x91
	mem.Data[0xFF41] = 0x85
	mem.Data[0xFF42] = 0x00
	mem.Data[0xFF43] = 0x00
	mem.Data[0xFF45] = 0x00
	mem.Data[0xFF47] = 0xFC
	mem.Data[0xFF48] = 0xFF
	mem.Data[0xFF49] = 0xFF
	mem.Data[0xFF4A] = 0x00
	mem.Data[0xFF4B] = 0x00
	mem.Data[0xFFFF] = 0x00
}

func (mem *Memory) LoadCart(loc string) error {
	mem.Cart = &Cartridge{}
	return mem.Cart.Load(loc)
}

func (mem *Memory) Write(address uint16, value byte) {
	switch {
	case address >= 0xFF10 && address <= 0xFF26:
		mem.GB.Sound.Write(address, value)

	case address >= 0xFF30 && address <= 0xFF3F:
		// Writing to channel 3 waveform RAM.
		mem.Data[address] = value
		soundIndex := (address - 0xFF30) * 2
		mem.GB.Sound.WaveformRam[soundIndex] = int8((value >> 4) & 0xF)
		mem.GB.Sound.WaveformRam[soundIndex+1] = int8(value & 0xF)

	case address == TMC:
		// Timer control
		currentFreq := mem.GB.GetClockFreq()
		mem.Data[TMC] = value
		newFreq := mem.GB.GetClockFreq()

		if currentFreq != newFreq {
			mem.GB.SetClockFreq()
		}

	case address == 0xFF02:
		// Serial transfer control
		if value == 0x81 {
			f := mem.GB.TransferFunction
			if f != nil {
				f(mem.Read(0xFF01))
			}
		}

	case address == 0xFF04:
		// Trap divider register
		mem.Data[0xFF04] = 0

	case address == 0xFF44:
		// Trap scanline register
		mem.Data[0xFF44] = 0

	case address == 0xFF46:
		// DMA transfer
		mem.DMATransfer(value)

	case address < 0x8000:
		// Write to the cartridge ROM (banking)
		mem.Cart.Write(address, value)

	case address >= 0xA000 && address < 0xC000:
		// Write to the cartridge ram
		mem.Cart.WriteRAM(address, value)

	// ECHO RAM
	case address >= 0xE000 && address < 0xFE00:
		mem.Data[address] = value
		mem.Write(address-0x2000, value)

	// Restricted
	case address >= 0xFEA0 && address < 0xFEFF:
		return

	// Not restricted RAM
	default:
		mem.Data[address] = value
	}
	mem.Data[address] = value
}

// Read from memory. Will go and read from cartridge memory if the
// requested address is mapped to that space.
func (mem *Memory) Read(address uint16) byte {
	switch {
	// Joypad address
	case address == 0xFF00:
		return mem.GB.JoypadValue(mem.Data[0xFF00])

	case address == 0xFF0F:
		return mem.Data[0xFF0F] | 0xE0

	case address <= 0x7FFF || address >= 0xA000 && address <= 0xBFFF:
		return mem.Cart.Read(address)

	// Else return memory
	default:
		return mem.Data[address]
	}
}

// Perform a DMA transfer.
func (mem *Memory) DMATransfer(value byte) {
	// TODO: This may need to be done instead of CPU ticks
	address := uint16(value) << 8 // (data * 100)

	var i uint16
	for i = 0; i < 0xA0; i++ {
		// TODO: Check this doesn't prevent
		mem.Write(0xFE00+i, mem.Read(address+i))
	}
}
