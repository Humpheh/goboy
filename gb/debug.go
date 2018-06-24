package gb

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

func compare(val1 uint16, val2 uint16) string {
	if val1 != val2 {
		return "\033[1;31m!!!!!!\033[0m"
	} else {
		return "      "
	}
}

func cpuCompareString(cpu *CPU, other *CPU) string {
	return fmt.Sprintf("%5s       %s      %s      %s      %s      %s      %s",
		"     ",
		compare(cpu.AF.HiLo(), other.AF.HiLo()),
		compare(cpu.BC.HiLo(), other.BC.HiLo()),
		compare(cpu.DE.HiLo(), other.DE.HiLo()),
		compare(cpu.HL.HiLo(), other.HL.HiLo()),
		compare(cpu.PC, other.PC),
		compare(cpu.SP.HiLo(), other.SP.HiLo()),
	)
}

func isEqual(cpu *CPU, other *CPU) bool {
	return cpu.AF.HiLo() == other.AF.HiLo() &&
		cpu.BC.HiLo() == other.BC.HiLo() &&
		cpu.DE.HiLo() == other.DE.HiLo() &&
		cpu.HL.HiLo() == other.HL.HiLo() &&
		cpu.PC == other.PC &&
		cpu.SP.HiLo() == other.SP.HiLo()
}

func cpuStateString(cpu *CPU, label string) string {
	return fmt.Sprintf("%5v - AF: %0#4x  BC: %0#4x  DE: %0#4x  HL: %0#4x  PC: %0#4x  SP: %0#4x",
		label, cpu.AF.HiLo(), cpu.BC.HiLo(), cpu.DE.HiLo(), cpu.HL.HiLo(),
		cpu.PC, cpu.SP.HiLo(),
	)
}

func GetDebugScanner(filename string) (*bufio.Scanner, error) {
	// Load debug file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open debug file: %s", err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	return scanner, nil
}

func logOpcode(gb *Gameboy) {
	pc := gb.CPU.PC
	opcode := gb.Memory.Read(pc)

	next := gb.Memory.Read(pc + 1)
	fmt.Printf("[%0#2x]: %3v %-20v %0#4x  [[", opcode, gb.ScanlineCounter, GetOpcodeName(opcode, next), pc)

	for i := math.Max(0, float64(pc)-5); i < float64(pc); i++ {
		fmt.Printf(" %02x", gb.Memory.Read(uint16(i)))
	}
	fmt.Printf(" \033[1;31m%02x\033[0m", opcode)
	for i := float64(pc) + 1; i < float64(pc)+6; i++ {
		fmt.Printf(" %02x", gb.Memory.Read(uint16(i)))
	}
	fmt.Print(" ]]\n")
}

func getDebugNum(scanner *bufio.Scanner) (CPU, uint16) {
	scanner.Scan()
	line := scanner.Text()

	split := strings.Split(line, " ")
	val1, _ := strconv.ParseUint(split[0], 10, 16)
	val2, _ := strconv.ParseUint(split[1], 10, 16)
	val3, _ := strconv.ParseUint(split[2], 10, 16)
	val4, _ := strconv.ParseUint(split[3], 10, 16)
	val5, _ := strconv.ParseUint(split[4], 10, 16)
	val6, _ := strconv.ParseUint(split[5], 10, 16)
	val7, _ := strconv.ParseUint(split[6], 10, 16)

	return CPU{
		PC: uint16(val1),
		AF: register{Val: uint16(val3)},
		BC: register{Val: uint16(val4)},
		DE: register{Val: uint16(val5)},
		HL: register{Val: uint16(val6)},
		SP: register{Val: uint16(val7)},
	}, uint16(val2)
}

func waitForInput() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter text: ")
	reader.ReadString('\n')
}
