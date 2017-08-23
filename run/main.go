package main

import (
	"github.com/humpheh/gob"
	"log"
)

func main() {
	cpu := gob.CPU{}
	mem := gob.Memory{}
	gb := gob.Gameboy{}

	gb.CPU = &cpu
	gb.Memory = &mem

	mem.GB = &gb
	mem.LoadCart("C:\\Users\\Humphrey\\go\\src\\github.com\\humpheh\\gob\\run\\tetris.gb")

	gb.Init()

	cycles := 0
	for true {
		cycles += gb.Update()
		log.Print(cycles)
	}
}
