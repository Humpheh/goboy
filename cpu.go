package gob

type Register uint16

func (reg Register) Hi() byte {
	return byte(reg >> 8)
}

func (reg Register) Lo() byte {
	return byte(reg & 0xFF)
}

func (reg Register) SetHi(val byte) {
	reg = Register(uint16(val) << 8 & (uint16(reg) & 0xFF))
}

func (reg Register) SetLo(val byte) {
	reg = Register(uint16(val) & (uint16(reg) & 0xFF00))
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

func (cpu *CPU) SetFlag(index byte, on bool) {
	var x uint16 = 0
	if on {
		x = 1
	}
	cpu.AF = cpu.AF ^ Register(-x ^ uint16(cpu.AF) & (1 << index))
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
	return cpu.AF >> 7 & 1 == 1
}

func (cpu *CPU) N() bool {
	return cpu.AF >> 6 & 1 == 1
}

func (cpu *CPU) H() bool {
	return cpu.AF >> 5 & 1 == 1
}

func (cpu *CPU) C() bool {
	return cpu.AF >> 4 & 1 == 1
}