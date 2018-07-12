package cart

import "log"

func NewROM(data []byte) BankingController {
	return &ROM{
		rom: data,
	}
}

// ROM is a basic Gameboy cartridge.
type ROM struct {
	rom []byte
}

// Read returns a value at a memory address in the ROM.
func (r *ROM) Read(address uint16) byte {
	return r.rom[address]
}

func (r *ROM) WriteROM(address uint16, value byte) {
	log.Print("at")
}

func (r *ROM) WriteRAM(address uint16, value byte) {
	log.Print("at")
}
