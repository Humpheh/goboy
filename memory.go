package gob

import (
	"github.com/humpheh/gob/bits"
	"fmt"
)

type Memory struct {
	GB   *Gameboy
	Cart *Cartridge
	Data [0x10000]byte

	EnableRAM  bool
	ROMBanking bool
}

func (mem *Memory) LoadCart(loc string) {
	mem.Cart = &Cartridge{}
	mem.Cart.Load(loc)
}

func (mem *Memory) Write(address uint16, value byte) {
	//if address == 0xFF80 {
	//	mem.GB.HALT("0xFF80")
	//}

	switch {
	// Timer control
	case address == TMC:
		current_freq := mem.GB.GetClockFreq()
		mem.Data[TMC] = value
		new_freq := mem.GB.GetClockFreq()

		if current_freq != new_freq {
			mem.GB.SetClockFreq()
		}
	case address == 0xFF02:
		if value == 0x81 {
			fmt.Print(string(mem.Read(0xFF01)))
		}
	// Trap divider register
	case address == 0xFF04:
		mem.Data[0xFF04] = 0

	// Trap scanline register
	case address == 0xFF44:
		mem.Data[0xFF44] = 0

	// DMA transfer
	case address == 0xFF46:
		mem.DMATransfer(value)

	// ROM
	case address < 0x8000:
		mem.HandleBanking(address, value)

	case address >= 0xA000 && address < 0xC000:
		if mem.EnableRAM {
			new_address := address - 0xA000
			mem.Cart.RAM[new_address + mem.Cart.RAMBank * 0x2000] = value
		}

	// ECHO RAM
	case address >= 0xE000 && address < 0xFE00:
		mem.Data[address] = value
		mem.Write(address - 0x2000, value)

	// Restricted
	case address >= 0xFEA0 && address < 0xFEFF:
		return

	// Not restricted RAM
	default:
		mem.Data[address] = value
	}
	mem.Data[address] = value
}

func (mem *Memory) Read(address uint16) byte {
	switch {
	case address < 0x4000:
		return mem.Cart.Data[address]

	// Reading from ROM memory bank
	case address >= 0x4000 && address <= 0x7FFF:
		new_address := address - 0x4000
		return mem.Cart.Data[new_address + (mem.Cart.ROMBank * 0x4000)]

	// Reading from RAM memory bank
	case address >= 0xA000 && address <= 0xBFFF:
		new_address := address - 0xA000
		return mem.Cart.RAM[new_address + (mem.Cart.RAMBank * 0x2000)]

	// Else return memory
	default:
		return mem.Data[address]
	}
}

func (mem *Memory) HandleBanking(address uint16, value byte) {
	switch {
	// Enable RAM
	case address < 0x2000:
		if mem.Cart.MBC1 || mem.Cart.MBC2 {
			mem.enableRAMBank(address, value)
		}

	// Change ROM bank
	case address >= 0x200 && address < 0x4000:
		if mem.Cart.MBC1 || mem.Cart.MBC2 {
			mem.changeLoROMBank(value)
		}

	// Change ROM or RAM
	case address >= 0x4000 && address < 0x6000:
		if mem.Cart.MBC1 {
			if mem.ROMBanking {
				mem.changeHiROMBank(value)
			} else {
				mem.changeRAMBank(value)
			}
		}

	// Change if ROM/RAM banking
	case address >= 0x6000 && address < 0x8000:
		if mem.Cart.MBC1 {
			mem.changeROMRAMMode(value)
		}
	}
}

func (mem *Memory) enableRAMBank(address uint16, value byte) {
	if mem.Cart.MBC2 {
		if bits.Test(byte(address), 4) {
			return
		}
	}

	var test byte = value & 0xF
	if test == 0xA {
		mem.EnableRAM = true
	} else if test == 0x0 {
		mem.EnableRAM = false
	}
}

func (mem *Memory) changeLoROMBank(value byte) {
	if mem.Cart.MBC2 {
		mem.Cart.ROMBank = uint16(value & 0xF)
	} else {
		var lower byte = value & 31
		mem.Cart.ROMBank &= 224 // turn off the lower 5
		mem.Cart.ROMBank |= uint16(lower)
	}

	if mem.Cart.ROMBank == 0 {
		mem.Cart.ROMBank++
	}
}

func (mem *Memory) changeHiROMBank(value byte) {
	mem.Cart.ROMBank &= 31 // turn off upper 3 bits

	value &= 224 // turn off lower 5 bits of data
	mem.Cart.ROMBank |= uint16(value)

	if mem.Cart.ROMBank == 0 {
		mem.Cart.ROMBank++
	}
}

func (mem *Memory) changeRAMBank(value byte) {
	mem.Cart.RAMBank = uint16(value & 0x3)
}

func (mem *Memory) changeROMRAMMode(value byte) {
	if value & 0x1 == 0 {
		mem.ROMBanking = true
		mem.Cart.RAMBank = 0
	} else {
		mem.ROMBanking = false
	}
}

func (mem *Memory) DMATransfer(value byte) {
	// TODO: This may need to be done instead of CPU ticks
	address := uint16(value) << 8 // (data * 100)

	var i uint16
	for i = 0; i < 0xA0; i++ {
		// TODO: Check this doesn't prevent
		mem.Write(0xFE00 + i, mem.Read(address + i))
	}
}