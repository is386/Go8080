package i8080

import (
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	BUTTONS = map[sdl.Keycode]uint8{
		sdl.K_SPACE:  0x1,
		sdl.K_a:      0x20,
		sdl.K_d:      0x40,
		sdl.K_j:      0x10,
		sdl.K_RETURN: 0x04,
	}
)

type InvadersMachine struct {
	cpu           *CPU
	shift0        uint8
	shift1        uint8
	offset        uint8
	port1         uint8
	interrupt     int
	lastTimer     float64
	nextInterrupt float64
}

func NewInvadersMachine(filename string, pcStart uint16, showDebug bool, isTest bool) *InvadersMachine {
	im := &InvadersMachine{}
	cpu := NewCPU(pcStart, showDebug, isTest, im.PortIn, im.PortOut)
	cpu.LoadRom(filename, pcStart)
	im.setEmu(cpu)
	return im
}

func (im *InvadersMachine) setEmu(c *CPU) {
	im.cpu = c
}

func (im *InvadersMachine) Run() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Input", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		return
	}
	defer window.Destroy()

	running := true
	for running {
		now := getTime()

		if im.lastTimer == 0 {
			im.lastTimer = now
			im.nextInterrupt = im.lastTimer + 16000.0
			im.interrupt = 1
		}

		if im.cpu.intEnable == 1 && now > im.nextInterrupt {
			if im.interrupt == 1 {
				im.cpu.interrupt(1)
				im.interrupt = 2
			} else {
				im.cpu.interrupt(2)
				im.interrupt = 1
			}
			im.nextInterrupt = now + 8000.0
		}

		sinceLast := now - im.lastTimer
		cyclesMax := 2 * sinceLast
		cycles := 0.0
		c := 0

		for cyclesMax > cycles && running {
			running, c = im.cpu.Execute()
			cycles += float64(c)

			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch e := event.(type) {
				case *sdl.QuitEvent:
					running = false
				case *sdl.KeyboardEvent:
					switch e.Type {
					case sdl.KEYDOWN:
						im.keyDown(e.Keysym.Sym)
					case sdl.KEYUP:
						im.keyUp(e.Keysym.Sym)
					}
				}
			}
		}
		im.lastTimer = now
	}
}

func getTime() float64 {
	return float64(time.Now().UnixNano()) / (float64(time.Millisecond) / float64(time.Nanosecond))
}

func (im *InvadersMachine) fetchNext() uint8 {
	return im.cpu.mem[im.cpu.pc+1]
}

func (im *InvadersMachine) setAccumulator(val uint8) {
	im.cpu.reg.A = val
}

func (im *InvadersMachine) getAccumulator() uint8 {
	return im.cpu.reg.A
}

func (im *InvadersMachine) PortIn() {
	port := im.fetchNext()
	switch port {
	case 0:
		im.setAccumulator(1)
	case 1:
		im.setAccumulator(0)
	case 3:
		v16 := (uint16(im.shift1) << 8) | uint16(im.shift0)
		v8 := uint8((v16 >> (8 - im.offset)) & 0xff)
		im.setAccumulator(v8)
	}
}

func (im *InvadersMachine) PortOut() {
	a := im.getAccumulator()
	port := im.fetchNext()
	switch port {
	case 2:
		im.offset = a & 0x7
	case 4:
		im.shift0 = im.shift1
		im.shift1 = a
	}
}

func (im *InvadersMachine) keyDown(key sdl.Keycode) {
	im.port1 |= BUTTONS[key]
}

func (im *InvadersMachine) keyUp(key sdl.Keycode) {
	im.port1 &= ^BUTTONS[key]
}
