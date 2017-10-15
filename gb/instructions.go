package gb

import (
	"log"
	"fmt"
)

var OpcodeCycles = []int{
	//  0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f
	1, 3, 2, 2, 1, 1, 2, 1, 5, 2, 2, 2, 1, 1, 2, 1, // 0
	1, 3, 2, 2, 1, 1, 2, 1, 3, 2, 2, 2, 1, 1, 2, 1, // 1
	2, 3, 2, 2, 1, 1, 2, 1, 2, 2, 2, 2, 1, 1, 2, 1, // 2
	2, 3, 2, 2, 3, 3, 3, 1, 2, 2, 2, 2, 1, 1, 2, 1, // 3
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 4
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 5
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 6
	2, 2, 2, 2, 2, 2, 1, 2, 1, 1, 1, 1, 1, 1, 2, 1, // 7
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 8
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 9
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // a
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // b
	2, 3, 3, 4, 3, 4, 2, 4, 2, 4, 3, 0, 3, 6, 2, 4, // c
	2, 3, 3, 1, 3, 4, 2, 4, 2, 4, 3, 1, 3, 1, 2, 4, // d
	3, 3, 2, 1, 1, 4, 2, 4, 4, 1, 4, 1, 1, 1, 2, 4, // e
	3, 3, 2, 1, 1, 4, 2, 4, 3, 2, 4, 1, 0, 1, 2, 4, // f
}

var CBOpcodeCycles = []int{
	//  0  1  2  3  4  5  6  7  8  9  A  B  C  D  E  F
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2, // 0
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2, // 1
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2, // 2
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2, // 3
	2, 2, 2, 2, 2, 2, 3, 2, 2, 2, 2, 2, 2, 2, 3, 2, // 4
	2, 2, 2, 2, 2, 2, 3, 2, 2, 2, 2, 2, 2, 2, 3, 2, // 5
	2, 2, 2, 2, 2, 2, 3, 2, 2, 2, 2, 2, 2, 2, 3, 2, // 6
	2, 2, 2, 2, 2, 2, 3, 2, 2, 2, 2, 2, 2, 2, 3, 2, // 7
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2, // 8
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2, // 9
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2, // A
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2, // B
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2, // C
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2, // D
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2, // E
	2, 2, 2, 2, 2, 2, 4, 2, 2, 2, 2, 2, 2, 2, 4, 2, // F
}

// Get the value at the current PC address, increment the PC, update
// the CPU ticks and execute the opcode.
func (gb *Gameboy) ExecuteNextOpcode() int {
	if gb.DebugScanner != nil {
		expectedCPU, ff0f := getDebugNum(gb.DebugScanner)
		myff0f := uint16(gb.Memory.Read(0xFF05))
		logOpcode(gb)
		fmt.Println(cpuStateString(gb.CPU, "GoBoy"), fmt.Sprintf("%b", myff0f))
		fmt.Println(cpuStateString(&expectedCPU, "Exp"), fmt.Sprintf("%b", ff0f))
		fmt.Println(cpuCompareString(gb.CPU, &expectedCPU))

		if !isEqual(gb.CPU, &expectedCPU) || myff0f != ff0f {
			waitForInput()
		}
	}

	opcode := gb.popPC()
	gb.thisCpuTicks = OpcodeCycles[opcode] * 4
	gb.ExecuteOpcode(opcode)
	return gb.thisCpuTicks
}

// Read the value at the PC and increment the PC.
func (gb *Gameboy) popPC() byte {
	opcode := gb.Memory.Read(gb.CPU.PC)
	gb.CPU.PC++
	return opcode
}

// Read the next 16bit value at the PC.
func (gb *Gameboy) popPC16() uint16 {
	b1 := uint16(gb.popPC())
	b2 := uint16(gb.popPC())
	return b2<<8 | b1
}

// Large switch statement containing the opcode operations.
func (gb *Gameboy) ExecuteOpcode(opcode byte) {
	switch opcode {
	// LD B, n
	case 0x06:
		gb.CPU.BC.SetHi(gb.popPC())

	// LD C, n
	case 0x0E:
		gb.CPU.BC.SetLo(gb.popPC())

	// LD D, n
	case 0x16:
		gb.CPU.DE.SetHi(gb.popPC())

	// LD E, n
	case 0x1E:
		gb.CPU.DE.SetLo(gb.popPC())

	// LD H, n
	case 0x26:
		gb.CPU.HL.SetHi(gb.popPC())

	// LD L, n
	case 0x2E:
		gb.CPU.HL.SetLo(gb.popPC())

	// LD A,A
	case 0x7F:
		gb.CPU.AF.SetHi(gb.CPU.AF.Hi())

	// LD A,B
	case 0x78:
		gb.CPU.AF.SetHi(gb.CPU.BC.Hi())

	// LD A,C
	case 0x79:
		gb.CPU.AF.SetHi(gb.CPU.BC.Lo())

	// LD A,D
	case 0x7A:
		gb.CPU.AF.SetHi(gb.CPU.DE.Hi())

	// LD A,E
	case 0x7B:
		gb.CPU.AF.SetHi(gb.CPU.DE.Lo())

	// LD A,H
	case 0x7C:
		gb.CPU.AF.SetHi(gb.CPU.HL.Hi())

	// LD A,L
	case 0x7D:
		gb.CPU.AF.SetHi(gb.CPU.HL.Lo())

	// LD A,(BC)
	case 0x0A:
		val := gb.Memory.Read(gb.CPU.BC.HiLo())
		gb.CPU.AF.SetHi(val)

	// LD A,(DE)
	case 0x1A:
		val := gb.Memory.Read(gb.CPU.DE.HiLo())
		gb.CPU.AF.SetHi(val)

	// LD A,(HL)
	case 0x7E:
		val := gb.Memory.Read(gb.CPU.HL.HiLo())
		gb.CPU.AF.SetHi(val)

	// LD A,(nn)
	case 0xFA:
		val := gb.Memory.Read(gb.popPC16())
		gb.CPU.AF.SetHi(val)

	// LD A,(nn)
	case 0x3E:
		val := gb.popPC()
		gb.CPU.AF.SetHi(val)

	// LD B,A
	case 0x47:
		gb.CPU.BC.SetHi(gb.CPU.AF.Hi())

	// LD B,B
	case 0x40:
		gb.CPU.BC.SetHi(gb.CPU.BC.Hi())

	// LD B,C
	case 0x41:
		gb.CPU.BC.SetHi(gb.CPU.BC.Lo())

	// LD B,D
	case 0x42:
		gb.CPU.BC.SetHi(gb.CPU.DE.Hi())

	// LD B,E
	case 0x43:
		gb.CPU.BC.SetHi(gb.CPU.DE.Lo())

	// LD B,H
	case 0x44:
		gb.CPU.BC.SetHi(gb.CPU.HL.Hi())

	// LD B,L
	case 0x45:
		gb.CPU.BC.SetHi(gb.CPU.HL.Lo())

	// LD B,(HL)
	case 0x46:
		val := gb.Memory.Read(gb.CPU.HL.HiLo())
		gb.CPU.BC.SetHi(val)

	// LD C,A
	case 0x4F:
		gb.CPU.BC.SetLo(gb.CPU.AF.Hi())

	// LD C,B
	case 0x48:
		gb.CPU.BC.SetLo(gb.CPU.BC.Hi())

	// LD C,C
	case 0x49:
		gb.CPU.BC.SetLo(gb.CPU.BC.Lo())

	// LD C,D
	case 0x4A:
		gb.CPU.BC.SetLo(gb.CPU.DE.Hi())

	// LD C,E
	case 0x4B:
		gb.CPU.BC.SetLo(gb.CPU.DE.Lo())

	// LD C,H
	case 0x4C:
		gb.CPU.BC.SetLo(gb.CPU.HL.Hi())

	// LD C,L
	case 0x4D:
		gb.CPU.BC.SetLo(gb.CPU.HL.Lo())

	// LD C,(HL)
	case 0x4E:
		val := gb.Memory.Read(gb.CPU.HL.HiLo())
		gb.CPU.BC.SetLo(val)

	// LD D,A
	case 0x57:
		gb.CPU.DE.SetHi(gb.CPU.AF.Hi())

	// LD D,B
	case 0x50:
		gb.CPU.DE.SetHi(gb.CPU.BC.Hi())

	// LD D,C
	case 0x51:
		gb.CPU.DE.SetHi(gb.CPU.BC.Lo())

	// LD D,D
	case 0x52:
		gb.CPU.DE.SetHi(gb.CPU.DE.Hi())

	// LD D,E
	case 0x53:
		gb.CPU.DE.SetHi(gb.CPU.DE.Lo())

	// LD D,H
	case 0x54:
		gb.CPU.DE.SetHi(gb.CPU.HL.Hi())

	// LD D,L
	case 0x55:
		gb.CPU.DE.SetHi(gb.CPU.HL.Lo())

	// LD D,(HL)
	case 0x56:
		val := gb.Memory.Read(gb.CPU.HL.HiLo())
		gb.CPU.DE.SetHi(val)

	// LD E,A
	case 0x5F:
		gb.CPU.DE.SetLo(gb.CPU.AF.Hi())

	// LD E,B
	case 0x58:
		gb.CPU.DE.SetLo(gb.CPU.BC.Hi())

	// LD E,C
	case 0x59:
		gb.CPU.DE.SetLo(gb.CPU.BC.Lo())

	// LD E,D
	case 0x5A:
		gb.CPU.DE.SetLo(gb.CPU.DE.Hi())

	// LD E,E
	case 0x5B:
		gb.CPU.DE.SetLo(gb.CPU.DE.Lo())

	// LD E,H
	case 0x5C:
		gb.CPU.DE.SetLo(gb.CPU.HL.Hi())

	// LD E,L
	case 0x5D:
		gb.CPU.DE.SetLo(gb.CPU.HL.Lo())

	// LD E,(HL)
	case 0x5E:
		val := gb.Memory.Read(gb.CPU.HL.HiLo())
		gb.CPU.DE.SetLo(val)

	// LD H,A
	case 0x67:
		gb.CPU.HL.SetHi(gb.CPU.AF.Hi())

	// LD H,B
	case 0x60:
		gb.CPU.HL.SetHi(gb.CPU.BC.Hi())

	// LD H,C
	case 0x61:
		gb.CPU.HL.SetHi(gb.CPU.BC.Lo())

	// LD H,D
	case 0x62:
		gb.CPU.HL.SetHi(gb.CPU.DE.Hi())

	// LD H,E
	case 0x63:
		gb.CPU.HL.SetHi(gb.CPU.DE.Lo())

	// LD H,H
	case 0x64:
		gb.CPU.HL.SetHi(gb.CPU.HL.Hi())

	// LD H,L
	case 0x65:
		gb.CPU.HL.SetHi(gb.CPU.HL.Lo())

	// LD H,(HL)
	case 0x66:
		val := gb.Memory.Read(gb.CPU.HL.HiLo())
		gb.CPU.HL.SetHi(val)

	// LD L,A
	case 0x6F:
		gb.CPU.HL.SetLo(gb.CPU.AF.Hi())

	// LD L,B
	case 0x68:
		gb.CPU.HL.SetLo(gb.CPU.BC.Hi())

	// LD L,C
	case 0x69:
		gb.CPU.HL.SetLo(gb.CPU.BC.Lo())

	// LD L,D
	case 0x6A:
		gb.CPU.HL.SetLo(gb.CPU.DE.Hi())

	// LD L,E
	case 0x6B:
		gb.CPU.HL.SetLo(gb.CPU.DE.Lo())

	// LD L,H
	case 0x6C:
		gb.CPU.HL.SetLo(gb.CPU.HL.Hi())

	// LD L,L
	case 0x6D:
		gb.CPU.HL.SetLo(gb.CPU.HL.Lo())

	// LD L,(HL)
	case 0x6E:
		val := gb.Memory.Read(gb.CPU.HL.HiLo())
		gb.CPU.HL.SetLo(val)

	// LD (HL),A
	case 0x77:
		val := gb.CPU.AF.Hi()
		gb.Memory.Write(gb.CPU.HL.HiLo(), val)

	// LD (HL),B
	case 0x70:
		val := gb.CPU.BC.Hi()
		gb.Memory.Write(gb.CPU.HL.HiLo(), val)

	// LD (HL),C
	case 0x71:
		val := gb.CPU.BC.Lo()
		gb.Memory.Write(gb.CPU.HL.HiLo(), val)

	// LD (HL),D
	case 0x72:
		val := gb.CPU.DE.Hi()
		gb.Memory.Write(gb.CPU.HL.HiLo(), val)

	// LD (HL),E
	case 0x73:
		val := gb.CPU.DE.Lo()
		gb.Memory.Write(gb.CPU.HL.HiLo(), val)

	// LD (HL),H
	case 0x74:
		val := gb.CPU.HL.Hi()
		gb.Memory.Write(gb.CPU.HL.HiLo(), val)

	// LD (HL),L
	case 0x75:
		val := gb.CPU.HL.Lo()
		gb.Memory.Write(gb.CPU.HL.HiLo(), val)

	// LD (HL),n 36
	case 0x36:
		val := gb.popPC()
		gb.Memory.Write(gb.CPU.HL.HiLo(), val)

	// LD (BC),A
	case 0x02:
		val := gb.CPU.AF.Hi()
		gb.Memory.Write(gb.CPU.BC.HiLo(), val)

	// LD (DE),A
	case 0x12:
		val := gb.CPU.AF.Hi()
		gb.Memory.Write(gb.CPU.DE.HiLo(), val)

	// LD (nn),A
	case 0xEA:
		val := gb.CPU.AF.Hi()
		gb.Memory.Write(gb.popPC16(), val)

	// LD A,(C)
	case 0xF2:
		val := 0xFF00 + uint16(gb.CPU.BC.Lo())
		gb.CPU.AF.SetHi(gb.Memory.Read(val))

	// LD (C),A
	case 0xE2:
		val := gb.CPU.AF.Hi()
		mem := 0xFF00 + uint16(gb.CPU.BC.Lo())
		gb.Memory.Write(mem, val)

	// LDD A,(HL)
	case 0x3A:
		val := gb.Memory.Read(gb.CPU.HL.HiLo())
		gb.CPU.AF.SetHi(val)
		gb.CPU.HL.Set(gb.CPU.HL.HiLo() - 1)

	// LDD (HL),A
	case 0x32:
		val := gb.CPU.HL.HiLo()
		gb.Memory.Write(val, gb.CPU.AF.Hi())
		gb.CPU.HL.Set(gb.CPU.HL.HiLo() - 1)

	// LDI A,(HL)
	case 0x2A:
		val := gb.Memory.Read(gb.CPU.HL.HiLo())
		gb.CPU.AF.SetHi(val)
		gb.CPU.HL.Set(gb.CPU.HL.HiLo() + 1)

	// LDI (HL),A
	case 0x22:
		val := gb.CPU.HL.HiLo()
		gb.Memory.Write(val, gb.CPU.AF.Hi())
		gb.CPU.HL.Set(gb.CPU.HL.HiLo() + 1)

	// LD (0xFF00+n),A
	case 0xE0:
		val := 0xFF00 + uint16(gb.popPC())
		gb.Memory.Write(val, gb.CPU.AF.Hi())

	// LD A,(0xFF00+n)
	case 0xF0:
		val := gb.Memory.Read(0xFF00 + uint16(gb.popPC()))
		gb.CPU.AF.SetHi(val)

	// ========== 16-Bit Loads ===========

	// LD BC,nn
	case 0x01:
		val := gb.popPC16()
		gb.CPU.BC.Set(val)

	// LD DE,nn
	case 0x11:
		val := gb.popPC16()
		gb.CPU.DE.Set(val)

	// LD HL,nn
	case 0x21:
		val := gb.popPC16()
		gb.CPU.HL.Set(val)

	// LD SP,nn
	case 0x31:
		val := gb.popPC16()
		gb.CPU.SP.Set(val)

	// LD SP,HL
	case 0xF9:
		val := gb.CPU.HL
		gb.CPU.SP = val

	// LD HL,SP+n
	case 0xF8:
		val1 := int32(gb.CPU.SP.HiLo())
		val2 := int32(int8(gb.popPC()))
		result := val1 + val2
		gb.CPU.HL.Set(uint16(result))
		tempVal := val1 ^ val2 ^ result
		gb.CPU.SetZ(false)
		gb.CPU.SetN(false)
		// TODO: Probably chec ktehse
		gb.CPU.SetH((tempVal & 0x10) == 0x10)
		gb.CPU.SetC((tempVal & 0x100) == 0x100)

	// LD (nn),SP
	case 0x08:
		address := gb.popPC16()
		gb.Memory.Write(address, gb.CPU.SP.Lo())
		gb.Memory.Write(address+1, gb.CPU.SP.Hi())

	// PUSH AF
	case 0xF5:
		gb.pushStack(gb.CPU.AF.HiLo())

	// PUSH BC
	case 0xC5:
		gb.pushStack(gb.CPU.BC.HiLo())

	// PUSH DE
	case 0xD5:
		gb.pushStack(gb.CPU.DE.HiLo())

	// PUSH HL
	case 0xE5:
		gb.pushStack(gb.CPU.HL.HiLo())

	// POP AF
	case 0xF1:
		gb.CPU.AF.Set(gb.popStack())

	// POP BC
	case 0xC1:
		gb.CPU.BC.Set(gb.popStack())

	// POP DE
	case 0xD1:
		gb.CPU.DE.Set(gb.popStack())

	// POP HL
	case 0xE1:
		gb.CPU.HL.Set(gb.popStack())

	// ========== 8-Bit ALU ===========

	// ADD A,A
	case 0x87:
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi(), false)

	// ADD A,B
	case 0x80:
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.BC.Hi(), gb.CPU.AF.Hi(), false)

	// ADD A,C
	case 0x81:
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.BC.Lo(), gb.CPU.AF.Hi(), false)

	// ADD A,D
	case 0x82:
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.DE.Hi(), gb.CPU.AF.Hi(), false)

	// ADD A,E
	case 0x83:
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.DE.Lo(), gb.CPU.AF.Hi(), false)

	// ADD A,H
	case 0x84:
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.HL.Hi(), gb.CPU.AF.Hi(), false)

	// ADD A,L
	case 0x85:
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.HL.Lo(), gb.CPU.AF.Hi(), false)

	// ADD A,(HL)
	case 0x86:
		gb.instAdd(gb.CPU.AF.SetHi, gb.Memory.Read(gb.CPU.HL.HiLo()), gb.CPU.AF.Hi(), false)

	// ADD A,#
	case 0xC6:
		gb.instAdd(gb.CPU.AF.SetHi, gb.popPC(), gb.CPU.AF.Hi(), false)

	// ADC A,A
	case 0x8F:
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi(), true)

	// ADC A,B
	case 0x88:
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.BC.Hi(), gb.CPU.AF.Hi(), true)

	// ADC A,C
	case 0x89:
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.BC.Lo(), gb.CPU.AF.Hi(), true)

	// ADC A,D
	case 0x8A:
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.DE.Hi(), gb.CPU.AF.Hi(), true)

	// ADC A,E
	case 0x8B:
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.DE.Lo(), gb.CPU.AF.Hi(), true)

	// ADC A,H
	case 0x8C:
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.HL.Hi(), gb.CPU.AF.Hi(), true)

	// ADC A,L
	case 0x8D:
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.HL.Lo(), gb.CPU.AF.Hi(), true)

	// ADC A,(HL)
	case 0x8E:
		gb.instAdd(gb.CPU.AF.SetHi, gb.Memory.Read(gb.CPU.HL.HiLo()), gb.CPU.AF.Hi(), true)

	// ADC A,#
	case 0xCE:
		gb.instAdd(gb.CPU.AF.SetHi, gb.popPC(), gb.CPU.AF.Hi(), true)

	// SUB A,A
	case 0x97:
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi(), false)

	// SUB A,B
	case 0x90:
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.BC.Hi(), false)

	// SUB A,C
	case 0x91:
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.BC.Lo(), false)

	// SUB A,D
	case 0x92:
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.DE.Hi(), false)

	// SUB A,E
	case 0x93:
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.DE.Lo(), false)

	// SUB A,H
	case 0x94:
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.HL.Hi(), false)

	// SUB A,L
	case 0x95:
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.HL.Lo(), false)

	// SUB A,(HL)
	case 0x96:
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.Memory.Read(gb.CPU.HL.HiLo()), false)

	// SUB A,#
	case 0xD6:
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.popPC(), false)

	// SBC A,A
	case 0x9F:
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi(), true)

	// SBC A,B
	case 0x98:
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.BC.Hi(), true)

	// SBC A,C
	case 0x99:
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.BC.Lo(), true)

	// SBC A,D
	case 0x9A:
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.DE.Hi(), true)

	// SBC A,E
	case 0x9B:
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.DE.Lo(), true)

	// SBC A,H
	case 0x9C:
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.HL.Hi(), true)

	// SBC A,L
	case 0x9D:
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.HL.Lo(), true)

	// SBC A,(HL)
	case 0x9E:
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.Memory.Read(gb.CPU.HL.HiLo()), true)

	// SBC A,#
	case 0xDE:
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.popPC(), true)

	// AND A,A
	case 0xA7:
		gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi())

	// AND A,B
	case 0xA0:
		gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.BC.Hi(), gb.CPU.AF.Hi())

	// AND A,C
	case 0xA1:
		gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.BC.Lo(), gb.CPU.AF.Hi())

	// AND A,D
	case 0xA2:
		gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.DE.Hi(), gb.CPU.AF.Hi())

	// AND A,E
	case 0xA3:
		gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.DE.Lo(), gb.CPU.AF.Hi())

	// AND A,H
	case 0xA4:
		gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.HL.Hi(), gb.CPU.AF.Hi())

	// AND A,L
	case 0xA5:
		gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.HL.Lo(), gb.CPU.AF.Hi())

	// AND A,(HL)
	case 0xA6:
		gb.instAnd(gb.CPU.AF.SetHi, gb.Memory.Read(gb.CPU.HL.HiLo()), gb.CPU.AF.Hi())

	// AND A,#
	case 0xE6:
		gb.instAnd(gb.CPU.AF.SetHi, gb.popPC(), gb.CPU.AF.Hi())

	// OR A,A
	case 0xB7:
		gb.instOr(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi())

	// OR A,B
	case 0xB0:
		gb.instOr(gb.CPU.AF.SetHi, gb.CPU.BC.Hi(), gb.CPU.AF.Hi())

	// OR A,C
	case 0xB1:
		gb.instOr(gb.CPU.AF.SetHi, gb.CPU.BC.Lo(), gb.CPU.AF.Hi())

	// OR A,D
	case 0xB2:
		gb.instOr(gb.CPU.AF.SetHi, gb.CPU.DE.Hi(), gb.CPU.AF.Hi())

	// OR A,E
	case 0xB3:
		gb.instOr(gb.CPU.AF.SetHi, gb.CPU.DE.Lo(), gb.CPU.AF.Hi())

	// OR A,H
	case 0xB4:
		gb.instOr(gb.CPU.AF.SetHi, gb.CPU.HL.Hi(), gb.CPU.AF.Hi())

	// OR A,L
	case 0xB5:
		gb.instOr(gb.CPU.AF.SetHi, gb.CPU.HL.Lo(), gb.CPU.AF.Hi())

	// OR A,(HL)
	case 0xB6:
		gb.instOr(gb.CPU.AF.SetHi, gb.Memory.Read(gb.CPU.HL.HiLo()), gb.CPU.AF.Hi())

	// OR A,#
	case 0xF6:
		gb.instOr(gb.CPU.AF.SetHi, gb.popPC(), gb.CPU.AF.Hi())

	// XOR A,A
	case 0xAF:
		gb.instXor(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi())

	// XOR A,B
	case 0xA8:
		gb.instXor(gb.CPU.AF.SetHi, gb.CPU.BC.Hi(), gb.CPU.AF.Hi())

	// XOR A,C
	case 0xA9:
		gb.instXor(gb.CPU.AF.SetHi, gb.CPU.BC.Lo(), gb.CPU.AF.Hi())

	// XOR A,D
	case 0xAA:
		gb.instXor(gb.CPU.AF.SetHi, gb.CPU.DE.Hi(), gb.CPU.AF.Hi())

	// XOR A,E
	case 0xAB:
		gb.instXor(gb.CPU.AF.SetHi, gb.CPU.DE.Lo(), gb.CPU.AF.Hi())

	// XOR A,H
	case 0xAC:
		gb.instXor(gb.CPU.AF.SetHi, gb.CPU.HL.Hi(), gb.CPU.AF.Hi())

	// XOR A,L
	case 0xAD:
		gb.instXor(gb.CPU.AF.SetHi, gb.CPU.HL.Lo(), gb.CPU.AF.Hi())

	// XOR A,(HL)
	case 0xAE:
		gb.instXor(gb.CPU.AF.SetHi, gb.Memory.Read(gb.CPU.HL.HiLo()), gb.CPU.AF.Hi())

	// XOR A,#
	case 0xEE:
		gb.instXor(gb.CPU.AF.SetHi, gb.popPC(), gb.CPU.AF.Hi())

	// CP A,A
	case 0xBF:
		gb.instCp(gb.CPU.AF.Hi(), gb.CPU.AF.Hi())

	// CP A,B
	case 0xB8:
		gb.instCp(gb.CPU.BC.Hi(), gb.CPU.AF.Hi())

	// CP A,C
	case 0xB9:
		gb.instCp(gb.CPU.BC.Lo(), gb.CPU.AF.Hi())

	// CP A,D
	case 0xBA:
		gb.instCp(gb.CPU.DE.Hi(), gb.CPU.AF.Hi())

	// CP A,E
	case 0xBB:
		gb.instCp(gb.CPU.DE.Lo(), gb.CPU.AF.Hi())

	// CP A,H
	case 0xBC:
		gb.instCp(gb.CPU.HL.Hi(), gb.CPU.AF.Hi())

	// CP A,L
	case 0xBD:
		gb.instCp(gb.CPU.HL.Lo(), gb.CPU.AF.Hi())

	// CP A,(HL)
	case 0xBE:
		gb.instCp(gb.Memory.Read(gb.CPU.HL.HiLo()), gb.CPU.AF.Hi())

	// CP A,#
	case 0xFE:
		gb.instCp(gb.popPC(), gb.CPU.AF.Hi())

	// INC A
	case 0x3C:
		gb.instInc(gb.CPU.AF.SetHi, gb.CPU.AF.Hi())

	// INC B
	case 0x04:
		gb.instInc(gb.CPU.BC.SetHi, gb.CPU.BC.Hi())

	// INC C
	case 0x0C:
		gb.instInc(gb.CPU.BC.SetLo, gb.CPU.BC.Lo())

	// INC D
	case 0x14:
		gb.instInc(gb.CPU.DE.SetHi, gb.CPU.DE.Hi())

	// INC E
	case 0x1C:
		gb.instInc(gb.CPU.DE.SetLo, gb.CPU.DE.Lo())

	// INC H
	case 0x24:
		gb.instInc(gb.CPU.HL.SetHi, gb.CPU.HL.Hi())

	// INC L
	case 0x2C:
		gb.instInc(gb.CPU.HL.SetLo, gb.CPU.HL.Lo())

	// INC (HL)
	case 0x34:
		addr := gb.CPU.HL.HiLo()
		gb.instInc(func(val byte) { gb.Memory.Write(addr, val) }, gb.Memory.Read(addr))

	// DEC A
	case 0x3D:
		gb.instDec(gb.CPU.AF.SetHi, gb.CPU.AF.Hi())

	// DEC B
	case 0x05:
		gb.instDec(gb.CPU.BC.SetHi, gb.CPU.BC.Hi())

	// DEC C
	case 0x0D:
		gb.instDec(gb.CPU.BC.SetLo, gb.CPU.BC.Lo())

	// DEC D
	case 0x15:
		gb.instDec(gb.CPU.DE.SetHi, gb.CPU.DE.Hi())

	// DEC E
	case 0x1D:
		gb.instDec(gb.CPU.DE.SetLo, gb.CPU.DE.Lo())

	// DEC H
	case 0x25:
		gb.instDec(gb.CPU.HL.SetHi, gb.CPU.HL.Hi())

	// DEC L
	case 0x2D:
		gb.instDec(gb.CPU.HL.SetLo, gb.CPU.HL.Lo())

	// DEC (HL)
	case 0x35:
		addr := gb.CPU.HL.HiLo()
		gb.instDec(func(val byte) { gb.Memory.Write(addr, val) }, gb.Memory.Read(addr))

	// ========== 16-Bit ALU ===========

	// ADD HL,BC
	case 0x09:
		gb.instAdd16(gb.CPU.HL.Set, gb.CPU.HL.HiLo(), gb.CPU.BC.HiLo())

	// ADD HL,DE
	case 0x19:
		gb.instAdd16(gb.CPU.HL.Set, gb.CPU.HL.HiLo(), gb.CPU.DE.HiLo())

	// ADD HL,HL
	case 0x29:
		gb.instAdd16(gb.CPU.HL.Set, gb.CPU.HL.HiLo(), gb.CPU.HL.HiLo())

	// ADD HL,SP
	case 0x39:
		gb.instAdd16(gb.CPU.HL.Set, gb.CPU.HL.HiLo(), gb.CPU.SP.HiLo())

	// ADD SP,n
	case 0xE8:
		gb.instAdd16Signed(gb.CPU.SP.Set, gb.CPU.SP.HiLo(), int8(gb.popPC()))
		gb.CPU.SetZ(false)

	// INC BC
	case 0x03:
		gb.instInc16(gb.CPU.BC.Set, gb.CPU.BC.HiLo())

	// INC DE
	case 0x13:
		gb.instInc16(gb.CPU.DE.Set, gb.CPU.DE.HiLo())

	// INC HL
	case 0x23:
		gb.instInc16(gb.CPU.HL.Set, gb.CPU.HL.HiLo())

	// INC SP
	case 0x33:
		gb.instInc16(gb.CPU.SP.Set, gb.CPU.SP.HiLo())

	// DEC BC
	case 0x0B:
		gb.instDec16(gb.CPU.BC.Set, gb.CPU.BC.HiLo())

	// DEC DE
	case 0x1B:
		gb.instDec16(gb.CPU.DE.Set, gb.CPU.DE.HiLo())

	// DEC HL
	case 0x2B:
		gb.instDec16(gb.CPU.HL.Set, gb.CPU.HL.HiLo())

	// DEC SP
	case 0x3B:
		gb.instDec16(gb.CPU.SP.Set, gb.CPU.SP.HiLo())

	// DAA
	case 0x27:
		/*
			When this instruction is executed, the A register is BCD
			corrected using the contents of the flags. The exact process
			is the following: if the least significant four bits of A
			contain a non-BCD digit (i. e. it is greater than 9) or the
			H flag is set, then $06 is added to the register. Then the
			four most significant bits are checked. If this more significant
			digit also happens to be greater than 9 or the C flag is set,
			then $60 is added.
		*/
		// TODO: This could be more efficient?
		if !gb.CPU.N() {
			if gb.CPU.C() || gb.CPU.AF.Hi() > 0x99 {
				gb.CPU.AF.SetHi(gb.CPU.AF.Hi() + 0x60)
				gb.CPU.SetC(true)
			}
			if gb.CPU.H() || gb.CPU.AF.Hi()&0xF > 0x9 {
				gb.CPU.AF.SetHi(gb.CPU.AF.Hi() + 0x06)
				gb.CPU.SetH(false)
			}
		} else if gb.CPU.C() && gb.CPU.H() {
			gb.CPU.AF.SetHi(gb.CPU.AF.Hi() + 0x9A)
			gb.CPU.SetH(false)
		} else if gb.CPU.C() {
			gb.CPU.AF.SetHi(gb.CPU.AF.Hi() + 0xA0)
		} else if gb.CPU.H() {
			gb.CPU.AF.SetHi(gb.CPU.AF.Hi() + 0xFA)
			gb.CPU.SetH(false)
		}
		gb.CPU.SetZ(gb.CPU.AF.Hi() == 0)

	// CPL
	case 0x2F:
		gb.CPU.AF.SetHi(0xFF ^ gb.CPU.AF.Hi())
		gb.CPU.SetN(true)
		gb.CPU.SetH(true)

	// CCF
	case 0x3F:
		gb.CPU.SetN(false)
		gb.CPU.SetH(false)
		gb.CPU.SetC(!gb.CPU.C())

	// SCF
	case 0x37:
		gb.CPU.SetN(false)
		gb.CPU.SetH(false)
		gb.CPU.SetC(true)

	// NOP
	case 0x00:
		break

	// HALT
	case 0x76:
		gb.Halted = true

	// STOP
	case 0x10:
		//gb.Halted = true
		log.Print("0x10 (STOP) unimplemented (is 0x00 follows)")

	// DI
	case 0xF3:
		gb.InterruptsOn = false

	// EI
	case 0xFB:
		gb.InterruptsEnabling = true

	// RLCA
	case 0x07:
		value := gb.CPU.AF.Hi()
		result := byte(value<<1) | (value >> 7)
		gb.CPU.AF.SetHi(result)
		gb.CPU.SetZ(false)
		gb.CPU.SetN(false)
		gb.CPU.SetH(false)
		gb.CPU.SetC(value > 0x7F)

	// RLA
	case 0x17:
		value := gb.CPU.AF.Hi()
		var carry byte = 0
		if gb.CPU.C() {
			carry = 1
		}
		result := byte(value<<1) + carry
		gb.CPU.AF.SetHi(result)
		gb.CPU.SetZ(false)
		gb.CPU.SetN(false)
		gb.CPU.SetH(false)
		gb.CPU.SetC(value > 0x7F)

	// RRCA
	case 0x0F:
		value := gb.CPU.AF.Hi()
		result := byte(value>>1) | byte((value&1)<<7)
		gb.CPU.AF.SetHi(result)
		gb.CPU.SetZ(false)
		gb.CPU.SetN(false)
		gb.CPU.SetH(false)
		gb.CPU.SetC(result > 0x7F)

	// RRA
	case 0x1F:
		value := gb.CPU.AF.Hi()
		var carry byte = 0
		if gb.CPU.C() {
			carry = 0x80
		}
		result := byte(value>>1) | carry
		gb.CPU.AF.SetHi(result)
		gb.CPU.SetZ(false)
		gb.CPU.SetN(false)
		gb.CPU.SetH(false)
		gb.CPU.SetC((1 & value) == 1)

	// JP nn
	case 0xC3:
		gb.instJump(gb.popPC16())

	// JP NZ,nn
	case 0xC2:
		next := gb.popPC16()
		if !gb.CPU.Z() {
			gb.instJump(next)
			gb.thisCpuTicks += 4
		}

	// JP Z,nn
	case 0xCA:
		next := gb.popPC16()
		if gb.CPU.Z() {
			gb.instJump(next)
			gb.thisCpuTicks += 4
		}

	// JP NC,nn
	case 0xD2:
		next := gb.popPC16()
		if !gb.CPU.C() {
			gb.instJump(next)
			gb.thisCpuTicks += 4
		}

	// JP C,nn
	case 0xDA:
		next := gb.popPC16()
		if gb.CPU.C() {
			gb.instJump(next)
			gb.thisCpuTicks += 4
		}

	// JP HL
	case 0xE9:
		gb.instJump(gb.CPU.HL.HiLo())

	// JR n
	case 0x18:
		addr := int32(gb.CPU.PC) + int32(int8(gb.popPC()))
		gb.instJump(uint16(addr))

	// JR NZ,n
	case 0x20:
		next := int8(gb.popPC())
		if !gb.CPU.Z() {
			addr := int32(gb.CPU.PC) + int32(next)
			gb.instJump(uint16(addr))
			gb.thisCpuTicks += 4
		}

	// JR Z,n
	case 0x28:
		next := int8(gb.popPC())
		if gb.CPU.Z() {
			addr := int32(gb.CPU.PC) + int32(next)
			gb.instJump(uint16(addr))
			gb.thisCpuTicks += 4
		}

	// JR NC,n
	case 0x30:
		next := int8(gb.popPC())
		if !gb.CPU.C() {
			addr := int32(gb.CPU.PC) + int32(next)
			gb.instJump(uint16(addr))
			gb.thisCpuTicks += 4
		}

	// JR C,n
	case 0x38:
		next := int8(gb.popPC())
		if gb.CPU.C() {
			addr := int32(gb.CPU.PC) + int32(next)
			gb.instJump(uint16(addr))
			gb.thisCpuTicks += 4
		}

	// CALL nn
	case 0xCD:
		gb.instCall(gb.popPC16())

	// CALL NZ,nn
	case 0xC4:
		next := gb.popPC16()
		if !gb.CPU.Z() {
			gb.instCall(next)
			gb.thisCpuTicks += 12
		}

	// CALL Z,nn
	case 0xCC:
		next := gb.popPC16()
		if gb.CPU.Z() {
			gb.instCall(next)
			gb.thisCpuTicks += 12
		}

	// CALL NC,nn
	case 0xD4:
		next := gb.popPC16()
		if !gb.CPU.C() {
			gb.instCall(next)
			gb.thisCpuTicks += 12
		}

	// CALL C,nn
	case 0xDC:
		next := gb.popPC16()
		if gb.CPU.C() {
			gb.instCall(next)
			gb.thisCpuTicks += 12
		}

	// RST 0x00
	case 0xC7:
		gb.instCall(0x0000)

	// RST 0x08
	case 0xCF:
		gb.instCall(0x0008)

	// RST 0x10
	case 0xD7:
		gb.instCall(0x0010)

	// RST 0x18
	case 0xDF:
		gb.instCall(0x0018)

	// RST 0x20
	case 0xE7:
		gb.instCall(0x0020)

	// RST 0x28
	case 0xEF:
		gb.instCall(0x0028)

	// RST 0x30
	case 0xF7:
		gb.instCall(0x0030)

	// RST 0x38
	case 0xFF:
		gb.instCall(0x0038)

	// RET
	case 0xC9:
		gb.instRet()

	// RET NZ
	case 0xC0:
		if !gb.CPU.Z() {
			gb.instRet()
			gb.thisCpuTicks += 12
		}

	// RET Z
	case 0xC8:
		if gb.CPU.Z() {
			gb.instRet()
			gb.thisCpuTicks += 12
		}

	// RET NC
	case 0xD0:
		if !gb.CPU.C() {
			gb.instRet()
			gb.thisCpuTicks += 12
		}

	// RET C
	case 0xD8:
		if gb.CPU.C() {
			gb.instRet()
			gb.thisCpuTicks += 12
		}

	// RETI
	case 0xD9:
		gb.instRet()
		gb.InterruptsOn = true

	// CB!
	case 0xCB:
		nextInst := gb.popPC()
		gb.thisCpuTicks += CBOpcodeCycles[nextInst] * 4
		gb.CBInst[nextInst]()

	default:
		log.Printf("Unimplemented opcode: %#2x", opcode)
		waitForInput()
	}
}
