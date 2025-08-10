package gb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cpuTimingTest(t *testing.T, options ...GameboyOption) {
	output := ""
	transferOption := WithTransferFunction(func(val byte) {
		output += string(val)
	})
	options = append(options, transferOption)
	gb, err := New("./../../roms/blargg/instr_timing.gb", options...)
	require.NoError(t, err, "error in init gb %v", err)

	expected := "instr_timing\n\n\nPassed"

	// Run the CPU until maxIterations iterations have passed.
	for i := 0; i < maxIterations; i++ {
		gb.Update()
		if output == expected {
			break
		}
	}
	assert.Equal(t, expected, output, "Output does not match expected: '%v'", output)
}

// TestInstructionTimingCGB tests that the CPU passes all of the timing tests
// in the instr_timing rom in CGB mode.
func TestInstructionTimingCGB(t *testing.T) {
	cpuTimingTest(t, WithCGBEnabled())
}
