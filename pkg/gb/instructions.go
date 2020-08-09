package gb

import "log"

// OpcodeCycles is the number of cpu cycles for each normal opcode.
var OpcodeCycles = []int{
	1, 3, 2, 2, 1, 1, 2, 1, 5, 2, 2, 2, 1, 1, 2, 1, // 0
	0, 3, 2, 2, 1, 1, 2, 1, 3, 2, 2, 2, 1, 1, 2, 1, // 1
	2, 3, 2, 2, 1, 1, 2, 1, 2, 2, 2, 2, 1, 1, 2, 1, // 2
	2, 3, 2, 2, 3, 3, 3, 1, 2, 2, 2, 2, 1, 1, 2, 1, // 3
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 4
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 5
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 6
	2, 2, 2, 2, 2, 2, 0, 2, 1, 1, 1, 1, 1, 1, 2, 1, // 7
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 8
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // 9
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // a
	1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 2, 1, // b
	2, 3, 3, 4, 3, 4, 2, 4, 2, 4, 3, 0, 3, 6, 2, 4, // c
	2, 3, 3, 0, 3, 4, 2, 4, 2, 4, 3, 0, 3, 0, 2, 4, // d
	3, 3, 2, 0, 0, 4, 2, 4, 4, 1, 4, 0, 0, 0, 2, 4, // e
	3, 3, 2, 1, 0, 4, 2, 4, 3, 2, 4, 1, 0, 0, 2, 4, // f
} //0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f

// CBOpcodeCycles is the number of cpu cycles for each CB opcode.
var CBOpcodeCycles = []int{
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
} //0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f

// ExecuteNextOpcode gets the value at the current PC address, increments the PC,
// updates the CPU ticks and executes the opcode.
func (gb *Gameboy) ExecuteNextOpcode() int {
	opcode := gb.popPC()
	gb.thisCpuTicks = OpcodeCycles[opcode] * 4
	instructions[opcode](gb)
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

func init() {
	// Fill the empty elements of the array with a noop function to eliminate null checks.
	for k, v := range instructions {
		if v == nil {
			opcode := k
			instructions[k] = func(gb *Gameboy) {
				log.Printf("Unimplemented opcode: %#2x", opcode)
				WaitForInput()
			}
		}
	}
}

var instructions = [0x100]func(*Gameboy){
	0x06: func(gb *Gameboy) {
		// LD B, n
		gb.CPU.BC.SetHi(gb.popPC())
	},
	0x0E: func(gb *Gameboy) {
		// LD C, n
		gb.CPU.BC.SetLo(gb.popPC())
	},
	0x16: func(gb *Gameboy) {
		// LD D, n
		gb.CPU.DE.SetHi(gb.popPC())
	},
	0x1E: func(gb *Gameboy) {
		// LD E, n
		gb.CPU.DE.SetLo(gb.popPC())
	},
	0x26: func(gb *Gameboy) {
		// LD H, n
		gb.CPU.HL.SetHi(gb.popPC())
	},
	0x2E: func(gb *Gameboy) {
		// LD L, n
		gb.CPU.HL.SetLo(gb.popPC())
	},
	0x7F: func(gb *Gameboy) {
		// LD A,A
		gb.CPU.AF.SetHi(gb.CPU.AF.Hi())
	},
	0x78: func(gb *Gameboy) {
		// LD A,B
		gb.CPU.AF.SetHi(gb.CPU.BC.Hi())
	},
	0x79: func(gb *Gameboy) {
		// LD A,C
		gb.CPU.AF.SetHi(gb.CPU.BC.Lo())
	},
	0x7A: func(gb *Gameboy) {
		// LD A,D
		gb.CPU.AF.SetHi(gb.CPU.DE.Hi())
	},
	0x7B: func(gb *Gameboy) {
		// LD A,E
		gb.CPU.AF.SetHi(gb.CPU.DE.Lo())
	},
	0x7C: func(gb *Gameboy) {
		// LD A,H
		gb.CPU.AF.SetHi(gb.CPU.HL.Hi())
	},
	0x7D: func(gb *Gameboy) {
		// LD A,L
		gb.CPU.AF.SetHi(gb.CPU.HL.Lo())
	},
	0x0A: func(gb *Gameboy) {
		// LD A,(BC)
		val := gb.Memory.Read(gb.CPU.BC.HiLo())
		gb.CPU.AF.SetHi(val)
	},
	0x1A: func(gb *Gameboy) {
		// LD A,(DE)
		val := gb.Memory.Read(gb.CPU.DE.HiLo())
		gb.CPU.AF.SetHi(val)
	},
	0x7E: func(gb *Gameboy) {
		// LD A,(HL)
		val := gb.Memory.Read(gb.CPU.HL.HiLo())
		gb.CPU.AF.SetHi(val)
	},
	0xFA: func(gb *Gameboy) {
		// LD A,(nn)
		val := gb.Memory.Read(gb.popPC16())
		gb.CPU.AF.SetHi(val)
	},
	0x3E: func(gb *Gameboy) {
		// LD A,(nn)
		val := gb.popPC()
		gb.CPU.AF.SetHi(val)
	},
	0x47: func(gb *Gameboy) {
		// LD B,A
		gb.CPU.BC.SetHi(gb.CPU.AF.Hi())
	},
	0x40: func(gb *Gameboy) {
		// LD B,B
		gb.CPU.BC.SetHi(gb.CPU.BC.Hi())
	},
	0x41: func(gb *Gameboy) {
		// LD B,C
		gb.CPU.BC.SetHi(gb.CPU.BC.Lo())
	},
	0x42: func(gb *Gameboy) {
		// LD B,D
		gb.CPU.BC.SetHi(gb.CPU.DE.Hi())
	},
	0x43: func(gb *Gameboy) {
		// LD B,E
		gb.CPU.BC.SetHi(gb.CPU.DE.Lo())
	},
	0x44: func(gb *Gameboy) {
		// LD B,H
		gb.CPU.BC.SetHi(gb.CPU.HL.Hi())
	},
	0x45: func(gb *Gameboy) {
		// LD B,L
		gb.CPU.BC.SetHi(gb.CPU.HL.Lo())
	},
	0x46: func(gb *Gameboy) {
		// LD B,(HL)
		val := gb.Memory.Read(gb.CPU.HL.HiLo())
		gb.CPU.BC.SetHi(val)
	},
	0x4F: func(gb *Gameboy) {
		// LD C,A
		gb.CPU.BC.SetLo(gb.CPU.AF.Hi())
	},
	0x48: func(gb *Gameboy) {
		// LD C,B
		gb.CPU.BC.SetLo(gb.CPU.BC.Hi())
	},
	0x49: func(gb *Gameboy) {
		// LD C,C
		gb.CPU.BC.SetLo(gb.CPU.BC.Lo())
	},
	0x4A: func(gb *Gameboy) {
		// LD C,D
		gb.CPU.BC.SetLo(gb.CPU.DE.Hi())
	},
	0x4B: func(gb *Gameboy) {
		// LD C,E
		gb.CPU.BC.SetLo(gb.CPU.DE.Lo())
	},
	0x4C: func(gb *Gameboy) {
		// LD C,H
		gb.CPU.BC.SetLo(gb.CPU.HL.Hi())
	},
	0x4D: func(gb *Gameboy) {
		// LD C,L
		gb.CPU.BC.SetLo(gb.CPU.HL.Lo())
	},
	0x4E: func(gb *Gameboy) {
		// LD C,(HL)
		val := gb.Memory.Read(gb.CPU.HL.HiLo())
		gb.CPU.BC.SetLo(val)
	},
	0x57: func(gb *Gameboy) {
		// LD D,A
		gb.CPU.DE.SetHi(gb.CPU.AF.Hi())
	},
	0x50: func(gb *Gameboy) {
		// LD D,B
		gb.CPU.DE.SetHi(gb.CPU.BC.Hi())
	},
	0x51: func(gb *Gameboy) {
		// LD D,C
		gb.CPU.DE.SetHi(gb.CPU.BC.Lo())
	},
	0x52: func(gb *Gameboy) {
		// LD D,D
		gb.CPU.DE.SetHi(gb.CPU.DE.Hi())
	},
	0x53: func(gb *Gameboy) {
		// LD D,E
		gb.CPU.DE.SetHi(gb.CPU.DE.Lo())
	},
	0x54: func(gb *Gameboy) {
		// LD D,H
		gb.CPU.DE.SetHi(gb.CPU.HL.Hi())
	},
	0x55: func(gb *Gameboy) {
		// LD D,L
		gb.CPU.DE.SetHi(gb.CPU.HL.Lo())
	},
	0x56: func(gb *Gameboy) {
		// LD D,(HL)
		val := gb.Memory.Read(gb.CPU.HL.HiLo())
		gb.CPU.DE.SetHi(val)
	},
	0x5F: func(gb *Gameboy) {
		// LD E,A
		gb.CPU.DE.SetLo(gb.CPU.AF.Hi())
	},
	0x58: func(gb *Gameboy) {
		// LD E,B
		gb.CPU.DE.SetLo(gb.CPU.BC.Hi())
	},
	0x59: func(gb *Gameboy) {
		// LD E,C
		gb.CPU.DE.SetLo(gb.CPU.BC.Lo())
	},
	0x5A: func(gb *Gameboy) {
		// LD E,D
		gb.CPU.DE.SetLo(gb.CPU.DE.Hi())
	},
	0x5B: func(gb *Gameboy) {
		// LD E,E
		gb.CPU.DE.SetLo(gb.CPU.DE.Lo())
	},
	0x5C: func(gb *Gameboy) {
		// LD E,H
		gb.CPU.DE.SetLo(gb.CPU.HL.Hi())
	},
	0x5D: func(gb *Gameboy) {
		// LD E,L
		gb.CPU.DE.SetLo(gb.CPU.HL.Lo())
	},
	0x5E: func(gb *Gameboy) {
		// LD E,(HL)
		val := gb.Memory.Read(gb.CPU.HL.HiLo())
		gb.CPU.DE.SetLo(val)
	},
	0x67: func(gb *Gameboy) {
		// LD H,A
		gb.CPU.HL.SetHi(gb.CPU.AF.Hi())
	},
	0x60: func(gb *Gameboy) {
		// LD H,B
		gb.CPU.HL.SetHi(gb.CPU.BC.Hi())
	},
	0x61: func(gb *Gameboy) {
		// LD H,C
		gb.CPU.HL.SetHi(gb.CPU.BC.Lo())
	},
	0x62: func(gb *Gameboy) {
		// LD H,D
		gb.CPU.HL.SetHi(gb.CPU.DE.Hi())
	},
	0x63: func(gb *Gameboy) {
		// LD H,E
		gb.CPU.HL.SetHi(gb.CPU.DE.Lo())
	},
	0x64: func(gb *Gameboy) {
		// LD H,H
		gb.CPU.HL.SetHi(gb.CPU.HL.Hi())
	},
	0x65: func(gb *Gameboy) {
		// LD H,L
		gb.CPU.HL.SetHi(gb.CPU.HL.Lo())
	},
	0x66: func(gb *Gameboy) {
		// LD H,(HL)
		val := gb.Memory.Read(gb.CPU.HL.HiLo())
		gb.CPU.HL.SetHi(val)
	},
	0x6F: func(gb *Gameboy) {
		// LD L,A
		gb.CPU.HL.SetLo(gb.CPU.AF.Hi())
	},
	0x68: func(gb *Gameboy) {
		// LD L,B
		gb.CPU.HL.SetLo(gb.CPU.BC.Hi())
	},
	0x69: func(gb *Gameboy) {
		// LD L,C
		gb.CPU.HL.SetLo(gb.CPU.BC.Lo())
	},
	0x6A: func(gb *Gameboy) {
		// LD L,D
		gb.CPU.HL.SetLo(gb.CPU.DE.Hi())
	},
	0x6B: func(gb *Gameboy) {
		// LD L,E
		gb.CPU.HL.SetLo(gb.CPU.DE.Lo())
	},
	0x6C: func(gb *Gameboy) {
		// LD L,H
		gb.CPU.HL.SetLo(gb.CPU.HL.Hi())
	},
	0x6D: func(gb *Gameboy) {
		// LD L,L
		gb.CPU.HL.SetLo(gb.CPU.HL.Lo())
	},
	0x6E: func(gb *Gameboy) {
		// LD L,(HL)
		val := gb.Memory.Read(gb.CPU.HL.HiLo())
		gb.CPU.HL.SetLo(val)
	},
	0x77: func(gb *Gameboy) {
		// LD (HL),A
		val := gb.CPU.AF.Hi()
		gb.Memory.Write(gb.CPU.HL.HiLo(), val)
	},
	0x70: func(gb *Gameboy) {
		// LD (HL),B
		val := gb.CPU.BC.Hi()
		gb.Memory.Write(gb.CPU.HL.HiLo(), val)
	},
	0x71: func(gb *Gameboy) {
		// LD (HL),C
		val := gb.CPU.BC.Lo()
		gb.Memory.Write(gb.CPU.HL.HiLo(), val)
	},
	0x72: func(gb *Gameboy) {
		// LD (HL),D
		val := gb.CPU.DE.Hi()
		gb.Memory.Write(gb.CPU.HL.HiLo(), val)
	},
	0x73: func(gb *Gameboy) {
		// LD (HL),E
		val := gb.CPU.DE.Lo()
		gb.Memory.Write(gb.CPU.HL.HiLo(), val)
	},
	0x74: func(gb *Gameboy) {
		// LD (HL),H
		val := gb.CPU.HL.Hi()
		gb.Memory.Write(gb.CPU.HL.HiLo(), val)
	},
	0x75: func(gb *Gameboy) {
		// LD (HL),L
		val := gb.CPU.HL.Lo()
		gb.Memory.Write(gb.CPU.HL.HiLo(), val)
	},
	0x36: func(gb *Gameboy) {
		// LD (HL),n 36
		val := gb.popPC()
		gb.Memory.Write(gb.CPU.HL.HiLo(), val)
	},
	0x02: func(gb *Gameboy) {
		// LD (BC),A
		val := gb.CPU.AF.Hi()
		gb.Memory.Write(gb.CPU.BC.HiLo(), val)
	},
	0x12: func(gb *Gameboy) {
		// LD (DE),A
		val := gb.CPU.AF.Hi()
		gb.Memory.Write(gb.CPU.DE.HiLo(), val)
	},
	0xEA: func(gb *Gameboy) {
		// LD (nn),A
		val := gb.CPU.AF.Hi()
		gb.Memory.Write(gb.popPC16(), val)
	},
	0xF2: func(gb *Gameboy) {
		// LD A,(C)
		val := 0xFF00 + uint16(gb.CPU.BC.Lo())
		gb.CPU.AF.SetHi(gb.Memory.Read(val))
	},
	0xE2: func(gb *Gameboy) {
		// LD (C),A
		val := gb.CPU.AF.Hi()
		mem := 0xFF00 + uint16(gb.CPU.BC.Lo())
		gb.Memory.Write(mem, val)
	},
	0x3A: func(gb *Gameboy) {
		// LDD A,(HL)
		val := gb.Memory.Read(gb.CPU.HL.HiLo())
		gb.CPU.AF.SetHi(val)
		gb.CPU.HL.Set(gb.CPU.HL.HiLo() - 1)
	},
	0x32: func(gb *Gameboy) {
		// LDD (HL),A
		val := gb.CPU.HL.HiLo()
		gb.Memory.Write(val, gb.CPU.AF.Hi())
		gb.CPU.HL.Set(gb.CPU.HL.HiLo() - 1)
	},
	0x2A: func(gb *Gameboy) {
		// LDI A,(HL)
		val := gb.Memory.Read(gb.CPU.HL.HiLo())
		gb.CPU.AF.SetHi(val)
		gb.CPU.HL.Set(gb.CPU.HL.HiLo() + 1)
	},
	0x22: func(gb *Gameboy) {
		// LDI (HL),A
		val := gb.CPU.HL.HiLo()
		gb.Memory.Write(val, gb.CPU.AF.Hi())
		gb.CPU.HL.Set(gb.CPU.HL.HiLo() + 1)
	},
	0xE0: func(gb *Gameboy) {
		// LD (0xFF00+n),A
		val := 0xFF00 + uint16(gb.popPC())
		gb.Memory.Write(val, gb.CPU.AF.Hi())
	},
	0xF0: func(gb *Gameboy) {
		// LD A,(0xFF00+n)
		val := gb.Memory.ReadHighRam(0xFF00 + uint16(gb.popPC()))
		gb.CPU.AF.SetHi(val)
	},
	// ========== 16-Bit Loads ===========
	0x01: func(gb *Gameboy) {
		// LD BC,nn
		val := gb.popPC16()
		gb.CPU.BC.Set(val)
	},
	0x11: func(gb *Gameboy) {
		// LD DE,nn
		val := gb.popPC16()
		gb.CPU.DE.Set(val)
	},
	0x21: func(gb *Gameboy) {
		// LD HL,nn
		val := gb.popPC16()
		gb.CPU.HL.Set(val)
	},
	0x31: func(gb *Gameboy) {
		// LD SP,nn
		val := gb.popPC16()
		gb.CPU.SP.Set(val)
	},
	0xF9: func(gb *Gameboy) {
		// LD SP,HL
		val := gb.CPU.HL
		gb.CPU.SP = val
	},
	0xF8: func(gb *Gameboy) {
		// LD HL,SP+n
		gb.instAdd16Signed(gb.CPU.HL.Set, gb.CPU.SP.HiLo(), int8(gb.popPC()))
	},
	0x08: func(gb *Gameboy) {
		// LD (nn),SP
		address := gb.popPC16()
		gb.Memory.Write(address, gb.CPU.SP.Lo())
		gb.Memory.Write(address+1, gb.CPU.SP.Hi())
	},
	0xF5: func(gb *Gameboy) {
		// PUSH AF
		gb.pushStack(gb.CPU.AF.HiLo())
	},
	0xC5: func(gb *Gameboy) {
		// PUSH BC
		gb.pushStack(gb.CPU.BC.HiLo())
	},
	0xD5: func(gb *Gameboy) {
		// PUSH DE
		gb.pushStack(gb.CPU.DE.HiLo())
	},
	0xE5: func(gb *Gameboy) {
		// PUSH HL
		gb.pushStack(gb.CPU.HL.HiLo())
	},
	0xF1: func(gb *Gameboy) {
		// POP AF
		gb.CPU.AF.Set(gb.popStack())
	},
	0xC1: func(gb *Gameboy) {
		// POP BC
		gb.CPU.BC.Set(gb.popStack())
	},
	0xD1: func(gb *Gameboy) {
		// POP DE
		gb.CPU.DE.Set(gb.popStack())
	},
	0xE1: func(gb *Gameboy) {
		// POP HL
		gb.CPU.HL.Set(gb.popStack())
	},
	// ========== 8-Bit ALU ===========
	0x87: func(gb *Gameboy) {
		// ADD A,A
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi(), false)
	},
	0x80: func(gb *Gameboy) {
		// ADD A,B
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.BC.Hi(), gb.CPU.AF.Hi(), false)
	},
	0x81: func(gb *Gameboy) {
		// ADD A,C
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.BC.Lo(), gb.CPU.AF.Hi(), false)
	},
	0x82: func(gb *Gameboy) {
		// ADD A,D
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.DE.Hi(), gb.CPU.AF.Hi(), false)
	},
	0x83: func(gb *Gameboy) {
		// ADD A,E
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.DE.Lo(), gb.CPU.AF.Hi(), false)
	},
	0x84: func(gb *Gameboy) {
		// ADD A,H
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.HL.Hi(), gb.CPU.AF.Hi(), false)
	},
	0x85: func(gb *Gameboy) {
		// ADD A,L
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.HL.Lo(), gb.CPU.AF.Hi(), false)
	},
	0x86: func(gb *Gameboy) {
		// ADD A,(HL)
		gb.instAdd(gb.CPU.AF.SetHi, gb.Memory.Read(gb.CPU.HL.HiLo()), gb.CPU.AF.Hi(), false)
	},
	0xC6: func(gb *Gameboy) {
		// ADD A,#
		gb.instAdd(gb.CPU.AF.SetHi, gb.popPC(), gb.CPU.AF.Hi(), false)
	},
	0x8F: func(gb *Gameboy) {
		// ADC A,A
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi(), true)
	},
	0x88: func(gb *Gameboy) {
		// ADC A,B
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.BC.Hi(), gb.CPU.AF.Hi(), true)
	},
	0x89: func(gb *Gameboy) {
		// ADC A,C
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.BC.Lo(), gb.CPU.AF.Hi(), true)
	},
	0x8A: func(gb *Gameboy) {
		// ADC A,D
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.DE.Hi(), gb.CPU.AF.Hi(), true)
	},
	0x8B: func(gb *Gameboy) {
		// ADC A,E
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.DE.Lo(), gb.CPU.AF.Hi(), true)
	},
	0x8C: func(gb *Gameboy) {
		// ADC A,H
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.HL.Hi(), gb.CPU.AF.Hi(), true)
	},
	0x8D: func(gb *Gameboy) {
		// ADC A,L
		gb.instAdd(gb.CPU.AF.SetHi, gb.CPU.HL.Lo(), gb.CPU.AF.Hi(), true)
	},
	0x8E: func(gb *Gameboy) {
		// ADC A,(HL)
		gb.instAdd(gb.CPU.AF.SetHi, gb.Memory.Read(gb.CPU.HL.HiLo()), gb.CPU.AF.Hi(), true)
	},
	0xCE: func(gb *Gameboy) {
		// ADC A,#
		gb.instAdd(gb.CPU.AF.SetHi, gb.popPC(), gb.CPU.AF.Hi(), true)
	},
	0x97: func(gb *Gameboy) {
		// SUB A,A
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi(), false)
	},
	0x90: func(gb *Gameboy) {
		// SUB A,B
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.BC.Hi(), false)
	},
	0x91: func(gb *Gameboy) {
		// SUB A,C
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.BC.Lo(), false)
	},
	0x92: func(gb *Gameboy) {
		// SUB A,D
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.DE.Hi(), false)
	},
	0x93: func(gb *Gameboy) {
		// SUB A,E
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.DE.Lo(), false)
	},
	0x94: func(gb *Gameboy) {
		// SUB A,H
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.HL.Hi(), false)
	},
	0x95: func(gb *Gameboy) {
		// SUB A,L
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.HL.Lo(), false)
	},
	0x96: func(gb *Gameboy) {
		// SUB A,(HL)
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.Memory.Read(gb.CPU.HL.HiLo()), false)
	},
	0xD6: func(gb *Gameboy) {
		// SUB A,#
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.popPC(), false)
	},
	0x9F: func(gb *Gameboy) {
		// SBC A,A
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi(), true)
	},
	0x98: func(gb *Gameboy) {
		// SBC A,B
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.BC.Hi(), true)
	},
	0x99: func(gb *Gameboy) {
		// SBC A,C
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.BC.Lo(), true)
	},
	0x9A: func(gb *Gameboy) {
		// SBC A,D
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.DE.Hi(), true)
	},
	0x9B: func(gb *Gameboy) {
		// SBC A,E
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.DE.Lo(), true)
	},
	0x9C: func(gb *Gameboy) {
		// SBC A,H
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.HL.Hi(), true)
	},
	0x9D: func(gb *Gameboy) {
		// SBC A,L
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.HL.Lo(), true)
	},
	0x9E: func(gb *Gameboy) {
		// SBC A,(HL)
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.Memory.Read(gb.CPU.HL.HiLo()), true)
	},
	0xDE: func(gb *Gameboy) {
		// SBC A,#
		gb.instSub(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.popPC(), true)
	},
	0xA7: func(gb *Gameboy) {
		// AND A,A
		gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi())
	},
	0xA0: func(gb *Gameboy) {
		// AND A,B
		gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.BC.Hi(), gb.CPU.AF.Hi())
	},
	0xA1: func(gb *Gameboy) {
		// AND A,C
		gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.BC.Lo(), gb.CPU.AF.Hi())
	},
	0xA2: func(gb *Gameboy) {
		// AND A,D
		gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.DE.Hi(), gb.CPU.AF.Hi())
	},
	0xA3: func(gb *Gameboy) {
		// AND A,E
		gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.DE.Lo(), gb.CPU.AF.Hi())
	},
	0xA4: func(gb *Gameboy) {
		// AND A,H
		gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.HL.Hi(), gb.CPU.AF.Hi())
	},
	0xA5: func(gb *Gameboy) {
		// AND A,L
		gb.instAnd(gb.CPU.AF.SetHi, gb.CPU.HL.Lo(), gb.CPU.AF.Hi())
	},
	0xA6: func(gb *Gameboy) {
		// AND A,(HL)
		gb.instAnd(gb.CPU.AF.SetHi, gb.Memory.Read(gb.CPU.HL.HiLo()), gb.CPU.AF.Hi())
	},
	0xE6: func(gb *Gameboy) {
		// AND A,#
		gb.instAnd(gb.CPU.AF.SetHi, gb.popPC(), gb.CPU.AF.Hi())
	},
	0xB7: func(gb *Gameboy) {
		// OR A,A
		gb.instOr(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi())
	},
	0xB0: func(gb *Gameboy) {
		// OR A,B
		gb.instOr(gb.CPU.AF.SetHi, gb.CPU.BC.Hi(), gb.CPU.AF.Hi())
	},
	0xB1: func(gb *Gameboy) {
		// OR A,C
		gb.instOr(gb.CPU.AF.SetHi, gb.CPU.BC.Lo(), gb.CPU.AF.Hi())
	},
	0xB2: func(gb *Gameboy) {
		// OR A,D
		gb.instOr(gb.CPU.AF.SetHi, gb.CPU.DE.Hi(), gb.CPU.AF.Hi())
	},
	0xB3: func(gb *Gameboy) {
		// OR A,E
		gb.instOr(gb.CPU.AF.SetHi, gb.CPU.DE.Lo(), gb.CPU.AF.Hi())
	},
	0xB4: func(gb *Gameboy) {
		// OR A,H
		gb.instOr(gb.CPU.AF.SetHi, gb.CPU.HL.Hi(), gb.CPU.AF.Hi())
	},
	0xB5: func(gb *Gameboy) {
		// OR A,L
		gb.instOr(gb.CPU.AF.SetHi, gb.CPU.HL.Lo(), gb.CPU.AF.Hi())
	},
	0xB6: func(gb *Gameboy) {
		// OR A,(HL)
		gb.instOr(gb.CPU.AF.SetHi, gb.Memory.Read(gb.CPU.HL.HiLo()), gb.CPU.AF.Hi())
	},
	0xF6: func(gb *Gameboy) {
		// OR A,#
		gb.instOr(gb.CPU.AF.SetHi, gb.popPC(), gb.CPU.AF.Hi())
	},
	0xAF: func(gb *Gameboy) {
		// XOR A,A
		gb.instXor(gb.CPU.AF.SetHi, gb.CPU.AF.Hi(), gb.CPU.AF.Hi())
	},
	0xA8: func(gb *Gameboy) {
		// XOR A,B
		gb.instXor(gb.CPU.AF.SetHi, gb.CPU.BC.Hi(), gb.CPU.AF.Hi())
	},
	0xA9: func(gb *Gameboy) {
		// XOR A,C
		gb.instXor(gb.CPU.AF.SetHi, gb.CPU.BC.Lo(), gb.CPU.AF.Hi())
	},
	0xAA: func(gb *Gameboy) {
		// XOR A,D
		gb.instXor(gb.CPU.AF.SetHi, gb.CPU.DE.Hi(), gb.CPU.AF.Hi())
	},
	0xAB: func(gb *Gameboy) {
		// XOR A,E
		gb.instXor(gb.CPU.AF.SetHi, gb.CPU.DE.Lo(), gb.CPU.AF.Hi())
	},
	0xAC: func(gb *Gameboy) {
		// XOR A,H
		gb.instXor(gb.CPU.AF.SetHi, gb.CPU.HL.Hi(), gb.CPU.AF.Hi())
	},
	0xAD: func(gb *Gameboy) {
		// XOR A,L
		gb.instXor(gb.CPU.AF.SetHi, gb.CPU.HL.Lo(), gb.CPU.AF.Hi())
	},
	0xAE: func(gb *Gameboy) {
		// XOR A,(HL)
		gb.instXor(gb.CPU.AF.SetHi, gb.Memory.Read(gb.CPU.HL.HiLo()), gb.CPU.AF.Hi())
	},
	0xEE: func(gb *Gameboy) {
		// XOR A,#
		gb.instXor(gb.CPU.AF.SetHi, gb.popPC(), gb.CPU.AF.Hi())
	},
	0xBF: func(gb *Gameboy) {
		// CP A,A
		gb.instCp(gb.CPU.AF.Hi(), gb.CPU.AF.Hi())
	},
	0xB8: func(gb *Gameboy) {
		// CP A,B
		gb.instCp(gb.CPU.BC.Hi(), gb.CPU.AF.Hi())
	},
	0xB9: func(gb *Gameboy) {
		// CP A,C
		gb.instCp(gb.CPU.BC.Lo(), gb.CPU.AF.Hi())
	},
	0xBA: func(gb *Gameboy) {
		// CP A,D
		gb.instCp(gb.CPU.DE.Hi(), gb.CPU.AF.Hi())
	},
	0xBB: func(gb *Gameboy) {
		// CP A,E
		gb.instCp(gb.CPU.DE.Lo(), gb.CPU.AF.Hi())
	},
	0xBC: func(gb *Gameboy) {
		// CP A,H
		gb.instCp(gb.CPU.HL.Hi(), gb.CPU.AF.Hi())
	},
	0xBD: func(gb *Gameboy) {
		// CP A,L
		gb.instCp(gb.CPU.HL.Lo(), gb.CPU.AF.Hi())
	},
	0xBE: func(gb *Gameboy) {
		// CP A,(HL)
		gb.instCp(gb.Memory.Read(gb.CPU.HL.HiLo()), gb.CPU.AF.Hi())
	},
	0xFE: func(gb *Gameboy) {
		// CP A,#
		gb.instCp(gb.popPC(), gb.CPU.AF.Hi())
	},
	0x3C: func(gb *Gameboy) {
		// INC A
		gb.instInc(gb.CPU.AF.SetHi, gb.CPU.AF.Hi())
	},
	0x04: func(gb *Gameboy) {
		// INC B
		gb.instInc(gb.CPU.BC.SetHi, gb.CPU.BC.Hi())
	},
	0x0C: func(gb *Gameboy) {
		// INC C
		gb.instInc(gb.CPU.BC.SetLo, gb.CPU.BC.Lo())
	},
	0x14: func(gb *Gameboy) {
		// INC D
		gb.instInc(gb.CPU.DE.SetHi, gb.CPU.DE.Hi())
	},
	0x1C: func(gb *Gameboy) {
		// INC E
		gb.instInc(gb.CPU.DE.SetLo, gb.CPU.DE.Lo())
	},
	0x24: func(gb *Gameboy) {
		// INC H
		gb.instInc(gb.CPU.HL.SetHi, gb.CPU.HL.Hi())
	},
	0x2C: func(gb *Gameboy) {
		// INC L
		gb.instInc(gb.CPU.HL.SetLo, gb.CPU.HL.Lo())
	},
	0x34: func(gb *Gameboy) {
		// INC (HL)
		addr := gb.CPU.HL.HiLo()
		gb.instInc(func(val byte) { gb.Memory.Write(addr, val) }, gb.Memory.Read(addr))
	},
	0x3D: func(gb *Gameboy) {
		// DEC A
		gb.instDec(gb.CPU.AF.SetHi, gb.CPU.AF.Hi())
	},
	0x05: func(gb *Gameboy) {
		// DEC B
		gb.instDec(gb.CPU.BC.SetHi, gb.CPU.BC.Hi())
	},
	0x0D: func(gb *Gameboy) {
		// DEC C
		gb.instDec(gb.CPU.BC.SetLo, gb.CPU.BC.Lo())
	},
	0x15: func(gb *Gameboy) {
		// DEC D
		gb.instDec(gb.CPU.DE.SetHi, gb.CPU.DE.Hi())
	},
	0x1D: func(gb *Gameboy) {
		// DEC E
		gb.instDec(gb.CPU.DE.SetLo, gb.CPU.DE.Lo())
	},
	0x25: func(gb *Gameboy) {
		// DEC H
		gb.instDec(gb.CPU.HL.SetHi, gb.CPU.HL.Hi())
	},
	0x2D: func(gb *Gameboy) {
		// DEC L
		gb.instDec(gb.CPU.HL.SetLo, gb.CPU.HL.Lo())
	},
	0x35: func(gb *Gameboy) {
		// DEC (HL)
		addr := gb.CPU.HL.HiLo()
		gb.instDec(func(val byte) { gb.Memory.Write(addr, val) }, gb.Memory.Read(addr))
	},
	// ========== 16-Bit ALU ===========
	0x09: func(gb *Gameboy) {
		// ADD HL,BC
		gb.instAdd16(gb.CPU.HL.Set, gb.CPU.HL.HiLo(), gb.CPU.BC.HiLo())
	},
	0x19: func(gb *Gameboy) {
		// ADD HL,DE
		gb.instAdd16(gb.CPU.HL.Set, gb.CPU.HL.HiLo(), gb.CPU.DE.HiLo())
	},
	0x29: func(gb *Gameboy) {
		// ADD HL,HL
		gb.instAdd16(gb.CPU.HL.Set, gb.CPU.HL.HiLo(), gb.CPU.HL.HiLo())
	},
	0x39: func(gb *Gameboy) {
		// ADD HL,SP
		gb.instAdd16(gb.CPU.HL.Set, gb.CPU.HL.HiLo(), gb.CPU.SP.HiLo())
	},
	0xE8: func(gb *Gameboy) {
		// ADD SP,n
		gb.instAdd16Signed(gb.CPU.SP.Set, gb.CPU.SP.HiLo(), int8(gb.popPC()))
		gb.CPU.SetZ(false)
	},
	0x03: func(gb *Gameboy) {
		// INC BC
		gb.instInc16(gb.CPU.BC.Set, gb.CPU.BC.HiLo())
	},
	0x13: func(gb *Gameboy) {
		// INC DE
		gb.instInc16(gb.CPU.DE.Set, gb.CPU.DE.HiLo())
	},
	0x23: func(gb *Gameboy) {
		// INC HL
		gb.instInc16(gb.CPU.HL.Set, gb.CPU.HL.HiLo())
	},
	0x33: func(gb *Gameboy) {
		// INC SP
		gb.instInc16(gb.CPU.SP.Set, gb.CPU.SP.HiLo())
	},
	0x0B: func(gb *Gameboy) {
		// DEC BC
		gb.instDec16(gb.CPU.BC.Set, gb.CPU.BC.HiLo())
	},
	0x1B: func(gb *Gameboy) {
		// DEC DE
		gb.instDec16(gb.CPU.DE.Set, gb.CPU.DE.HiLo())
	},
	0x2B: func(gb *Gameboy) {
		// DEC HL
		gb.instDec16(gb.CPU.HL.Set, gb.CPU.HL.HiLo())
	},
	0x3B: func(gb *Gameboy) {
		// DEC SP
		gb.instDec16(gb.CPU.SP.Set, gb.CPU.SP.HiLo())
	},
	0x27: func(gb *Gameboy) {
		// DAA

		// When this instruction is executed, the A register is BCD
		// corrected using the contents of the flags. The exact process
		// is the following: if the least significant four bits of A
		// contain a non-BCD digit (i. e. it is greater than 9) or the
		// H flag is set, then 0x60 is added to the register. Then the
		// four most significant bits are checked. If this more significant
		// digit also happens to be greater than 9 or the C flag is set,
		// then 0x60 is added.
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
	},
	0x2F: func(gb *Gameboy) {
		// CPL
		gb.CPU.AF.SetHi(0xFF ^ gb.CPU.AF.Hi())
		gb.CPU.SetN(true)
		gb.CPU.SetH(true)
	},
	0x3F: func(gb *Gameboy) {
		// CCF
		gb.CPU.SetN(false)
		gb.CPU.SetH(false)
		gb.CPU.SetC(!gb.CPU.C())
	},
	0x37: func(gb *Gameboy) {
		// SCF
		gb.CPU.SetN(false)
		gb.CPU.SetH(false)
		gb.CPU.SetC(true)
	},
	0x00: func(gb *Gameboy) {
		// NOP
	},
	0x76: func(gb *Gameboy) {
		// HALT
		gb.halted = true
	},
	0x10: func(gb *Gameboy) {
		// STOP
		gb.halted = true
		if gb.IsCGB() {
			// Handle switching to double speed mode
			gb.checkSpeedSwitch()
		}

		// Pop the next value as the STOP instruction is 2 bytes long. The second value
		// can be ignored, although generally it is expected to be 0x00 and any other
		// value is counted as a corrupted STOP instruction.
		gb.popPC()
	},
	0xF3: func(gb *Gameboy) {
		// DI
		gb.interruptsOn = false
	},
	0xFB: func(gb *Gameboy) {
		// EI
		gb.interruptsEnabling = true
	},
	0x07: func(gb *Gameboy) {
		// RLCA
		value := gb.CPU.AF.Hi()
		result := byte(value<<1) | (value >> 7)
		gb.CPU.AF.SetHi(result)
		gb.CPU.SetZ(false)
		gb.CPU.SetN(false)
		gb.CPU.SetH(false)
		gb.CPU.SetC(value > 0x7F)
	},
	0x17: func(gb *Gameboy) {
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
	0x0F: func(gb *Gameboy) {
		// RRCA
		value := gb.CPU.AF.Hi()
		result := byte(value>>1) | byte((value&1)<<7)
		gb.CPU.AF.SetHi(result)
		gb.CPU.SetZ(false)
		gb.CPU.SetN(false)
		gb.CPU.SetH(false)
		gb.CPU.SetC(result > 0x7F)
	},
	0x1F: func(gb *Gameboy) {
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
	0xC3: func(gb *Gameboy) {
		// JP nn
		gb.instJump(gb.popPC16())
	},
	0xC2: func(gb *Gameboy) {
		// JP NZ,nn
		next := gb.popPC16()
		if !gb.CPU.Z() {
			gb.instJump(next)
			gb.thisCpuTicks += 4
		}
	},
	0xCA: func(gb *Gameboy) {
		// JP Z,nn
		next := gb.popPC16()
		if gb.CPU.Z() {
			gb.instJump(next)
			gb.thisCpuTicks += 4
		}
	},
	0xD2: func(gb *Gameboy) {
		// JP NC,nn
		next := gb.popPC16()
		if !gb.CPU.C() {
			gb.instJump(next)
			gb.thisCpuTicks += 4
		}
	},
	0xDA: func(gb *Gameboy) {
		// JP C,nn
		next := gb.popPC16()
		if gb.CPU.C() {
			gb.instJump(next)
			gb.thisCpuTicks += 4
		}
	},
	0xE9: func(gb *Gameboy) {
		// JP HL
		gb.instJump(gb.CPU.HL.HiLo())
	},
	0x18: func(gb *Gameboy) {
		// JR n
		addr := int32(gb.CPU.PC) + int32(int8(gb.popPC()))
		gb.instJump(uint16(addr))
	},
	0x20: func(gb *Gameboy) {
		// JR NZ,n
		next := int8(gb.popPC())
		if !gb.CPU.Z() {
			addr := int32(gb.CPU.PC) + int32(next)
			gb.instJump(uint16(addr))
			gb.thisCpuTicks += 4
		}
	},
	0x28: func(gb *Gameboy) {
		// JR Z,n
		next := int8(gb.popPC())
		if gb.CPU.Z() {
			addr := int32(gb.CPU.PC) + int32(next)
			gb.instJump(uint16(addr))
			gb.thisCpuTicks += 4
		}
	},
	0x30: func(gb *Gameboy) {
		// JR NC,n
		next := int8(gb.popPC())
		if !gb.CPU.C() {
			addr := int32(gb.CPU.PC) + int32(next)
			gb.instJump(uint16(addr))
			gb.thisCpuTicks += 4
		}
	},
	0x38: func(gb *Gameboy) {
		// JR C,n
		next := int8(gb.popPC())
		if gb.CPU.C() {
			addr := int32(gb.CPU.PC) + int32(next)
			gb.instJump(uint16(addr))
			gb.thisCpuTicks += 4
		}
	},
	0xCD: func(gb *Gameboy) {
		// CALL nn
		gb.instCall(gb.popPC16())
	},
	0xC4: func(gb *Gameboy) {
		// CALL NZ,nn
		next := gb.popPC16()
		if !gb.CPU.Z() {
			gb.instCall(next)
			gb.thisCpuTicks += 12
		}
	},
	0xCC: func(gb *Gameboy) {
		// CALL Z,nn
		next := gb.popPC16()
		if gb.CPU.Z() {
			gb.instCall(next)
			gb.thisCpuTicks += 12
		}
	},
	0xD4: func(gb *Gameboy) {
		// CALL NC,nn
		next := gb.popPC16()
		if !gb.CPU.C() {
			gb.instCall(next)
			gb.thisCpuTicks += 12
		}
	},
	0xDC: func(gb *Gameboy) {
		// CALL C,nn
		next := gb.popPC16()
		if gb.CPU.C() {
			gb.instCall(next)
			gb.thisCpuTicks += 12
		}
	},
	0xC7: func(gb *Gameboy) {
		// RST 0x00
		gb.instCall(0x0000)
	},
	0xCF: func(gb *Gameboy) {
		// RST 0x08
		gb.instCall(0x0008)
	},
	0xD7: func(gb *Gameboy) {
		// RST 0x10
		gb.instCall(0x0010)
	},
	0xDF: func(gb *Gameboy) {
		// RST 0x18
		gb.instCall(0x0018)
	},
	0xE7: func(gb *Gameboy) {
		// RST 0x20
		gb.instCall(0x0020)
	},
	0xEF: func(gb *Gameboy) {
		// RST 0x28
		gb.instCall(0x0028)
	},
	0xF7: func(gb *Gameboy) {
		// RST 0x30
		gb.instCall(0x0030)
	},
	0xFF: func(gb *Gameboy) {
		// RST 0x38
		gb.instCall(0x0038)
	},
	0xC9: func(gb *Gameboy) {
		// RET
		gb.instRet()
	},
	0xC0: func(gb *Gameboy) {
		// RET NZ
		if !gb.CPU.Z() {
			gb.instRet()
			gb.thisCpuTicks += 12
		}
	},
	0xC8: func(gb *Gameboy) {
		// RET Z
		if gb.CPU.Z() {
			gb.instRet()
			gb.thisCpuTicks += 12
		}
	},
	0xD0: func(gb *Gameboy) {
		// RET NC
		if !gb.CPU.C() {
			gb.instRet()
			gb.thisCpuTicks += 12
		}
	},
	0xD8: func(gb *Gameboy) {
		// RET C
		if gb.CPU.C() {
			gb.instRet()
			gb.thisCpuTicks += 12
		}
	},
	0xD9: func(gb *Gameboy) {
		// RETI
		gb.instRet()
		gb.interruptsEnabling = true
	},
	0xCB: func(gb *Gameboy) {
		// CB
		nextInst := gb.popPC()
		gb.thisCpuTicks += CBOpcodeCycles[nextInst] * 4
		gb.cbInst[nextInst]()
	},
}
