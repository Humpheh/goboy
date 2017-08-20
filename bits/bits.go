package bits

func Test(value byte, bit byte) bool {
	return (value >> bit) & 1 == 1
}