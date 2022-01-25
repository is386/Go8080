package i8080

import (
	"fmt"
	"io/ioutil"
)

type Emulator struct {
	memory    [64 * 1024]uint8
	registers *Registers
	pc        uint16
	sp        uint16
	intEnable uint8
	flags     *Flags
}

func NewEmulator() *Emulator {
	return &Emulator{registers: &Registers{}, flags: &Flags{}, pc: 0x100}
}

func (e *Emulator) LoadRom(filename string, offset uint16) {
	rom, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(rom); i++ {
		e.memory[offset+uint16(i)] = rom[i]
	}

	// CPU DIAg
	e.memory[0x7] = 0xc9
	e.memory[0x59c] = 0xc3 //JMP
	e.memory[0x59d] = 0xc2
	e.memory[0x59e] = 0x05
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
	if (val & 0xff) == 0 {
		e.flags.Z = 1
	} else {
		e.flags.Z = 0
	}
}

func (e *Emulator) setSign(val uint16) {
	if (val & 0x80) != 0 {
		e.flags.S = 1
	} else {
		e.flags.S = 0
	}
}

func parity(num uint16) uint8 {
	ones := uint16(0)
	for i := 0; i < 8; i++ {
		ones += ((num >> i) & 1)
	}
	if (ones % 2) == 0 {
		return 1
	}
	return 0
}

func (e *Emulator) setParity(val uint16) {
	e.flags.P = parity(val)
}

func (e *Emulator) setZSP(val uint8) {
	e.setZero(uint16(val))
	e.setSign(uint16(val))
	e.setParity(uint16(val))
}

func (e *Emulator) setAuxCarry(a uint8, b uint8, cy uint8) {
	result := uint16(a + b + cy)
	carry := result ^ uint16(a) ^ uint16(b)
	if (carry & (1 << 4)) == 1 {
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

func (e *Emulator) addToAccumulator(val uint8, cy uint8) {
	ans := uint16(e.registers.A) + uint16(val) + uint16(cy)
	e.setZero(ans)
	e.setSign(ans)
	if ans > 0xff {
		e.flags.CY = 1
	} else {
		e.flags.CY = 0
	}
	e.setParity(ans & 0xff)
	e.setAuxCarry(e.registers.A, val, cy)
	e.registers.A = uint8(ans & 0xff)
}

func (e *Emulator) subFromAccumulator(val uint8, cy uint8) {
	ans := uint16(e.registers.A) - uint16(val) - uint16(cy)
	e.setZero(ans)
	e.setSign(ans)
	if ans > 0xff {
		e.flags.CY = 1
	} else {
		e.flags.CY = 0
	}
	e.setParity(ans & 0xff)
	e.setAuxCarry(e.registers.A, val, cy)
	e.registers.A = uint8(ans)
}

func (e *Emulator) andAccumulator(val uint8) {
	ans := uint16(e.registers.A & val)
	e.setZero(ans)
	e.setSign(ans)
	e.flags.CY = 0
	e.setParity(ans & 0xff)
	if ((e.registers.A | val) & 0x08) != 0 {
		e.flags.AC = 1
	} else {
		e.flags.AC = 0
	}
	e.registers.A = uint8(ans & 0xff)
}

func (e *Emulator) xorAccumulator(val uint8) {
	ans := uint16(e.registers.A ^ val)
	e.setZero(ans)
	e.setSign(ans)
	e.flags.CY = 0
	e.setParity(ans & 0xff)
	e.flags.AC = 0
	e.registers.A = uint8(ans & 0xff)
}

func (e *Emulator) orAccumulator(val uint8) {
	ans := uint16(e.registers.A | val)
	e.setZero(ans)
	e.setSign(ans)
	e.flags.CY = 0
	e.setParity(ans & 0xff)
	e.flags.AC = 0
	e.registers.A = uint8(ans & 0xff)
}

func (e *Emulator) compareAccumulator(val uint8) {
	ans := uint16(e.registers.A) - uint16(val)
	e.setZero(ans)
	e.setSign(ans)
	if ans > 0xff {
		e.flags.CY = 1
	} else {
		e.flags.CY = 0
	}
	e.setParity(ans)
	ac := uint16(e.registers.A) ^ ans ^ uint16(val)
	if ac == 1 {
		ac = 0
	} else {
		ac = 1
	}
	if (ac & 0x10) == 1 {
		e.flags.AC = 0
	} else {
		e.flags.AC = 1
	}
}

func (e *Emulator) fetch() uint8 {
	return e.memory[e.pc]
}

func (e *Emulator) decode(opcode uint8) func(*Emulator) uint16 {
	return INSTRUCTIONS[opcode]
}

func (e *Emulator) Execute(count int) bool {
	opcode := e.fetch()
	fmt.Printf("\nPC:%04x OP:%02x SP:%04x BC:%04x DE:%04x HL:%04x A:%02x Cy:%d AC:%d S:%d Z:%d P:%d",
		e.pc, opcode, e.sp, e.getBC(), e.getDE(), e.getHL(), e.registers.A, e.flags.CY, e.flags.AC,
		e.flags.S, e.flags.Z, e.flags.P)
	instr := e.decode(opcode)
	steps := instr(e)
	if steps == 5 {
		fmt.Printf("unimplemented instruction: %x\n", opcode)
		return false
	}
	e.pc += steps
	if e.pc == 5 {
		if e.registers.C == 9 {
			fmt.Println()
			offset := e.getDE()
			str := e.memory[offset]
			for str != '$' {
				fmt.Printf("%c", str)
				offset += 1
				str = e.memory[offset]
			}
		} else if e.registers.C == 2 {
			fmt.Printf("%c", e.registers.E)
		}
	} else if e.pc == 0 {
		fmt.Printf("\n\n%x", e.memory[275])
		return false
	}
	return true
}

func unimplemented(e *Emulator) uint16 {
	return 5
}

func noOp(e *Emulator) uint16 {
	return 1
}

func addB(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.B, 0)
	return 1
}

func addC(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.C, 0)
	return 1
}

func addD(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.D, 0)
	return 1
}

func addE(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.E, 0)
	return 1
}

func addH(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.H, 0)
	return 1
}

func addL(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.L, 0)
	return 1
}

func addM(e *Emulator) uint16 {
	offset := e.getHL()
	e.addToAccumulator(e.memory[offset], 0)
	return 1
}

func addA(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.A, 0)
	return 1
}

func adi(e *Emulator) uint16 {
	e.addToAccumulator(e.memory[e.pc+1], 0)
	return 2
}

func adcB(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.B, e.flags.CY)
	return 1
}

func adcC(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.C, e.flags.CY)
	return 1
}

func adcD(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.D, e.flags.CY)
	return 1
}

func adcE(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.E, e.flags.CY)
	return 1
}

func adcH(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.H, e.flags.CY)
	return 1
}

func adcL(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.L, e.flags.CY)
	return 1
}

func adcA(e *Emulator) uint16 {
	e.addToAccumulator(e.registers.A, e.flags.CY)
	return 1
}

func adcM(e *Emulator) uint16 {
	offset := e.getHL()
	e.addToAccumulator(e.memory[offset], e.flags.CY)
	return 1
}

func aci(e *Emulator) uint16 {
	e.addToAccumulator(e.memory[e.pc+1], e.flags.CY)
	return 2
}

func subB(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.B, 0)
	return 1
}

func subC(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.C, 0)
	return 1
}

func subD(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.D, 0)
	return 1
}

func subE(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.E, 0)
	return 1
}

func subH(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.H, 0)
	return 1
}

func subL(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.L, 0)
	return 1
}

func subA(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.A, 0)
	return 1
}

func subM(e *Emulator) uint16 {
	offset := e.getHL()
	e.subFromAccumulator(e.memory[offset], 0)
	return 1
}

func sui(e *Emulator) uint16 {
	e.subFromAccumulator(e.memory[e.pc+1], 0)
	return 2
}

func sbbB(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.B, e.flags.CY)
	return 1
}

func sbbC(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.C, e.flags.CY)
	return 1
}

func sbbD(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.D, e.flags.CY)
	return 1
}

func sbbE(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.E, e.flags.CY)
	return 1
}

func sbbH(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.H, e.flags.CY)
	return 1
}

func sbbL(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.L, e.flags.CY)
	return 1
}

func sbbA(e *Emulator) uint16 {
	e.subFromAccumulator(e.registers.A, e.flags.CY)
	return 1
}

func sbbM(e *Emulator) uint16 {
	offset := e.getHL()
	e.subFromAccumulator(e.memory[offset], e.flags.CY)
	return 1
}

func sbi(e *Emulator) uint16 {
	e.subFromAccumulator(e.memory[e.pc+1], e.flags.CY)
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
	ans := uint32(e.getHL()) + uint32(e.getBC())
	e.setHL(uint16(ans))
	if (ans & 0xffff0000) > 0 {
		e.flags.CY = 1
	} else {
		e.flags.CY = 0
	}
	return 1
}

func dadD(e *Emulator) uint16 {
	ans := uint32(e.getHL()) + uint32(e.getDE())
	e.setHL(uint16(ans))
	if (ans & 0xffff0000) > 0 {
		e.flags.CY = 1
	} else {
		e.flags.CY = 0
	}
	return 1
}

func dadH(e *Emulator) uint16 {
	ans := uint32(e.getHL()) + uint32(e.getHL())
	e.setHL(uint16(ans))
	if (ans & 0xffff0000) > 0 {
		e.flags.CY = 1
	} else {
		e.flags.CY = 0
	}
	return 1
}

func dadSP(e *Emulator) uint16 {
	ans := uint32(e.getHL()) + uint32(e.sp)
	e.setHL(uint16(ans))
	if (ans & 0xffff0000) > 0 {
		e.flags.CY = 1
	} else {
		e.flags.CY = 0
	}
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

	e.addToAccumulator(uint8(correction), 0)
	e.flags.CY = cy
	return 1
}

func jmp(e *Emulator) uint16 {
	e.pc = (uint16(e.memory[e.pc+2]) << 8) | uint16(e.memory[e.pc+1])
	return 0
}

func jnz(e *Emulator) uint16 {
	if e.flags.Z == 0 {
		return jmp(e)
	}
	return 3
}

func jz(e *Emulator) uint16 {
	if e.flags.Z == 1 {
		return jmp(e)
	}
	return 3
}

func jnc(e *Emulator) uint16 {
	if e.flags.CY == 0 {
		return jmp(e)
	}
	return 3
}

func jc(e *Emulator) uint16 {
	if e.flags.CY == 1 {
		return jmp(e)
	}
	return 3
}

func jpo(e *Emulator) uint16 {
	if e.flags.P == 0 {
		return jmp(e)
	}
	return 3
}

func jpe(e *Emulator) uint16 {
	if e.flags.P == 1 {
		return jmp(e)
	}
	return 3
}

func jp(e *Emulator) uint16 {
	if e.flags.S == 0 {
		return jmp(e)
	}
	return 3
}

func jm(e *Emulator) uint16 {
	if e.flags.S == 1 {
		return jmp(e)
	}
	return 3
}

func ret(e *Emulator) uint16 {
	e.pc = (uint16(e.memory[e.sp]) | (uint16(e.memory[e.sp+1]) << 8))
	e.sp += 2
	return 1
}

func rz(e *Emulator) uint16 {
	if e.flags.Z == 1 {
		return ret(e)
	}
	return 1
}

func rnz(e *Emulator) uint16 {
	if e.flags.Z == 0 {
		return ret(e)
	}
	return 1
}

func rnc(e *Emulator) uint16 {
	if e.flags.CY == 0 {
		return ret(e)
	}
	return 1
}

func rc(e *Emulator) uint16 {
	if e.flags.CY == 1 {
		return ret(e)
	}
	return 1
}

func rpo(e *Emulator) uint16 {
	if e.flags.P == 0 {
		return ret(e)
	}
	return 1
}

func rpe(e *Emulator) uint16 {
	if e.flags.P == 1 {
		return ret(e)
	}
	return 1
}

func rp(e *Emulator) uint16 {
	if e.flags.S == 0 {
		return ret(e)
	}
	return 1
}

func rm(e *Emulator) uint16 {
	if e.flags.S == 1 {
		return ret(e)
	}
	return 1
}

func mviB(e *Emulator) uint16 {
	e.registers.B = e.memory[e.pc+1]
	return 2
}

func mviC(e *Emulator) uint16 {
	e.registers.C = e.memory[e.pc+1]
	return 2
}

func mviD(e *Emulator) uint16 {
	e.registers.D = e.memory[e.pc+1]
	return 2
}

func mviE(e *Emulator) uint16 {
	e.registers.E = e.memory[e.pc+1]
	return 2
}

func mviH(e *Emulator) uint16 {
	e.registers.H = e.memory[e.pc+1]
	return 2
}

func mviL(e *Emulator) uint16 {
	e.registers.L = e.memory[e.pc+1]
	return 2
}

func mviA(e *Emulator) uint16 {
	e.registers.A = e.memory[e.pc+1]
	return 2
}

func mviM(e *Emulator) uint16 {
	e.memory[e.getHL()] = e.memory[e.pc+1]
	return 2
}

func call(e *Emulator) uint16 {
	ret := e.pc + 2
	e.memory[e.sp-1] = uint8(ret>>8) & uint8(0xff)
	e.memory[e.sp-2] = uint8(ret) & uint8(0xff)
	e.sp = e.sp - 2
	e.pc = (uint16(e.memory[e.pc+2]) << 8) | uint16(e.memory[e.pc+1])
	return 0
}

func cz(e *Emulator) uint16 {
	if e.flags.Z == 1 {
		return call(e)
	}
	return 3
}

func cnz(e *Emulator) uint16 {
	if e.flags.Z == 0 {
		return call(e)
	}
	return 3
}

func cnc(e *Emulator) uint16 {
	if e.flags.CY == 0 {
		return call(e)
	}
	return 3
}

func cc(e *Emulator) uint16 {
	if e.flags.CY == 1 {
		return call(e)
	}
	return 3
}

func cpo(e *Emulator) uint16 {
	if e.flags.P == 0 {
		return call(e)
	}
	return 3
}

func cpe(e *Emulator) uint16 {
	if e.flags.P == 1 {
		return call(e)
	}
	return 3
}

func cp(e *Emulator) uint16 {
	if e.flags.S == 0 {
		return call(e)
	}
	return 3
}

func cm(e *Emulator) uint16 {
	if e.flags.S == 1 {
		return call(e)
	}
	return 3
}

func lxiB(e *Emulator) uint16 {
	e.registers.C = e.memory[e.pc+1]
	e.registers.B = e.memory[e.pc+2]
	return 3
}

func lxiD(e *Emulator) uint16 {
	e.registers.E = e.memory[e.pc+1]
	e.registers.D = e.memory[e.pc+2]
	return 3
}

func lxiH(e *Emulator) uint16 {
	e.registers.L = e.memory[e.pc+1]
	e.registers.H = e.memory[e.pc+2]
	return 3
}

func lxiSP(e *Emulator) uint16 {
	e.sp = (uint16(e.memory[e.pc+2]) << 8) | uint16(e.memory[e.pc+1])
	return 3
}

func lda(e *Emulator) uint16 {
	addr := (uint16(e.memory[e.pc+2]) << 8) | uint16(e.memory[e.pc+1])
	e.registers.A = e.memory[addr]
	return 3
}

func ldaxB(e *Emulator) uint16 {
	e.registers.A = e.memory[e.getBC()]
	return 1
}

func ldaxD(e *Emulator) uint16 {
	e.registers.A = e.memory[e.getDE()]
	return 1
}

func movBB(e *Emulator) uint16 {
	return 1
}

func movBC(e *Emulator) uint16 {
	e.registers.B = e.registers.C
	return 1
}

func movBD(e *Emulator) uint16 {
	e.registers.B = e.registers.D
	return 1
}

func movBE(e *Emulator) uint16 {
	e.registers.B = e.registers.E
	return 1
}

func movBH(e *Emulator) uint16 {
	e.registers.B = e.registers.H
	return 1
}

func movBL(e *Emulator) uint16 {
	e.registers.B = e.registers.L
	return 1
}

func movBM(e *Emulator) uint16 {
	offset := e.getHL()
	e.registers.B = e.memory[offset]
	return 1
}

func movBA(e *Emulator) uint16 {
	e.registers.B = e.registers.A
	return 1
}

func movCB(e *Emulator) uint16 {
	e.registers.C = e.registers.B
	return 1
}

func movCC(e *Emulator) uint16 {
	return 1
}

func movCD(e *Emulator) uint16 {
	e.registers.C = e.registers.D
	return 1
}

func movCE(e *Emulator) uint16 {
	e.registers.C = e.registers.E
	return 1
}

func movCH(e *Emulator) uint16 {
	e.registers.C = e.registers.H
	return 1
}

func movCL(e *Emulator) uint16 {
	e.registers.C = e.registers.L
	return 1
}

func movCM(e *Emulator) uint16 {
	offset := e.getHL()
	e.registers.C = e.memory[offset]
	return 1
}

func movCA(e *Emulator) uint16 {
	e.registers.C = e.registers.A
	return 1
}

func movDB(e *Emulator) uint16 {
	e.registers.D = e.registers.B
	return 1
}

func movDC(e *Emulator) uint16 {
	e.registers.D = e.registers.C
	return 1
}

func movDD(e *Emulator) uint16 {
	return 1
}

func movDE(e *Emulator) uint16 {
	e.registers.D = e.registers.E
	return 1
}

func movDH(e *Emulator) uint16 {
	e.registers.D = e.registers.H
	return 1
}

func movDL(e *Emulator) uint16 {
	e.registers.D = e.registers.L
	return 1
}

func movDM(e *Emulator) uint16 {
	offset := e.getHL()
	e.registers.D = e.memory[offset]
	return 1
}

func movDA(e *Emulator) uint16 {
	e.registers.D = e.registers.A
	return 1
}

func movEB(e *Emulator) uint16 {
	e.registers.E = e.registers.B
	return 1
}

func movEC(e *Emulator) uint16 {
	e.registers.E = e.registers.C
	return 1
}

func movED(e *Emulator) uint16 {
	e.registers.E = e.registers.D
	return 1
}

func movEE(e *Emulator) uint16 {
	return 1
}

func movEH(e *Emulator) uint16 {
	e.registers.E = e.registers.H
	return 1
}

func movEL(e *Emulator) uint16 {
	e.registers.E = e.registers.L
	return 1
}

func movEM(e *Emulator) uint16 {
	offset := e.getHL()
	e.registers.E = e.memory[offset]
	return 1
}

func movEA(e *Emulator) uint16 {
	e.registers.E = e.registers.A
	return 1
}

func movHB(e *Emulator) uint16 {
	e.registers.H = e.registers.B
	return 1
}

func movHC(e *Emulator) uint16 {
	e.registers.H = e.registers.C
	return 1
}

func movHD(e *Emulator) uint16 {
	e.registers.H = e.registers.D
	return 1
}

func movHE(e *Emulator) uint16 {
	e.registers.H = e.registers.E
	return 1
}

func movHH(e *Emulator) uint16 {
	return 1
}

func movHL(e *Emulator) uint16 {
	e.registers.H = e.registers.L
	return 1
}

func movHM(e *Emulator) uint16 {
	offset := e.getHL()
	e.registers.H = e.memory[offset]
	return 1
}

func movHA(e *Emulator) uint16 {
	e.registers.H = e.registers.A
	return 1
}

func movLB(e *Emulator) uint16 {
	e.registers.L = e.registers.B
	return 1
}

func movLC(e *Emulator) uint16 {
	e.registers.L = e.registers.C
	return 1
}

func movLD(e *Emulator) uint16 {
	e.registers.L = e.registers.D
	return 1
}

func movLE(e *Emulator) uint16 {
	e.registers.L = e.registers.E
	return 1
}

func movLH(e *Emulator) uint16 {
	e.registers.L = e.registers.H
	return 1
}

func movLL(e *Emulator) uint16 {
	return 1
}

func movLM(e *Emulator) uint16 {
	offset := e.getHL()
	e.registers.L = e.memory[offset]
	return 1
}

func movLA(e *Emulator) uint16 {
	e.registers.L = e.registers.A
	return 1
}

func movAB(e *Emulator) uint16 {
	e.registers.A = e.registers.B
	return 1
}

func movAC(e *Emulator) uint16 {
	e.registers.A = e.registers.C
	return 1
}

func movAD(e *Emulator) uint16 {
	e.registers.A = e.registers.D
	return 1
}

func movAE(e *Emulator) uint16 {
	e.registers.A = e.registers.E
	return 1
}

func movAH(e *Emulator) uint16 {
	e.registers.A = e.registers.H
	return 1
}

func movAL(e *Emulator) uint16 {
	e.registers.A = e.registers.L
	return 1
}

func movAM(e *Emulator) uint16 {
	offset := e.getHL()
	e.registers.A = e.memory[offset]
	return 1
}

func movAA(e *Emulator) uint16 {
	return 1
}

func movMB(e *Emulator) uint16 {
	e.memory[e.getHL()] = e.registers.B
	return 1
}

func movMC(e *Emulator) uint16 {
	e.memory[e.getHL()] = e.registers.C
	return 1
}

func movMD(e *Emulator) uint16 {
	e.memory[e.getHL()] = e.registers.D
	return 1
}

func movME(e *Emulator) uint16 {
	e.memory[e.getHL()] = e.registers.E
	return 1
}

func movMH(e *Emulator) uint16 {
	e.memory[e.getHL()] = e.registers.H
	return 1
}

func movML(e *Emulator) uint16 {
	e.memory[e.getHL()] = e.registers.L
	return 1
}

func movMA(e *Emulator) uint16 {
	e.memory[e.getHL()] = e.registers.A
	return 1
}

func sta(e *Emulator) uint16 {
	addr := (uint16(e.memory[e.pc+2]) << 8) | uint16(e.memory[e.pc+1])
	e.memory[addr] = e.registers.A
	return 3
}

func staxB(e *Emulator) uint16 {
	e.memory[e.getBC()] = e.registers.A
	return 1
}

func staxD(e *Emulator) uint16 {
	e.memory[e.getDE()] = e.registers.A
	return 1
}

func anaB(e *Emulator) uint16 {
	e.andAccumulator(e.registers.B)
	return 1
}

func anaC(e *Emulator) uint16 {
	e.andAccumulator(e.registers.C)
	return 1
}

func anaD(e *Emulator) uint16 {
	e.andAccumulator(e.registers.D)
	return 1
}

func anaE(e *Emulator) uint16 {
	e.andAccumulator(e.registers.E)
	return 1
}

func anaH(e *Emulator) uint16 {
	e.andAccumulator(e.registers.H)
	return 1
}

func anaL(e *Emulator) uint16 {
	e.andAccumulator(e.registers.L)
	return 1
}

func anaA(e *Emulator) uint16 {
	e.andAccumulator(e.registers.A)
	return 1
}

func anaM(e *Emulator) uint16 {
	offset := e.getHL()
	e.andAccumulator(e.memory[offset])
	return 1
}

func ani(e *Emulator) uint16 {
	e.andAccumulator(e.memory[e.pc+1])
	return 2
}

func oraB(e *Emulator) uint16 {
	e.orAccumulator(e.registers.B)
	return 1
}

func oraC(e *Emulator) uint16 {
	e.orAccumulator(e.registers.C)
	return 1
}

func oraD(e *Emulator) uint16 {
	e.orAccumulator(e.registers.D)
	return 1
}

func oraE(e *Emulator) uint16 {
	e.orAccumulator(e.registers.E)
	return 1
}

func oraH(e *Emulator) uint16 {
	e.orAccumulator(e.registers.H)
	return 1
}

func oraL(e *Emulator) uint16 {
	e.orAccumulator(e.registers.L)
	return 1
}

func oraA(e *Emulator) uint16 {
	e.orAccumulator(e.registers.A)
	return 1
}

func oraM(e *Emulator) uint16 {
	offset := e.getHL()
	e.orAccumulator(e.memory[offset])
	return 1
}

func ori(e *Emulator) uint16 {
	e.orAccumulator(e.memory[e.pc+1])
	return 2
}

func xraB(e *Emulator) uint16 {
	e.xorAccumulator(e.registers.B)
	return 1
}

func xraC(e *Emulator) uint16 {
	e.xorAccumulator(e.registers.C)
	return 1
}

func xraD(e *Emulator) uint16 {
	e.xorAccumulator(e.registers.D)
	return 1
}

func xraE(e *Emulator) uint16 {
	e.xorAccumulator(e.registers.E)
	return 1
}

func xraH(e *Emulator) uint16 {
	e.xorAccumulator(e.registers.H)
	return 1
}

func xraL(e *Emulator) uint16 {
	e.xorAccumulator(e.registers.L)
	return 1
}

func xraA(e *Emulator) uint16 {
	e.xorAccumulator(e.registers.A)
	return 1
}

func xraM(e *Emulator) uint16 {
	offset := e.getHL()
	e.xorAccumulator(e.memory[offset])
	return 1
}

func xri(e *Emulator) uint16 {
	e.xorAccumulator(e.memory[e.pc+1])
	return 2
}

func cmpB(e *Emulator) uint16 {
	e.compareAccumulator(e.registers.B)
	return 1
}

func cmpC(e *Emulator) uint16 {
	e.compareAccumulator(e.registers.C)
	return 1
}

func cmpD(e *Emulator) uint16 {
	e.compareAccumulator(e.registers.D)
	return 1
}

func cmpE(e *Emulator) uint16 {
	e.compareAccumulator(e.registers.E)
	return 1
}

func cmpH(e *Emulator) uint16 {
	e.compareAccumulator(e.registers.H)
	return 1
}

func cmpL(e *Emulator) uint16 {
	e.compareAccumulator(e.registers.L)
	return 1
}

func cmpA(e *Emulator) uint16 {
	e.compareAccumulator(e.registers.A)
	return 1
}

func cmpM(e *Emulator) uint16 {
	offset := e.getHL()
	e.compareAccumulator(e.memory[offset])
	return 1
}

func cpi(e *Emulator) uint16 {
	e.compareAccumulator(e.memory[e.pc+1])
	return 2
}

func in(e *Emulator) uint16 {
	return 2
}

func out(e *Emulator) uint16 {
	return 2
}

func ei(e *Emulator) uint16 {
	e.intEnable = 1
	return 1
}

func di(e *Emulator) uint16 {
	e.intEnable = 0
	return 1
}

func pushB(e *Emulator) uint16 {
	e.memory[e.sp-1] = e.registers.B
	e.memory[e.sp-2] = e.registers.C
	e.sp -= 2
	return 1
}

func pushD(e *Emulator) uint16 {
	e.memory[e.sp-1] = e.registers.D
	e.memory[e.sp-2] = e.registers.E
	e.sp -= 2
	return 1
}

func pushH(e *Emulator) uint16 {
	e.memory[e.sp-1] = e.registers.H
	e.memory[e.sp-2] = e.registers.L
	e.sp -= 2
	return 1
}

func pushPSW(e *Emulator) uint16 {
	e.memory[e.sp-1] = e.registers.A
	psw := (e.flags.Z | (e.flags.S << 1) | (e.flags.P << 2) | (e.flags.CY << 3) | (e.flags.AC << 4))
	e.memory[e.sp-2] = psw
	e.sp -= 2
	return 1
}

func popB(e *Emulator) uint16 {
	e.registers.C = e.memory[e.sp]
	e.registers.B = e.memory[e.sp+1]
	e.sp += 2
	return 1
}

func popD(e *Emulator) uint16 {
	e.registers.E = e.memory[e.sp]
	e.registers.D = e.memory[e.sp+1]
	e.sp += 2
	return 1
}

func popH(e *Emulator) uint16 {
	e.registers.L = e.memory[e.sp]
	e.registers.H = e.memory[e.sp+1]
	e.sp += 2
	return 1
}

func popPSW(e *Emulator) uint16 {
	e.registers.A = e.memory[e.sp+1]
	psw := e.memory[e.sp]
	if (psw & 0x01) == 0x01 {
		e.flags.Z = 1
	} else {
		e.flags.Z = 0
	}
	if (psw & 0x02) == 0x02 {
		e.flags.S = 1
	} else {
		e.flags.S = 0
	}
	if (psw & 0x04) == 0x04 {
		e.flags.P = 1
	} else {
		e.flags.P = 0
	}
	if (psw & 0x08) == 0x05 {
		e.flags.CY = 1
	} else {
		e.flags.CY = 0
	}
	if (psw & 0x10) == 0x10 {
		e.flags.AC = 1
	} else {
		e.flags.AC = 0
	}
	e.memory[e.sp-2] = psw
	e.sp += 2
	return 1
}

func lhld(e *Emulator) uint16 {
	addr := (uint16(e.memory[e.pc+2]) << 8) | uint16(e.memory[e.pc+1])
	e.registers.L = e.memory[addr]
	e.registers.H = e.memory[addr+1]
	return 3
}

func shld(e *Emulator) uint16 {
	addr := (uint16(e.memory[e.pc+2]) << 8) | uint16(e.memory[e.pc+1])
	e.memory[addr] = e.registers.L
	e.memory[addr+1] = e.registers.H
	return 3
}

func xchg(e *Emulator) uint16 {
	H := e.registers.H
	D := e.registers.D
	L := e.registers.L
	E := e.registers.E
	e.registers.D = H
	e.registers.H = D
	e.registers.L = E
	e.registers.E = L
	return 1
}

func xthl(e *Emulator) uint16 {
	H := e.registers.H
	L := e.registers.L
	sp1 := e.memory[e.sp]
	sp2 := e.memory[e.sp+1]
	e.memory[e.sp] = L
	e.memory[e.sp+1] = H
	e.registers.H = sp2
	e.registers.L = sp1
	return 1
}

func sphl(e *Emulator) uint16 {
	e.sp = e.getHL()
	return 1
}

func pchl(e *Emulator) uint16 {
	e.pc = e.getHL()
	return 0
}

func rlc(e *Emulator) uint16 {
	e.flags.CY = e.registers.A >> 7
	e.registers.A = (e.registers.A << 1) | e.flags.CY
	return 1
}

func rrc(e *Emulator) uint16 {
	e.flags.CY = e.registers.A & 1
	e.registers.A = (e.registers.A >> 1) | (e.flags.CY << 7)
	return 1
}

func ral(e *Emulator) uint16 {
	cy := e.flags.CY
	e.flags.CY = e.registers.A >> 7
	e.registers.A = (e.registers.A << 1) | cy
	return 1
}

func rar(e *Emulator) uint16 {
	cy := e.flags.CY
	e.flags.CY = e.registers.A & 1
	e.registers.A = (e.registers.A >> 1) | (cy << 7)
	return 1
}

func stc(e *Emulator) uint16 {
	e.flags.CY = 1
	return 1
}

func cmc(e *Emulator) uint16 {
	e.flags.CY ^= 1
	return 1
}

func cma(e *Emulator) uint16 {
	e.registers.A ^= 255
	return 1
}
