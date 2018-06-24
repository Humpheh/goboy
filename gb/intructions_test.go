package gb

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test that the CPU passes all of the test instructions
// in the cpu_instrs rom.
func TestInstructions(t *testing.T) {
	output := ""
	transferOption := WithTransferFunction(func(val byte) {
		output += string(val)
	})
	gb, err := NewGameboy("./../roms/cpu_instrs.gb", transferOption)
	require.NoError(t, err, "error in init gb %v", err)

	// Expect the output to be 68 characters long
	expected := 106

	// Run the CPU until the output has matched the expected
	// or until 4000 iterations have passed.
	for i := 0; i < 4000; i++ {
		gb.Update()
		if len(output) >= expected {
			break
		}
	}
	require.Equal(t, len(output), expected, "did not finish getting output in 4000 iterations")

	startLen := len("cpu_instr   ")
	testOutput := output[startLen:]
	for i := int64(0); i < 11; i++ {
		t.Run(fmt.Sprintf("Test %02v", i), func(t *testing.T) {
			testString := testOutput[0:7]
			testOutput = testOutput[7:]

			testNum, err := strconv.ParseInt(testString[:2], 10, 8)
			assert.NoError(t, err, "error in parsing number: %s", testString[:2])
			assert.Equal(t, i+1, testNum, "unexpected test number")

			status := testString[3:5]
			assert.Equal(t, "ok", status, "status was not ok")
		})
	}
}
