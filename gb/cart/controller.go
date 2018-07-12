package cart

type BankingController interface {
	Read(address uint16) byte
	WriteROM(address uint16, value byte)
	WriteRAM(address uint16, value byte)
}
