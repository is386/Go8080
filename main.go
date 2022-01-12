package main

import "github.com/is386/Go8080/i8080"

var (
	FILENAME = "roms/invaders.rom"
)

func main() {
	emu := i8080.NewEmulator()
	emu.LoadRom(FILENAME, 0x0000)
	running := true

	for running {
		running = emu.Execute()
	}
}
