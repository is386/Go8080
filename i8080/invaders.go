package i8080

type InvadersMachine struct {
	emu    *Emulator
	shift0 uint8
	shift1 uint8
	offset uint8
}

func NewInvadersMachine(filename string, pcStart uint16, showDebug bool, isTest bool) *InvadersMachine {
	im := &InvadersMachine{}
	emu := NewEmulator(pcStart, showDebug, isTest, im.PortIn, im.PortOut)
	emu.LoadRom(filename, pcStart)
	im.setEmu(emu)
	return im
}

func (im *InvadersMachine) setEmu(e *Emulator) {
	im.emu = e
}

func (im *InvadersMachine) Run() {
	running := true
	for running {
		running = im.emu.Execute()
	}
}

func (im *InvadersMachine) PortIn() {
}

func (im *InvadersMachine) PortOut() {
}
