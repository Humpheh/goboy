package bits

// Test if a bit is set.
func Test(value byte, bit byte) bool {
	return (value>>bit)&1 == 1
}

// Get the value of a bit.
func Val(value byte, bit byte) byte {
	return (value >> bit) & 1
}

// Set a bit and return the new value.
func Set(value byte, bit byte) byte {
	return value | (1 << bit)
}

// Reset a bit and return the new value.
func Reset(value byte, bit byte) byte {
	return value & ^(1 << bit)
}

// Half carry two values.
func HalfCarryAdd(val1 byte, val2 byte) bool {
	return (val1&0xF)+(val2&0xF) > 0xF
}

// Transform a bool into a 1/0 byte.
func B(val bool) byte {
	if val {
		return 1
	} else {
		return 0
	}
}
