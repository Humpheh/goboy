package cart

import (
	"archive/zip"
	"errors"
	"io/ioutil"
	"log"
	"log/slog"
	"strings"
	"time"
)

// Mode represents the types of mode the GameBoy can run in.
type Mode int

const (
	// DMG is the mode for the original GameBoy.
	DMG Mode = 1 << iota
	// CGB is the mode for the GameBoy Color.
	CGB
)

// BankingController provides methods for accessing and writing data to a
// cartridge, which provides different banking functionality depending on
// the implementation.
type BankingController interface {
	// Read returns the value from the cartridges ROM or RAM depending on
	// the banking.
	Read(address uint16) byte

	// WriteROM attempts to write a value to an address in ROM. This is
	// generally used for switching memory banks depending on the implementation.
	WriteROM(address uint16, value byte)

	// WriteRAM sets a value on an address in the internal cartridge RAM.
	// Like the ROM, this can be banked depending on the implementation
	// of the memory controller. Furthermore, if the cartridge supports
	// RAM+BATTERY, then this data can be saved between sessions.
	WriteRAM(address uint16, value byte)

	// GetSaveData returns the save data for this banking controller. In
	// general this will the contents of the RAM, however controllers may
	// choose to store this data in their own format.
	GetSaveData() []byte

	// LoadSaveData loads some save data into the cartridge. The banking
	// controller implementation can decide how this data should be loaded.
	LoadSaveData(data []byte)
}

// Cart represents a GameBoy cartridge.
//
// The cartridge is an extension of a banking controller which determines how the cart
// reacts with memory banking. The banking controller provides methods for reading and
// writing data to the cartridge, along with extra functionality such as RTC (real
// time clock).
type Cart struct {
	BankingController
	title    string
	filename string
	mode     Mode
}

// GetName returns the name of the cartridge. This is retrieved from the memory location
// [0x134,0x142) on the cartridge. The function will cache the result of the read from
// the cartridge.
func (c *Cart) GetName() string {
	if c.title == "" {
		// We have not loaded the ROM name yet, so go get it
		for i := uint16(0x134); i < 0x142; i++ {
			chr := c.Read(i)
			if chr != 0x00 {
				c.title += string(chr)
			}
		}
		c.title = strings.TrimSpace(c.title)
	}
	return c.title
}

// GetSaveFilename returns the name of the file that the game should be saved to. This is
// used for saving and loading save data to the cartridge.
// TODO: do something better here
func (c *Cart) GetSaveFilename() string {
	return c.filename + ".sav"
}

// GetMode returns the modes that this cart can run in.
func (c *Cart) GetMode() Mode {
	return c.mode
}

// Attempt to load a save game from the expected location.
func (c *Cart) initGameSaves() {
	saveData, err := ioutil.ReadFile(c.GetSaveFilename())
	if err == nil {
		c.LoadSaveData(saveData)
	}
	// Write the RAM to file every second
	// TODO: improve this behaviour
	ticker := time.NewTicker(time.Second)
	go func() {
		for range ticker.C {
			c.Save()
		}
	}()
}

// Save dumps the carts RAM to the save location.
func (c *Cart) Save() {
	data := c.BankingController.GetSaveData()
	if len(data) > 0 {
		err := ioutil.WriteFile(c.GetSaveFilename(), data, 0644)
		if err != nil {
			log.Printf("Error saving cartridge RAM: %v", err)
		}
	}
}

// NewCartFromFile loads a cartridge ROM from a file.
func NewCartFromFile(filename string) (*Cart, error) {
	rom, err := loadROMData(filename)
	if err != nil {
		return nil, err
	}
	return NewCart(rom, filename), nil
}

// NewCart loads a cartridge ROM from a byte array and returns a new cartridge with
// the correct memory banking controller. If the game supports saves, then the
// save file for the cartridge will also be loaded, and the saving loop will be
// started to write the save data back to file.
//
// The function will use the following list to determine which MBC to use. Not
// all of the controllers are supported, and the function will only start the
// save loop for controllers which support RAM+BATTERY.
//
//     0x00  ROM ONLY
//     0x01  MBC1
//     0x02  MBC1+RAM
//     0x03  MBC1+RAM+BATTERY
//     0x05  MBC2
//     0x06  MBC2+BATTERY
//     0x08  ROM+RAM
//     0x09  ROM+RAM+BATTERY
//     0x0B  MMM01
//     0x0C  MMM01+RAM
//     0x0D  MMM01+RAM+BATTERY
//     0x0F  MBC3+TIMER+BATTERY
//     0x10  MBC3+TIMER+RAM+BATTERY
//     0x11  MBC3
//     0x12  MBC3+RAM
//     0x13  MBC3+RAM+BATTERY
//     0x15  MBC4
//     0x16  MBC4+RAM
//     0x17  MBC4+RAM+BATTERY
//     0x19  MBC5
//     0x1A  MBC5+RAM
//     0x1B  MBC5+RAM+BATTERY
//     0x1C  MBC5+RUMBLE
//     0x1D  MBC5+RUMBLE+RAM
//     0x1E  MBC5+RUMBLE+RAM+BATTERY
//     0xFC  POCKET CAMERA
//     0xFD  BANDAI TAMA5
//     0xFE  HuC3
//     0xFF  HuC1+RAM+BATTERY
func NewCart(rom []byte, filename string) *Cart {
	cartridge := Cart{
		filename: filename,
	}

	// Check for GB mode
	switch rom[0x0143] {
	case 0x80:
		cartridge.mode = DMG | CGB
	case 0xC0:
		cartridge.mode = CGB
	default:
		cartridge.mode = DMG
	}

	// Determine cartridge type
	mbcFlag := rom[0x147]
	cartType := "Unknown"
	switch mbcFlag {
	case 0x00, 0x08, 0x09, 0x0B, 0x0C, 0x0D:
		cartType = "ROM"
		cartridge.BankingController = NewROM(rom)
	default:
		switch {
		case mbcFlag <= 0x03:
			cartridge.BankingController = NewMBC1(rom)
			cartType = "MBC1"
		case mbcFlag <= 0x06:
			cartridge.BankingController = NewMBC2(rom)
			cartType = "MBC2"
		case mbcFlag <= 0x13:
			cartridge.BankingController = NewMBC3(rom)
			cartType = "MBC3"
		case mbcFlag < 0x17:
			log.Println("Warning: MBC4 carts are not supported.")
			cartridge.BankingController = NewMBC1(rom)
			cartType = "MBC4"
		case mbcFlag < 0x1F:
			cartridge.BankingController = NewMBC5(rom)
			cartType = "MBC5"
		default:
			log.Printf("Warning: This cart may not be supported: %02x", mbcFlag)
			cartridge.BankingController = NewMBC1(rom)
		}
	}

	slog.Debug("Loaded ROM type", slog.String("type", cartType), slog.Int("mbcFlag", int(mbcFlag)))

	switch mbcFlag {
	case 0x3, 0x6, 0x9, 0xD, 0xF, 0x10, 0x13, 0x17, 0x1B, 0x1E, 0xFF:
		cartridge.initGameSaves()
	}
	return &cartridge
}

// Open the file and load the data out of it as an array of bytes. If the file is
// a zip file containing one file, then open that as the rom instead.
func loadROMData(filename string) ([]byte, error) {
	var data []byte
	if strings.HasSuffix(filename, ".zip") {
		return loadZIPData(filename)
	}
	// Load the file as a rom
	var err error
	data, err = ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Load a zip file with a single rom in it.
func loadZIPData(filename string) ([]byte, error) {
	// Load the rom from a zip file
	reader, err := zip.OpenReader(filename)
	if err != nil {
		return nil, err
	}
	if len(reader.File) != 1 {
		return nil, errors.New("zip must contain one file")
	}
	f := reader.File[0]
	fo, err := f.Open()
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(fo)
}
