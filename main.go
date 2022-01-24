package main

import (
	"github.com/is386/Go8080/i8080"
)

var (
	FILENAME = "roms/TST8080.COM"
)

func main() {
	emu := i8080.NewEmulator()
	emu.LoadRom(FILENAME, 0x100)
	//emu.LoadRom(FILENAME, 0x0)
	running := true
	count := 0
	for running {
		running = emu.Execute(count)
		count++
	}
}
