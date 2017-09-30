package gb

import (
	"github.com/humpheh/goboy/bits"
)

type Register struct {
	// The value of the register.
	val uint16
	// A mask over the possible values in the register.
	// Only used for the AF register where lower bits of
	// F cannot be set.
	mask uint16
}

// Get the higher byte of the register.
func (reg *Register) Hi() byte {
	return byte(reg.val >> 8)
}

// Get the lower byte of the register.
func (reg *Register) Lo() byte {
	return byte(reg.val & 0xFF)
}

// Get the 2 byte value of the register.
func (reg *Register) HiLo() uint16 {
	return reg.val
}

// Set the higher byte of the register.
func (reg *Register) SetHi(val byte) {
	reg.val = uint16(val)<<8 | (uint16(reg.val) & 0xFF)
	reg.updateMask()
}

// Set the lower byte of the register.
func (reg *Register) SetLo(val byte) {
	reg.val = uint16(val) | (uint16(reg.val) & 0xFF00)
	reg.updateMask()
}

// Set hte value of the register.
func (reg *Register) Set(val uint16) {
	reg.val = val
	reg.updateMask()
}

// Mask the value if one is set on this register.
func (reg *Register) updateMask() {
	if reg.mask != 0 {
		reg.val &= reg.mask
	}
}

// CPU contains the registers used for program execution and
// provides methods for setting flags.
type CPU struct {
	AF Register
	BC Register
	DE Register
	HL Register

	PC uint16
	SP Register

	Divider int
}

// Init CPU and its registers to the initial values.
func (cpu *CPU) Init() {
	cpu.PC = 0x100
	cpu.AF.Set(0x01B0)
	cpu.BC.Set(0x0013)
	cpu.DE.Set(0x00D8)
	cpu.HL.Set(0x014D)
	cpu.SP.Set(0xFFFE)

	cpu.AF.mask = 0xFFF0
}

// Internally set the value of a flag on the flag register.
func (cpu *CPU) setFlag(index byte, on bool) {
	if on {
		cpu.AF.SetLo(bits.Set(cpu.AF.Lo(), index))
	} else {
		cpu.AF.SetLo(bits.Reset(cpu.AF.Lo(), index))
	}
}

// Set the value of the Z flag.
func (cpu *CPU) SetZ(on bool) {
	cpu.setFlag(7, on)
}

// Set the value of the N flag.
func (cpu *CPU) SetN(on bool) {
	cpu.setFlag(6, on)
}

// Set the value of the H flag.
func (cpu *CPU) SetH(on bool) {
	cpu.setFlag(5, on)
}

// Set the value of the C flag.
func (cpu *CPU) SetC(on bool) {
	cpu.setFlag(4, on)
}

// Get the value of the Z flag.
func (cpu *CPU) Z() bool {
	return cpu.AF.HiLo()>>7&1 == 1
}

// Get the value of the N flag.
func (cpu *CPU) N() bool {
	return cpu.AF.HiLo()>>6&1 == 1
}

// Get the value of the H flag.
func (cpu *CPU) H() bool {
	return cpu.AF.HiLo()>>5&1 == 1
}

// Get the value of the C flag.
func (cpu *CPU) C() bool {
	return cpu.AF.HiLo()>>4&1 == 1
}
