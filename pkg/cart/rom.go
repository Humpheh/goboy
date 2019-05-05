package cart

// NewROM returns a new ROM cartridge.
func NewROM(data []byte) BankingController {
	return &ROM{
		rom: data,
	}
}

// ROM is a basic Gameboy cartridge that contains a fixed rom and no
// banking or RAM.
type ROM struct {
	rom []byte
}

// Read returns a value at a memory address in the ROM.
func (r *ROM) Read(address uint16) byte {
	return r.rom[address]
}

// WriteROM would switch between cartridge banks, however a ROM cart does
// not support banking.
func (r *ROM) WriteROM(address uint16, value byte) {}

// WriteRAM would write data to the cartridge RAM, however a ROM cart does
// not contain RAM so this is a noop.
func (r *ROM) WriteRAM(address uint16, value byte) {}

// GetSaveData returns the save data for this banking controller. As RAM is
// not supported on this memory controller, this is a noop.
func (r *ROM) GetSaveData() []byte {
	return []byte{}
}

// LoadSaveData loads the save data into the cartridge. As RAM is not supported
// on this memory controller, this is a noop.
func (r *ROM) LoadSaveData([]byte) {}
