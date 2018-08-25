package gb

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/Humpheh/goboy/pkg/debug"
)

// CpuStateString is a debug function to get a string of the state of the CPU.
func CpuStateString(cpu *CPU, label string) string {
	return fmt.Sprintf("%5v - AF: %0#4x  BC: %0#4x  DE: %0#4x  HL: %0#4x  PC: %0#4x  SP: %0#4x",
		label, cpu.AF.HiLo(), cpu.BC.HiLo(), cpu.DE.HiLo(), cpu.HL.HiLo(),
		cpu.PC, cpu.SP.HiLo(),
	)
}

// LogOpcode is a debug function to log the the current state of the gameboys CPU and next memory.
func LogOpcode(gb *Gameboy, short bool) {
	pc := gb.CPU.PC
	opcode := gb.Memory.Read(pc)

	next := gb.Memory.Read(pc + 1)
	fmt.Printf("[%0#2x]: %3v %-20v %0#4x", opcode, gb.GetScanlineCounter(), debug.GetOpcodeName(opcode, next), pc)

	if !short {
		fmt.Printf("  [[")
		for i := math.Max(0, float64(pc)-5); i < float64(pc); i++ {
			fmt.Printf(" %02x", gb.Memory.Read(uint16(i)))
		}
		fmt.Printf(" \033[1;31m%02x\033[0m", opcode)
		for i := float64(pc) + 1; i < float64(pc)+6; i++ {
			fmt.Printf(" %02x", gb.Memory.Read(uint16(i)))
		}
		fmt.Print(" ]]\n")
	}
}

// LogMemory is a debug function to log some arbitrary memory.
func LogMemory(gb *Gameboy, start uint16, len uint16) {
	fmt.Printf(" [[")
	for i := start; i < start+len; i++ {
		fmt.Printf(" %02x", gb.Memory.Read(i))
	}
	fmt.Print(" ]]\n")
}

// WaitForInput is a debug function which blocks and waits for some input before continuing.
func WaitForInput() uint16 {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter text: ")
	str, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error: %v", err)
		return WaitForInput()
	}
	trimmed := strings.TrimSpace(str)
	if trimmed == "" {
		return 0
	}
	d, err := strconv.ParseInt("0x"+trimmed, 0, 64)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return WaitForInput()
	}
	log.Printf("Entered: %04x", d)
	return uint16(d)
}
