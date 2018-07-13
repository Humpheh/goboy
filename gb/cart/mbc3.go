package cart

func NewMBC3(data []byte) BankingController {
	return &MBC3{
		rom:        data,
		romBank:    1,
		ram:        make([]byte, 0x8000),
		rtc:        make([]byte, 0x10),
		latchedRtc: make([]byte, 0x10),
	}
}

// ROM is a basic Gameboy cartridge.
type MBC3 struct {
	rom     []byte
	romBank uint32

	ram        []byte
	ramBank    uint32
	ramEnabled bool

	rtc        []byte
	latchedRtc []byte
	latched    bool
}

// Read returns a value at a memory address in the ROM.
func (r *MBC3) Read(address uint16) byte {
	switch {
	case address < 0x4000:
		return r.rom[address] // Bank 0 is fixed
	case address < 0x8000:
		return r.rom[uint32(address-0x4000)+(r.romBank*0x4000)] // Use selected rom bank
	default:
		if r.ramBank >= 0x4 {
			if r.latched {
				return r.latchedRtc[r.ramBank]
			}
			return r.rtc[r.ramBank]
		}
		return r.ram[(0x2000*r.ramBank)+uint32(address-0xA000)] // Use selected ram bank
	}
}

//var donebanks = map[uint32]bool{}

// Write is not supported on a ROM cart.
func (r *MBC3) WriteROM(address uint16, value byte) {
	switch {
	case address < 0x2000:
		// RAM enable
		r.ramEnabled = (value & 0xA) != 0
	case address < 0x4000:
		// ROM bank number (lower 5)
		r.romBank = uint32(value & 0x7F)
		if r.romBank == 0x00 {
			r.romBank++
		}
		//if !donebanks[r.romBank] {
		//	donebanks[r.romBank] = true
		//log.Printf("New ROM Bank: %2x (%4x: %2x)", r.romBank, address, value)
		//}
	case address < 0x6000:
		r.ramBank = uint32(value)
	case address < 0x8000:
		if value == 0x1 {
			r.latched = false
		} else if value == 0x0 {
			r.latched = true
			copy(r.rtc, r.latchedRtc)
		}
	}
}

func (r *MBC3) WriteRAM(address uint16, value byte) {
	if r.ramEnabled {
		if r.ramBank >= 0x4 {
			r.rtc[r.ramBank] = value
		} else {
			r.ram[(0x2000*r.ramBank)+uint32(address-0xA000)] = value
		}
	}
	// TODO: what do if disabled
}
