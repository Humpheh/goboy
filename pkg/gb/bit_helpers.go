package gb

// bitTest if a bit is set.
func bitTest(value byte, bit byte) bool {
	return (value>>bit)&1 == 1
}

// bitGet returns the value of a bytes bit at an index.
func bitGet(value byte, bit byte) byte {
	return (value >> bit) & 1
}

// bitSet a bit and return the new value.
func bitSet(value byte, bit byte) byte {
	return value | (1 << bit)
}

// bitReset a bit and return the new value.
func bitReset(value byte, bit byte) byte {
	return value & ^(1 << bit)
}

// bitHalfCarryAdd half carries two values.
func bitHalfCarryAdd(val1 byte, val2 byte) bool {
	return (val1&0xF)+(val2&0xF) > 0xF
}

// boolToBitByte transforms a bool into a 1/0 byte.
func boolToBitByte(val bool) byte {
	if val {
		return 1
	}
	return 0
}
