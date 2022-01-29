package main

import (
	"github.com/is386/Go8080/i8080"
)

var (
	FILENAME = "roms/TST8080.COM"
	PC       = uint16(0x100)
	DEBUG    = false
	TEST     = true
)

func main() {
	im := i8080.NewInvadersMachine(FILENAME, PC, DEBUG, TEST)
	im.Run()
}
