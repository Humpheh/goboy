package cart

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestROM_Read(t *testing.T) {
	rom := NewROM([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	for i := 0; i < 10; i++ {
		assert.Equal(t, byte(i), rom.Read(uint16(i)))
	}
}
