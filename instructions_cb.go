package gob

import (
	"github.com/humpheh/gob/bits"
)

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
	carry := val & 1
	rot := (val & 64) | (val >> 1)
	setter(rot)

	gb.CPU.SetZ(rot == 0)
	gb.CPU.SetN(false)
	gb.CPU.SetH(false)
	gb.CPU.SetC(carry == 1)
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
	gb.CPU.SetZ((val >> bit) & 1 == 1)
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

	for x := 0; x < len(get_map); x++ {
		var i = x
		instructions[byte(0x00 + i)] = func() {
			gb.instRlc(set_map[i], get_map[i]())
		}
		instructions[byte(0x08 + i)] = func() {
			gb.instRrc(set_map[i], get_map[i]())
		}
		instructions[byte(0x10 + i)] = func() {
			gb.instRl(set_map[i], get_map[i]())
		}
		instructions[byte(0x18 + i)] = func() {
			gb.instRr(set_map[i], get_map[i]())
		}
		instructions[byte(0x20 + i)] = func() {
			gb.instSla(set_map[i], get_map[i]())
		}
		instructions[byte(0x28 + i)] = func() {
			gb.instSra(set_map[i], get_map[i]())
		}
		instructions[byte(0x30 + i)] = func() {
			gb.instSwap(set_map[i], get_map[i]())
		}
		instructions[byte(0x38 + i)] = func() {
			gb.instSrl(set_map[i], get_map[i]())
		}

		// BIT instructions
		instructions[byte(0x40 + i)] = func() {
			gb.instBit(0, get_map[i]())
		}
		instructions[byte(0x48 + i)] = func() {
			gb.instBit(1, get_map[i]())
		}
		instructions[byte(0x50 + i)] = func() {
			gb.instBit(2, get_map[i]())
		}
		instructions[byte(0x58 + i)] = func() {
			gb.instBit(3, get_map[i]())
		}
		instructions[byte(0x60 + i)] = func() {
			gb.instBit(4, get_map[i]())
		}
		instructions[byte(0x68 + i)] = func() {
			gb.instBit(5, get_map[i]())
		}
		instructions[byte(0x70 + i)] = func() {
			gb.instBit(6, get_map[i]())
		}
		instructions[byte(0x78 + i)] = func() {
			gb.instBit(7, get_map[i]())
		}

		// RES instructions
		instructions[byte(0x80 + i)] = func() {
			set_map[i](bits.Reset(get_map[i](), 0))
		}
		instructions[byte(0x88 + i)] = func() {
			set_map[i](bits.Reset(get_map[i](), 1))
		}
		instructions[byte(0x90 + i)] = func() {
			set_map[i](bits.Reset(get_map[i](), 2))
		}
		instructions[byte(0x98 + i)] = func() {
			set_map[i](bits.Reset(get_map[i](), 3))
		}
		instructions[byte(0xA0 + i)] = func() {
			set_map[i](bits.Reset(get_map[i](), 4))
		}
		instructions[byte(0xA8 + i)] = func() {
			set_map[i](bits.Reset(get_map[i](), 5))
		}
		instructions[byte(0xB0 + i)] = func() {
			set_map[i](bits.Reset(get_map[i](), 6))
		}
		instructions[byte(0xB8 + i)] = func() {
			set_map[i](bits.Reset(get_map[i](), 7))
		}

		// SET instructions
		instructions[byte(0xC0 + i)] = func() {
			set_map[i](bits.Set(get_map[i](), 0))
		}
		instructions[byte(0xC8 + i)] = func() {
			set_map[i](bits.Set(get_map[i](), 1))
		}
		instructions[byte(0xD0 + i)] = func() {
			set_map[i](bits.Set(get_map[i](), 2))
		}
		instructions[byte(0xD8 + i)] = func() {
			set_map[i](bits.Set(get_map[i](), 3))
		}
		instructions[byte(0xE0 + i)] = func() {
			set_map[i](bits.Set(get_map[i](), 4))
		}
		instructions[byte(0xE8 + i)] = func() {
			set_map[i](bits.Set(get_map[i](), 5))
		}
		instructions[byte(0xF0 + i)] = func() {
			set_map[i](bits.Set(get_map[i](), 6))
		}
		instructions[byte(0xF8 + i)] = func() {
			set_map[i](bits.Set(get_map[i](), 7))
		}
	}
	return instructions
}