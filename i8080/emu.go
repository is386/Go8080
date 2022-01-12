package i8080

import (
	"fmt"
	"io/ioutil"
	"math/bits"
)

// xra, ora sets AC to 0
// ana sets AC to ((c->a | val) & 0x08) != 0
// cmp sets AC to ~(c->a ^ result ^ val) & 0x10;
// pop psw sets AC to something weird

type Emulator struct {
	memory    [64 * 1024]uint8
	registers *Registers
	pc        uint16
	sp        uint16
	intEnable uint8
	flags     *Flags
}

func NewEmulator() *Emulator {
	return &Emulator{}
}

func (e *Emulator) LoadRom(filename string, offset uint16) {
	rom, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(rom); i++ {
		e.memory[offset+uint16(i)] = rom[i]
	}
}

func (e *Emulator) getBC() uint16 {
	return (uint16(e.registers.B) << 8) | uint16(e.registers.C)
}

func (e *Emulator) getDE() uint16 {
	return (uint16(e.registers.D) << 8) | uint16(e.registers.E)
}

func (e *Emulator) getHL() uint16 {
	return (uint16(e.registers.H) << 8) | uint16(e.registers.L)
}

func (e *Emulator) setBC(val uint16) {
	e.registers.B = uint8(val >> 8)
	e.registers.C = uint8(val & 0xff)
}

func (e *Emulator) setDE(val uint16) {
	e.registers.D = uint8(val >> 8)
	e.registers.E = uint8(val & 0xff)
}

func (e *Emulator) setHL(val uint16) {
	e.registers.H = uint8(val >> 8)
	e.registers.L = uint8(val & 0xff)
}

func (e *Emulator) setZero(val uint16) {
	if (val & 0x80) == 1 {
		e.flags.S = 1
	} else {
		e.flags.S = 0
	}
}

func (e *Emulator) setCarry(val uint16) {
	if val > 0xff {
		e.flags.CY = 1
	} else {
		e.flags.CY = 0
	}
}

func (e *Emulator) setSign(val uint16) {
	if (val & 0xff) == 0 {
		e.flags.Z = 1
	} else {
		e.flags.Z = 0
	}
}

func (e *Emulator) setZSP(val uint8) {
	e.setZero(uint16(val))
	e.setSign(uint16(val))
	e.setParity(uint16(val))
}

func (e *Emulator) setParity(val uint16) {
	e.flags.P = parity(uint(val & 0xff))
}

func (e *Emulator) setAuxCarry(val1 uint16, val2 uint16, total uint16) {
	if ((total ^ val1 ^ val2) & 0x10) > 0 {
		e.flags.AC = 1
	} else {
		e.flags.AC = 0
	}
}

func (e *Emulator) setAuxCarryInr(val uint8) {
	if (val & 0xf) == 0 {
		e.flags.AC = 1
	} else {
		e.flags.AC = 0
	}
}

func (e *Emulator) setAuxCarryDcr(val uint8) {
	if (val & 0xf) == 0xf {
		e.flags.AC = 0
	} else {
		e.flags.AC = 1
	}
}

func (e *Emulator) addToAccumulator(val uint8) {
	ans := uint16(e.registers.A) + uint16(val)
	e.setZero(ans)
	e.setSign(ans)
	e.setCarry(ans)
	e.setParity(ans)
	e.setAuxCarry(uint16(e.registers.A), uint16(val), ans)
	e.registers.A = uint8(ans & 0xff)
}

func (e *Emulator) subFromAccumulator(val uint8) {
	ans := uint16(e.registers.A) - uint16(val)
	e.setZero(ans)
	e.setSign(ans)
	e.setCarry(ans)
	e.setParity(ans)
	e.setAuxCarry(uint16(e.registers.A), uint16(val), ans)
	e.registers.A = uint8(ans & 0xff)
}

func (e *Emulator) fetch() uint8 {
	return e.memory[e.pc]
}

func (e *Emulator) decode(opcode uint8) func(*Emulator) uint16 {
	return INSTRUCTIONS[opcode]
}

func (e *Emulator) Execute() bool {
	opcode := e.fetch()
	instr := e.decode(opcode)
	steps := instr(e)
	if steps == 0 {
		fmt.Printf("unimplemented instruction: %x\n", opcode)
		return false
	}
	e.pc += steps
	return true
}

func unimplemented(e *Emulator) uint16 {
	return 0
}

func noOp(e *Emulator) uint16 {
	return 1
}

func lxiB(e *Emulator) uint16 {
	e.registers.C = e.memory[e.pc+1]
	e.registers.B = e.memory[e.pc+2]
	return 3
}

func parity(num uint) uint8 {
	ones := bits.OnesCount(num)
	if ones%2 == 0 {
		return 1
	}
	return 0
}

func addB(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.B)
	return 1
}

func addC(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.C)
	return 1
}

func addD(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.D)
	return 1
}

func addE(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.E)
	return 1
}

func addH(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.H)
	return 1
}

func addL(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.L)
	return 1
}

func addM(e *Emulator) uint16 {
	offset := e.getHL()
	e.addToAccumulator(e.memory[offset])
	return 1
}

func addA(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.A)
	return 1
}

func adi(e *Emulator) uint16 {
	e.addToAccumulator(e.memory[e.pc+1])
	return 2
}

func adcB(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.B + e.flags.CY)
	return 1
}

func adcC(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.C + e.flags.CY)
	return 1
}

func adcD(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.D + e.flags.CY)
	return 1
}

func adcE(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.E + e.flags.CY)
	return 1
}

func adcH(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.H + e.flags.CY)
	return 1
}

func adcL(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.L + e.flags.CY)
	return 1
}

func adcA(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.A + e.flags.CY)
	return 1
}

func adcM(e *Emulator) uint16 {
	offset := e.getHL()
	e.addToAccumulator(e.memory[offset])
	return 1
}

func aci(e *Emulator) uint16 {
	e.addToAccumulator(e.memory[e.pc+1] + e.flags.CY)
	return 2
}

func subB(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.B)
	return 1
}

func subC(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.C)
	return 1
}

func subD(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.D)
	return 1
}

func subE(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.E)
	return 1
}

func subH(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.H)
	return 1
}

func subL(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.L)
	return 1
}

func subA(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.A)
	return 1
}

func subM(e *Emulator) uint16 {
	offset := e.getHL()
	e.subFromAccumulator(e.memory[offset])
	return 1
}

func sui(e *Emulator) uint16 {
	e.subFromAccumulator(e.memory[e.pc+1])
	return 2
}

func sbbB(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.B - e.flags.CY)
	return 1
}

func sbbC(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.C - e.flags.CY)
	return 1
}

func sbbD(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.D - e.flags.CY)
	return 1
}

func sbbE(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.E - e.flags.CY)
	return 1
}

func sbbH(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.H - e.flags.CY)
	return 1
}

func sbbL(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.L - e.flags.CY)
	return 1
}

func sbbA(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.A - e.flags.CY)
	return 1
}

func sbbM(e *Emulator) uint16 {
	offset := e.getHL()
	e.subFromAccumulator(e.memory[offset] - e.flags.CY)
	return 1
}

func sbi(e *Emulator) uint16 {
	e.subFromAccumulator(e.memory[e.pc+1] - e.flags.CY)
	return 2
}

func inrB(e *Emulator) uint16 {
	e.registers.B += 1
	e.setZSP(e.registers.B)
	e.setAuxCarryInr(e.registers.B)
	return 1
}

func inrC(e *Emulator) uint16 {
	e.registers.C += 1
	e.setZSP(e.registers.C)
	e.setAuxCarryInr(e.registers.C)
	return 1
}

func inrD(e *Emulator) uint16 {
	e.registers.D += 1
	e.setZSP(e.registers.D)
	e.setAuxCarryInr(e.registers.D)
	return 1
}

func inrE(e *Emulator) uint16 {
	e.registers.E += 1
	e.setZSP(e.registers.E)
	e.setAuxCarryInr(e.registers.E)
	return 1
}

func inrH(e *Emulator) uint16 {
	e.registers.H += 1
	e.setZSP(e.registers.H)
	e.setAuxCarryInr(e.registers.H)
	return 1
}

func inrL(e *Emulator) uint16 {
	e.registers.L += 1
	e.setZSP(e.registers.L)
	e.setAuxCarryInr(e.registers.L)
	return 1
}

func inrA(e *Emulator) uint16 {
	e.registers.A += 1
	e.setZSP(e.registers.A)
	e.setAuxCarryInr(e.registers.A)
	return 1
}

func inrM(e *Emulator) uint16 {
	offset := e.getHL()
	e.memory[offset] += 1
	e.setZSP(e.memory[offset])
	e.setAuxCarryInr(e.memory[offset])
	return 1
}

func dcrB(e *Emulator) uint16 {
	e.registers.B -= 1
	e.setZSP(e.registers.B)
	e.setAuxCarryDcr(e.registers.B)
	return 1
}

func dcrC(e *Emulator) uint16 {
	e.registers.C -= 1
	e.setZSP(e.registers.C)
	e.setAuxCarryDcr(e.registers.C)
	return 1
}

func dcrD(e *Emulator) uint16 {
	e.registers.D -= 1
	e.setZSP(e.registers.D)
	e.setAuxCarryDcr(e.registers.D)
	return 1
}

func dcrE(e *Emulator) uint16 {
	e.registers.E -= 1
	e.setZSP(e.registers.E)
	e.setAuxCarryDcr(e.registers.E)
	return 1
}

func dcrH(e *Emulator) uint16 {
	e.registers.H -= 1
	e.setZSP(e.registers.H)
	e.setAuxCarryDcr(e.registers.H)
	return 1
}

func dcrL(e *Emulator) uint16 {
	e.registers.L -= 1
	e.setZSP(e.registers.L)
	e.setAuxCarryDcr(e.registers.L)
	return 1
}

func dcrA(e *Emulator) uint16 {
	e.registers.A -= 1
	e.setZSP(e.registers.A)
	e.setAuxCarryDcr(e.registers.A)
	return 1
}

func dcrM(e *Emulator) uint16 {
	offset := e.getHL()
	e.memory[offset] -= 1
	e.setZSP(e.memory[offset])
	e.setAuxCarryInr(e.memory[offset])
	return 1
}

func inxB(e *Emulator) uint16 {
	e.setBC(e.getBC() + 1)
	return 1
}

func inxD(e *Emulator) uint16 {
	e.setDE(e.getDE() + 1)
	return 1
}

func inxH(e *Emulator) uint16 {
	e.setHL(e.getHL() + 1)
	return 1
}

func inxSP(e *Emulator) uint16 {
	e.sp += 1
	return 1
}

func dcxB(e *Emulator) uint16 {
	e.setBC(e.getBC() - 1)
	return 1
}

func dcxD(e *Emulator) uint16 {
	e.setDE(e.getDE() - 1)
	return 1
}

func dcxH(e *Emulator) uint16 {
	e.setHL(e.getHL() - 1)
	return 1
}

func dcxSP(e *Emulator) uint16 {
	e.sp -= 1
	return 1
}

func dadB(e *Emulator) uint16 {
	e.setHL(e.getHL() + e.getBC())
	e.setCarry(e.getHL())
	return 1
}

func dadD(e *Emulator) uint16 {
	e.setHL(e.getHL() + e.getDE())
	e.setCarry(e.getHL())
	return 1
}

func dadH(e *Emulator) uint16 {
	e.setHL(e.getHL() + e.getHL())
	e.setCarry(e.getHL())
	return 1
}

func dadSP(e *Emulator) uint16 {
	e.setHL(e.getHL() + e.sp)
	e.setCarry(e.getHL())
	return 1
}

func daa(e *Emulator) uint16 {
	cy := e.flags.CY
	lsb := e.registers.A & 0x0f
	msb := e.registers.A >> 4
	correction := 0

	if lsb > 9 || e.flags.AC == 1 {
		correction += 0x06
	}

	if (e.flags.CY == 1 || msb > 9) || (msb >= 9 && lsb > 9) {
		correction += 0x60
		cy = 1
	}

	e.addToAccumulator(uint8(correction))
	e.flags.CY = cy
	return 1
}
