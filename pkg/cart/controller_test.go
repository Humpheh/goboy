package cart

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func appendBytes(by ...[]byte) (out []byte) {
	for _, b := range by {
		out = append(out, b...)
	}
	return out
}

func TestCart_GetName(t *testing.T) {
	romData := appendBytes(
		bytes.Repeat([]byte{0}, 0x134),
		[]byte("CartridgeName!"),
		bytes.Repeat([]byte{1}, 0xFF),
	)
	rom := NewCart(romData, "test")
	assert.Equal(t, "CartridgeName!", rom.GetName())
	// Run second time to assert that caching working correctly.
	assert.Equal(t, "CartridgeName!", rom.GetName())
}

func TestCart_GetMode(t *testing.T) {
	modeRom := func(val byte) []byte {
		return appendBytes(
			bytes.Repeat([]byte{0}, 0x143),
			[]byte{val},
			bytes.Repeat([]byte{1}, 0xFF),
		)
	}

	t.Run("Dual Mode", func(t *testing.T) {
		romData := modeRom(0x80)
		rom := NewCart(romData, "test")
		assert.Equal(t, rom.GetMode(), DMG|CGB)
	})

	t.Run("CGB Mode", func(t *testing.T) {
		romData := modeRom(0xC0)
		rom := NewCart(romData, "test")
		assert.Equal(t, rom.GetMode(), CGB)
	})

	t.Run("DMG Mode", func(t *testing.T) {
		romData := modeRom(0x00)
		rom := NewCart(romData, "test")
		assert.Equal(t, rom.GetMode(), DMG)
	})
}
