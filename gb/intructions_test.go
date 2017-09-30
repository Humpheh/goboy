package gb

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

// Test that the CPU passes all of the test instructions
// in the cpu_instrs rom.
func TestInstructions(t *testing.T) {
	gb := Gameboy{}
	err := gb.Init("./../roms/cpu_instrs.gb")
	require.NoError(t, err, "error in init gb2 %v", err)

	// Expect the output to be 68 characters long
	expected := 106

	output := ""
	gb.TransferFunction = func(val byte) {
		output += string(val)
	}

	// Run the CPU until the output has matched the expected
	// or until 4000 iterations have passed.
	for i := 0; i < 4000; i++ {
		gb.Update()
		if len(output) >= expected {
			break
		}
	}
	require.Equal(t, len(output), expected, "did not finish getting output in 4000 iterations")

	start_len := len("cpu_instr   ")
	test_output := output[start_len:]
	for i := int64(0); i < 11; i++ {
		t.Run(fmt.Sprintf("Test %02v", i), func(t *testing.T) {
			test_str := test_output[0:7]
			test_output = test_output[7:]

			test_num, err := strconv.ParseInt(test_str[:2], 10, 8)
			assert.NoError(t, err, "error in parsing number: %s", test_str[:2])
			assert.Equal(t, i+1, test_num, "unexpected test number")

			status := test_str[3:5]
			assert.Equal(t, "ok", status, "status was not ok")
		})
	}
}
