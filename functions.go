package gob

import (
	"github.com/humpheh/gob/bits"
)

func (gb *Gameboy) instAdd(set func(byte), val1 byte, val2 byte, addCarry bool) {
	if gb.CPU.C() && addCarry {
		val1 += 1
	}
	total := val1 + val2
	set(total)
	gb.CPU.SetZ(total == 0)
	gb.CPU.SetN(false)
	gb.CPU.SetH(bits.HalfCarryAdd(val1, val2))
	gb.CPU.SetC(bits.CarryAdd(val1, val2))
}

func (gb *Gameboy) instSub(set func(byte), val1 byte, val2 byte, addCarry bool) {
	if gb.CPU.C() && addCarry {
		val1 += 1
	}
	total := val1 - val2
	set(total)
	gb.CPU.SetZ(total == 0)
	gb.CPU.SetN(true)

	// TODO: check these
	gb.CPU.SetH((val1 & 0x0f) > (val2 & 0x0f))
	gb.CPU.SetC(val1 > val2)
}

func (gb *Gameboy) instAnd(set func(byte), val1 byte, val2 byte) {
	total := val1 & val2
	set(total)
	gb.CPU.SetZ(total == 0)
	gb.CPU.SetN(false)
	gb.CPU.SetH(true)
	gb.CPU.SetC(false)
}

func (gb *Gameboy) instOr(set func(byte), val1 byte, val2 byte) {
	total := val1 | val2
	set(total)
	gb.CPU.SetZ(total == 0)
	gb.CPU.SetN(false)
	gb.CPU.SetH(false)
	gb.CPU.SetC(false)
}

func (gb *Gameboy) instXor(set func(byte), val1 byte, val2 byte) {
	total := val1 ^ val2
	set(total)
	gb.CPU.SetZ(total == 0)
	gb.CPU.SetN(false)
	gb.CPU.SetH(false)
	gb.CPU.SetC(false)
}

func (gb *Gameboy) instCp(val1 byte, val2 byte) {
	total := val2 - val1
	gb.CPU.SetZ(total == 0)
	gb.CPU.SetN(true)
	gb.CPU.SetH((val1 & 0x0f) > (val2 & 0x0f)) // TODO: check
	gb.CPU.SetC(val1 > val2)
}

func (gb *Gameboy) instInc(set func(byte), org byte) {
	total := org + 1
	set(total)

	gb.CPU.SetZ(total == 0)
	gb.CPU.SetN(false)
	gb.CPU.SetH(bits.HalfCarryAdd(org, 1))
}

func (gb *Gameboy) instDec(set func(byte), org byte) {
	total := org - 1
	set(total)

	gb.CPU.SetZ(total == 0)
	gb.CPU.SetN(true)
	gb.CPU.SetH(org & 0x0F == 0) // TODO: Check
}

func (gb *Gameboy) instAdd16(set func(uint16), val1 uint16, val2 uint16) {
	total := val1 + val2
	set(total)
	gb.CPU.SetZ(total == 0) // TODO: 0xE8 assumes this will never == 0 (always reset)
	gb.CPU.SetN(false)
	gb.CPU.SetH(bits.HalfCarryAdd16(val1, val2))
	gb.CPU.SetC(bits.CarryAdd16(val1, val2))
}

func (gb *Gameboy) instInc16(set func(uint16 uint16), org uint16) {
	set(org + 1)
	// TODO: Apparently no flags are set from this op?
}

func (gb *Gameboy) instDec16(set func(uint16 uint16), org uint16) {
	set(org - 1)
	// TODO: Apparently no flags are set from this op?
}

func (gb *Gameboy) instJump(next uint16) {
	gb.CPU.PC = next
}

func (gb *Gameboy) instCall(next uint16) {
	gb.PushStack(gb.CPU.PC)
	gb.CPU.PC = next
}

func (gb *Gameboy) instRet() {
	gb.CPU.PC = gb.PopStack()
}