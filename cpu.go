package gob

import (
	"github.com/humpheh/gob/bits"
)

type Register struct {
	isAF bool // TODO: change this?
	val  uint16
}

func (reg *Register) Hi() byte {
	return byte(reg.val >> 8)
}

func (reg *Register) Lo() byte {
	return byte(reg.val & 0xFF)
}

func (reg *Register) HiLo() uint16 {
	return reg.val
}

func (reg *Register) SetHi(val byte) {
	reg.val = uint16(val) << 8 | (uint16(reg.val) & 0xFF)
	reg.AfterSet()
}

func (reg *Register) SetLo(val byte) {
	reg.val = uint16(val) | (uint16(reg.val) & 0xFF00)
	reg.AfterSet()
}

func (reg *Register) Set(val uint16) {
	reg.val = val
	reg.AfterSet()
}

func (reg *Register) AfterSet() {
	if reg.isAF {
		reg.val &= 0xFFF0
	}
}

type CPU struct {
	AF Register
	BC Register
	DE Register
	HL Register

	PC uint16
	SP Register

	Divider int
}

// Init CPU and its registers to the initial values
func (cpu *CPU) Init() {
	cpu.PC = 0x100
	cpu.AF.Set(0x01B0)
	cpu.BC.Set(0x0013)
	cpu.DE.Set(0x00D8)
	cpu.HL.Set(0x014D)
	cpu.SP.Set(0xFFFE)

	cpu.AF.isAF = true
}

func (cpu *CPU) SetFlag(index byte, on bool) {
	if on {
		cpu.AF.SetLo(bits.Set(cpu.AF.Lo(), index))
	} else {
		cpu.AF.SetLo(bits.Reset(cpu.AF.Lo(), index))
	}
}

func (cpu *CPU) SetZ(on bool) {
	cpu.SetFlag(7, on)
}

func (cpu *CPU) SetN(on bool) {
	cpu.SetFlag(6, on)
}

func (cpu *CPU) SetH(on bool) {
	cpu.SetFlag(5, on)
}

func (cpu *CPU) SetC(on bool) {
	cpu.SetFlag(4, on)
}

func (cpu *CPU) Z() bool {
	return cpu.AF.HiLo() >> 7 & 1 == 1
}

func (cpu *CPU) N() bool {
	return cpu.AF.HiLo() >> 6 & 1 == 1
}

func (cpu *CPU) H() bool {
	return cpu.AF.HiLo() >> 5 & 1 == 1
}

func (cpu *CPU) C() bool {
	return cpu.AF.HiLo() >> 4 & 1 == 1
}
