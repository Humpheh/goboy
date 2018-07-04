package gb

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const maxIterations = 5000

func cpuInstTest(t *testing.T, options ...GameboyOption) {
	output := ""
	transferOption := WithTransferFunction(func(val byte) {
		output += string(val)
	})
	options = append(options, transferOption)
	gb, err := NewGameboy("./../roms/cpu_instrs.gb", options...)
	require.NoError(t, err, "error in init gb %v", err)

	// Expect the output to be 75 characters long
	expected := 106

	// Run the CPU until the output has matched the expected
	// or until maxIterations iterations have passed.
	for i := 0; i < maxIterations; i++ {
		gb.Update()
		if len(output) >= expected {
			break
		}
	}

	// Trim off the title and any whitespace
	trimmed := strings.TrimSpace(strings.TrimPrefix(output, "cpu_instrs"))
	require.True(t, len(trimmed) >= 94, "did not finish getting output in %v iterations: %v", maxIterations, trimmed)

	for i := int64(0); i < 11; i++ {
		t.Run(fmt.Sprintf("Test %02v", i), func(t *testing.T) {
			testString := trimmed[0:7]
			trimmed = trimmed[7:]

			testNum, err := strconv.ParseInt(testString[:2], 10, 8)
			assert.NoError(t, err, "error in parsing number: %s", testString[:2])
			assert.Equal(t, i+1, testNum, "unexpected test number")

			status := testString[3:5]
			assert.Equal(t, "ok", status, "status was not ok")
		})
	}
}

// TestInstructionsGB tests that the CPU passes all of the test instructions
// in the cpu_instrs rom in GB mode.
func TestInstructionsGB(t *testing.T) {
	cpuInstTest(t)
}

// TestInstructionsGB tests that the CPU passes all of the test instructions
// in the cpu_instrs rom in CGB mode (includes speed switches).
func TestInstructionsCGB(t *testing.T) {
	cpuInstTest(t, WithCGBEnabled())
}
