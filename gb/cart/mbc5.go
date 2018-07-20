package cart

// NewMBC5 returns a new MBC5 memory controller.
func NewMBC5(data []byte) BankingController {
	return &MBC5{
		rom:     data,
		romBank: 1,
		ram:     make([]byte, 0x20000),
	}
}

// MBC5 is a GameBoy cartridge that supports rom and ram banking.
type MBC5 struct {
	rom     []byte
	romBank uint32

	ram        []byte
	ramBank    uint32
	ramEnabled bool
}

// Read returns a value at a memory address in the ROM.
func (r *MBC5) Read(address uint16) byte {
	switch {
	case address < 0x4000:
		return r.rom[address] // Bank 0 is fixed
	case address < 0x8000:
		return r.rom[uint32(address-0x4000)+(r.romBank*0x4000)] // Use selected rom bank
	default:
		return r.ram[(0x2000*r.ramBank)+uint32(address-0xA000)] // Use selected ram bank
	}
}

// WriteROM attempts to switch the ROM or RAM bank.
func (r *MBC5) WriteROM(address uint16, value byte) {
	switch {
	case address < 0x2000:
		// RAM enable
		if value&0xF == 0xA {
			r.ramEnabled = true
		} else if value&0xF == 0x0 {
			r.ramEnabled = false
		}
	case address < 0x3000:
		// ROM bank number
		r.romBank = (r.romBank & 0x100) | uint32(value)
	case address < 0x4000:
		// ROM/RAM banking
		r.romBank = (r.romBank & 0xFF) | uint32(value&0x01)<<8
	case address < 0x6000:
		r.ramBank = uint32(value & 0xF)
	}
}

// WriteRAM writes data to the ram if it is enabled.
func (r *MBC5) WriteRAM(address uint16, value byte) {
	if r.ramEnabled {
		r.ram[(0x2000*r.ramBank)+uint32(address-0xA000)] = value
	}
}

// GetSaveData returns the save data for this banking controller.
func (r *MBC5) GetSaveData() []byte {
	data := make([]byte, len(r.ram))
	copy(data, r.ram)
	return data
}

// LoadSaveData loads the save data into the cartridge.
func (r *MBC5) LoadSaveData(data []byte) {
	r.ram = data
}
