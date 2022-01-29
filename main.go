package main

import "github.com/is386/Go8080/i8080Test"

var (
	FILENAME = "roms/TST8080.COM"
	DEBUG    = false
)

func main() {
	m := i8080Test.NewTestMachine(FILENAME, DEBUG)
	m.Run()
}
