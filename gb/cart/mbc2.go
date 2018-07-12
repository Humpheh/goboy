package cart

func NewMBC2(data []byte) BankingController {
	return &MBC2{
		rom:     data,
		romBank: 1,
		ram:     make([]byte, 0x2000),
	}
}

// ROM is a basic Gameboy cartridge.
type MBC2 struct {
	rom     []byte
	romBank uint32

	ram        []byte
	ramEnabled bool
}

// Read returns a value at a memory address in the ROM.
func (r *MBC2) Read(address uint16) byte {
	switch {
	case address < 0x4000:
		return r.rom[address] // Bank 0 is fixed
	case address < 0x8000:
		return r.rom[uint32(address-0x4000)+(r.romBank*0x4000)] // Use selected rom bank
	default:
		return r.ram[address-0xA000] // Use ram
	}
}

// Write is not supported on a ROM cart.
func (r *MBC2) WriteROM(address uint16, value byte) {
	switch {
	case address < 0x2000:
		// RAM enable
		if address&0x100 == 0 {
			if value&0xF == 0xA {
				r.ramEnabled = true
			} else if value&0xF == 0x0 {
				r.ramEnabled = false
			}
		}
		return
	case address < 0x4000:
		// ROM bank number (lower 4)
		if address&0x100 == 0x100 {
			r.romBank = uint32(value & 0xF)
			if r.romBank == 0x00 || r.romBank == 0x20 || r.romBank == 0x40 || r.romBank == 0x60 {
				r.romBank++
			}
		}
		return
	}
}

func (r *MBC2) WriteRAM(address uint16, value byte) {
	if r.ramEnabled {
		r.ram[address-0xA000] = value & 0xF
	}
	// TODO: what do if disabled
}
