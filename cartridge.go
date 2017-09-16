package gob

import (
	"io/ioutil"
	"log"
	"time"
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
	cart.Name = string(name_bytes)

	cart.Data = data

	// RAM banking
	cart.RAMBank = 0
	cart.RAM = make([]byte, 0x8000)

	// ROM banking
	mbc_flag := cart.Data[0x147]
	switch mbc_flag {
	case 1, 2, 3:
		cart.MBC1 = true
	case 5, 6:
		cart.MBC2 = true
	case 0x12, 0x13:
		cart.MBC3 = true
	case 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E:
		cart.MBC5 = true
	}
	cart.ROMBank = 1

	switch mbc_flag {
	case 0x3, 0x6, 0x9, 0xD, 0x13, 0x1B, 0x1E:
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
