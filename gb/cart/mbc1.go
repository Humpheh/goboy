package cart

func NewMBC1(data []byte) BankingController {
	return &MBC1{
		rom:     data,
		romBank: 1,
		ram:     make([]byte, 0x8000),
	}
}

// ROM is a basic Gameboy cartridge.
type MBC1 struct {
	rom     []byte
	romBank uint32

	ram        []byte
	ramBank    uint32
	ramEnabled bool

	romBanking bool
}

// Read returns a value at a memory address in the ROM.
func (r *MBC1) Read(address uint16) byte {
	switch {
	case address < 0x4000:
		return r.rom[address] // Bank 0 is fixed
	case address < 0x8000:
		return r.rom[uint32(address-0x4000)+(r.romBank*0x4000)] // Use selected rom bank
	default:
		return r.ram[(0x2000*r.ramBank)+uint32(address-0xA000)] // Use selected ram bank
	}
}

// Write is not supported on a ROM cart.
func (r *MBC1) WriteROM(address uint16, value byte) {
	switch {
	case address < 0x2000:
		// RAM enable
		if value&0xF == 0xA {
			r.ramEnabled = true
		} else if value&0xF == 0x0 {
			r.ramEnabled = false
		}
	case address < 0x4000:
		// ROM bank number (lower 5)
		r.romBank = (r.romBank & 0xe0) | uint32(value&0x1f)
		if r.romBank == 0x00 || r.romBank == 0x20 || r.romBank == 0x40 || r.romBank == 0x60 {
			r.romBank++
		}
	case address < 0x6000:
		// ROM/RAM banking
		if r.romBanking {
			r.romBank = (r.romBank & 0x1F) | uint32(value&0xe0)
			if r.romBank == 0x00 || r.romBank == 0x20 || r.romBank == 0x40 || r.romBank == 0x60 {
				r.romBank++
			}
		} else {
			r.ramBank = uint32(value & 0x3)
		}
	case address < 0x8000:
		// ROM/RAM select mode
		r.romBanking = value&0x1 == 0x00
		if r.romBanking {
			r.ramBank = 0
		} else {
			r.romBank = r.romBank & 0x1F
		}
	}
}

func (r *MBC1) WriteRAM(address uint16, value byte) {
	if r.ramEnabled {
		r.ram[(0x2000*r.ramBank)+uint32(address-0xA000)] = value
	}
	// TODO: what do if disabled
}
