package gb

import (
	"log"

	"github.com/Humpheh/goboy/bits"
	"github.com/Humpheh/goboy/gb/cart"
)

// Memory stores gameboy ROM, RAM and cartridge data. It manages the
// banking of these data banks.
type Memory struct {
	gb      *Gameboy
	Cart    *cart.Cart
	HighRAM [0x100]byte
	// VRAM bank 1-2 data
	VRAM [0x4000]byte
	// Index of the current VRAM bank
	VRAMBank byte

	// WRAM bank 0-7 data
	WRAM [0x9000]byte
	// Index of the current WRAM bank
	WRAMBank byte

	OAM [0x100]byte

	// H-Blank DMA transfer variables
	hbDMADestination uint16
	hbDMASource      uint16
	hbDMALength      byte
	hbDMAActive      bool
	// TODO: Block bank change when active
}

// Init the gb memory to the post-boot values.
func (mem *Memory) Init(gameboy *Gameboy) {
	mem.gb = gameboy

	// Set the default values
	mem.HighRAM[0x04] = 0x1E
	mem.HighRAM[0x05] = 0x00
	mem.HighRAM[0x06] = 0x00
	mem.HighRAM[0x07] = 0xF8
	mem.HighRAM[0x0F] = 0xE1
	mem.HighRAM[0x10] = 0x80
	mem.HighRAM[0x11] = 0xBF
	mem.HighRAM[0x12] = 0xF3
	mem.HighRAM[0x14] = 0xBF
	mem.HighRAM[0x16] = 0x3F
	mem.HighRAM[0x17] = 0x00
	mem.HighRAM[0x19] = 0xBF
	mem.HighRAM[0x1A] = 0x7F
	mem.HighRAM[0x1B] = 0xFF
	mem.HighRAM[0x1C] = 0x9F
	mem.HighRAM[0x1E] = 0xBF
	mem.HighRAM[0x20] = 0xFF
	mem.HighRAM[0x21] = 0x00
	mem.HighRAM[0x22] = 0x00
	mem.HighRAM[0x23] = 0xBF
	mem.HighRAM[0x24] = 0x77
	mem.HighRAM[0x25] = 0xF3
	mem.HighRAM[0x26] = 0xF1

	// Sets wave ram to
	// 00 FF 00 FF  00 FF 00 FF  00 FF 00 FF  00 FF 00 FF
	for x := 0x30; x < 0x3F; x++ {
		if x&2 == 0 {
			mem.HighRAM[x] = 0x00
		} else {
			mem.HighRAM[x] = 0xFF
		}
	}
	mem.HighRAM[0x40] = 0x91
	mem.HighRAM[0x41] = 0x85
	mem.HighRAM[0x42] = 0x00
	mem.HighRAM[0x43] = 0x00
	mem.HighRAM[0x45] = 0x00
	mem.HighRAM[0x47] = 0xFC
	mem.HighRAM[0x48] = 0xFF
	mem.HighRAM[0x49] = 0xFF
	mem.HighRAM[0x4A] = 0x00
	mem.HighRAM[0x4B] = 0x00
	mem.HighRAM[0xFF] = 0x00

	mem.WRAMBank = 1
}

// LoadCart load a cart rom into memory.
func (mem *Memory) LoadCart(loc string) (bool, error) {
	var err error
	mem.Cart, err = cart.NewCart(loc)
	return mem.Cart.GetMode()&cart.DMG == 0, err
}

func (mem *Memory) WriteHighRam(address uint16, value byte) {
	switch {
	case address >= 0xFEA0 && address < 0xFEFF:
		// Restricted RAM
		return

	case address >= 0xFF10 && address <= 0xFF26:
		mem.HighRAM[address-0xFF00] = value // & soundMask[byte(address-0xFF10)]
		mem.gb.Sound.Write(address, value)

	case address >= 0xFF30 && address <= 0xFF3F:
		// Writing to channel 3 waveform RAM.
		mem.HighRAM[address-0xFF00] = value
		soundIndex := (address - 0xFF30) * 2
		mem.gb.Sound.waveformRam[soundIndex] = int8((value >> 4) & 0xF)
		mem.gb.Sound.waveformRam[soundIndex+1] = int8(value & 0xF)

	case address == DIV:
		// Trap divider register
		mem.gb.setClockFreq()
		mem.gb.CPU.Divider = 0
		mem.HighRAM[DIV-0xFF00] = 0

	case address == TIMA:
		mem.HighRAM[TIMA-0xFF00] = value

	case address == TMA:
		mem.HighRAM[TMA-0xFF00] = value

	case address == TAC:
		// Timer control
		currentFreq := mem.gb.getClockFreq()
		mem.HighRAM[TAC-0xFF00] = value | 0xF8
		newFreq := mem.gb.getClockFreq()

		if currentFreq != newFreq {
			mem.gb.setClockFreq()
		}

	case address == 0xFF02:
		// Serial transfer control
		if value == 0x81 {
			f := mem.gb.options.transferFunction
			if f != nil {
				f(mem.Read(0xFF01))
			}
		}

	case address == 0xFF41:
		mem.HighRAM[0x41] = value | 0x80

	case address == 0xFF44:
		// Trap scanline register
		mem.HighRAM[0x44] = 0

	case address == 0xFF46:
		// DMA transfer
		mem.doDMATransfer(value)

	case address == 0xFF55:
		mem.doHDMATransfer(value)

	case address == 0xFF4F:
		// VRAM bank (CGB only)
		if mem.gb.IsCGB() {
			mem.VRAMBank = value & 0x1
		}

	case address == 0xFF70:
		// WRAM1 bank (CGB mode)
		if mem.gb.IsCGB() {
			mem.WRAMBank = value & 0x7
			if mem.WRAMBank == 0 {
				mem.WRAMBank = 1
			}
		}

	case address == 0xFF68:
		// BG palette index
		if mem.gb.IsCGB() {
			mem.gb.BGPalette.updateIndex(value)
		}

	case address == 0xFF69:
		// BG Palette data
		if mem.gb.IsCGB() {
			mem.gb.BGPalette.write(value)
		}

	case address == 0xFF6A:
		// Sprite palette index
		if mem.gb.IsCGB() {
			mem.gb.SpritePalette.updateIndex(value)
		}

	case address == 0xFF6B:
		// Sprite Palette data
		if mem.gb.IsCGB() {
			mem.gb.SpritePalette.write(value)
		}

	case address == 0xFF4D:
		// CGB speed change
		if mem.gb.IsCGB() {
			log.Print("Change speed")
			mem.gb.prepareSpeed = bits.Test(value, 0)
		}

	case address >= 0xFF72 && address <= 0xFF77:
		log.Print("write to ", address)

	default:
		mem.HighRAM[address-0xFF00] = value
	}
}

// Write a value at an address to the relevant location based on the
// current state of the gameboy. This handles banking and side effects
// of writing to certain addresses.
func (mem *Memory) Write(address uint16, value byte) {
	switch {
	case address < 0x8000:
		// Write to the cartridge ROM (banking)
		mem.Cart.WriteROM(address, value)

	case address < 0xA000:
		// VRAM Banking
		bankOffset := uint16(mem.VRAMBank) * 0x2000
		mem.VRAM[address-0x8000+bankOffset] = value

	case address < 0xC000:
		// Cartridge ram
		mem.Cart.WriteRAM(address, value)

	case address < 0xD000:
		// Internal RAM - Bank 0
		mem.WRAM[address-0xC000] = value

	case address < 0xE000:
		// Internal RAM Bank 1-7
		mem.WRAM[(address-0xC000)+(uint16(mem.WRAMBank)*0x1000)] = value

	case address < 0xFE00:
		// Echo RAM
		// TODO: re-enable echo RAM?
		//mem.Data[address] = value
		//mem.Write(address-0x2000, value)

	case address < 0xFEA0:
		// Object Attribute Memory
		mem.OAM[address-0xFE00] = value

	case address < 0xFF00:
		// Unusable memory
		break

	default:
		// High RAM
		mem.WriteHighRam(address, value)
	}
}

func (mem *Memory) ReadHighRam(address uint16) byte {
	switch {
	// Joypad address
	case address == 0xFF00:
		return mem.gb.joypadValue(mem.HighRAM[0x00])

	case address == 0xFF0F:
		return mem.HighRAM[0x0F] | 0xE0

	case address >= 0xFF72 && address <= 0xFF77:
		log.Print("read from ", address)
		return 0

	case address == 0xFF68:
		// BG palette index
		if mem.gb.IsCGB() {
			return mem.gb.BGPalette.readIndex()
		}
		return 0

	case address == 0xFF69:
		// BG Palette data
		if mem.gb.IsCGB() {
			return mem.gb.BGPalette.read()
		}
		return 0

	case address == 0xFF6A:
		// Sprite palette index
		if mem.gb.IsCGB() {
			return mem.gb.SpritePalette.readIndex()
		}
		return 0

	case address == 0xFF6B:
		// Sprite Palette data
		if mem.gb.IsCGB() {
			return mem.gb.SpritePalette.read()
		}
		return 0

	case address == 0xFF4D:
		// Speed switch data
		return mem.gb.currentSpeed<<7 | bits.B(mem.gb.prepareSpeed)

	case address == 0xFF4F:
		return mem.VRAMBank

	case address == 0xFF70:
		return mem.WRAMBank

	default:
		return mem.HighRAM[address-0xFF00]
	}
}

// Read from memory. Will go and read from cartridge memory if the
// requested address is mapped to that space.
func (mem *Memory) Read(address uint16) byte {
	switch {
	case address < 0x8000:
		// Cartridge ROM
		return mem.Cart.Read(address)

	case address < 0xA000:
		// VRAM Banking
		// TODO: check this is correct
		bankOffset := uint16(mem.VRAMBank) * 0x2000
		return mem.VRAM[address-0x8000+bankOffset]

	case address < 0xC000:
		// Cartridge RAM
		return mem.Cart.Read(address)

	case address < 0xD000:
		// Internal RAM - Bank 0
		return mem.WRAM[address-0xC000]

	case address < 0xE000:
		// Internal RAM Bank 1-7
		return mem.WRAM[(address-0xC000)+(uint16(mem.WRAMBank)*0x1000)]

	case address < 0xFE00:
		// Echo RAM
		// TODO: re-enable echo RAM?
		//mem.Data[address] = value
		//mem.Write(address-0x2000, value)
		return 0xFF

	case address < 0xFEA0:
		// Object Attribute Memory
		return mem.OAM[address-0xFE00]

	case address < 0xFF00:
		// Unusable memory
		return 0xFF

	default:
		return mem.ReadHighRam(address)
	}
}

// Perform a DMA transfer.
func (mem *Memory) doDMATransfer(value byte) {
	// TODO: This may need to be done instead of CPU ticks
	address := uint16(value) << 8 // (data * 100)

	var i uint16
	for i = 0; i < 0xA0; i++ {
		// TODO: Check this doesn't prevent
		mem.Write(0xFE00+i, mem.Read(address+i))
	}
}

// Start a HDMA transfer.
func (mem *Memory) doHDMATransfer(value byte) {
	if mem.hbDMAActive && bits.Val(value, 7) == 0 {
		// Abort a HDMA transfer
		mem.hbDMAActive = false
		mem.HighRAM[0x55] |= 0x80 // Set bit 7
		return
	}

	source := (uint16(mem.HighRAM[0x51])<<8 | uint16(mem.HighRAM[0x52])) & 0xFFF0
	destination := (uint16(mem.HighRAM[0x53])<<8 | uint16(mem.HighRAM[0x54])) & 0x1FF0
	destination += 0x8000

	length := ((uint16(value) & 0x7F) + 1) * 0x10

	dmaMode := value >> 7
	if dmaMode == 0 {
		// General purpose DMA
		var i uint16
		for i = 0; i < length; i++ {
			mem.Write(destination+i, mem.Read(source+i))
		}
		mem.HighRAM[0x55] = 0xFF
	} else {
		// H-Blank DMA
		mem.hbDMADestination = destination
		mem.hbDMASource = source
		mem.hbDMALength = byte(value)
		mem.hbDMAActive = true
	}
}

// Perform a HDMA transfer during a HBlank period.
func (mem *Memory) hbHDMATransfer() {
	if !mem.hbDMAActive {
		return
	}
	var i uint16
	for i = 0; i < 0x10; i++ {
		mem.Write(mem.hbDMADestination, mem.Read(mem.hbDMASource))
		mem.hbDMADestination++
		mem.hbDMASource++
	}
	if mem.hbDMALength > 0 {
		mem.hbDMALength--
		mem.HighRAM[0x55] = mem.hbDMALength
	} else {
		// DMA has finished
		mem.HighRAM[0x55] = 0xFF
		mem.hbDMAActive = false
	}
}
