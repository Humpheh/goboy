package gb

import (
	"io/ioutil"
	"log"
	"strings"
	"time"

	"archive/zip"

	"github.com/Humpheh/goboy/bits"
	"github.com/Humpheh/goboy/gb/cart"
	"github.com/pkg/errors"
)

const (
	MBC1 = iota + 1
	MBC2
	MBC3
	MBC5
)

type CGBMode int

const (
	DMG CGBMode = iota + 1
	CGB
	DUAL
)

type Cart struct {
	cart.BankingController
	title string
	mode  CGBMode
}

func loadROMData(filename string) ([]byte, error) {
	var data []byte
	if strings.HasSuffix(filename, ".zip") {
		// Load the rom from a zip file
		reader, err := zip.OpenReader(filename)
		if err != nil {
			return nil, err
		}
		if len(reader.File) != 1 {
			return nil, errors.New("Zip must contain one file")
		}
		for _, f := range reader.File {
			fo, err := f.Open()
			if err != nil {
				return nil, err
			}
			data, err = ioutil.ReadAll(fo)
			if err != nil {
				return nil, err
			}
		}
	} else {
		// Load the file as a rom
		var err error
		data, err = ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

/*
	00h  ROM ONLY                 13h  MBC3+RAM+BATTERY
	01h  MBC1                     15h  MBC4
	02h  MBC1+RAM                 16h  MBC4+RAM
	03h  MBC1+RAM+BATTERY         17h  MBC4+RAM+BATTERY
	05h  MBC2                     19h  MBC5
	06h  MBC2+BATTERY             1Ah  MBC5+RAM
	08h  ROM+RAM                  1Bh  MBC5+RAM+BATTERY
	09h  ROM+RAM+BATTERY          1Ch  MBC5+RUMBLE
	0Bh  MMM01                    1Dh  MBC5+RUMBLE+RAM
	0Ch  MMM01+RAM                1Eh  MBC5+RUMBLE+RAM+BATTERY
	0Dh  MMM01+RAM+BATTERY        FCh  POCKET CAMERA
	0Fh  MBC3+TIMER+BATTERY       FDh  BANDAI TAMA5
	10h  MBC3+TIMER+RAM+BATTERY   FEh  HuC3
	11h  MBC3                     FFh  HuC1+RAM+BATTERY
	12h  MBC3+RAM
*/
func GetCart(filename string) (*Cart, error) {
	rom, err := loadROMData(filename)
	if err != nil {
		return nil, err
	}

	cartridge := Cart{
		title: GetRomName(rom),
	}

	// Check for GB mode
	switch rom[0x0143] {
	case 0x80:
		cartridge.mode = DUAL
	case 0xC0:
		cartridge.mode = CGB
	default:
		cartridge.mode = DMG
	}

	// Determine cartridge type
	mbcFlag := rom[0x147]
	log.Printf("Cart type: %#02x", mbcFlag)
	switch mbcFlag {
	case 0x00, 0x08, 0x09, 0x0B, 0x0C, 0x0D:
		log.Println("ROM/MMM01")
		cartridge.BankingController = cart.NewROM(rom)
	default:
		switch {
		case mbcFlag <= 0x03:
			cartridge.BankingController = cart.NewMBC1(rom)
			log.Println("MBC1")
		case mbcFlag <= 0x06:
			cartridge.BankingController = cart.NewMBC2(rom)
			log.Println("MBC2")
		case mbcFlag <= 0x13:
			cartridge.BankingController = cart.NewMBC3(rom)
			log.Println("MBC3")
		//case mbcFlag < 0x17:
		//	log.Println("Warning: MBC4 carts are not supported.")
		//case mbcFlag < 0x1F:
		//	cart.Type = MBC5
		//	log.Println("MBC5")
		default:
			log.Printf("Warning: This cart may not be supported: %02x", mbcFlag)
		}
	}

	//switch mbcFlag {
	//case 0x3, 0x6, 0x9, 0xD, 0xF, 0x10, 0x13, 0x17, 0x1B, 0x1E:
	//	cart.initGameSaves()
	//}
	return &cartridge, nil
}

// Get the name of the game from the rom file.
// The name is stored as a series of bytes between [0x134,0x142).
func GetRomName(data []byte) string {
	cartName := string(data[0x0134:0x0142])
	return strings.Replace(cartName, "\x00", "", -1)
}

type Cartridge struct {
	// Cartridge ROM data.
	ROM []byte
	// Cartridge RAM data.
	RAM []byte
	// Type of cartridge.
	Type int
	// Current active ROM bank.
	ROMBank uint16
	// Current active RAM bank.
	RAMBank uint16
	// The name of the cartridge from 0x0134-0x0142.
	Name string

	// The filename of the cartridge was loaded from.
	filename         string
	enableRAM        bool
	enableROMBanking bool
	writtenRAM       bool
}

// Returns the name of the file which should contain the save data.
func (cart *Cartridge) GetSaveFilename() string {
	return cart.filename + ".sav"
}

// Load a bin file as a cartridge. Returns a boolean if the cart
// has a cgb mode.
func (cart *Cartridge) Load(filename string) (bool, error) {
	// Load the file into ROM
	var err error
	cart.filename = filename
	cart.ROM, err = cart.loadROMData(filename)
	if err != nil {
		return false, err
	}

	// Get the cart name from the rom
	cart.Name = GetRomName(cart.ROM)

	// RAM banking
	cart.RAMBank = 0
	cart.RAM = make([]byte, 0x8000)

	// ROM banking
	mbcFlag := cart.ROM[0x147]

	/*
		00h  ROM ONLY                 13h  MBC3+RAM+BATTERY
		01h  MBC1                     15h  MBC4
		02h  MBC1+RAM                 16h  MBC4+RAM
		03h  MBC1+RAM+BATTERY         17h  MBC4+RAM+BATTERY
		05h  MBC2                     19h  MBC5
		06h  MBC2+BATTERY             1Ah  MBC5+RAM
		08h  ROM+RAM                  1Bh  MBC5+RAM+BATTERY
		09h  ROM+RAM+BATTERY          1Ch  MBC5+RUMBLE
		0Bh  MMM01                    1Dh  MBC5+RUMBLE+RAM
		0Ch  MMM01+RAM                1Eh  MBC5+RUMBLE+RAM+BATTERY
		0Dh  MMM01+RAM+BATTERY        FCh  POCKET CAMERA
		0Fh  MBC3+TIMER+BATTERY       FDh  BANDAI TAMA5
		10h  MBC3+TIMER+RAM+BATTERY   FEh  HuC3
		11h  MBC3                     FFh  HuC1+RAM+BATTERY
		12h  MBC3+RAM
	*/

	log.Printf("Cart type: %#02x", mbcFlag)

	// Check for GB mode
	hasCGB := false
	switch cart.ROM[0x0143] {
	case 0x80:
		log.Print("DMG/CGB mode")
		hasCGB = true
	case 0xC0:
		log.Print("CGB only mode")
		hasCGB = true
	default:
		log.Print("DMG mode")
	}

	// Determine cartridge type
	switch mbcFlag {
	case 0x00, 0x08, 0x09, 0x0B, 0x0C, 0x0D:
		log.Println("ROM/MMM01")
	default:
		switch {
		case mbcFlag <= 0x03:
			cart.Type = MBC1
			log.Println("MBC1")
		case mbcFlag <= 0x06:
			cart.Type = MBC2
			log.Println("MBC2")
		case mbcFlag <= 0x13:
			cart.Type = MBC3
			log.Println("MBC3")
		case mbcFlag < 0x17:
			log.Println("Warning: MBC4 carts are not supported.")
		case mbcFlag < 0x1F:
			cart.Type = MBC5
			log.Println("MBC5")
		default:
			log.Printf("Warning: This cart may not be supported: %02x", mbcFlag)
		}
	}
	cart.ROMBank = 1

	switch mbcFlag {
	case 0x3, 0x6, 0x9, 0xD, 0xF, 0x10, 0x13, 0x17, 0x1B, 0x1E:
		cart.initGameSaves()
	}
	return hasCGB, nil
}

// Load the bytes from a ROM file. The file can also be a zip file containing a single
// rom file.
func (cart *Cartridge) loadROMData(filename string) ([]byte, error) {
	var data []byte
	if strings.HasSuffix(filename, ".zip") {
		// Load the rom from a zip file
		reader, err := zip.OpenReader(filename)
		if err != nil {
			return nil, err
		}
		if len(reader.File) != 1 {
			return nil, errors.New("Zip must contain one file")
		}
		for _, f := range reader.File {
			fo, err := f.Open()
			if err != nil {
				return nil, err
			}
			data, err = ioutil.ReadAll(fo)
			if err != nil {
				return nil, err
			}
		}
	} else {
		// Load the file as a rom
		var err error
		data, err = ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

// Attempt to load a save game from the expected location.
func (cart *Cartridge) initGameSaves() {
	saveData, err := ioutil.ReadFile(cart.GetSaveFilename())
	if err == nil {
		cart.RAM = saveData
	}
	// Write the RAM to file every second if it has changed
	ticker := time.NewTicker(time.Second)
	go func() {
		for range ticker.C {
			if cart.writtenRAM {
				cart.writtenRAM = false
				log.Println("saving cartridge RAM")
				cart.Save()
			}
		}
	}()
}

// Dump the carts RAM to the save location.
func (cart *Cartridge) Save() {
	err := ioutil.WriteFile(cart.GetSaveFilename(), cart.RAM, 0644)
	if err != nil {
		log.Printf("error in saving the game: %v", err)
	}
}

// Read data from the cartridge ROM/RAM.
func (cart *Cartridge) Read(address uint16) byte {
	switch {
	case address < 0x4000:
		return cart.ROM[address]

	case address >= 0x4000 && address <= 0x7FFF:
		// Reading from ROM memory bank
		newAddress := uint32(address) - 0x4000
		return cart.ROM[newAddress+(uint32(cart.ROMBank)*0x4000)]

	case address >= 0xA000 && address <= 0xBFFF:
		// Reading from RAM memory bank
		newAddress := address - 0xA000
		return cart.RAM[newAddress+(cart.RAMBank*0x2000)]
	}
	log.Printf("Trying to access address not in cartridge space: %v", address)
	return 0
}

// Handle writing to the cartridge at an address with a value.
func (cart *Cartridge) Write(address uint16, value byte) {
	switch cart.Type {
	case MBC1:
		cart.doMBC1(address, value)
	case MBC2:
		cart.doMBC2(address, value)
	case MBC3:
		cart.doMBC3(address, value)
	case MBC5:
		cart.doMBC5(address, value)
	}
	return
}

func (cart *Cartridge) doMBC1(address uint16, value byte) {
	switch {
	case address < 0x2000:
		// Enable RAM
		cart.enableRAMBank(address, value)

	case address >= 0x200 && address < 0x4000:
		// Change ROM bank
		cart.changeLoROMBank(value, false)

	case address >= 0x4000 && address < 0x6000:
		// Change ROM or RAM
		if cart.enableROMBanking {
			cart.changeHiROMBank(value)
		} else {
			cart.RAMBank = uint16(value & 0x3)
		}

	case address >= 0x6000 && address < 0x8000:
		// Change if ROM/RAM banking
		if value&0x1 == 0 {
			cart.enableROMBanking = true
			cart.RAMBank = 0
		} else {
			cart.enableROMBanking = false
		}
	}
}

func (cart *Cartridge) doMBC2(address uint16, value byte) {
	switch {
	case address < 0x2000:
		// Enable RAM
		cart.enableRAMBank(address, value)

	case address >= 0x200 && address < 0x4000:
		// Change ROM bank
		cart.changeLoROMBank(value, false)
	}
}

func (cart *Cartridge) doMBC3(address uint16, value byte) {
	switch {
	case address < 0x2000:
		// Enable RAM bank
		cart.enableRAMBank(address, value)

	case address < 0x4000:
		// Switch ROM bank
		var lower byte = value & 127
		cart.ROMBank = uint16(lower)
		if cart.ROMBank == 0 {
			cart.ROMBank++
		}

	case address < 0x6000:
		// Switch RAM bank
		cart.RAMBank = uint16(value & 0x3)
	}
}

func (cart *Cartridge) doMBC5(address uint16, value byte) {
	switch {
	case address < 0x2000:
		// Enable RAM bank
		cart.enableRAMBank(address, value)

	case address < 0x3000:
		// Switch ROM bank lower bits
		cart.ROMBank = cart.ROMBank&0x100 | uint16(value)

	case address < 0x4000:
		// Switch ROM bank upper bits
		cart.ROMBank = cart.ROMBank&0xFF | (uint16(value&1) << 8)

	case address < 0x6000:
		// Switch RAM bank
		cart.RAMBank = uint16(value & 0xF)
	}
}

// Attempt to write to the cartridges RAM. If it is not enabled this
// will have no function.
func (cart *Cartridge) WriteRAM(address uint16, value byte) {
	if cart.enableRAM {
		cart.writtenRAM = true
		newAddress := address - 0xA000
		cart.RAM[newAddress+cart.RAMBank*0x2000] = value
	}
}

// Enable RAM bank for MBC1 and MBC2.
func (cart *Cartridge) enableRAMBank(address uint16, value byte) {
	if cart.Type == MBC2 {
		if bits.Test(byte(address), 4) {
			return
		}
	}

	var test = value & 0xF
	if test == 0xA {
		cart.enableRAM = true
	} else if test == 0x0 {
		cart.enableRAM = false
	}
}

// Change the lower ROM bank for MBC1 and MBC2.
func (cart *Cartridge) changeLoROMBank(value byte, allowZero bool) {
	if cart.Type == MBC2 {
		cart.ROMBank = uint16(value & 0xF)
	} else {
		var lower = value & 31
		cart.ROMBank &= 224 // turn off the lower 5
		cart.ROMBank |= uint16(lower)
	}
	if cart.ROMBank == 0 && !allowZero {
		cart.ROMBank++
	}
}

// Change the higher ROM bank for MBC1.
func (cart *Cartridge) changeHiROMBank(value byte) {
	cart.ROMBank &= 31 // turn off upper 3 bits

	value &= 224 // turn off lower 5 bits of data
	cart.ROMBank |= uint16(value)

	if cart.ROMBank == 0 {
		cart.ROMBank++
	}
}
