package gob

import (
	"github.com/humpheh/gob/bits"
)

var CB_NAMES map[byte]string = map[byte]string{}

func (gb *Gameboy) instRlc(setter func(byte), val byte) {
	carry := val >> 7
	rot := (val << 1) & 0xFF | carry
	setter(rot)

	gb.CPU.SetZ(rot == 0)
	gb.CPU.SetN(false)
	gb.CPU.SetH(false)
	gb.CPU.SetC(carry == 1)
}

func (gb *Gameboy) instRl(setter func(byte), val byte) {
	new_carry := val >> 7
	old_carry := bits.B(gb.CPU.C())
	rot := (val << 1) & 0xFF | old_carry
	setter(rot)

	gb.CPU.SetZ(rot == 0)
	gb.CPU.SetN(false)
	gb.CPU.SetH(false)
	gb.CPU.SetC(new_carry == 1)
}

func (gb *Gameboy) instRrc(setter func(byte), val byte) {
	carry := val & 1
	rot := (val >> 1) | (carry << 7)
	setter(rot)

	gb.CPU.SetZ(rot == 0)
	gb.CPU.SetN(false)
	gb.CPU.SetH(false)
	gb.CPU.SetC(carry == 1)
}

func (gb *Gameboy) instRr(setter func(byte), val byte) {
	new_carry := val & 1
	old_carry := bits.B(gb.CPU.C())
	rot := (val >> 1) | (old_carry << 7)
	setter(rot)

	gb.CPU.SetZ(rot == 0)
	gb.CPU.SetN(false)
	gb.CPU.SetH(false)
	gb.CPU.SetC(new_carry == 1)
}

func (gb *Gameboy) instSla(setter func(byte), val byte) {
	carry := val >> 7
	rot := (val << 1) & 0xFF
	setter(rot)

	gb.CPU.SetZ(rot == 0)
	gb.CPU.SetN(false)
	gb.CPU.SetH(false)
	gb.CPU.SetC(carry == 1)
}

func (gb *Gameboy) instSra(setter func(byte), val byte) {
	rot := (val & 128) | (val >> 1)
	setter(rot)

	gb.CPU.SetZ(rot == 0)
	gb.CPU.SetN(false)
	gb.CPU.SetH(false)
	gb.CPU.SetC(val & 1 == 1)
}

func (gb *Gameboy) instSrl(setter func(byte), val byte) {
	carry := val & 1
	rot := val >> 1
	setter(rot)

	gb.CPU.SetZ(rot == 0)
	gb.CPU.SetN(false)
	gb.CPU.SetH(false)
	gb.CPU.SetC(carry == 1)
}

func (gb *Gameboy) instBit(bit byte, val byte) {
	gb.CPU.SetZ((val >> bit) & 1 == 0)
	gb.CPU.SetN(false)
	gb.CPU.SetH(true)
}

func (gb *Gameboy) instSwap(setter func(byte), val byte) {
	swapped := val << 4 & 240 | val >> 4
	setter(swapped)

	gb.CPU.SetZ(swapped == 0)
	gb.CPU.SetN(false)
	gb.CPU.SetH(false)
	gb.CPU.SetC(false)
}

func (gb *Gameboy) CBInstructions() map[byte]func() {
	instructions := map[byte]func(){}

	get_map := []func() byte{
		gb.CPU.BC.Hi,
		gb.CPU.BC.Lo,
		gb.CPU.DE.Hi,
		gb.CPU.DE.Lo,
		gb.CPU.HL.Hi,
		gb.CPU.HL.Lo,
		func() byte { return gb.Memory.Read(gb.CPU.HL.HiLo()) },
		gb.CPU.AF.Hi,
	}
	set_map := []func(byte) {
		gb.CPU.BC.SetHi,
		gb.CPU.BC.SetLo,
		gb.CPU.DE.SetHi,
		gb.CPU.DE.SetLo,
		gb.CPU.HL.SetHi,
		gb.CPU.HL.SetLo,
		func(v byte) { gb.Memory.Write(gb.CPU.HL.HiLo(), v) },
		gb.CPU.AF.SetHi,
	}
	str_map := []string{
		"B", "C", "D", "E", "H", "L", "(HL)", "A",
	}

	for x := 0; x < len(get_map); x++ {
		var i = x
		instructions[byte(0x00 + i)] = func() {
			gb.instRlc(set_map[i], get_map[i]())
		}
		CB_NAMES[byte(0x00 + i)] = "RLC " + str_map[i]

		instructions[byte(0x08 + i)] = func() {
			gb.instRrc(set_map[i], get_map[i]())
		}
		CB_NAMES[byte(0x08 + i)] = "RRC " + str_map[i]

		instructions[byte(0x10 + i)] = func() {
			gb.instRl(set_map[i], get_map[i]())
		}
		CB_NAMES[byte(0x10 + i)] = "RL " + str_map[i]

		instructions[byte(0x18 + i)] = func() {
			gb.instRr(set_map[i], get_map[i]())
		}
		CB_NAMES[byte(0x18 + i)] = "RR " + str_map[i]

		instructions[byte(0x20 + i)] = func() {
			gb.instSla(set_map[i], get_map[i]())
		}
		CB_NAMES[byte(0x20 + i)] = "SLA " + str_map[i]

		instructions[byte(0x28 + i)] = func() {
			gb.instSra(set_map[i], get_map[i]())
		}
		CB_NAMES[byte(0x28 + i)] = "SRA " + str_map[i]

		instructions[byte(0x30 + i)] = func() {
			gb.instSwap(set_map[i], get_map[i]())
		}
		CB_NAMES[byte(0x30 + i)] = "SWAP " + str_map[i]

		instructions[byte(0x38 + i)] = func() {
			gb.instSrl(set_map[i], get_map[i]())
		}
		CB_NAMES[byte(0x38 + i)] = "SRL " + str_map[i]

		// BIT instructions
		instructions[byte(0x40 + i)] = func() {
			gb.instBit(0, get_map[i]())
		}
		CB_NAMES[byte(0x40 + i)] = "BIT 0," + str_map[i]

		instructions[byte(0x48 + i)] = func() {
			gb.instBit(1, get_map[i]())
		}
		CB_NAMES[byte(0x48 + i)] = "BIT 1," + str_map[i]

		instructions[byte(0x50 + i)] = func() {
			gb.instBit(2, get_map[i]())
		}
		CB_NAMES[byte(0x50 + i)] = "BIT 2," + str_map[i]

		instructions[byte(0x58 + i)] = func() {
			gb.instBit(3, get_map[i]())
		}
		CB_NAMES[byte(0x58 + i)] = "BIT 3," + str_map[i]

		instructions[byte(0x60 + i)] = func() {
			gb.instBit(4, get_map[i]())
		}
		CB_NAMES[byte(0x60 + i)] = "BIT 4," + str_map[i]

		instructions[byte(0x68 + i)] = func() {
			gb.instBit(5, get_map[i]())
		}
		CB_NAMES[byte(0x68 + i)] = "BIT 5," + str_map[i]

		instructions[byte(0x70 + i)] = func() {
			gb.instBit(6, get_map[i]())
		}
		CB_NAMES[byte(0x70 + i)] = "BIT 6," + str_map[i]

		instructions[byte(0x78 + i)] = func() {
			gb.instBit(7, get_map[i]())
		}
		CB_NAMES[byte(0x78 + i)] = "BIT 7," + str_map[i]

		// RES instructions
		instructions[byte(0x80 + i)] = func() {
			set_map[i](bits.Reset(get_map[i](), 0))
		}
		CB_NAMES[byte(0x80 + i)] = "RES 0," + str_map[i]

		instructions[byte(0x88 + i)] = func() {
			set_map[i](bits.Reset(get_map[i](), 1))
		}
		CB_NAMES[byte(0x88 + i)] = "RES 1," + str_map[i]

		instructions[byte(0x90 + i)] = func() {
			set_map[i](bits.Reset(get_map[i](), 2))
		}
		CB_NAMES[byte(0x90 + i)] = "RES 2," + str_map[i]

		instructions[byte(0x98 + i)] = func() {
			set_map[i](bits.Reset(get_map[i](), 3))
		}
		CB_NAMES[byte(0x98 + i)] = "RES 3," + str_map[i]

		instructions[byte(0xA0 + i)] = func() {
			set_map[i](bits.Reset(get_map[i](), 4))
		}
		CB_NAMES[byte(0xA0 + i)] = "RES 4," + str_map[i]

		instructions[byte(0xA8 + i)] = func() {
			set_map[i](bits.Reset(get_map[i](), 5))
		}
		CB_NAMES[byte(0xA8 + i)] = "RES 5," + str_map[i]

		instructions[byte(0xB0 + i)] = func() {
			set_map[i](bits.Reset(get_map[i](), 6))
		}
		CB_NAMES[byte(0xB0 + i)] = "RES 6," + str_map[i]

		instructions[byte(0xB8 + i)] = func() {
			set_map[i](bits.Reset(get_map[i](), 7))
		}
		CB_NAMES[byte(0xD8 + i)] = "RES 7," + str_map[i]

		// SET instructions
		instructions[byte(0xC0 + i)] = func() {
			set_map[i](bits.Set(get_map[i](), 0))
		}
		CB_NAMES[byte(0xC0 + i)] = "SET 1," + str_map[i]

		instructions[byte(0xC8 + i)] = func() {
			set_map[i](bits.Set(get_map[i](), 1))
		}
		CB_NAMES[byte(0xC8 + i)] = "SET 1," + str_map[i]

		instructions[byte(0xD0 + i)] = func() {
			set_map[i](bits.Set(get_map[i](), 2))
		}
		CB_NAMES[byte(0xD0 + i)] = "SET 2," + str_map[i]

		instructions[byte(0xD8 + i)] = func() {
			set_map[i](bits.Set(get_map[i](), 3))
		}
		CB_NAMES[byte(0xD8 + i)] = "SET 3," + str_map[i]

		instructions[byte(0xE0 + i)] = func() {
			set_map[i](bits.Set(get_map[i](), 4))
		}
		CB_NAMES[byte(0xE0 + i)] = "SET 4," + str_map[i]

		instructions[byte(0xE8 + i)] = func() {
			set_map[i](bits.Set(get_map[i](), 5))
		}
		CB_NAMES[byte(0xE8 + i)] = "SET 5," + str_map[i]

		instructions[byte(0xF0 + i)] = func() {
			set_map[i](bits.Set(get_map[i](), 6))
		}
		CB_NAMES[byte(0xF0 + i)] = "SET 6," + str_map[i]

		instructions[byte(0xF8 + i)] = func() {
			set_map[i](bits.Set(get_map[i](), 7))
		}
		CB_NAMES[byte(0xF8 + i)] = "SET 7," + str_map[i]
	}
	return instructions
}