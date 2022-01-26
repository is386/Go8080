package i8080

import (
	"fmt"
	"io/ioutil"
)

type Emulator struct {
	mem       [64 * 1024]uint8
	reg       *Registers
	pc        uint16
	sp        uint16
	intEnable uint8
	flags     *Flags
	halt      bool
}

func NewEmulator(pcStart uint16) *Emulator {
	return &Emulator{reg: &Registers{}, flags: &Flags{}, pc: pcStart, halt: false}
}

func (e *Emulator) LoadRom(filename string, offset uint16) {
	rom, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(rom); i++ {
		e.mem[offset+uint16(i)] = rom[i]
	}

	// CPU DIAG
	e.mem[0x7] = 0xc9
}

func (e *Emulator) fetch() uint8 {
	return e.mem[e.pc]
}

func (e *Emulator) decode(opcode uint8) func(*Emulator) uint16 {
	return INSTRUCTIONS[opcode]
}

func (e *Emulator) Execute() bool {
	opcode := e.fetch()
	instr := e.decode(opcode)
	out := e.showDebug(opcode)
	steps := instr(e)
	e.pc += steps
	return !e.halt && out
}

func (e *Emulator) showDebug(opcode uint8) bool {
	// fmt.Printf("\nPC:%04X OP:%02X SP:%04X BC:%04X DE:%04X HL:%04X A:%02X Cy:%dAC:%d S:%d Z:%d P:%d",
	// 	e.pc, opcode, e.sp, e.getBC(), e.getDE(), e.getHL(), e.reg.A, e.flags.CY, e.flags.AC,
	// 	e.flags.S, e.flags.Z, e.flags.P)
	// PC: 0100, AF: 0002, BC: 0000, DE: 0000, HL: 0000, SP: 0000, CYC: 0	(3E 01 FE 02)
	// f := uint8(0)
	// f |= e.flags.S << 7
	// f |= e.flags.Z << 6
	// f |= e.flags.AC << 4
	// f |= e.flags.P << 2
	// f |= 1 << 1
	// f |= e.flags.CY << 0

	// fmt.Printf("\nPC: %04X, AF: %04X, BC: %04X, DE: %04X, HL: %04X, SP: %04X (%02X %02X %02X %02X)",
	// 	e.pc, uint16(e.reg.A)<<8|uint16(f), e.getBC(), e.getDE(), e.getHL(), e.sp, opcode,
	// 	e.mem[e.pc+1], e.mem[e.pc+2], e.mem[e.pc+3])
	if e.pc == 5 {
		if e.reg.C == 9 {
			fmt.Println()
			offset := e.getDE()
			str := e.mem[offset]
			for str != '$' {
				fmt.Printf("%c", str)
				offset += 1
				str = e.mem[offset]
			}
		} else if e.reg.C == 2 {
			fmt.Printf("%c", e.reg.E)
		}
	} else if e.pc == 0 {
		return false
	}

	return true
}

func (e *Emulator) getNextTwoBytes() uint16 {
	return (uint16(e.mem[e.pc+2]) << 8) | uint16(e.mem[e.pc+1])
}

func (e *Emulator) getBC() uint16 {
	return (uint16(e.reg.B) << 8) | uint16(e.reg.C)
}

func (e *Emulator) getDE() uint16 {
	return (uint16(e.reg.D) << 8) | uint16(e.reg.E)
}

func (e *Emulator) getHL() uint16 {
	return (uint16(e.reg.H) << 8) | uint16(e.reg.L)
}

func (e *Emulator) setBC(val uint16) {
	e.reg.B = uint8(val >> 8)
	e.reg.C = uint8(val & 0xff)
}

func (e *Emulator) setDE(val uint16) {
	e.reg.D = uint8(val >> 8)
	e.reg.E = uint8(val & 0xff)
}

func (e *Emulator) setHL(val uint16) {
	e.reg.H = uint8(val >> 8)
	e.reg.L = uint8(val & 0xff)
}

func (e *Emulator) setCarryArith(val uint16) {
	if val > 0xff {
		e.flags.CY = 1
	} else {
		e.flags.CY = 0
	}
}

func (e *Emulator) setCarryDad(val uint32) {
	if (val & 0xffff0000) > 0 {
		e.flags.CY = 1
	} else {
		e.flags.CY = 0
	}
}

func (e *Emulator) setAuxCarry(b uint16, res uint16) {
	if ((uint16(e.reg.A) ^ b ^ res) & 0x10) > 0 {
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

func (e *Emulator) setParity(val uint16) {
	e.flags.P = parity(val)
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

func (e *Emulator) setZSP(val uint8) {
	e.setZero(uint16(val))
	e.setSign(uint16(val))
	e.setParity(uint16(val))
}

func flip(val uint8) uint8 {
	if val == 1 {
		return 0
	} else {
		return 1
	}
}

func (e *Emulator) addToAccumulator(val uint8, cy uint8) {
	ans := uint16(e.reg.A) + uint16(val) + uint16(cy)
	e.setZSP(uint8(ans))
	e.setCarryArith(ans)
	e.setAuxCarry(uint16(val), ans)
	e.reg.A = uint8(ans)
}

func (e *Emulator) subFromAccumulator(val uint8, cy uint8) {
	cy = flip(cy)
	e.addToAccumulator(^val, cy)
	e.flags.CY = flip(e.flags.CY)
}

func (e *Emulator) andAccumulator(val uint8) {
	ans := uint16(e.reg.A & val)
	e.setZSP(uint8(ans))
	e.flags.CY = 0
	if ((e.reg.A | val) & 0x08) > 0 {
		e.flags.AC = 1
	} else {
		e.flags.AC = 0
	}
	e.reg.A = uint8(ans)
}

func (e *Emulator) xorAccumulator(val uint8) {
	ans := uint16(e.reg.A ^ val)
	e.setZSP(uint8(ans))
	e.flags.CY = 0
	e.flags.AC = 0
	e.reg.A = uint8(ans)
}

func (e *Emulator) orAccumulator(val uint8) {
	ans := uint16(e.reg.A | val)
	e.setZSP(uint8(ans))
	e.flags.CY = 0
	e.flags.AC = 0
	e.reg.A = uint8(ans)
}

func (e *Emulator) compareAccumulator(val uint8) {
	ans := uint16(e.reg.A) - uint16(val)
	e.setZSP(uint8(ans))
	e.setCarryArith(ans)
	if (^(uint16(e.reg.A) ^ ans ^ uint16(val)) & 0x10) > 0 {
		e.flags.AC = 1
	} else {
		e.flags.AC = 0
	}
}

func noOp(e *Emulator) uint16 {
	return 1
}

func addB(e *Emulator) uint16 {
	e.addToAccumulator(e.reg.B, 0)
	return 1
}

func addC(e *Emulator) uint16 {
	e.addToAccumulator(e.reg.C, 0)
	return 1
}

func addD(e *Emulator) uint16 {
	e.addToAccumulator(e.reg.D, 0)
	return 1
}

func addE(e *Emulator) uint16 {
	e.addToAccumulator(e.reg.E, 0)
	return 1
}

func addH(e *Emulator) uint16 {
	e.addToAccumulator(e.reg.H, 0)
	return 1
}

func addL(e *Emulator) uint16 {
	e.addToAccumulator(e.reg.L, 0)
	return 1
}

func addM(e *Emulator) uint16 {
	offset := e.getHL()
	e.addToAccumulator(e.mem[offset], 0)
	return 1
}

func addA(e *Emulator) uint16 {
	e.addToAccumulator(e.reg.A, 0)
	return 1
}

func adi(e *Emulator) uint16 {
	e.addToAccumulator(e.mem[e.pc+1], 0)
	return 2
}

func adcB(e *Emulator) uint16 {
	e.addToAccumulator(e.reg.B, e.flags.CY)
	return 1
}

func adcC(e *Emulator) uint16 {
	e.addToAccumulator(e.reg.C, e.flags.CY)
	return 1
}

func adcD(e *Emulator) uint16 {
	e.addToAccumulator(e.reg.D, e.flags.CY)
	return 1
}

func adcE(e *Emulator) uint16 {
	e.addToAccumulator(e.reg.E, e.flags.CY)
	return 1
}

func adcH(e *Emulator) uint16 {
	e.addToAccumulator(e.reg.H, e.flags.CY)
	return 1
}

func adcL(e *Emulator) uint16 {
	e.addToAccumulator(e.reg.L, e.flags.CY)
	return 1
}

func adcA(e *Emulator) uint16 {
	e.addToAccumulator(e.reg.A, e.flags.CY)
	return 1
}

func adcM(e *Emulator) uint16 {
	offset := e.getHL()
	e.addToAccumulator(e.mem[offset], e.flags.CY)
	return 1
}

func aci(e *Emulator) uint16 {
	e.addToAccumulator(e.mem[e.pc+1], e.flags.CY)
	return 2
}

func subB(e *Emulator) uint16 {
	e.subFromAccumulator(e.reg.B, 0)
	return 1
}

func subC(e *Emulator) uint16 {
	e.subFromAccumulator(e.reg.C, 0)
	return 1
}

func subD(e *Emulator) uint16 {
	e.subFromAccumulator(e.reg.D, 0)
	return 1
}

func subE(e *Emulator) uint16 {
	e.subFromAccumulator(e.reg.E, 0)
	return 1
}

func subH(e *Emulator) uint16 {
	e.subFromAccumulator(e.reg.H, 0)
	return 1
}

func subL(e *Emulator) uint16 {
	e.subFromAccumulator(e.reg.L, 0)
	return 1
}

func subA(e *Emulator) uint16 {
	e.subFromAccumulator(e.reg.A, 0)
	return 1
}

func subM(e *Emulator) uint16 {
	offset := e.getHL()
	e.subFromAccumulator(e.mem[offset], 0)
	return 1
}

func sui(e *Emulator) uint16 {
	e.subFromAccumulator(e.mem[e.pc+1], 0)
	return 2
}

func sbbB(e *Emulator) uint16 {
	e.subFromAccumulator(e.reg.B, e.flags.CY)
	return 1
}

func sbbC(e *Emulator) uint16 {
	e.subFromAccumulator(e.reg.C, e.flags.CY)
	return 1
}

func sbbD(e *Emulator) uint16 {
	e.subFromAccumulator(e.reg.D, e.flags.CY)
	return 1
}

func sbbE(e *Emulator) uint16 {
	e.subFromAccumulator(e.reg.E, e.flags.CY)
	return 1
}

func sbbH(e *Emulator) uint16 {
	e.subFromAccumulator(e.reg.H, e.flags.CY)
	return 1
}

func sbbL(e *Emulator) uint16 {
	e.subFromAccumulator(e.reg.L, e.flags.CY)
	return 1
}

func sbbA(e *Emulator) uint16 {
	e.subFromAccumulator(e.reg.A, e.flags.CY)
	return 1
}

func sbbM(e *Emulator) uint16 {
	offset := e.getHL()
	e.subFromAccumulator(e.mem[offset], e.flags.CY)
	return 1
}

func sbi(e *Emulator) uint16 {
	e.subFromAccumulator(e.mem[e.pc+1], e.flags.CY)
	return 2
}

func inrB(e *Emulator) uint16 {
	e.reg.B += 1
	e.setZSP(e.reg.B)
	e.setAuxCarryInr(e.reg.B)
	return 1
}

func inrC(e *Emulator) uint16 {
	e.reg.C += 1
	e.setZSP(e.reg.C)
	e.setAuxCarryInr(e.reg.C)
	return 1
}

func inrD(e *Emulator) uint16 {
	e.reg.D += 1
	e.setZSP(e.reg.D)
	e.setAuxCarryInr(e.reg.D)
	return 1
}

func inrE(e *Emulator) uint16 {
	e.reg.E += 1
	e.setZSP(e.reg.E)
	e.setAuxCarryInr(e.reg.E)
	return 1
}

func inrH(e *Emulator) uint16 {
	e.reg.H += 1
	e.setZSP(e.reg.H)
	e.setAuxCarryInr(e.reg.H)
	return 1
}

func inrL(e *Emulator) uint16 {
	e.reg.L += 1
	e.setZSP(e.reg.L)
	e.setAuxCarryInr(e.reg.L)
	return 1
}

func inrA(e *Emulator) uint16 {
	e.reg.A += 1
	e.setZSP(e.reg.A)
	e.setAuxCarryInr(e.reg.A)
	return 1
}

func inrM(e *Emulator) uint16 {
	offset := e.getHL()
	e.mem[offset] += 1
	e.setZSP(e.mem[offset])
	e.setAuxCarryInr(e.mem[offset])
	return 1
}

func dcrB(e *Emulator) uint16 {
	e.reg.B -= 1
	e.setZSP(e.reg.B)
	e.setAuxCarryDcr(e.reg.B)
	return 1
}

func dcrC(e *Emulator) uint16 {
	e.reg.C -= 1
	e.setZSP(e.reg.C)
	e.setAuxCarryDcr(e.reg.C)
	return 1
}

func dcrD(e *Emulator) uint16 {
	e.reg.D -= 1
	e.setZSP(e.reg.D)
	e.setAuxCarryDcr(e.reg.D)
	return 1
}

func dcrE(e *Emulator) uint16 {
	e.reg.E -= 1
	e.setZSP(e.reg.E)
	e.setAuxCarryDcr(e.reg.E)
	return 1
}

func dcrH(e *Emulator) uint16 {
	e.reg.H -= 1
	e.setZSP(e.reg.H)
	e.setAuxCarryDcr(e.reg.H)
	return 1
}

func dcrL(e *Emulator) uint16 {
	e.reg.L -= 1
	e.setZSP(e.reg.L)
	e.setAuxCarryDcr(e.reg.L)
	return 1
}

func dcrA(e *Emulator) uint16 {
	e.reg.A -= 1
	e.setZSP(e.reg.A)
	e.setAuxCarryDcr(e.reg.A)
	return 1
}

func dcrM(e *Emulator) uint16 {
	offset := e.getHL()
	e.mem[offset] -= 1
	e.setZSP(e.mem[offset])
	e.setAuxCarryDcr(e.mem[offset])
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
	e.setCarryDad(ans)
	return 1
}

func dadD(e *Emulator) uint16 {
	ans := uint32(e.getHL()) + uint32(e.getDE())
	e.setHL(uint16(ans))
	e.setCarryDad(ans)
	return 1
}

func dadH(e *Emulator) uint16 {
	ans := uint32(e.getHL()) + uint32(e.getHL())
	e.setHL(uint16(ans))
	e.setCarryDad(ans)
	return 1
}

func dadSP(e *Emulator) uint16 {
	ans := uint32(e.getHL()) + uint32(e.sp)
	e.setHL(uint16(ans))
	e.setCarryDad(ans)
	return 1
}

func daa(e *Emulator) uint16 {
	cy := e.flags.CY
	lsb := e.reg.A & 0x0f
	msb := e.reg.A >> 4
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
	e.pc = e.getNextTwoBytes()
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
	e.pc = (uint16(e.mem[e.sp]) | (uint16(e.mem[e.sp+1]) << 8))
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
	e.reg.B = e.mem[e.pc+1]
	return 2
}

func mviC(e *Emulator) uint16 {
	e.reg.C = e.mem[e.pc+1]
	return 2
}

func mviD(e *Emulator) uint16 {
	e.reg.D = e.mem[e.pc+1]
	return 2
}

func mviE(e *Emulator) uint16 {
	e.reg.E = e.mem[e.pc+1]
	return 2
}

func mviH(e *Emulator) uint16 {
	e.reg.H = e.mem[e.pc+1]
	return 2
}

func mviL(e *Emulator) uint16 {
	e.reg.L = e.mem[e.pc+1]
	return 2
}

func mviA(e *Emulator) uint16 {
	e.reg.A = e.mem[e.pc+1]
	return 2
}

func mviM(e *Emulator) uint16 {
	e.mem[e.getHL()] = e.mem[e.pc+1]
	return 2
}

func call(e *Emulator) uint16 {
	ret := e.pc + 2
	e.mem[e.sp-1] = uint8(ret>>8) & uint8(0xff)
	e.mem[e.sp-2] = uint8(ret) & uint8(0xff)
	e.sp = e.sp - 2
	e.pc = e.getNextTwoBytes()
	return 0
}

func callRst(e *Emulator, addr uint16) uint16 {
	ret := e.pc + 2
	e.mem[e.sp-1] = uint8(ret>>8) & uint8(0xff)
	e.mem[e.sp-2] = uint8(ret) & uint8(0xff)
	e.sp = e.sp - 2
	e.pc = addr
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
	e.reg.C = e.mem[e.pc+1]
	e.reg.B = e.mem[e.pc+2]
	return 3
}

func lxiD(e *Emulator) uint16 {
	e.reg.E = e.mem[e.pc+1]
	e.reg.D = e.mem[e.pc+2]
	return 3
}

func lxiH(e *Emulator) uint16 {
	e.reg.L = e.mem[e.pc+1]
	e.reg.H = e.mem[e.pc+2]
	return 3
}

func lxiSP(e *Emulator) uint16 {
	e.sp = e.getNextTwoBytes()
	return 3
}

func lda(e *Emulator) uint16 {
	e.reg.A = e.mem[e.getNextTwoBytes()]
	return 3
}

func ldaxB(e *Emulator) uint16 {
	e.reg.A = e.mem[e.getBC()]
	return 1
}

func ldaxD(e *Emulator) uint16 {
	e.reg.A = e.mem[e.getDE()]
	return 1
}

func movBB(e *Emulator) uint16 {
	return 1
}

func movBC(e *Emulator) uint16 {
	e.reg.B = e.reg.C
	return 1
}

func movBD(e *Emulator) uint16 {
	e.reg.B = e.reg.D
	return 1
}

func movBE(e *Emulator) uint16 {
	e.reg.B = e.reg.E
	return 1
}

func movBH(e *Emulator) uint16 {
	e.reg.B = e.reg.H
	return 1
}

func movBL(e *Emulator) uint16 {
	e.reg.B = e.reg.L
	return 1
}

func movBM(e *Emulator) uint16 {
	offset := e.getHL()
	e.reg.B = e.mem[offset]
	return 1
}

func movBA(e *Emulator) uint16 {
	e.reg.B = e.reg.A
	return 1
}

func movCB(e *Emulator) uint16 {
	e.reg.C = e.reg.B
	return 1
}

func movCC(e *Emulator) uint16 {
	return 1
}

func movCD(e *Emulator) uint16 {
	e.reg.C = e.reg.D
	return 1
}

func movCE(e *Emulator) uint16 {
	e.reg.C = e.reg.E
	return 1
}

func movCH(e *Emulator) uint16 {
	e.reg.C = e.reg.H
	return 1
}

func movCL(e *Emulator) uint16 {
	e.reg.C = e.reg.L
	return 1
}

func movCM(e *Emulator) uint16 {
	offset := e.getHL()
	e.reg.C = e.mem[offset]
	return 1
}

func movCA(e *Emulator) uint16 {
	e.reg.C = e.reg.A
	return 1
}

func movDB(e *Emulator) uint16 {
	e.reg.D = e.reg.B
	return 1
}

func movDC(e *Emulator) uint16 {
	e.reg.D = e.reg.C
	return 1
}

func movDD(e *Emulator) uint16 {
	return 1
}

func movDE(e *Emulator) uint16 {
	e.reg.D = e.reg.E
	return 1
}

func movDH(e *Emulator) uint16 {
	e.reg.D = e.reg.H
	return 1
}

func movDL(e *Emulator) uint16 {
	e.reg.D = e.reg.L
	return 1
}

func movDM(e *Emulator) uint16 {
	offset := e.getHL()
	e.reg.D = e.mem[offset]
	return 1
}

func movDA(e *Emulator) uint16 {
	e.reg.D = e.reg.A
	return 1
}

func movEB(e *Emulator) uint16 {
	e.reg.E = e.reg.B
	return 1
}

func movEC(e *Emulator) uint16 {
	e.reg.E = e.reg.C
	return 1
}

func movED(e *Emulator) uint16 {
	e.reg.E = e.reg.D
	return 1
}

func movEE(e *Emulator) uint16 {
	return 1
}

func movEH(e *Emulator) uint16 {
	e.reg.E = e.reg.H
	return 1
}

func movEL(e *Emulator) uint16 {
	e.reg.E = e.reg.L
	return 1
}

func movEM(e *Emulator) uint16 {
	offset := e.getHL()
	e.reg.E = e.mem[offset]
	return 1
}

func movEA(e *Emulator) uint16 {
	e.reg.E = e.reg.A
	return 1
}

func movHB(e *Emulator) uint16 {
	e.reg.H = e.reg.B
	return 1
}

func movHC(e *Emulator) uint16 {
	e.reg.H = e.reg.C
	return 1
}

func movHD(e *Emulator) uint16 {
	e.reg.H = e.reg.D
	return 1
}

func movHE(e *Emulator) uint16 {
	e.reg.H = e.reg.E
	return 1
}

func movHH(e *Emulator) uint16 {
	return 1
}

func movHL(e *Emulator) uint16 {
	e.reg.H = e.reg.L
	return 1
}

func movHM(e *Emulator) uint16 {
	offset := e.getHL()
	e.reg.H = e.mem[offset]
	return 1
}

func movHA(e *Emulator) uint16 {
	e.reg.H = e.reg.A
	return 1
}

func movLB(e *Emulator) uint16 {
	e.reg.L = e.reg.B
	return 1
}

func movLC(e *Emulator) uint16 {
	e.reg.L = e.reg.C
	return 1
}

func movLD(e *Emulator) uint16 {
	e.reg.L = e.reg.D
	return 1
}

func movLE(e *Emulator) uint16 {
	e.reg.L = e.reg.E
	return 1
}

func movLH(e *Emulator) uint16 {
	e.reg.L = e.reg.H
	return 1
}

func movLL(e *Emulator) uint16 {
	return 1
}

func movLM(e *Emulator) uint16 {
	offset := e.getHL()
	e.reg.L = e.mem[offset]
	return 1
}

func movLA(e *Emulator) uint16 {
	e.reg.L = e.reg.A
	return 1
}

func movAB(e *Emulator) uint16 {
	e.reg.A = e.reg.B
	return 1
}

func movAC(e *Emulator) uint16 {
	e.reg.A = e.reg.C
	return 1
}

func movAD(e *Emulator) uint16 {
	e.reg.A = e.reg.D
	return 1
}

func movAE(e *Emulator) uint16 {
	e.reg.A = e.reg.E
	return 1
}

func movAH(e *Emulator) uint16 {
	e.reg.A = e.reg.H
	return 1
}

func movAL(e *Emulator) uint16 {
	e.reg.A = e.reg.L
	return 1
}

func movAM(e *Emulator) uint16 {
	offset := e.getHL()
	e.reg.A = e.mem[offset]
	return 1
}

func movAA(e *Emulator) uint16 {
	return 1
}

func movMB(e *Emulator) uint16 {
	e.mem[e.getHL()] = e.reg.B
	return 1
}

func movMC(e *Emulator) uint16 {
	e.mem[e.getHL()] = e.reg.C
	return 1
}

func movMD(e *Emulator) uint16 {
	e.mem[e.getHL()] = e.reg.D
	return 1
}

func movME(e *Emulator) uint16 {
	e.mem[e.getHL()] = e.reg.E
	return 1
}

func movMH(e *Emulator) uint16 {
	e.mem[e.getHL()] = e.reg.H
	return 1
}

func movML(e *Emulator) uint16 {
	e.mem[e.getHL()] = e.reg.L
	return 1
}

func movMA(e *Emulator) uint16 {
	e.mem[e.getHL()] = e.reg.A
	return 1
}

func sta(e *Emulator) uint16 {
	e.mem[e.getNextTwoBytes()] = e.reg.A
	return 3
}

func staxB(e *Emulator) uint16 {
	e.mem[e.getBC()] = e.reg.A
	return 1
}

func staxD(e *Emulator) uint16 {
	e.mem[e.getDE()] = e.reg.A
	return 1
}

func anaB(e *Emulator) uint16 {
	e.andAccumulator(e.reg.B)
	return 1
}

func anaC(e *Emulator) uint16 {
	e.andAccumulator(e.reg.C)
	return 1
}

func anaD(e *Emulator) uint16 {
	e.andAccumulator(e.reg.D)
	return 1
}

func anaE(e *Emulator) uint16 {
	e.andAccumulator(e.reg.E)
	return 1
}

func anaH(e *Emulator) uint16 {
	e.andAccumulator(e.reg.H)
	return 1
}

func anaL(e *Emulator) uint16 {
	e.andAccumulator(e.reg.L)
	return 1
}

func anaA(e *Emulator) uint16 {
	e.andAccumulator(e.reg.A)
	return 1
}

func anaM(e *Emulator) uint16 {
	offset := e.getHL()
	e.andAccumulator(e.mem[offset])
	return 1
}

func ani(e *Emulator) uint16 {
	e.andAccumulator(e.mem[e.pc+1])
	return 2
}

func oraB(e *Emulator) uint16 {
	e.orAccumulator(e.reg.B)
	return 1
}

func oraC(e *Emulator) uint16 {
	e.orAccumulator(e.reg.C)
	return 1
}

func oraD(e *Emulator) uint16 {
	e.orAccumulator(e.reg.D)
	return 1
}

func oraE(e *Emulator) uint16 {
	e.orAccumulator(e.reg.E)
	return 1
}

func oraH(e *Emulator) uint16 {
	e.orAccumulator(e.reg.H)
	return 1
}

func oraL(e *Emulator) uint16 {
	e.orAccumulator(e.reg.L)
	return 1
}

func oraA(e *Emulator) uint16 {
	e.orAccumulator(e.reg.A)
	return 1
}

func oraM(e *Emulator) uint16 {
	offset := e.getHL()
	e.orAccumulator(e.mem[offset])
	return 1
}

func ori(e *Emulator) uint16 {
	e.orAccumulator(e.mem[e.pc+1])
	return 2
}

func xraB(e *Emulator) uint16 {
	e.xorAccumulator(e.reg.B)
	return 1
}

func xraC(e *Emulator) uint16 {
	e.xorAccumulator(e.reg.C)
	return 1
}

func xraD(e *Emulator) uint16 {
	e.xorAccumulator(e.reg.D)
	return 1
}

func xraE(e *Emulator) uint16 {
	e.xorAccumulator(e.reg.E)
	return 1
}

func xraH(e *Emulator) uint16 {
	e.xorAccumulator(e.reg.H)
	return 1
}

func xraL(e *Emulator) uint16 {
	e.xorAccumulator(e.reg.L)
	return 1
}

func xraA(e *Emulator) uint16 {
	e.xorAccumulator(e.reg.A)
	return 1
}

func xraM(e *Emulator) uint16 {
	offset := e.getHL()
	e.xorAccumulator(e.mem[offset])
	return 1
}

func xri(e *Emulator) uint16 {
	e.xorAccumulator(e.mem[e.pc+1])
	return 2
}

func cmpB(e *Emulator) uint16 {
	e.compareAccumulator(e.reg.B)
	return 1
}

func cmpC(e *Emulator) uint16 {
	e.compareAccumulator(e.reg.C)
	return 1
}

func cmpD(e *Emulator) uint16 {
	e.compareAccumulator(e.reg.D)
	return 1
}

func cmpE(e *Emulator) uint16 {
	e.compareAccumulator(e.reg.E)
	return 1
}

func cmpH(e *Emulator) uint16 {
	e.compareAccumulator(e.reg.H)
	return 1
}

func cmpL(e *Emulator) uint16 {
	e.compareAccumulator(e.reg.L)
	return 1
}

func cmpA(e *Emulator) uint16 {
	e.compareAccumulator(e.reg.A)
	return 1
}

func cmpM(e *Emulator) uint16 {
	offset := e.getHL()
	e.compareAccumulator(e.mem[offset])
	return 1
}

func cpi(e *Emulator) uint16 {
	e.compareAccumulator(e.mem[e.pc+1])
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
	e.mem[e.sp-1] = e.reg.B
	e.mem[e.sp-2] = e.reg.C
	e.sp -= 2
	return 1
}

func pushD(e *Emulator) uint16 {
	e.mem[e.sp-1] = e.reg.D
	e.mem[e.sp-2] = e.reg.E
	e.sp -= 2
	return 1
}

func pushH(e *Emulator) uint16 {
	e.mem[e.sp-1] = e.reg.H
	e.mem[e.sp-2] = e.reg.L
	e.sp -= 2
	return 1
}

func pushPSW(e *Emulator) uint16 {
	e.mem[e.sp-1] = e.reg.A
	psw := (e.flags.Z | (e.flags.S << 1) | (e.flags.P << 2) | (e.flags.CY << 3) | (e.flags.AC << 4))
	e.mem[e.sp-2] = psw
	e.sp -= 2
	return 1
}

func popB(e *Emulator) uint16 {
	e.reg.C = e.mem[e.sp]
	e.reg.B = e.mem[e.sp+1]
	e.sp += 2
	return 1
}

func popD(e *Emulator) uint16 {
	e.reg.E = e.mem[e.sp]
	e.reg.D = e.mem[e.sp+1]
	e.sp += 2
	return 1
}

func popH(e *Emulator) uint16 {
	e.reg.L = e.mem[e.sp]
	e.reg.H = e.mem[e.sp+1]
	e.sp += 2
	return 1
}

func popPSW(e *Emulator) uint16 {
	e.reg.A = e.mem[e.sp+1]
	psw := e.mem[e.sp]
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
	e.sp += 2
	return 1
}

func lhld(e *Emulator) uint16 {
	e.reg.L = e.mem[e.getNextTwoBytes()]
	e.reg.H = e.mem[e.getNextTwoBytes()+1]
	return 3
}

func shld(e *Emulator) uint16 {
	e.mem[e.getNextTwoBytes()] = e.reg.L
	e.mem[e.getNextTwoBytes()+1] = e.reg.H
	return 3
}

func xchg(e *Emulator) uint16 {
	H := e.reg.H
	D := e.reg.D
	L := e.reg.L
	E := e.reg.E
	e.reg.D = H
	e.reg.H = D
	e.reg.L = E
	e.reg.E = L
	return 1
}

func xthl(e *Emulator) uint16 {
	H := e.reg.H
	L := e.reg.L
	sp1 := e.mem[e.sp]
	sp2 := e.mem[e.sp+1]
	e.mem[e.sp] = L
	e.mem[e.sp+1] = H
	e.reg.H = sp2
	e.reg.L = sp1
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
	e.flags.CY = e.reg.A >> 7
	e.reg.A = (e.reg.A << 1) | e.flags.CY
	return 1
}

func rrc(e *Emulator) uint16 {
	e.flags.CY = e.reg.A & 1
	e.reg.A = (e.reg.A >> 1) | (e.flags.CY << 7)
	return 1
}

func ral(e *Emulator) uint16 {
	cy := e.flags.CY
	e.flags.CY = e.reg.A >> 7
	e.reg.A = (e.reg.A << 1) | cy
	return 1
}

func rar(e *Emulator) uint16 {
	cy := e.flags.CY
	e.flags.CY = e.reg.A & 1
	e.reg.A = (e.reg.A >> 1) | (cy << 7)
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
	e.reg.A ^= 255
	return 1
}

func hlt(e *Emulator) uint16 {
	e.halt = true
	return 1
}

func rst0(e *Emulator) uint16 {
	callRst(e, 0x00)
	return 0
}

func rst1(e *Emulator) uint16 {
	callRst(e, 0x08)
	return 0
}

func rst2(e *Emulator) uint16 {
	callRst(e, 0x10)
	return 0
}

func rst3(e *Emulator) uint16 {
	callRst(e, 0x18)
	return 0
}

func rst4(e *Emulator) uint16 {
	callRst(e, 0x20)
	return 0
}

func rst5(e *Emulator) uint16 {
	callRst(e, 0x28)
	return 0
}

func rst6(e *Emulator) uint16 {
	callRst(e, 0x30)
	return 0
}

func rst7(e *Emulator) uint16 {
	callRst(e, 0x38)
	return 0
}
