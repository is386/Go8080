package i8080

type InvadersMachine struct {
	cpu    *CPU
	shift0 uint8
	shift1 uint8
	offset uint8
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
	running := true
	for running {
		running = im.cpu.Execute()
	}
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
