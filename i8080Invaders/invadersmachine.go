package i8080Invaders

import (
	"fmt"

	"github.com/is386/Go8080/i8080"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	FILENAME         = "i8080Invaders/INVADERS.COM"
	SPEED            = 1
	FPS              = 59.541985
	CLOCK_SPEED      = 1996800
	CYCLES_PER_FRAME = float64(CLOCK_SPEED) / FPS
	BUTTONS          = map[sdl.Keycode]uint8{
		sdl.K_SPACE:  0x1,
		sdl.K_a:      0x20,
		sdl.K_d:      0x40,
		sdl.K_j:      0x10,
		sdl.K_RETURN: 0x04,
	}
)

type InvadersMachine struct {
	cpu                             *i8080.CPU
	screen                          *Screen
	nextInt                         uint8
	port1                           uint8
	shiftMsb, shiftLsb, shiftOffset uint8
	showDebug                       bool
}

func NewInvadersMachine(showDebug bool) *InvadersMachine {
	im := &InvadersMachine{nextInt: 0xCF, screen: NewScreen(), showDebug: showDebug}
	cpu := i8080.NewCPU(0x0, im.PortIn, im.PortOut)
	cpu.LoadRom(FILENAME)
	im.cpu = cpu
	return im
}

func (im *InvadersMachine) Run() {
	lastTime := uint32(0)
	running := true

	for running {
		currentTime := sdl.GetTicks()
		dt := currentTime - lastTime
		lastTime = currentTime

		running = im.pollSDL()
		if im.showDebug {
			im.printState()
		}

		im.runCPU(dt * uint32(SPEED))
		im.screen.Update()
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

func (im *InvadersMachine) runCPU(ms uint32) {
	count := uint32(0)
	for count < (ms * uint32(CLOCK_SPEED) / 1000) {
		cyc := im.cpu.GetCycles()
		im.cpu.Execute()
		elapsed := im.cpu.GetCycles() - cyc
		count += uint32(elapsed)

		if im.cpu.GetCycles() >= int(CYCLES_PER_FRAME/2) {
			im.cpu.SubtractCycles(int(CYCLES_PER_FRAME / 2))
			im.sendInterrupt()
		}
	}
}

func (im *InvadersMachine) sendInterrupt() {
	im.cpu.Interrupt(im.nextInt)
	if im.nextInt == 0xD7 {
		im.screen.Draw(im)
	}
	if im.nextInt == 0xCF {
		im.nextInt = 0xD7
	} else {
		im.nextInt = 0xCF
	}
}

func (im *InvadersMachine) PortIn(port uint8) {
	val := uint8(0xFF)
	switch port {
	case 0:
		break
	case 1:
		val = im.port1
	case 3:
		shift := (uint16(im.shiftMsb) << 8) | uint16(im.shiftLsb)
		val = uint8((shift >> (8 - im.shiftOffset)) & 0xFF)
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
	}
}

func (im *InvadersMachine) keyDown(key sdl.Keycode) {
	if val, ok := BUTTONS[key]; ok {
		im.port1 |= val
	}

}

func (im *InvadersMachine) keyUp(key sdl.Keycode) {
	if val, ok := BUTTONS[key]; ok {
		im.port1 &= ^val
	}
}

func (tm *InvadersMachine) printState() {
	mem := tm.cpu.GetMemory()
	pc := tm.cpu.GetPC()
	sp := tm.cpu.GetSP()
	cyc := tm.cpu.GetCycles()
	af := tm.cpu.GetAF()
	bc := tm.cpu.GetBC()
	de := tm.cpu.GetDE()
	hl := tm.cpu.GetHL()

	fmt.Printf("\nPC: %04X, AF: %04X, BC: %04X, DE: %04X, HL: %04X, SP: %04X, CYC: %04d	(%02X %02X %02X %02X)",
		pc, af, bc, de, hl, sp, cyc, mem[pc], mem[pc+1], mem[pc+2], mem[pc+3])
}
