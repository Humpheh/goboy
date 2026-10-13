package gb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Value of the blargg test status byte at $A000 while the test is still running.
const blarggStatusRunning = 0x80

// blarggResult reads the status and text output from blargg test rom.
func blarggResult(gb *Gameboy) (status byte, text string, ok bool) {
	if gb.memory.Read(0xA001) != 0xDE || gb.memory.Read(0xA002) != 0xB0 || gb.memory.Read(0xA003) != 0x61 {
		return 0, "", false
	}
	status = gb.memory.Read(0xA000)
	if status == blarggStatusRunning {
		return 0, "", false
	}
	for addr := uint16(0xA004); addr < 0xC000 && gb.memory.Read(addr) != 0; addr++ {
		text += string(gb.memory.Read(addr))
	}
	return status, text, true
}

// Run a blarrg test rom that outputs to memory until it finishes or til
// maxIterations.
func runBlarrgTestRomWithMemoryOutput(t *testing.T, rom string) {
	gb, err := New(rom, WithCGBEnabled())
	require.NoError(t, err, "error in init gb %v", err)

	// Run until the rom reports a result
	var status byte
	var text string
	var done bool
	for i := 0; i < maxIterations && !done; i++ {
		gb.Update()
		status, text, done = blarggResult(gb)
	}
	require.True(t, done, "test rom did not finish")
	assert.Equal(t, byte(0), status, "test rom failed with output:\n%v", text)
}
