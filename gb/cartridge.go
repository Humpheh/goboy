package gb

import (
	"io/ioutil"
	"log"
	"time"
	"bytes"
)

type Cartridge struct {
	Data     []byte // max len 0x200000
	ROMBank  uint16
	MBC1     bool
	MBC2     bool
	MBC3     bool
	MBC5     bool
	RAM      []byte
	RAMBank  uint16
	Name     string
	filename string
}

func (cart *Cartridge) GetSaveFilename() string {
	return cart.filename + ".sav"
}

func (cart *Cartridge) Load(filename string) error {
	cart.filename = filename
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// Get the cart name from the rom
	name_bytes := data[0x0134:0x0142]
	cart.Name = string(bytes.Trim(name_bytes, "\x00"))

	cart.Data = data

	// RAM banking
	cart.RAMBank = 0
	cart.RAM = make([]byte, 0x8000)

	// ROM banking
	mbc_flag := cart.Data[0x147]

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

	log.Printf("Cart type: %#02x", mbc_flag)

	if cart.Data[0x0143] == 0x80 {
		log.Print("GBC Mode")
	}

	switch mbc_flag {
	case 0x00, 0x08, 0x09, 0x0B, 0x0C, 0x0D:
		log.Println("ROM/MMM01")
	default:
		switch {
		case mbc_flag <= 0x03:
			cart.MBC1 = true
			log.Println("MBC1")
		case mbc_flag <= 0x06:
			cart.MBC2 = true
			log.Println("MBC2")
		case mbc_flag <= 0x13:
			cart.MBC3 = true
			log.Println("MBC3")
		case mbc_flag < 0x17:
			log.Println("Warning: MBC4 carts are not supported.")
		case mbc_flag < 0x1E:
			cart.MBC5 = true
			log.Println("MBC5")
		default:
			log.Printf("Warning: This cart may not be supported: %02x", mbc_flag)
		}
	}
	cart.ROMBank = 1

	switch mbc_flag {
	case 0x3, 0x6, 0x9, 0xD, 0xF, 0x10, 0x13, 0x17, 0x1B, 0x1E:
		cart.initGameSaves()
	}

	return nil
}

func (cart *Cartridge) initGameSaves() {
	save_data, err := ioutil.ReadFile(cart.GetSaveFilename())
	if err == nil {
		cart.RAM = save_data
	}
	// Write the RAM to file every second
	ticker := time.NewTicker(time.Second)
	go func() {
		for range ticker.C {
			log.Println("saving")
			cart.Save()
		}
	}()
}

func (cart *Cartridge) Save() {
	err := ioutil.WriteFile(cart.GetSaveFilename(), cart.RAM, 0644)
	if err != nil {
		panic("Couldn't save?")
	}
}
