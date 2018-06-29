package gb

import (
	"log"

	"github.com/Humpheh/goboy/bits"
)

// Memory stores gameboy ROM, RAM and cartridge data. It manages the
// banking of these data banks.
type Memory struct {
	gb   *Gameboy
	Cart *Cartridge
	Data [0x10000]byte
	// VRAM bank 1-2 data
	VRAM [0x4000]byte
	// Index of the current VRAM bank
	VRAMBank byte
	// WRAM bank 1-7 data
	WRAM1 [0x7000]byte
	// Index of the current WRAM bank
	WRAM1Bank byte

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
		if x&2 == 0 {
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

	mem.WRAM1Bank = 1
}

// LoadCart load a cart rom into memory.
func (mem *Memory) LoadCart(loc string) (bool, error) {
	mem.Cart = &Cartridge{}
	return mem.Cart.Load(loc)
}

// Write a value at an address to the relevant location based on the
// current state of the gameboy. This handles banking and side effects
// of writing to certain addresses.
func (mem *Memory) Write(address uint16, value byte) {
	switch {
	case address >= 0xFF10 && address <= 0xFF26:
		mem.gb.Sound.Write(address, value)

	case address >= 0xFF30 && address <= 0xFF3F:
		// Writing to channel 3 waveform RAM.
		mem.Data[address] = value
		soundIndex := (address - 0xFF30) * 2
		mem.gb.Sound.waveformRam[soundIndex] = int8((value >> 4) & 0xF)
		mem.gb.Sound.waveformRam[soundIndex+1] = int8(value & 0xF)

	case address == TAC:
		//log.Printf("TAC: %0#2v", value)
		// Timer control
		currentFreq := mem.gb.getClockFreq()
		mem.Data[TAC] = value
		newFreq := mem.gb.getClockFreq()

		if currentFreq != newFreq {
			mem.gb.setClockFreq()
		}

	case address == TIMA:
		//log.Printf("TIMA WRITE %0#2v", value)
		//mem.gb.setClockFreq()
		mem.Data[TIMA] = value

	case address == 0xFF02:
		// Serial transfer control
		if value == 0x81 {
			f := mem.gb.options.transferFunction
			if f != nil {
				f(mem.Read(0xFF01))
			}
		}

	case address == 0xFF04:
		// Trap divider register
		mem.gb.setClockFreq()
		mem.gb.CPU.Divider = 0
		mem.Data[0xFF04] = 0

	case address == 0xFF44:
		// Trap scanline register
		mem.Data[0xFF44] = 0

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
			mem.WRAM1Bank = value & 0x7
			if mem.WRAM1Bank == 0 {
				mem.WRAM1Bank = 1
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

	case address < 0x8000:
		// Write to the cartridge ROM (banking)
		mem.Cart.Write(address, value)

	case address >= 0xA000 && address < 0xC000:
		// Write to the cartridge ram
		mem.Cart.WriteRAM(address, value)

	case address >= 0xE000 && address < 0xFE00:
		// Echo RAM
		mem.Data[address] = value
		mem.Write(address-0x2000, value)

	case address >= 0xFEA0 && address < 0xFEFF:
		// Restricted RAM
		return

	case address >= 0x8000 && address < 0xA000:
		// VRAM Banking
		bankOffset := uint16(mem.VRAMBank) * 0x2000
		mem.VRAM[address-0x8000+bankOffset] = value

	case address >= 0xD000 && address < 0xE000:
		// WRAM Bank 1-7
		bankOffset := uint16(mem.WRAM1Bank-1) * 0x1000
		mem.WRAM1[address-0xD000+bankOffset] = value

	default:
		mem.Data[address] = value
	}
}

// Read from memory. Will go and read from cartridge memory if the
// requested address is mapped to that space.
func (mem *Memory) Read(address uint16) byte {
	switch {
	// Joypad address
	case address == 0xFF00:
		return mem.gb.joypadValue(mem.Data[0xFF00])

	case address == 0xFF0F:
		return mem.Data[0xFF0F] | 0xE0

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
		return mem.WRAM1Bank

	case address <= 0x7FFF || address >= 0xA000 && address <= 0xBFFF:
		return mem.Cart.Read(address)

	case address >= 0x8000 && address < 0xA000:
		// VRAM Banking
		bankOffset := uint16(mem.VRAMBank) * 0x2000
		return mem.VRAM[address-0x8000+bankOffset]

	case address >= 0xD000 && address < 0xE000:
		// WRAM Bank 1-7
		bankOffset := uint16(mem.WRAM1Bank-1) * 0x1000
		return mem.WRAM1[address-0xD000+bankOffset]

		// Else return memory
	default:
		return mem.Data[address]
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
		mem.Data[0xFF55] |= 0x80 // Set bit 7
		return
	}

	source := (uint16(mem.Data[0xFF51])<<8 | uint16(mem.Data[0xFF52])) & 0xFFF0
	destination := (uint16(mem.Data[0xFF53])<<8 | uint16(mem.Data[0xFF54])) & 0x1FF0

	length := ((uint16(value) & 0x7F) + 1) * 0x10

	dmaMode := value >> 7
	if dmaMode == 0 {
		// General purpose DMA
		var i uint16
		for i = 0; i < length; i++ {
			mem.VRAM[destination+i] = mem.Read(source + i)
		}
		mem.Data[0xFF55] = 0xFF
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
		mem.VRAM[mem.hbDMADestination] = mem.Read(mem.hbDMASource)
		mem.hbDMADestination++
		mem.hbDMASource++
	}
	if mem.hbDMALength > 0 {
		mem.hbDMALength--
		mem.Data[0xFF55] = mem.hbDMALength
	} else {
		// DMA has finished
		mem.Data[0xFF55] = 0xFF
		mem.hbDMAActive = false
	}
}
