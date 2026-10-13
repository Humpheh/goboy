package gb

import (
	"log"

	"github.com/Humpheh/goboy/pkg/cart"
)

// memory stores gameboy ROM, RAM and cartridge data. It manages the
// banking of these data banks.
type memory struct {
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

	// CGB HDMA transfer variables
	hdmaLength byte
	hdmaActive bool

	// OAM DMA transfer state. A requested transfer starts once dmaStartDelay
	// has counted down to zero, and copies the byte at dmaIndex each M-cycle
	// while dmaActive is set.
	dmaPendingSource uint16
	dmaStartDelay    int
	dmaSource        uint16
	dmaIndex         uint16
	dmaActive        bool
}

// Init the gb memory to the post-boot values.
func (mem *memory) Init(gameboy *Gameboy) {
	mem.gb = gameboy // TODO: this is a bit dirty

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
func (mem *memory) LoadCart(loc string) (bool, error) {
	var err error
	mem.Cart, err = cart.NewCartFromFile(loc)
	if err != nil {
		return false, err
	}
	return mem.Cart.GetMode()&cart.CGB != 0, nil
}

// WriteHighRam writes to the range 0xFF00-0xFFFF in the memory address
// space. The range includes both HRAM and the hardware registers.
func (mem *memory) WriteHighRam(address uint16, value byte) {
	switch {
	case address >= 0xFEA0 && address < 0xFEFF:
		// Restricted RAM
		return

	case address >= 0xFF10 && address <= 0xFF26:
		mem.gb.sound.Write(address, value)

	case address >= 0xFF30 && address <= 0xFF3F:
		// Writing to channel 3 waveform RAM.
		mem.gb.sound.WriteWaveform(address, value)

	case address == 0xFF02:
		// Serial transfer control
		if value == 0x81 {
			f := mem.gb.options.transferFunction
			if f != nil {
				f(mem.ReadHighRam(0xFF01))
			}
		}

	case address == DIV:
		// Writing to DIV resets the whole internal counter
		mem.gb.setTimerCounter(0)

	case address == TIMA:
		mem.gb.writeTIMA(value)

	case address == TMA:
		mem.gb.writeTMA(value)

	case address == TAC:
		mem.gb.writeTAC(value)

	case address == LCDC:
		wasEnabled := mem.gb.isLCDEnabled()
		mem.HighRAM[0x40] = value
		if enabled := mem.gb.isLCDEnabled(); enabled && !wasEnabled {
			mem.gb.lcdOn()
		} else if !enabled && wasEnabled {
			mem.gb.lcdOff()
		}

	case address == 0xFF41:
		// The mode and LY=LYC flag are read only
		mem.HighRAM[0x41] = value&0x78 | mem.HighRAM[0x41]&0x07 | 0x80

	case address == 0xFF44:
		// LY is read only

	case address == 0xFF46:
		// OAM DMA transfer, which starts two M-cycles after the write. Any
		// transfer already running carries on until then.
		mem.HighRAM[0x46] = value
		mem.dmaPendingSource = uint16(value) << 8
		mem.dmaStartDelay = 2

	case address == 0xFF4D:
		// CGB speed change
		if mem.gb.IsCGB() {
			mem.gb.prepareSpeed = bitTest(value, 0)
		}

	case address == 0xFF4F:
		// VRAM bank (CGB only), blocked when HDMA is active
		if mem.gb.IsCGB() && !mem.hdmaActive {
			mem.VRAMBank = value & 0x1
		}

	case address == 0xFF55:
		// CGB DMA transfer
		if mem.gb.IsCGB() {
			mem.doNewDMATransfer(value)
		}

	case address == 0xFF68:
		// BG palette index
		if mem.gb.IsCGB() {
			mem.gb.bgPalette.updateIndex(value)
		}

	case address == 0xFF69:
		// BG Palette data
		if mem.gb.IsCGB() {
			mem.gb.bgPalette.write(value)
		}

	case address == 0xFF6A:
		// Sprite palette index
		if mem.gb.IsCGB() {
			mem.gb.spritePalette.updateIndex(value)
		}

	case address == 0xFF6B:
		// Sprite Palette data
		if mem.gb.IsCGB() {
			mem.gb.spritePalette.write(value)
		}

	case address == 0xFF70:
		// WRAM1 bank (CGB mode)
		if mem.gb.IsCGB() {
			mem.WRAMBank = value & 0x7
			if mem.WRAMBank == 0 {
				mem.WRAMBank = 1
			}
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
func (mem *memory) Write(address uint16, value byte) {
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
		mem.Write(address-0x2000, value)

	case address < 0xFEA0:
		// Object Attribute Memory, which can't be accessed during OAM DMA
		if !mem.dmaActive {
			mem.OAM[address-0xFE00] = value
		}

	case address < 0xFF00:
		// Unusable memory
		break

	default:
		// High RAM
		mem.WriteHighRam(address, value)
	}
}

// Read from memory. Will go and read from cartridge memory if the
// requested address is mapped to that space.
func (mem *memory) Read(address uint16) byte {
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
		return mem.Read(address - 0x2000)

	case address < 0xFEA0:
		// Object Attribute Memory, which can't be accessed during OAM DMA
		if mem.dmaActive {
			return 0xFF
		}
		return mem.OAM[address-0xFE00]

	case address < 0xFF00:
		// Unusable memory
		return 0xFF

	default:
		return mem.ReadHighRam(address)
	}
}

// ReadHighRam reads from 0xFF00-0xFFFF in the memory address space. The range
// includes both HRAM and the hardware registers.
func (mem *memory) ReadHighRam(address uint16) byte {
	switch {
	// Joypad address
	case address == 0xFF00:
		return mem.gb.joypadValue(mem.HighRAM[0x00])

	case address >= 0xFF10 && address <= 0xFF26:
		return mem.gb.sound.Read(address)

	case address >= 0xFF30 && address <= 0xFF3F:
		// Writing to channel 3 waveform RAM.
		return mem.gb.sound.Read(address)

	case address == 0xFF0F:
		return mem.HighRAM[0x0F] | 0xE0

	case address >= 0xFF72 && address <= 0xFF77:
		//log.Print("read from ", address)
		return 0

	case address == 0xFF68:
		// BG palette index
		if mem.gb.IsCGB() {
			return mem.gb.bgPalette.readIndex()
		}
		return 0

	case address == 0xFF69:
		// BG Palette data
		if mem.gb.IsCGB() {
			return mem.gb.bgPalette.read()
		}
		return 0

	case address == 0xFF6A:
		// Sprite palette index
		if mem.gb.IsCGB() {
			return mem.gb.spritePalette.readIndex()
		}
		return 0

	case address == 0xFF6B:
		// Sprite Palette data
		if mem.gb.IsCGB() {
			return mem.gb.spritePalette.read()
		}
		return 0

	case address == 0xFF4D:
		// Speed switch data
		return mem.gb.currentSpeed<<7 | boolToBitByte(mem.gb.prepareSpeed)

	case address == 0xFF4F:
		return mem.VRAMBank

	case address == 0xFF70:
		return mem.WRAMBank

	default:
		return mem.HighRAM[address-0xFF00]
	}
}

// stepDMACycles advances the OAM DMA transfer by a number of cycles.
func (mem *memory) stepDMACycles(cycles int) {
	for c := 0; c < cycles && (mem.dmaActive || mem.dmaStartDelay > 0); c += 4 {
		mem.stepDMA()
	}
}

// stepDMA advances the OAM DMA transfer by one M-cycle, starting a requested
// transfer once its delay has passed and copying one byte per cycle.
func (mem *memory) stepDMA() {
	// OAM stays inaccessible until the cycle after the last byte is copied
	if mem.dmaActive && mem.dmaIndex == 0xA0 {
		mem.dmaActive = false
	}
	if mem.dmaStartDelay > 0 {
		mem.dmaStartDelay--
		if mem.dmaStartDelay == 0 {
			mem.dmaActive = true
			mem.dmaSource = mem.dmaPendingSource
			mem.dmaIndex = 0
		}
	}
	if mem.dmaActive {
		mem.OAM[mem.dmaIndex] = mem.Read(mem.dmaSource + mem.dmaIndex)
		mem.dmaIndex++
	}
}

// Start a CGB DMA transfer.
func (mem *memory) doNewDMATransfer(value byte) {
	if mem.hdmaActive && bitGet(value, 7) == 0 {
		// Abort a HDMA transfer
		mem.hdmaActive = false
		mem.HighRAM[0x55] |= 0x80 // Set bit 7
		return
	}

	length := ((uint16(value) & 0x7F) + 1) * 0x10

	// The 7th bit is DMA mode
	if value>>7 == 0 {
		// Mode 0, general purpose DMA
		mem.performNewDMATransfer(length)
		mem.HighRAM[0x55] = 0xFF
	} else {
		// Mode 1, H-Blank DMA
		mem.hdmaLength = byte(value)
		mem.hdmaActive = true
	}
}

// Perform a HDMA transfer during a HBlank period.
func (mem *memory) doHDMATransfer() {
	if !mem.hdmaActive {
		return
	}

	mem.performNewDMATransfer(0x10)
	if mem.hdmaLength > 0 {
		mem.hdmaLength--
		mem.HighRAM[0x55] = mem.hdmaLength
	} else {
		// DMA has finished
		mem.HighRAM[0x55] = 0xFF
		mem.hdmaActive = false
	}
}

// Transfer a set amount of DMA data based on the current register values.
func (mem *memory) performNewDMATransfer(length uint16) {
	// Load the source and destination from RAM
	source := (uint16(mem.HighRAM[0x51])<<8 | uint16(mem.HighRAM[0x52])) & 0xFFF0
	destination := (uint16(mem.HighRAM[0x53])<<8 | uint16(mem.HighRAM[0x54])) & 0x1FF0
	destination += 0x8000

	// Transfer the data from the source to the destination
	for range length {
		mem.Write(destination, mem.Read(source))
		destination++
		source++
	}

	// Update the source and destination in RAM
	mem.HighRAM[0x51] = byte(source >> 8)
	mem.HighRAM[0x52] = byte(source & 0xFF)
	mem.HighRAM[0x53] = byte(destination >> 8)
	mem.HighRAM[0x54] = byte(destination & 0xF0)
}
