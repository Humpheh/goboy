package cart

func NewMBC3(data []byte) BankingController {
	return &MBC3{
		rom:        data,
		romBank:    1,
		ram:        make([]byte, 0x8000),
		rtc:        make([]byte, 0xD),
		latchedRtc: make([]byte, 0xD),
	}
}

// ROM is a basic Gameboy cartridge.
type MBC3 struct {
	rom     []byte
	romBank uint32

	ram        []byte
	ramBank    uint32
	ramEnabled bool

	rtc         []byte
	latchedRtc  []byte
	rtcRegister byte
	latched     bool
}

// Read returns a value at a memory address in the ROM.
func (r *MBC3) Read(address uint16) byte {
	switch {
	case address < 0x4000:
		return r.rom[address] // Bank 0 is fixed
	case address < 0x8000:
		return r.rom[uint32(address-0x4000)+(r.romBank*0x4000)] // Use selected rom bank
	default:
		if r.rtcRegister == 0 {
			return r.ram[(0x2000*r.ramBank)+uint32(address-0xA000)] // Use selected ram bank
		}
		if r.latched {
			return r.latchedRtc[r.rtcRegister]
		}
		return r.rtc[r.rtcRegister]
	}
}

// Write is not supported on a ROM cart.
func (r *MBC3) WriteROM(address uint16, value byte) {
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
		r.romBank = uint32(value & 0x7F)
		if r.romBank == 0x00 {
			r.romBank++
		}
	case address < 0x6000:
		// ROM/RAM banking
		switch {
		case value < 0x4:
			r.ramBank = uint32(value & 0x3)
			r.rtcRegister = 0
		case value < 0xD:
			r.rtcRegister = (value & 0xC) >> 2
			r.ramBank = 0
		}
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
		if r.rtcRegister == 0 {
			r.ram[(0x2000*r.ramBank)+uint32(address-0xA000)] = value
		} else {
			r.rtc[r.rtcRegister] = value
		}
	}
	// TODO: what do if disabled
}
