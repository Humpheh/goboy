package gb

import (
	"testing"
)

// TestInterruptTimingCGB tests that the CPU passes the blargg interrupt_time
// rom in CGB mode.
func TestInterruptTimingCGB(t *testing.T) {
	runBlarrgTestRomWithMemoryOutput(t, "./../../roms/blargg/interrupt_time.gb")
}
