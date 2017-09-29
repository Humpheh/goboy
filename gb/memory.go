package gb

import (
	"github.com/humpheh/goboy/bits"
	"log"
)

type Memory struct {
	GB   *Gameboy
	Cart *Cartridge
	Data [0x10000]byte

	EnableRAM  bool
	ROMBanking bool
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
		if x & 2 == 0 {
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
		mem.Data[address] = value
		sound_index := (address - 0xFF30) * 2
		mem.GB.Sound.WaveformRam[sound_index] = int8((value >> 4) & 0xF)
		mem.GB.Sound.WaveformRam[sound_index + 1] = int8(value & 0xF)

	// Timer control
	case address == TMC:
		current_freq := mem.GB.GetClockFreq()
		mem.Data[TMC] = value
		new_freq := mem.GB.GetClockFreq()

		if current_freq != new_freq {
			mem.GB.SetClockFreq()
		}

	// Serial transfer control
	case address == 0xFF02:
		if value == 0x81 {
			f := mem.GB.TransferFunction
			if f != nil {
				f(mem.Read(0xFF01))
			}
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

	case address == 0xFF4D:
		log.Print("WARN: 0xFF4D WRITTEN", value)

	// ROM
	case address < 0x8000:
		mem.HandleBanking(address, value)

	case address >= 0xA000 && address < 0xC000:
		if mem.EnableRAM {
			new_address := address - 0xA000
			mem.Cart.RAM[new_address+mem.Cart.RAMBank*0x2000] = value
		}

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

func (mem *Memory) Read(address uint16) byte {
	switch {
	// Joypad address
	case address == 0xFF00:
		return mem.GB.JoypadValue(mem.Data[0xFF00])

	case address == 0xFF0F:
		return mem.Data[0xFF0F] | 0xE0

	case address < 0x4000:
		return mem.Cart.Data[address]

	// Reading from ROM memory bank
	case address >= 0x4000 && address <= 0x7FFF:
		new_address := uint32(address) - 0x4000
		return mem.Cart.Data[new_address+(uint32(mem.Cart.ROMBank)*0x4000)]

	// Reading from RAM memory bank
	case address >= 0xA000 && address <= 0xBFFF:
		new_address := address - 0xA000
		return mem.Cart.RAM[new_address+(mem.Cart.RAMBank*0x2000)]

	// Else return memory
	default:
		return mem.Data[address]
	}
}

func (mem *Memory) HandleBanking(address uint16, value byte) {
	// MBC3 banking TODO: merge this into main banking switch
	if mem.Cart.Type == MBC3 {
		switch {
		// Enable RAM bank
		case address < 0x2000:
			mem.enableRAMBank(address, value)

		// Switch ROM bank
		case address < 0x4000:
			var lower byte = value & 127
			mem.Cart.ROMBank = uint16(lower)
			if mem.Cart.ROMBank == 0 {
				mem.Cart.ROMBank++
			}

		// Switch RAM bank
		case address < 0x6000:
			mem.Cart.RAMBank = uint16(value & 0x3)
		}
		return
	}
	/*
	4000-7FFF - ROM Bank 00-1FF (Read Only)
Same as for MBC1, except that accessing up to bank 1E0h is supported now. Also, bank 0 is actually bank 0.
	A000-BFFF - RAM Bank 00-0F, if any (Read/Write)
Same as for MBC1, except RAM sizes are 64kbit, 256kbit and 1mbit.

	 */
	if mem.Cart.Type == MBC5 {
		switch {
		// Enable RAM bank
		case address < 0x2000:
			mem.enableRAMBank(address, value)

		case address < 0x3000:
			mem.Cart.ROMBank = mem.Cart.ROMBank & 0x100 | uint16(value)

		case address < 0x4000:
			mem.Cart.ROMBank = mem.Cart.ROMBank & 0xFF | (uint16(value & 1) << 8)

		// Switch RAM bank
		case address < 0x6000:
			mem.Cart.RAMBank = uint16(value & 0xF)
		}
		return
	}

	switch {
	// Enable RAM
	case address < 0x2000:
		if mem.Cart.Type == MBC1 || mem.Cart.Type == MBC2 {
			mem.enableRAMBank(address, value)
		}

	// Change ROM bank
	case address >= 0x200 && address < 0x4000:
		if mem.Cart.Type == MBC1 || mem.Cart.Type == MBC2 {
			mem.changeLoROMBank(value, false)
		}

	// Change ROM or RAM
	case address >= 0x4000 && address < 0x6000:
		if mem.Cart.Type == MBC1 {
			if mem.ROMBanking {
				mem.changeHiROMBank(value)
			} else {
				mem.changeRAMBank(value)
			}
		}

	// Change if ROM/RAM banking
	case address >= 0x6000 && address < 0x8000:
		if mem.Cart.Type == MBC1 {
			mem.changeROMRAMMode(value)
		}
	}
}

func (mem *Memory) enableRAMBank(address uint16, value byte) {
	if mem.Cart.Type == MBC2 {
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

func (mem *Memory) changeLoROMBank(value byte, allowZero bool) {
	if mem.Cart.Type == MBC2 {
		mem.Cart.ROMBank = uint16(value & 0xF)
	} else {
		var lower byte = value & 31
		mem.Cart.ROMBank &= 224 // turn off the lower 5
		mem.Cart.ROMBank |= uint16(lower)
	}
	if mem.Cart.ROMBank == 0 && !allowZero {
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
	if value&0x1 == 0 {
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
		mem.Write(0xFE00+i, mem.Read(address+i))
	}
}
