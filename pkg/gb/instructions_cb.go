package gb

func (gb *Gameboy) instRlc(setter func(byte), val byte) {
	carry := val >> 7
	rot := (val<<1)&0xFF | carry
	setter(rot)

	gb.cpu.SetZ(rot == 0)
	gb.cpu.SetN(false)
	gb.cpu.SetH(false)
	gb.cpu.SetC(carry == 1)
}

func (gb *Gameboy) instRl(setter func(byte), val byte) {
	newCarry := val >> 7
	oldCarry := boolToBitByte(gb.cpu.C())
	rot := (val<<1)&0xFF | oldCarry
	setter(rot)

	gb.cpu.SetZ(rot == 0)
	gb.cpu.SetN(false)
	gb.cpu.SetH(false)
	gb.cpu.SetC(newCarry == 1)
}

func (gb *Gameboy) instRrc(setter func(byte), val byte) {
	carry := val & 1
	rot := (val >> 1) | (carry << 7)
	setter(rot)

	gb.cpu.SetZ(rot == 0)
	gb.cpu.SetN(false)
	gb.cpu.SetH(false)
	gb.cpu.SetC(carry == 1)
}

func (gb *Gameboy) instRr(setter func(byte), val byte) {
	newCarry := val & 1
	oldCarry := boolToBitByte(gb.cpu.C())
	rot := (val >> 1) | (oldCarry << 7)
	setter(rot)

	gb.cpu.SetZ(rot == 0)
	gb.cpu.SetN(false)
	gb.cpu.SetH(false)
	gb.cpu.SetC(newCarry == 1)
}

func (gb *Gameboy) instSla(setter func(byte), val byte) {
	carry := val >> 7
	rot := (val << 1) & 0xFF
	setter(rot)

	gb.cpu.SetZ(rot == 0)
	gb.cpu.SetN(false)
	gb.cpu.SetH(false)
	gb.cpu.SetC(carry == 1)
}

func (gb *Gameboy) instSra(setter func(byte), val byte) {
	rot := (val & 128) | (val >> 1)
	setter(rot)

	gb.cpu.SetZ(rot == 0)
	gb.cpu.SetN(false)
	gb.cpu.SetH(false)
	gb.cpu.SetC(val&1 == 1)
}

func (gb *Gameboy) instSrl(setter func(byte), val byte) {
	carry := val & 1
	rot := val >> 1
	setter(rot)

	gb.cpu.SetZ(rot == 0)
	gb.cpu.SetN(false)
	gb.cpu.SetH(false)
	gb.cpu.SetC(carry == 1)
}

func (gb *Gameboy) instBit(bit byte, val byte) {
	gb.cpu.SetZ((val>>bit)&1 == 0)
	gb.cpu.SetN(false)
	gb.cpu.SetH(true)
}

func (gb *Gameboy) instSwap(setter func(byte), val byte) {
	swapped := val<<4&240 | val>>4
	setter(swapped)

	gb.cpu.SetZ(swapped == 0)
	gb.cpu.SetN(false)
	gb.cpu.SetH(false)
	gb.cpu.SetC(false)
}

func (gb *Gameboy) cbInstructions() [0x100]func() {
	instructions := [0x100]func(){}

	getMap := [8]func() byte{
		gb.cpu.BC.Hi,
		gb.cpu.BC.Lo,
		gb.cpu.DE.Hi,
		gb.cpu.DE.Lo,
		gb.cpu.HL.Hi,
		gb.cpu.HL.Lo,
		func() byte { return gb.memory.Read(gb.cpu.HL.HiLo()) },
		gb.cpu.AF.Hi,
	}
	setMap := [8]func(byte){
		gb.cpu.BC.SetHi,
		gb.cpu.BC.SetLo,
		gb.cpu.DE.SetHi,
		gb.cpu.DE.SetLo,
		gb.cpu.HL.SetHi,
		gb.cpu.HL.SetLo,
		func(v byte) { gb.memory.Write(gb.cpu.HL.HiLo(), v) },
		gb.cpu.AF.SetHi,
	}

	for x := 0; x < 8; x++ {
		// Store x so it can be used in the function scopes
		var i = x

		instructions[0x00+i] = func() { gb.instRlc(setMap[i], getMap[i]()) }
		instructions[0x08+i] = func() { gb.instRrc(setMap[i], getMap[i]()) }
		instructions[0x10+i] = func() { gb.instRl(setMap[i], getMap[i]()) }
		instructions[0x18+i] = func() { gb.instRr(setMap[i], getMap[i]()) }
		instructions[0x20+i] = func() { gb.instSla(setMap[i], getMap[i]()) }
		instructions[0x28+i] = func() { gb.instSra(setMap[i], getMap[i]()) }
		instructions[0x30+i] = func() { gb.instSwap(setMap[i], getMap[i]()) }
		instructions[0x38+i] = func() { gb.instSrl(setMap[i], getMap[i]()) }

		// BIT instructions
		instructions[0x40+i] = func() { gb.instBit(0, getMap[i]()) }
		instructions[0x48+i] = func() { gb.instBit(1, getMap[i]()) }
		instructions[0x50+i] = func() { gb.instBit(2, getMap[i]()) }
		instructions[0x58+i] = func() { gb.instBit(3, getMap[i]()) }
		instructions[0x60+i] = func() { gb.instBit(4, getMap[i]()) }
		instructions[0x68+i] = func() { gb.instBit(5, getMap[i]()) }
		instructions[0x70+i] = func() { gb.instBit(6, getMap[i]()) }
		instructions[0x78+i] = func() { gb.instBit(7, getMap[i]()) }

		// RES instructions
		instructions[0x80+i] = func() { setMap[i](bitReset(getMap[i](), 0)) }
		instructions[0x88+i] = func() { setMap[i](bitReset(getMap[i](), 1)) }
		instructions[0x90+i] = func() { setMap[i](bitReset(getMap[i](), 2)) }
		instructions[0x98+i] = func() { setMap[i](bitReset(getMap[i](), 3)) }
		instructions[0xA0+i] = func() { setMap[i](bitReset(getMap[i](), 4)) }
		instructions[0xA8+i] = func() { setMap[i](bitReset(getMap[i](), 5)) }
		instructions[0xB0+i] = func() { setMap[i](bitReset(getMap[i](), 6)) }
		instructions[0xB8+i] = func() { setMap[i](bitReset(getMap[i](), 7)) }

		// SET instructions
		instructions[0xC0+i] = func() { setMap[i](bitSet(getMap[i](), 0)) }
		instructions[0xC8+i] = func() { setMap[i](bitSet(getMap[i](), 1)) }
		instructions[0xD0+i] = func() { setMap[i](bitSet(getMap[i](), 2)) }
		instructions[0xD8+i] = func() { setMap[i](bitSet(getMap[i](), 3)) }
		instructions[0xE0+i] = func() { setMap[i](bitSet(getMap[i](), 4)) }
		instructions[0xE8+i] = func() { setMap[i](bitSet(getMap[i](), 5)) }
		instructions[0xF0+i] = func() { setMap[i](bitSet(getMap[i](), 6)) }
		instructions[0xF8+i] = func() { setMap[i](bitSet(getMap[i](), 7)) }
	}

	return instructions
}
