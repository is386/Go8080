package i8080Invaders

import (
	"os"

	"github.com/is386/Go8080/i8080"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	FILE = "i8080Invaders/INVADERS.COM"
	CPS  = 2000000 / 60
)

type InvadersMachine struct {
	cpu                             *i8080.CPU
	screen                          *Screen
	port1, port2                    uint8
	shiftMsb, shiftLsb, shiftOffset uint8
}

func NewInvadersMachine() *InvadersMachine {
	im := &InvadersMachine{screen: NewScreen()}
	cpu := i8080.NewCPU(0x0, 0x2000, 0x4000, im.PortIn, im.PortOut)
	cpu.LoadRom(FILE)
	im.cpu = cpu
	return im
}

func (im *InvadersMachine) Run() {
	running := true
	for running {
		running = im.pollSDL()
		im.runCPU()
	}
	im.screen.Destroy()
}

func (im *InvadersMachine) pollSDL() bool {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		case *sdl.QuitEvent:
			return false
		case *sdl.KeyboardEvent:
			switch e.Type {
			case sdl.KEYDOWN:
				im.keyDown(e.Keysym.Sym)
			case sdl.KEYUP:
				im.keyUp(e.Keysym.Sym)
			}
		}
	}
	return true
}

func (im *InvadersMachine) runCPU() {
	for im.cpu.GetCycles() < CPS/2 {
		im.cpu.Execute()
	}
	if im.cpu.IsInterrupted() {
		im.cpu.Interrupt(0x8)
	}
	for im.cpu.GetCycles() < CPS {
		im.cpu.Execute()
	}
	if im.cpu.IsInterrupted() {
		im.cpu.Interrupt(0x10)
	}
	im.cpu.SubtractCycles(int(CPS))
	im.screen.Draw(im)
	im.screen.Update()
}

func (im *InvadersMachine) PortIn(port uint8) {
	val := uint8(0xFF)
	switch port {
	case 0:
		break
	case 1:
		val = im.port1
	case 2:
		val = im.port2
	case 3:
		shift := (uint16(im.shiftMsb) << 8) | uint16(im.shiftLsb)
		val = uint8((shift >> (8 - im.shiftOffset)) & 0xFF)
	case 6:
		val = 0
	}
	reg := im.cpu.GetRegisters()
	reg.A = val
}

func (im *InvadersMachine) PortOut(port uint8) {
	reg := im.cpu.GetRegisters()
	a := reg.A
	switch port {
	case 2:
		im.shiftOffset = a & 7
	case 4:
		im.shiftLsb = im.shiftMsb
		im.shiftMsb = a
	case 6:
		break
	}
}

func (im *InvadersMachine) keyDown(key sdl.Keycode) {
	switch key {
	case sdl.K_SPACE:
		im.port1 |= 0x01
	case sdl.K_1:
		im.port1 |= 0x04
	case sdl.K_2:
		im.port1 |= 0x02
	case sdl.K_j:
		im.port1 |= 0x10
		im.port2 |= 0x10
	case sdl.K_a:
		im.port1 |= 0x20
		im.port2 |= 0x20
	case sdl.K_d:
		im.port1 |= 0x40
		im.port2 |= 0x40
	case sdl.K_ESCAPE:
		os.Exit(0)
	}
}

func (im *InvadersMachine) keyUp(key sdl.Keycode) {
	switch key {
	case sdl.K_SPACE:
		im.port1 &= 0xFE
	case sdl.K_1:
		im.port1 &= 0xFD
	case sdl.K_2:
		im.port1 &= 0xFB
	case sdl.K_j:
		im.port1 &= 0xEF
		im.port2 &= 0xEF
	case sdl.K_a:
		im.port1 &= 0xDF
		im.port2 &= 0xDF
	case sdl.K_d:
		im.port1 &= 0xBF
		im.port2 &= 0xBF
	case sdl.K_ESCAPE:
		os.Exit(0)
	}
}
