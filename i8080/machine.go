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

func (im *InvadersMachine) setEmu(e *CPU) {
	im.cpu = e
}

func (im *InvadersMachine) Run() {
	running := true
	for running {
		running = im.cpu.Execute()
	}
}

func (im *InvadersMachine) PortIn() {
}

func (im *InvadersMachine) PortOut() {
}
