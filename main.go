package main

import (
	"github.com/is386/Go8080/i8080Invaders"
)

var (
	DEBUG = true
)

func main() {
	im := i8080Invaders.NewInvadersMachine(DEBUG)
	im.Run()
}
