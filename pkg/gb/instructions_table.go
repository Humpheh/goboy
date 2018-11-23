package gb

import (
	"log"
)

// mainInstructions creates and  returns a table of the main
// insruction set
func (gb *Gameboy) mainInstructions() [0x100]func() {
	// TODO: possibly faster if we derefernce
	// each register and gb methods in this scope

	// ret gets returned
	ret := [0x100]func(){
		0x06: func() {
			// LD B, n
			gb.CPU.BC.SetHi(gb.popPC())

		},
		0x0E: func() {
			// LD C, n
			gb.CPU.BC.SetLo(gb.popPC())

		},
		0x16: func() {
			// LD D, n
			gb.CPU.DE.SetHi(gb.popPC())

		},
		0x1E: func() {
			// LD E, n
			gb.CPU.DE.SetLo(gb.popPC())

		},
		0x26: func() {
			// LD H, n
			gb.CPU.HL.SetHi(gb.popPC())

		},
		0x2E: func() {
			// LD L, n
			gb.CPU.HL.SetLo(gb.popPC())

		},
		0x7F: func() {
			// LD A,A
			gb.CPU.AF.SetHi(gb.CPU.AF.Hi())

		},
		0x78: func() {
			// LD A,B
			gb.CPU.AF.SetHi(gb.CPU.BC.Hi())

		},
		0x79: func() {
			// LD A,C
			gb.CPU.AF.SetHi(gb.CPU.BC.Lo())

		},
		0x7A: func() {
			// LD A,D
			gb.CPU.AF.SetHi(gb.CPU.DE.Hi())

		},
		0x7B: func() {
			// LD A,E
			gb.CPU.AF.SetHi(gb.CPU.DE.Lo())

		},
		0x7C: func() {
			// LD A,H
			gb.CPU.AF.SetHi(gb.CPU.HL.Hi())

		},
		0x7D: func() {
			// LD A,L
			gb.CPU.AF.SetHi(gb.CPU.HL.Lo())

		},
		0x0A: func() {
			// LD A,(BC)
			val := gb.Memory.Read(gb.CPU.BC.HiLo())
			gb.CPU.AF.SetHi(val)

		},
		0x1A: func() {
			// LD A,(DE)
			val := gb.Memory.Read(gb.CPU.DE.HiLo())
			gb.CPU.AF.SetHi(val)

		},
		0x7E: func() {
			// LD A,(HL)
			val := gb.Memory.Read(gb.CPU.HL.HiLo())
			gb.CPU.AF.SetHi(val)

		},
		0xFA: func() {
			// LD A,(nn)
			val := gb.Memory.Read(gb.popPC16())
			gb.CPU.AF.SetHi(val)

		},
		0x3E: func() {
			// LD A,(nn)
			val := gb.popPC()
			gb.CPU.AF.SetHi(val)

		},
		0x47: func() {
			// LD B,A
			gb.CPU.BC.SetHi(gb.CPU.AF.Hi())

		},
		0x40: func() {
			// LD B,B
			gb.CPU.BC.SetHi(gb.CPU.BC.Hi())

		},
		0x41: func() {
			// LD B,C
			gb.CPU.BC.SetHi(gb.CPU.BC.Lo())

		},
		0x42: func() {
			// LD B,D
			gb.CPU.BC.SetHi(gb.CPU.DE.Hi())

		},
		0x43: func() {
			// LD B,E
			gb.CPU.BC.SetHi(gb.CPU.DE.Lo())

		},
		0x44: func() {
			// LD B,H
			gb.CPU.BC.SetHi(gb.CPU.HL.Hi())

		},
		0x45: func() {
			// LD B,L
			gb.CPU.BC.SetHi(gb.CPU.HL.Lo())

		},
		0x46: func() {
			// LD B,(HL)
			val := gb.Memory.Read(gb.CPU.HL.HiLo())
			gb.CPU.BC.SetHi(val)

		},
		0x4F: func() {
			// LD C,A
			gb.CPU.BC.SetLo(gb.CPU.AF.Hi())

		},
		0x48: func() {
			// LD C,B
			gb.CPU.BC.SetLo(gb.CPU.BC.Hi())

		},
		0x49: func() {
			// LD C,C
			gb.CPU.BC.SetLo(gb.CPU.BC.Lo())

		},
		0x4A: func() {
			// LD C,D
			gb.CPU.BC.SetLo(gb.CPU.DE.Hi())

		},
		0x4B: func() {
			// LD C,E
			gb.CPU.BC.SetLo(gb.CPU.DE.Lo())

		},
		0x4C: func() {
			// LD C,H
			gb.CPU.BC.SetLo(gb.CPU.HL.Hi())

		},
		0x4D: func() {
			// LD C,L
			gb.CPU.BC.SetLo(gb.CPU.HL.Lo())

		},
		0x4E: func() {
			// LD C,(HL)
			val := gb.Memory.Read(gb.CPU.HL.HiLo())
			gb.CPU.BC.SetLo(val)

		},
		0x57: func() {
			// LD D,A
			gb.CPU.DE.SetHi(gb.CPU.AF.Hi())

		},
		0x50: func() {
			// LD D,B
			gb.CPU.DE.SetHi(gb.CPU.BC.Hi())

		},
		0x51: func() {
			// LD D,C
			gb.CPU.DE.SetHi(gb.CPU.BC.Lo())

		},
		0x52: func() {
			// LD D,D
			gb.CPU.DE.SetHi(gb.CPU.DE.Hi())

		},
		0x53: func() {
			// LD D,E
			gb.CPU.DE.SetHi(gb.CPU.DE.Lo())

		},
		0x54: func() {
			// LD D,H
			gb.CPU.DE.SetHi(gb.CPU.HL.Hi())

		},
		0x55: func() {
			// LD D,L
			gb.CPU.DE.SetHi(gb.CPU.HL.Lo())

		},
		0x56: func() {
			// LD D,(HL)
			val := gb.Memory.Read(gb.CPU.HL.HiLo())
			gb.CPU.DE.SetHi(val)

		},
		0x5F: func() {
			// LD E,A
			gb.CPU.DE.SetLo(gb.CPU.AF.Hi())

		},
		0x58: func() {
			// LD E,B
			gb.CPU.DE.SetLo(gb.CPU.BC.Hi())

		},
		0x59: func() {
			// LD E,C
			gb.CPU.DE.SetLo(gb.CPU.BC.Lo())

		},
		0x5A: func() {
			// LD E,D
			gb.CPU.DE.SetLo(gb.CPU.DE.Hi())

		},
		0x5B: func() {
			// LD E,E
			gb.CPU.DE.SetLo(gb.CPU.DE.Lo())

		},
		0x5C: func() {
			// LD E,H
			gb.CPU.DE.SetLo(gb.CPU.HL.Hi())

		},
		0x5D: func() {
			// LD E,L
			gb.CPU.DE.SetLo(gb.CPU.HL.Lo())

		},
		0x5E: func() {
			// LD E,(HL)
			val := gb.Memory.Read(gb.CPU.HL.HiLo())
			gb.CPU.DE.SetLo(val)

		},
		0x67: func() {
			// LD H,A
			gb.CPU.HL.SetHi(gb.CPU.AF.Hi())

		},
		0x60: func() {
			// LD H,B
			gb.CPU.HL.SetHi(gb.CPU.BC.Hi())

		},
		0x61: func() {
			// LD H,C
			gb.CPU.HL.SetHi(gb.CPU.BC.Lo())

		},
		0x62: func() {
			// LD H,D
			gb.CPU.HL.SetHi(gb.CPU.DE.Hi())

		},
		0x63: func() {
			// LD H,E
			gb.CPU.HL.SetHi(gb.CPU.DE.Lo())

		},
		0x64: func() {
			// LD H,H
			gb.CPU.HL.SetHi(gb.CPU.HL.Hi())

		},
		0x65: func() {
			// LD H,L
			gb.CPU.HL.SetHi(gb.CPU.HL.Lo())

		},
		0x66: func() {
			// LD H,(HL)
			val := gb.Memory.Read(gb.CPU.HL.HiLo())
			gb.CPU.HL.SetHi(val)

		},
		0x6F: func() {
			// LD L,A
			gb.CPU.HL.SetLo(gb.CPU.AF.Hi())

		},
		0x68: func() {
			// LD L,B
			gb.CPU.HL.SetLo(gb.CPU.BC.Hi())

		},
		0x69: func() {
			// LD L,C
			gb.CPU.HL.SetLo(gb.CPU.BC.Lo())

		},
		0x6A: func() {
			// LD L,D
			gb.CPU.HL.SetLo(gb.CPU.DE.Hi())

		},
		0x6B: func() {
			// LD L,E
			gb.CPU.HL.SetLo(gb.CPU.DE.Lo())

		},
		0x6C: func() {
			// LD L,H
			gb.CPU.HL.SetLo(gb.CPU.HL.Hi())

		},
		0x6D: func() {
			// LD L,L
			gb.CPU.HL.SetLo(gb.CPU.HL.Lo())

		},
		0x6E: func() {
			// LD L,(HL)
			val := gb.Memory.Read(gb.CPU.HL.HiLo())
			gb.CPU.HL.SetLo(val)

		},
		0x77: func() {
			// LD (HL),A
			val := gb.CPU.AF.Hi()
			gb.Memory.Write(gb.CPU.HL.HiLo(), val)

		},
		0x70: func() {
			// LD (HL),B
			val := gb.CPU.BC.Hi()
			gb.Memory.Write(gb.CPU.HL.HiLo(), val)

		},
		0x71: func() {
			// LD (HL),C
			val := gb.CPU.BC.Lo()
			gb.Memory.Write(gb.CPU.HL.HiLo(), val)

		},
		0x72: func() {
			// LD (HL),D
			val := gb.CPU.DE.Hi()
			gb.Memory.Write(gb.CPU.HL.HiLo(), val)

		},
		0x73: func() {
			// LD (HL),E
			val := gb.CPU.DE.Lo()
			gb.Memory.Write(gb.CPU.HL.HiLo(), val)

		},
		0x74: func() {
			// LD (HL),H
			val := gb.CPU.HL.Hi()
			gb.Memory.Write(gb.CPU.HL.HiLo(), val)

		},
		0x75: func() {
			// LD (HL),L
			val := gb.CPU.HL.Lo()
			gb.Memory.Write(gb.CPU.HL.HiLo(), val)

		},
		0x36: func() {
			// LD (HL),n 36
			val := gb.popPC()
			gb.Memory.Write(gb.CPU.HL.HiLo(), val)

		},
		0x02: func() {
			// LD (BC),A
			val := gb.CPU.AF.Hi()
			gb.Memory.Write(gb.CPU.BC.HiLo(), val)

		},
		0x12: func() {
			// LD (DE),A
			val := gb.CPU.AF.Hi()
			gb.Memory.Write(gb.CPU.DE.HiLo(), val)

		},
		0xEA: func() {
			// LD (nn),A
			val := gb.CPU.AF.Hi()
			gb.Memory.Write(gb.popPC16(), val)

		},
		0xF2: func() {
			// LD A,(C)
			val := 0xFF00 + uint16(gb.CPU.BC.Lo())
			gb.CPU.AF.SetHi(gb.Memory.Read(val))

		},
		0xE2: func() {
			// LD (C),A
			val := gb.CPU.AF.Hi()
			mem := 0xFF00 + uint16(gb.CPU.BC.Lo())
			gb.Memory.Write(mem, val)

		},
		0x3A: func() {
			// LDD A,(HL)
			val := gb.Memory.Read(gb.CPU.HL.HiLo())
			gb.CPU.AF.SetHi(val)
			gb.CPU.HL.Set(gb.CPU.HL.HiLo() - 1)

		},
		0x32: func() {
			// LDD (HL),A
			val := gb.CPU.HL.HiLo()
			gb.Memory.Write(val, gb.CPU.AF.Hi())
			gb.CPU.HL.Set(gb.CPU.HL.HiLo() - 1)

		},
		0x2A: func() {
			// LDI A,(HL)
			val := gb.Memory.Read(gb.CPU.HL.HiLo())
			gb.CPU.AF.SetHi(val)
			gb.CPU.HL.Set(gb.CPU.HL.HiLo() + 1)

		},
		0x22: func() {
			// LDI (HL),A
			val := gb.CPU.HL.HiLo()
			gb.Memory.Write(val, gb.CPU.AF.Hi())
			gb.CPU.HL.Set(gb.CPU.HL.HiLo() + 1)

		},
		0xE0: func() {
			// LD (0xFF00+n),A
			val := 0xFF00 + uint16(gb.popPC())
			gb.Memory.Write(val, gb.CPU.AF.Hi())

		},
		0xF0: func() {
			// LD A,(0xFF00+n)
			val := gb.Memory.ReadHighRam(0xFF00 + uint16(gb.popPC()))
			gb.CPU.AF.SetHi(val)

			// ========== 16-Bit Loads ===========
		},
		0x01: func() {
			// LD BC,nn
			val := gb.popPC16()
			gb.CPU.BC.Set(val)

		},
		0x11: func() {
			// LD DE,nn
			val := gb.popPC16()
			gb.CPU.DE.Set(val)

		},
		0x21: func() {
			// LD HL,nn
			val := gb.popPC16()
			gb.CPU.HL.Set(val)

		},
		0x31: func() {
			// LD SP,nn
			val := gb.popPC16()
			gb.CPU.SP.Set(val)

		},
		0xF9: func() {
			// LD SP,HL
			val := gb.CPU.HL
			gb.CPU.SP = val

		},
		0xF8: func() {
			// LD HL,SP+n
			val1 := int32(gb.CPU.SP.HiLo())
			val2 := int32(int8(gb.popPC()))
			result := val1 + val2
			gb.CPU.HL.Set(uint16(result))
			tempVal := val1 ^ val2 ^ result
			gb.CPU.SetZ(false)
			gb.CPU.SetN(false)
			// TODO: Probably check these
			gb.CPU.SetH((tempVal & 0x10) == 0x10)
			gb.CPU.SetC((tempVal & 0x100) == 0x100)

		},
		0x08: func() {
			// LD (nn),SP
			address := gb.popPC16()
			gb.Memory.Write(address, gb.CPU.SP.Lo())
			gb.Memory.Write(address+1, gb.CPU.SP.Hi())

		},
		0xF5: func() {
			// PUSH AF
			gb.pushStack(gb.CPU.AF.HiLo())

		},
		0xC5: func() {
			// PUSH BC
			gb.pushStack(gb.CPU.BC.HiLo())

		},
		0xD5: func() {
			// PUSH DE
			gb.pushStack(gb.CPU.DE.HiLo())

		},
		0xE5: func() {
			// PUSH HL
			gb.pushStack(gb.CPU.HL.HiLo())

		},
		0xF1: func() {
			// POP AF
			gb.CPU.AF.Set(gb.popStack())

		},
		0xC1: func() {
			// POP BC
			gb.CPU.BC.Set(gb.popStack())

		},
		0xD1: func() {
			// POP DE
			gb.CPU.DE.Set(gb.popStack())

		},
		0xE1: func() {
			// POP HL
			gb.CPU.HL.Set(gb.popStack())

			// ========== 8-Bit ALU ===========
		},
		0x87: func() {
			// ADD A,A
			gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi(), false)

		},
		0x80: func() {
			// ADD A,B
			gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.BC.Hi(), gb.CPU.AF.Hi(), false)

		},
		0x81: func() {
			// ADD A,C
			gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.BC.Lo(), gb.CPU.AF.Hi(), false)

		},
		0x82: func() {
			// ADD A,D
			gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.DE.Hi(), gb.CPU.AF.Hi(), false)

		},
		0x83: func() {
			// ADD A,E
			gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.DE.Lo(), gb.CPU.AF.Hi(), false)

		},
		0x84: func() {
			// ADD A,H
			gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.HL.Hi(), gb.CPU.AF.Hi(), false)

		},
		0x85: func() {
			// ADD A,L
			gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.HL.Lo(), gb.CPU.AF.Hi(), false)

		},
		0x86: func() {
			// ADD A,(HL)
			gb.instAdd(gb.CPU.AF.SetHi, gb.Memory.Read(gb.CPU.HL.HiLo()), gb.CPU.AF.Hi(), false)

		},
		0xC6: func() {
			// ADD A,#
			gb.instAdd(gb.CPU.AF.SetHi, gb.popPC(), gb.CPU.AF.Hi(), false)

		},
		0x8F: func() {
			// ADC A,A
			gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi(), true)

		},
		0x88: func() {
			// ADC A,B
			gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.BC.Hi(), gb.CPU.AF.Hi(), true)

		},
		0x89: func() {
			// ADC A,C
			gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.BC.Lo(), gb.CPU.AF.Hi(), true)

		},
		0x8A: func() {
			// ADC A,D
			gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.DE.Hi(), gb.CPU.AF.Hi(), true)

		},
		0x8B: func() {
			// ADC A,E
			gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.DE.Lo(), gb.CPU.AF.Hi(), true)

		},
		0x8C: func() {
			// ADC A,H
			gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.HL.Hi(), gb.CPU.AF.Hi(), true)

		},
		0x8D: func() {
			// ADC A,L
			gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.HL.Lo(), gb.CPU.AF.Hi(), true)

		},
		0x8E: func() {
			// ADC A,(HL)
			gb.instAdd(gb.CPU.AF.SetHi, gb.Memory.Read(gb.CPU.HL.HiLo()), gb.CPU.AF.Hi(), true)

		},
		0xCE: func() {
			// ADC A,#
			gb.instAdd(gb.CPU.AF.SetHi, gb.popPC(), gb.CPU.AF.Hi(), true)

		},
		0x97: func() {
			// SUB A,A
			gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi(), false)

		},
		0x90: func() {
			// SUB A,B
			gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.BC.Hi(), false)

		},
		0x91: func() {
			// SUB A,C
			gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.BC.Lo(), false)

		},
		0x92: func() {
			// SUB A,D
			gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.DE.Hi(), false)

		},
		0x93: func() {
			// SUB A,E
			gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.DE.Lo(), false)

		},
		0x94: func() {
			// SUB A,H
			gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.HL.Hi(), false)

		},
		0x95: func() {
			// SUB A,L
			gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.HL.Lo(), false)

		},
		0x96: func() {
			// SUB A,(HL)
			gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.Memory.Read(gb.CPU.HL.HiLo()), false)

		},
		0xD6: func() {
			// SUB A,#
			gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.popPC(), false)

		},
		0x9F: func() {
			// SBC A,A
			gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi(), true)

		},
		0x98: func() {
			// SBC A,B
			gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.BC.Hi(), true)

		},
		0x99: func() {
			// SBC A,C
			gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.BC.Lo(), true)

		},
		0x9A: func() {
			// SBC A,D
			gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.DE.Hi(), true)

		},
		0x9B: func() {
			// SBC A,E
			gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.DE.Lo(), true)

		},
		0x9C: func() {
			// SBC A,H
			gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.HL.Hi(), true)

		},
		0x9D: func() {
			// SBC A,L
			gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.HL.Lo(), true)

		},
		0x9E: func() {
			// SBC A,(HL)
			gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.Memory.Read(gb.CPU.HL.HiLo()), true)

		},
		0xDE: func() {
			// SBC A,#
			gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.popPC(), true)

		},
		0xA7: func() {
			// AND A,A
			gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi())

		},
		0xA0: func() {
			// AND A,B
			gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.BC.Hi(), gb.CPU.AF.Hi())

		},
		0xA1: func() {
			// AND A,C
			gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.BC.Lo(), gb.CPU.AF.Hi())

		},
		0xA2: func() {
			// AND A,D
			gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.DE.Hi(), gb.CPU.AF.Hi())

		},
		0xA3: func() {
			// AND A,E
			gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.DE.Lo(), gb.CPU.AF.Hi())

		},
		0xA4: func() {
			// AND A,H
			gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.HL.Hi(), gb.CPU.AF.Hi())

		},
		0xA5: func() {
			// AND A,L
			gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.HL.Lo(), gb.CPU.AF.Hi())

		},
		0xA6: func() {
			// AND A,(HL)
			gb.instAnd(gb.CPU.AF.SetHi, gb.Memory.Read(gb.CPU.HL.HiLo()), gb.CPU.AF.Hi())

		},
		0xE6: func() {
			// AND A,#
			gb.instAnd(gb.CPU.AF.SetHi, gb.popPC(), gb.CPU.AF.Hi())

		},
		0xB7: func() {
			// OR A,A
			gb.instOr(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi())

		},
		0xB0: func() {
			// OR A,B
			gb.instOr(gb.CPU.AF.SetHi, gb.CPU.BC.Hi(), gb.CPU.AF.Hi())

		},
		0xB1: func() {
			// OR A,C
			gb.instOr(gb.CPU.AF.SetHi, gb.CPU.BC.Lo(), gb.CPU.AF.Hi())

		},
		0xB2: func() {
			// OR A,D
			gb.instOr(gb.CPU.AF.SetHi, gb.CPU.DE.Hi(), gb.CPU.AF.Hi())

		},
		0xB3: func() {
			// OR A,E
			gb.instOr(gb.CPU.AF.SetHi, gb.CPU.DE.Lo(), gb.CPU.AF.Hi())

		},
		0xB4: func() {
			// OR A,H
			gb.instOr(gb.CPU.AF.SetHi, gb.CPU.HL.Hi(), gb.CPU.AF.Hi())

		},
		0xB5: func() {
			// OR A,L
			gb.instOr(gb.CPU.AF.SetHi, gb.CPU.HL.Lo(), gb.CPU.AF.Hi())

		},
		0xB6: func() {
			// OR A,(HL)
			gb.instOr(gb.CPU.AF.SetHi, gb.Memory.Read(gb.CPU.HL.HiLo()), gb.CPU.AF.Hi())

		},
		0xF6: func() {
			// OR A,#
			gb.instOr(gb.CPU.AF.SetHi, gb.popPC(), gb.CPU.AF.Hi())

		},
		0xAF: func() {
			// XOR A,A
			gb.instXor(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi())

		},
		0xA8: func() {
			// XOR A,B
			gb.instXor(gb.CPU.AF.SetHi, gb.CPU.BC.Hi(), gb.CPU.AF.Hi())

		},
		0xA9: func() {
			// XOR A,C
			gb.instXor(gb.CPU.AF.SetHi, gb.CPU.BC.Lo(), gb.CPU.AF.Hi())

		},
		0xAA: func() {
			// XOR A,D
			gb.instXor(gb.CPU.AF.SetHi, gb.CPU.DE.Hi(), gb.CPU.AF.Hi())

		},
		0xAB: func() {
			// XOR A,E
			gb.instXor(gb.CPU.AF.SetHi, gb.CPU.DE.Lo(), gb.CPU.AF.Hi())

		},
		0xAC: func() {
			// XOR A,H
			gb.instXor(gb.CPU.AF.SetHi, gb.CPU.HL.Hi(), gb.CPU.AF.Hi())

		},
		0xAD: func() {
			// XOR A,L
			gb.instXor(gb.CPU.AF.SetHi, gb.CPU.HL.Lo(), gb.CPU.AF.Hi())

		},
		0xAE: func() {
			// XOR A,(HL)
			gb.instXor(gb.CPU.AF.SetHi, gb.Memory.Read(gb.CPU.HL.HiLo()), gb.CPU.AF.Hi())

		},
		0xEE: func() {
			// XOR A,#
			gb.instXor(gb.CPU.AF.SetHi, gb.popPC(), gb.CPU.AF.Hi())

		},
		0xBF: func() {
			// CP A,A
			gb.instCp(gb.CPU.AF.Hi(), gb.CPU.AF.Hi())

		},
		0xB8: func() {
			// CP A,B
			gb.instCp(gb.CPU.BC.Hi(), gb.CPU.AF.Hi())

		},
		0xB9: func() {
			// CP A,C
			gb.instCp(gb.CPU.BC.Lo(), gb.CPU.AF.Hi())

		},
		0xBA: func() {
			// CP A,D
			gb.instCp(gb.CPU.DE.Hi(), gb.CPU.AF.Hi())

		},
		0xBB: func() {
			// CP A,E
			gb.instCp(gb.CPU.DE.Lo(), gb.CPU.AF.Hi())

		},
		0xBC: func() {
			// CP A,H
			gb.instCp(gb.CPU.HL.Hi(), gb.CPU.AF.Hi())

		},
		0xBD: func() {
			// CP A,L
			gb.instCp(gb.CPU.HL.Lo(), gb.CPU.AF.Hi())

		},
		0xBE: func() {
			// CP A,(HL)
			gb.instCp(gb.Memory.Read(gb.CPU.HL.HiLo()), gb.CPU.AF.Hi())

		},
		0xFE: func() {
			// CP A,#
			gb.instCp(gb.popPC(), gb.CPU.AF.Hi())

		},
		0x3C: func() {
			// INC A
			gb.instInc(gb.CPU.AF.SetHi, gb.CPU.AF.Hi())

		},
		0x04: func() {
			// INC B
			gb.instInc(gb.CPU.BC.SetHi, gb.CPU.BC.Hi())

		},
		0x0C: func() {
			// INC C
			gb.instInc(gb.CPU.BC.SetLo, gb.CPU.BC.Lo())

		},
		0x14: func() {
			// INC D
			gb.instInc(gb.CPU.DE.SetHi, gb.CPU.DE.Hi())

		},
		0x1C: func() {
			// INC E
			gb.instInc(gb.CPU.DE.SetLo, gb.CPU.DE.Lo())

		},
		0x24: func() {
			// INC H
			gb.instInc(gb.CPU.HL.SetHi, gb.CPU.HL.Hi())

		},
		0x2C: func() {
			// INC L
			gb.instInc(gb.CPU.HL.SetLo, gb.CPU.HL.Lo())

		},
		0x34: func() {
			// INC (HL)
			addr := gb.CPU.HL.HiLo()
			gb.instInc(func(val byte) { gb.Memory.Write(addr, val) }, gb.Memory.Read(addr))

		},
		0x3D: func() {
			// DEC A
			gb.instDec(gb.CPU.AF.SetHi, gb.CPU.AF.Hi())

		},
		0x05: func() {
			// DEC B
			gb.instDec(gb.CPU.BC.SetHi, gb.CPU.BC.Hi())

		},
		0x0D: func() {
			// DEC C
			gb.instDec(gb.CPU.BC.SetLo, gb.CPU.BC.Lo())

		},
		0x15: func() {
			// DEC D
			gb.instDec(gb.CPU.DE.SetHi, gb.CPU.DE.Hi())

		},
		0x1D: func() {
			// DEC E
			gb.instDec(gb.CPU.DE.SetLo, gb.CPU.DE.Lo())

		},
		0x25: func() {
			// DEC H
			gb.instDec(gb.CPU.HL.SetHi, gb.CPU.HL.Hi())

		},
		0x2D: func() {
			// DEC L
			gb.instDec(gb.CPU.HL.SetLo, gb.CPU.HL.Lo())

		},
		0x35: func() {
			// DEC (HL)
			addr := gb.CPU.HL.HiLo()
			gb.instDec(func(val byte) { gb.Memory.Write(addr, val) }, gb.Memory.Read(addr))

			// ========== 16-Bit ALU ===========
		},
		0x09: func() {
			// ADD HL,BC
			gb.instAdd16(gb.CPU.HL.Set, gb.CPU.HL.HiLo(), gb.CPU.BC.HiLo())

		},
		0x19: func() {
			// ADD HL,DE
			gb.instAdd16(gb.CPU.HL.Set, gb.CPU.HL.HiLo(), gb.CPU.DE.HiLo())

		},
		0x29: func() {
			// ADD HL,HL
			gb.instAdd16(gb.CPU.HL.Set, gb.CPU.HL.HiLo(), gb.CPU.HL.HiLo())

		},
		0x39: func() {
			// ADD HL,SP
			gb.instAdd16(gb.CPU.HL.Set, gb.CPU.HL.HiLo(), gb.CPU.SP.HiLo())

		},
		0xE8: func() {
			// ADD SP,n
			gb.instAdd16Signed(gb.CPU.SP.Set, gb.CPU.SP.HiLo(), int8(gb.popPC()))
			gb.CPU.SetZ(false)

		},
		0x03: func() {
			// INC BC
			gb.instInc16(gb.CPU.BC.Set, gb.CPU.BC.HiLo())

		},
		0x13: func() {
			// INC DE
			gb.instInc16(gb.CPU.DE.Set, gb.CPU.DE.HiLo())

		},
		0x23: func() {
			// INC HL
			gb.instInc16(gb.CPU.HL.Set, gb.CPU.HL.HiLo())

		},
		0x33: func() {
			// INC SP
			gb.instInc16(gb.CPU.SP.Set, gb.CPU.SP.HiLo())

		},
		0x0B: func() {
			// DEC BC
			gb.instDec16(gb.CPU.BC.Set, gb.CPU.BC.HiLo())

		},
		0x1B: func() {
			// DEC DE
			gb.instDec16(gb.CPU.DE.Set, gb.CPU.DE.HiLo())

		},
		0x2B: func() {
			// DEC HL
			gb.instDec16(gb.CPU.HL.Set, gb.CPU.HL.HiLo())

		},
		0x3B: func() {
			// DEC SP
			gb.instDec16(gb.CPU.SP.Set, gb.CPU.SP.HiLo())

		},
		0x27: func() { /*
				// DAA
					When this instruction is executed, the A register is BCD
					corrected using the contents of the flags. The exact process
					is the following: if the least significant four bits of A
					contain a non-BCD digit (i. e. it is greater than 9) or the
					H flag is set, then $06 is added to the register. Then the
					four most significant bits are checked. If this more significant
					digit also happens to be greater than 9 or the C flag is set,
					then $60 is added.
			*/
			if !gb.CPU.N() {
				// TODO: This could be more efficient?
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

		},
		0x2F: func() {
			// CPL
			gb.CPU.AF.SetHi(0xFF ^ gb.CPU.AF.Hi())
			gb.CPU.SetN(true)
			gb.CPU.SetH(true)

		},
		0x3F: func() {
			// CCF
			gb.CPU.SetN(false)
			gb.CPU.SetH(false)
			gb.CPU.SetC(!gb.CPU.C())

		},
		0x37: func() {
			// SCF
			gb.CPU.SetN(false)
			gb.CPU.SetH(false)
			gb.CPU.SetC(true)

		},
		0x00: func() {
			// NOP

		},
		0x76: func() {
			// HALT
			gb.halted = true

		},
		0x10: func() { //gb.halted = true
			// STOP
			log.Print("0x10 (STOP) unimplemented (is 0x00 follows)")

		},
		0xF3: func() {
			// DI
			gb.interruptsOn = false

		},
		0xFB: func() {
			// EI
			gb.interruptsEnabling = true

		},
		0x07: func() {
			// RLCA
			value := gb.CPU.AF.Hi()
			result := byte(value<<1) | (value >> 7)
			gb.CPU.AF.SetHi(result)
			gb.CPU.SetZ(false)
			gb.CPU.SetN(false)
			gb.CPU.SetH(false)
			gb.CPU.SetC(value > 0x7F)

		},
		0x17: func() {
			// RLA
			value := gb.CPU.AF.Hi()
			var carry byte
			if gb.CPU.C() {
				carry = 1
			}
			result := byte(value<<1) + carry
			gb.CPU.AF.SetHi(result)
			gb.CPU.SetZ(false)
			gb.CPU.SetN(false)
			gb.CPU.SetH(false)
			gb.CPU.SetC(value > 0x7F)

		},
		0x0F: func() {
			// RRCA
			value := gb.CPU.AF.Hi()
			result := byte(value>>1) | byte((value&1)<<7)
			gb.CPU.AF.SetHi(result)
			gb.CPU.SetZ(false)
			gb.CPU.SetN(false)
			gb.CPU.SetH(false)
			gb.CPU.SetC(result > 0x7F)

		},
		0x1F: func() {
			// RRA
			value := gb.CPU.AF.Hi()
			var carry byte
			if gb.CPU.C() {
				carry = 0x80
			}
			result := byte(value>>1) | carry
			gb.CPU.AF.SetHi(result)
			gb.CPU.SetZ(false)
			gb.CPU.SetN(false)
			gb.CPU.SetH(false)
			gb.CPU.SetC((1 & value) == 1)

		},
		0xC3: func() {
			// JP nn
			gb.instJump(gb.popPC16())

		},
		0xC2: func() {
			// JP NZ,nn
			next := gb.popPC16()
			if !gb.CPU.Z() {
				gb.instJump(next)
				gb.thisCpuTicks += 4
			}

		},
		0xCA: func() {
			// JP Z,nn
			next := gb.popPC16()
			if gb.CPU.Z() {
				gb.instJump(next)
				gb.thisCpuTicks += 4
			}

		},
		0xD2: func() {
			// JP NC,nn
			next := gb.popPC16()
			if !gb.CPU.C() {
				gb.instJump(next)
				gb.thisCpuTicks += 4
			}

		},
		0xDA: func() {
			// JP C,nn
			next := gb.popPC16()
			if gb.CPU.C() {
				gb.instJump(next)
				gb.thisCpuTicks += 4
			}

		},
		0xE9: func() {
			// JP HL
			gb.instJump(gb.CPU.HL.HiLo())

		},
		0x18: func() {
			// JR n
			addr := int32(gb.CPU.PC) + int32(int8(gb.popPC()))
			gb.instJump(uint16(addr))

		},
		0x20: func() {
			// JR NZ,n
			next := int8(gb.popPC())
			if !gb.CPU.Z() {
				addr := int32(gb.CPU.PC) + int32(next)
				gb.instJump(uint16(addr))
				gb.thisCpuTicks += 4
			}

		},
		0x28: func() {
			// JR Z,n
			next := int8(gb.popPC())
			if gb.CPU.Z() {
				addr := int32(gb.CPU.PC) + int32(next)
				gb.instJump(uint16(addr))
				gb.thisCpuTicks += 4
			}

		},
		0x30: func() {
			// JR NC,n
			next := int8(gb.popPC())
			if !gb.CPU.C() {
				addr := int32(gb.CPU.PC) + int32(next)
				gb.instJump(uint16(addr))
				gb.thisCpuTicks += 4
			}

		},
		0x38: func() {
			// JR C,n
			next := int8(gb.popPC())
			if gb.CPU.C() {
				addr := int32(gb.CPU.PC) + int32(next)
				gb.instJump(uint16(addr))
				gb.thisCpuTicks += 4
			}

		},
		0xCD: func() {
			// CALL nn
			gb.instCall(gb.popPC16())

		},
		0xC4: func() {
			// CALL NZ,nn
			next := gb.popPC16()
			if !gb.CPU.Z() {
				gb.instCall(next)
				gb.thisCpuTicks += 12
			}

		},
		0xCC: func() {
			// CALL Z,nn
			next := gb.popPC16()
			if gb.CPU.Z() {
				gb.instCall(next)
				gb.thisCpuTicks += 12
			}

		},
		0xD4: func() {
			// CALL NC,nn
			next := gb.popPC16()
			if !gb.CPU.C() {
				gb.instCall(next)
				gb.thisCpuTicks += 12
			}

		},
		0xDC: func() {
			// CALL C,nn
			next := gb.popPC16()
			if gb.CPU.C() {
				gb.instCall(next)
				gb.thisCpuTicks += 12
			}

		},
		0xC7: func() {
			// RST 0x00
			gb.instCall(0x0000)

		},
		0xCF: func() {
			// RST 0x08
			gb.instCall(0x0008)

		},
		0xD7: func() {
			// RST 0x10
			gb.instCall(0x0010)

		},
		0xDF: func() {
			// RST 0x18
			gb.instCall(0x0018)

		},
		0xE7: func() {
			// RST 0x20
			gb.instCall(0x0020)

		},
		0xEF: func() {
			// RST 0x28
			gb.instCall(0x0028)

		},
		0xF7: func() {
			// RST 0x30
			gb.instCall(0x0030)

		},
		0xFF: func() {
			// RST 0x38
			gb.instCall(0x0038)

		},
		0xC9: func() {
			// RET
			gb.instRet()

		},
		0xC0: func() {
			// RET NZ
			if !gb.CPU.Z() {
				gb.instRet()
				gb.thisCpuTicks += 12
			}

		},
		0xC8: func() {
			// RET Z
			if gb.CPU.Z() {
				gb.instRet()
				gb.thisCpuTicks += 12
			}

		},
		0xD0: func() {
			// RET NC
			if !gb.CPU.C() {
				gb.instRet()
				gb.thisCpuTicks += 12
			}

		},
		0xD8: func() {
			// RET C
			if gb.CPU.C() {
				gb.instRet()
				gb.thisCpuTicks += 12
			}

		},
		0xD9: func() {
			// RETI
			gb.instRet()
			gb.interruptsEnabling = true

		},
		0xCB: func() {
			// CB
			nextInst := gb.popPC()
			gb.thisCpuTicks += CBOpcodeCycles[nextInst] * 4
			gb.cbInst[nextInst]()
		},
	}

	// fill the empty elements of the array
	// with a noop function to eliminate null checks
	for k, v := range ret {
		if v == nil {
			ret[k] = func() {
				log.Printf("Unimplemented opcode: %#2x", k)
				WaitForInput()
			}
		}
	}

	return ret
}
