package gob

type Register struct {
	val uint16
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
}

func (reg *Register) SetLo(val byte) {
	reg.val = uint16(val) | (uint16(reg.val) & 0xFF00)
}

func (reg *Register) Set(val uint16) {
	reg.val = val
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
	cpu.AF.Set(cpu.AF.HiLo() ^ (-x ^ uint16(cpu.AF.HiLo()) & (1 << index)))
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