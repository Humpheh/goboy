package gob

import "io/ioutil"

type Cartridge struct {
	Data    []byte // max len 0x200000
	ROMBank uint16

	MBC1 bool
	MBC2 bool

	RAM     [0x8000]byte
	RAMBank uint16
}

func (cart *Cartridge) Load(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	cart.Data = data

	// ROM banking
	mbc_flag := cart.Data[0x147]
	switch mbc_flag{
	case 1, 2, 3:
		cart.MBC1 = true
	case 5, 6:
		cart.MBC2 = true
	}
	cart.ROMBank = 1

	// RAM banking
	cart.RAMBank = 0
	cart.RAM = [0x8000]byte{}
	return nil
}
