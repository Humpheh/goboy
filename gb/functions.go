package gb

import (
	"github.com/humpheh/goboy/bits"
)

func (gb *Gameboy) instAdd(set func(byte), val1 byte, val2 byte, addCarry bool) {
	var carry int16 = 0
	if gb.CPU.C() && addCarry {
		carry = 1
	}
	total := int16(val1) + int16(val2) + carry
	set(byte(total))

	gb.CPU.SetZ(byte(total) == 0)
	gb.CPU.SetN(false)
	gb.CPU.SetH((val2&0xF)+(val1&0xF)+byte(carry) > 0xF)
	gb.CPU.SetC(total > 0xFF) // If result is greater than 255
}

func (gb *Gameboy) instSub(set func(byte), val1 byte, val2 byte, addCarry bool) {
	var carry int16 = 0
	if gb.CPU.C() && addCarry {
		carry = 1
	}
	dirtySum := int16(val1) - int16(val2) - carry
	total := byte(dirtySum)
	set(total)

	gb.CPU.SetZ(total == 0)
	gb.CPU.SetN(true)
	gb.CPU.SetH(int16(val1&0x0f)-int16(val2&0xF)-int16(carry) < 0) // TODO: WRONG?
	gb.CPU.SetC(dirtySum < 0)                                      // If result is less than 0
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
	gb.CPU.SetC(val1 > val2)                   // TODO: Check
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
	gb.CPU.SetH(org&0x0F == 0) // TODO: Check
}

func (gb *Gameboy) instAdd16(set func(uint16), val1 uint16, val2 uint16) {
	total := int32(val1) + int32(val2)
	set(uint16(total))
	//gb2.CPU.SetZ(total == 0)
	gb.CPU.SetN(false)
	gb.CPU.SetH(int32(val1&0xFFF) > (total & 0xFFF)) //bits.HalfCarryAdd16(val1, val2))
	gb.CPU.SetC(total > 0xFFFF)                      //bits.CarryAdd16(val1, val2))
}

func (gb *Gameboy) instAdd16Signed(set func(uint16), val1 uint16, val2 int8) {
	total := uint16(int32(val1) + int32(val2))
	set(total)
	tmpVal := val1 ^ uint16(val2) ^ total
	gb.CPU.SetZ(false)
	gb.CPU.SetN(false)
	// TODO: Check these!
	gb.CPU.SetH((tmpVal & 0x10) == 0x10)
	gb.CPU.SetC((tmpVal & 0x100) == 0x100)
}

func (gb *Gameboy) instInc16(set func(uint16 uint16), org uint16) {
	set(org + 1)
}

func (gb *Gameboy) instDec16(set func(uint16 uint16), org uint16) {
	set(org - 1)
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
