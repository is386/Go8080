package main

import (
	"github.com/is386/Go8080/i8080"
)

var (
	FILENAME = "roms/TST8080.COM"
	PC       = uint16(0x100)
)

func main() {
	emu := i8080.NewEmulator(PC)
	emu.LoadRom(FILENAME, PC)
	running := true
	for running {
		running = emu.Execute()
	}
}
