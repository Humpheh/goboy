package bits

func Test(value byte, bit byte) bool {
	return (value >> bit) & 1 == 1
}

func Val(value byte, bit byte) byte {
	return (value >> bit) & 1
}

func Set(value byte, bit byte) byte {
	return value | (1 << bit)
}

func Reset(value byte, bit byte) byte {
	return value & ^(1 << bit)
}

func HalfCarryAdd(val1 byte, val2 byte) bool {
	return ((val1&0xf)+(val2&0xf))&0x10 == 1
}

func CarryAdd(val1 byte, val2 byte) bool {
	return (uint16(val1&0xf0)+uint16(val2&0xf0))&0x100 == 1
}

func HalfCarryAdd16(val1 uint16, val2 uint16) bool {
	return ((val1&0xf00)+(val2&0xf00))&0x1000 == 1
}

func CarryAdd16(val1 uint16, val2 uint16) bool {
	return (uint32(val1&0xf000)+uint32(val2&0xf000))&0x10000 == 1
}

func B(val bool) byte {
	if val {
		return 1
	} else {
		return 0
	}
}
