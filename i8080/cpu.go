package i8080

import (
	"fmt"
	"io/ioutil"
)

type CPU struct {
	mem       [64 * 1024]uint8
	reg       *Registers
	pc        uint16
	sp        uint16
	intEnable uint8
	flags     *Flags
	halt      bool
	showDebug bool
	isTest    bool
	portIn    func()
	portOut   func()
}

func NewCPU(pcStart uint16, showDebug bool, isTest bool, portIn func(), portOut func()) *CPU {
	return &CPU{
		reg: &Registers{}, flags: &Flags{}, pc: pcStart, showDebug: showDebug,
		isTest: isTest, portIn: portIn, portOut: portOut}
}

func (c *CPU) LoadRom(filename string, offset uint16) {
	rom, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(rom); i++ {
		c.mem[offset+uint16(i)] = rom[i]
	}
	if c.isTest {
		c.mem[0x7] = 0xc9
	}
}

func (c *CPU) fetch() uint8 {
	return c.mem[c.pc]
}

func (c *CPU) decode(opcode uint8) func(*CPU) uint16 {
	return INSTRUCTIONS[opcode]
}

func (c *CPU) Execute() bool {
	opcode := c.fetch()
	instr := c.decode(opcode)
	out := true
	if c.showDebug {
		c.debugOutput(opcode)
	}
	if c.isTest {
		out = c.testOutput()
	}
	steps := instr(c)
	c.pc += steps
	return !c.halt && out
}

func (c *CPU) debugOutput(opcode uint8) {
	f := uint8(0)
	f |= c.flags.S << 7
	f |= c.flags.Z << 6
	f |= c.flags.AC << 4
	f |= c.flags.P << 2
	f |= 1 << 1
	f |= c.flags.CY << 0

	fmt.Printf("\nPC: %04X, AF: %04X, BC: %04X, DE: %04X, HL: %04X, SP: %04X (%02X %02X %02X %02X)",
		c.pc, uint16(c.reg.A)<<8|uint16(f), c.getBC(), c.getDE(), c.getHL(), c.sp, opcode,
		c.mem[c.pc+1], c.mem[c.pc+2], c.mem[c.pc+3])
}

func (c *CPU) testOutput() bool {
	if c.pc == 5 {
		if c.reg.C == 9 {
			fmt.Println()
			offset := c.getDE()
			str := c.mem[offset]
			for str != '$' {
				fmt.Printf("%c", str)
				offset += 1
				str = c.mem[offset]
			}
		} else if c.reg.C == 2 {
			fmt.Printf("%c", c.reg.E)
		}
	} else if c.pc == 0 {
		return false
	}
	return true
}

func (c *CPU) getNextTwoBytes() uint16 {
	return (uint16(c.mem[c.pc+2]) << 8) | uint16(c.mem[c.pc+1])
}

func (c *CPU) getBC() uint16 {
	return (uint16(c.reg.B) << 8) | uint16(c.reg.C)
}

func (c *CPU) getDE() uint16 {
	return (uint16(c.reg.D) << 8) | uint16(c.reg.E)
}

func (c *CPU) getHL() uint16 {
	return (uint16(c.reg.H) << 8) | uint16(c.reg.L)
}

func (c *CPU) setBC(val uint16) {
	c.reg.B = uint8(val >> 8)
	c.reg.C = uint8(val & 0xff)
}

func (c *CPU) setDE(val uint16) {
	c.reg.D = uint8(val >> 8)
	c.reg.E = uint8(val & 0xff)
}

func (c *CPU) setHL(val uint16) {
	c.reg.H = uint8(val >> 8)
	c.reg.L = uint8(val & 0xff)
}

func (c *CPU) setCarryArith(val uint16) {
	if val > 0xff {
		c.flags.CY = 1
	} else {
		c.flags.CY = 0
	}
}

func (c *CPU) setCarryDad(val uint32) {
	if (val & 0xffff0000) > 0 {
		c.flags.CY = 1
	} else {
		c.flags.CY = 0
	}
}

func (c *CPU) setAuxCarry(b uint16, res uint16) {
	if ((uint16(c.reg.A) ^ b ^ res) & 0x10) > 0 {
		c.flags.AC = 1
	} else {
		c.flags.AC = 0
	}
}

func (c *CPU) setAuxCarryInr(val uint8) {
	if (val & 0xf) == 0 {
		c.flags.AC = 1
	} else {
		c.flags.AC = 0
	}
}

func (c *CPU) setAuxCarryDcr(val uint8) {
	if (val & 0xf) == 0xf {
		c.flags.AC = 0
	} else {
		c.flags.AC = 1
	}
}

func (c *CPU) setZero(val uint16) {
	if (val & 0xff) == 0 {
		c.flags.Z = 1
	} else {
		c.flags.Z = 0
	}
}

func (c *CPU) setSign(val uint16) {
	if (val & 0x80) != 0 {
		c.flags.S = 1
	} else {
		c.flags.S = 0
	}
}

func (c *CPU) setParity(val uint16) {
	c.flags.P = parity(val)
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

func (c *CPU) setZSP(val uint8) {
	c.setZero(uint16(val))
	c.setSign(uint16(val))
	c.setParity(uint16(val))
}

func flip(val uint8) uint8 {
	if val == 1 {
		return 0
	} else {
		return 1
	}
}

func (c *CPU) addToAccumulator(val uint8, cy uint8) {
	ans := uint16(c.reg.A) + uint16(val) + uint16(cy)
	c.setZSP(uint8(ans))
	c.setCarryArith(ans)
	c.setAuxCarry(uint16(val), ans)
	c.reg.A = uint8(ans)
}

func (c *CPU) subFromAccumulator(val uint8, cy uint8) {
	cy = flip(cy)
	c.addToAccumulator(^val, cy)
	c.flags.CY = flip(c.flags.CY)
}

func (c *CPU) andAccumulator(val uint8) {
	ans := uint16(c.reg.A & val)
	c.setZSP(uint8(ans))
	c.flags.CY = 0
	if ((c.reg.A | val) & 0x08) > 0 {
		c.flags.AC = 1
	} else {
		c.flags.AC = 0
	}
	c.reg.A = uint8(ans)
}

func (c *CPU) xorAccumulator(val uint8) {
	ans := uint16(c.reg.A ^ val)
	c.setZSP(uint8(ans))
	c.flags.CY = 0
	c.flags.AC = 0
	c.reg.A = uint8(ans)
}

func (c *CPU) orAccumulator(val uint8) {
	ans := uint16(c.reg.A | val)
	c.setZSP(uint8(ans))
	c.flags.CY = 0
	c.flags.AC = 0
	c.reg.A = uint8(ans)
}

func (c *CPU) compareAccumulator(val uint8) {
	ans := uint16(c.reg.A) - uint16(val)
	c.setZSP(uint8(ans))
	c.setCarryArith(ans)
	if (^(uint16(c.reg.A) ^ ans ^ uint16(val)) & 0x10) > 0 {
		c.flags.AC = 1
	} else {
		c.flags.AC = 0
	}
}

func noOp(c *CPU) uint16 {
	return 1
}

func addB(c *CPU) uint16 {
	c.addToAccumulator(c.reg.B, 0)
	return 1
}

func addC(c *CPU) uint16 {
	c.addToAccumulator(c.reg.C, 0)
	return 1
}

func addD(c *CPU) uint16 {
	c.addToAccumulator(c.reg.D, 0)
	return 1
}

func addE(c *CPU) uint16 {
	c.addToAccumulator(c.reg.E, 0)
	return 1
}

func addH(c *CPU) uint16 {
	c.addToAccumulator(c.reg.H, 0)
	return 1
}

func addL(c *CPU) uint16 {
	c.addToAccumulator(c.reg.L, 0)
	return 1
}

func addM(c *CPU) uint16 {
	offset := c.getHL()
	c.addToAccumulator(c.mem[offset], 0)
	return 1
}

func addA(c *CPU) uint16 {
	c.addToAccumulator(c.reg.A, 0)
	return 1
}

func adi(c *CPU) uint16 {
	c.addToAccumulator(c.mem[c.pc+1], 0)
	return 2
}

func adcB(c *CPU) uint16 {
	c.addToAccumulator(c.reg.B, c.flags.CY)
	return 1
}

func adcC(c *CPU) uint16 {
	c.addToAccumulator(c.reg.C, c.flags.CY)
	return 1
}

func adcD(c *CPU) uint16 {
	c.addToAccumulator(c.reg.D, c.flags.CY)
	return 1
}

func adcE(c *CPU) uint16 {
	c.addToAccumulator(c.reg.E, c.flags.CY)
	return 1
}

func adcH(c *CPU) uint16 {
	c.addToAccumulator(c.reg.H, c.flags.CY)
	return 1
}

func adcL(c *CPU) uint16 {
	c.addToAccumulator(c.reg.L, c.flags.CY)
	return 1
}

func adcA(c *CPU) uint16 {
	c.addToAccumulator(c.reg.A, c.flags.CY)
	return 1
}

func adcM(c *CPU) uint16 {
	offset := c.getHL()
	c.addToAccumulator(c.mem[offset], c.flags.CY)
	return 1
}

func aci(c *CPU) uint16 {
	c.addToAccumulator(c.mem[c.pc+1], c.flags.CY)
	return 2
}

func subB(c *CPU) uint16 {
	c.subFromAccumulator(c.reg.B, 0)
	return 1
}

func subC(c *CPU) uint16 {
	c.subFromAccumulator(c.reg.C, 0)
	return 1
}

func subD(c *CPU) uint16 {
	c.subFromAccumulator(c.reg.D, 0)
	return 1
}

func subE(c *CPU) uint16 {
	c.subFromAccumulator(c.reg.E, 0)
	return 1
}

func subH(c *CPU) uint16 {
	c.subFromAccumulator(c.reg.H, 0)
	return 1
}

func subL(c *CPU) uint16 {
	c.subFromAccumulator(c.reg.L, 0)
	return 1
}

func subA(c *CPU) uint16 {
	c.subFromAccumulator(c.reg.A, 0)
	return 1
}

func subM(c *CPU) uint16 {
	offset := c.getHL()
	c.subFromAccumulator(c.mem[offset], 0)
	return 1
}

func sui(c *CPU) uint16 {
	c.subFromAccumulator(c.mem[c.pc+1], 0)
	return 2
}

func sbbB(c *CPU) uint16 {
	c.subFromAccumulator(c.reg.B, c.flags.CY)
	return 1
}

func sbbC(c *CPU) uint16 {
	c.subFromAccumulator(c.reg.C, c.flags.CY)
	return 1
}

func sbbD(c *CPU) uint16 {
	c.subFromAccumulator(c.reg.D, c.flags.CY)
	return 1
}

func sbbE(c *CPU) uint16 {
	c.subFromAccumulator(c.reg.E, c.flags.CY)
	return 1
}

func sbbH(c *CPU) uint16 {
	c.subFromAccumulator(c.reg.H, c.flags.CY)
	return 1
}

func sbbL(c *CPU) uint16 {
	c.subFromAccumulator(c.reg.L, c.flags.CY)
	return 1
}

func sbbA(c *CPU) uint16 {
	c.subFromAccumulator(c.reg.A, c.flags.CY)
	return 1
}

func sbbM(c *CPU) uint16 {
	offset := c.getHL()
	c.subFromAccumulator(c.mem[offset], c.flags.CY)
	return 1
}

func sbi(c *CPU) uint16 {
	c.subFromAccumulator(c.mem[c.pc+1], c.flags.CY)
	return 2
}

func inrB(c *CPU) uint16 {
	c.reg.B += 1
	c.setZSP(c.reg.B)
	c.setAuxCarryInr(c.reg.B)
	return 1
}

func inrC(c *CPU) uint16 {
	c.reg.C += 1
	c.setZSP(c.reg.C)
	c.setAuxCarryInr(c.reg.C)
	return 1
}

func inrD(c *CPU) uint16 {
	c.reg.D += 1
	c.setZSP(c.reg.D)
	c.setAuxCarryInr(c.reg.D)
	return 1
}

func inrE(c *CPU) uint16 {
	c.reg.E += 1
	c.setZSP(c.reg.E)
	c.setAuxCarryInr(c.reg.E)
	return 1
}

func inrH(c *CPU) uint16 {
	c.reg.H += 1
	c.setZSP(c.reg.H)
	c.setAuxCarryInr(c.reg.H)
	return 1
}

func inrL(c *CPU) uint16 {
	c.reg.L += 1
	c.setZSP(c.reg.L)
	c.setAuxCarryInr(c.reg.L)
	return 1
}

func inrA(c *CPU) uint16 {
	c.reg.A += 1
	c.setZSP(c.reg.A)
	c.setAuxCarryInr(c.reg.A)
	return 1
}

func inrM(c *CPU) uint16 {
	offset := c.getHL()
	c.mem[offset] += 1
	c.setZSP(c.mem[offset])
	c.setAuxCarryInr(c.mem[offset])
	return 1
}

func dcrB(c *CPU) uint16 {
	c.reg.B -= 1
	c.setZSP(c.reg.B)
	c.setAuxCarryDcr(c.reg.B)
	return 1
}

func dcrC(c *CPU) uint16 {
	c.reg.C -= 1
	c.setZSP(c.reg.C)
	c.setAuxCarryDcr(c.reg.C)
	return 1
}

func dcrD(c *CPU) uint16 {
	c.reg.D -= 1
	c.setZSP(c.reg.D)
	c.setAuxCarryDcr(c.reg.D)
	return 1
}

func dcrE(c *CPU) uint16 {
	c.reg.E -= 1
	c.setZSP(c.reg.E)
	c.setAuxCarryDcr(c.reg.E)
	return 1
}

func dcrH(c *CPU) uint16 {
	c.reg.H -= 1
	c.setZSP(c.reg.H)
	c.setAuxCarryDcr(c.reg.H)
	return 1
}

func dcrL(c *CPU) uint16 {
	c.reg.L -= 1
	c.setZSP(c.reg.L)
	c.setAuxCarryDcr(c.reg.L)
	return 1
}

func dcrA(c *CPU) uint16 {
	c.reg.A -= 1
	c.setZSP(c.reg.A)
	c.setAuxCarryDcr(c.reg.A)
	return 1
}

func dcrM(c *CPU) uint16 {
	offset := c.getHL()
	c.mem[offset] -= 1
	c.setZSP(c.mem[offset])
	c.setAuxCarryDcr(c.mem[offset])
	return 1
}

func inxB(c *CPU) uint16 {
	c.setBC(c.getBC() + 1)
	return 1
}

func inxD(c *CPU) uint16 {
	c.setDE(c.getDE() + 1)
	return 1
}

func inxH(c *CPU) uint16 {
	c.setHL(c.getHL() + 1)
	return 1
}

func inxSP(c *CPU) uint16 {
	c.sp += 1
	return 1
}

func dcxB(c *CPU) uint16 {
	c.setBC(c.getBC() - 1)
	return 1
}

func dcxD(c *CPU) uint16 {
	c.setDE(c.getDE() - 1)
	return 1
}

func dcxH(c *CPU) uint16 {
	c.setHL(c.getHL() - 1)
	return 1
}

func dcxSP(c *CPU) uint16 {
	c.sp -= 1
	return 1
}

func dadB(c *CPU) uint16 {
	ans := uint32(c.getHL()) + uint32(c.getBC())
	c.setHL(uint16(ans))
	c.setCarryDad(ans)
	return 1
}

func dadD(c *CPU) uint16 {
	ans := uint32(c.getHL()) + uint32(c.getDE())
	c.setHL(uint16(ans))
	c.setCarryDad(ans)
	return 1
}

func dadH(c *CPU) uint16 {
	ans := uint32(c.getHL()) + uint32(c.getHL())
	c.setHL(uint16(ans))
	c.setCarryDad(ans)
	return 1
}

func dadSP(c *CPU) uint16 {
	ans := uint32(c.getHL()) + uint32(c.sp)
	c.setHL(uint16(ans))
	c.setCarryDad(ans)
	return 1
}

func daa(c *CPU) uint16 {
	cy := c.flags.CY
	lsb := c.reg.A & 0x0f
	msb := c.reg.A >> 4
	correction := 0

	if lsb > 9 || c.flags.AC == 1 {
		correction += 0x06
	}

	if (c.flags.CY == 1 || msb > 9) || (msb >= 9 && lsb > 9) {
		correction += 0x60
		cy = 1
	}

	c.addToAccumulator(uint8(correction), 0)
	c.flags.CY = cy
	return 1
}

func jmp(c *CPU) uint16 {
	c.pc = c.getNextTwoBytes()
	return 0
}

func jnz(c *CPU) uint16 {
	if c.flags.Z == 0 {
		return jmp(c)
	}
	return 3
}

func jz(c *CPU) uint16 {
	if c.flags.Z == 1 {
		return jmp(c)
	}
	return 3
}

func jnc(c *CPU) uint16 {
	if c.flags.CY == 0 {
		return jmp(c)
	}
	return 3
}

func jc(c *CPU) uint16 {
	if c.flags.CY == 1 {
		return jmp(c)
	}
	return 3
}

func jpo(c *CPU) uint16 {
	if c.flags.P == 0 {
		return jmp(c)
	}
	return 3
}

func jpe(c *CPU) uint16 {
	if c.flags.P == 1 {
		return jmp(c)
	}
	return 3
}

func jp(c *CPU) uint16 {
	if c.flags.S == 0 {
		return jmp(c)
	}
	return 3
}

func jm(c *CPU) uint16 {
	if c.flags.S == 1 {
		return jmp(c)
	}
	return 3
}

func ret(c *CPU) uint16 {
	c.pc = (uint16(c.mem[c.sp]) | (uint16(c.mem[c.sp+1]) << 8))
	c.sp += 2
	return 1
}

func rz(c *CPU) uint16 {
	if c.flags.Z == 1 {
		return ret(c)
	}
	return 1
}

func rnz(c *CPU) uint16 {
	if c.flags.Z == 0 {
		return ret(c)
	}
	return 1
}

func rnc(c *CPU) uint16 {
	if c.flags.CY == 0 {
		return ret(c)
	}
	return 1
}

func rc(c *CPU) uint16 {
	if c.flags.CY == 1 {
		return ret(c)
	}
	return 1
}

func rpo(c *CPU) uint16 {
	if c.flags.P == 0 {
		return ret(c)
	}
	return 1
}

func rpe(c *CPU) uint16 {
	if c.flags.P == 1 {
		return ret(c)
	}
	return 1
}

func rp(c *CPU) uint16 {
	if c.flags.S == 0 {
		return ret(c)
	}
	return 1
}

func rm(c *CPU) uint16 {
	if c.flags.S == 1 {
		return ret(c)
	}
	return 1
}

func mviB(c *CPU) uint16 {
	c.reg.B = c.mem[c.pc+1]
	return 2
}

func mviC(c *CPU) uint16 {
	c.reg.C = c.mem[c.pc+1]
	return 2
}

func mviD(c *CPU) uint16 {
	c.reg.D = c.mem[c.pc+1]
	return 2
}

func mviE(c *CPU) uint16 {
	c.reg.E = c.mem[c.pc+1]
	return 2
}

func mviH(c *CPU) uint16 {
	c.reg.H = c.mem[c.pc+1]
	return 2
}

func mviL(c *CPU) uint16 {
	c.reg.L = c.mem[c.pc+1]
	return 2
}

func mviA(c *CPU) uint16 {
	c.reg.A = c.mem[c.pc+1]
	return 2
}

func mviM(c *CPU) uint16 {
	c.mem[c.getHL()] = c.mem[c.pc+1]
	return 2
}

func call(c *CPU) uint16 {
	ret := c.pc + 2
	c.mem[c.sp-1] = uint8(ret>>8) & uint8(0xff)
	c.mem[c.sp-2] = uint8(ret) & uint8(0xff)
	c.sp = c.sp - 2
	c.pc = c.getNextTwoBytes()
	return 0
}

func callRst(c *CPU, addr uint16) uint16 {
	ret := c.pc + 2
	c.mem[c.sp-1] = uint8(ret>>8) & uint8(0xff)
	c.mem[c.sp-2] = uint8(ret) & uint8(0xff)
	c.sp = c.sp - 2
	c.pc = addr
	return 0
}

func cz(c *CPU) uint16 {
	if c.flags.Z == 1 {
		return call(c)
	}
	return 3
}

func cnz(c *CPU) uint16 {
	if c.flags.Z == 0 {
		return call(c)
	}
	return 3
}

func cnc(c *CPU) uint16 {
	if c.flags.CY == 0 {
		return call(c)
	}
	return 3
}

func cc(c *CPU) uint16 {
	if c.flags.CY == 1 {
		return call(c)
	}
	return 3
}

func cpo(c *CPU) uint16 {
	if c.flags.P == 0 {
		return call(c)
	}
	return 3
}

func cpe(c *CPU) uint16 {
	if c.flags.P == 1 {
		return call(c)
	}
	return 3
}

func cp(c *CPU) uint16 {
	if c.flags.S == 0 {
		return call(c)
	}
	return 3
}

func cm(c *CPU) uint16 {
	if c.flags.S == 1 {
		return call(c)
	}
	return 3
}

func lxiB(c *CPU) uint16 {
	c.reg.C = c.mem[c.pc+1]
	c.reg.B = c.mem[c.pc+2]
	return 3
}

func lxiD(c *CPU) uint16 {
	c.reg.E = c.mem[c.pc+1]
	c.reg.D = c.mem[c.pc+2]
	return 3
}

func lxiH(c *CPU) uint16 {
	c.reg.L = c.mem[c.pc+1]
	c.reg.H = c.mem[c.pc+2]
	return 3
}

func lxiSP(c *CPU) uint16 {
	c.sp = c.getNextTwoBytes()
	return 3
}

func lda(c *CPU) uint16 {
	c.reg.A = c.mem[c.getNextTwoBytes()]
	return 3
}

func ldaxB(c *CPU) uint16 {
	c.reg.A = c.mem[c.getBC()]
	return 1
}

func ldaxD(c *CPU) uint16 {
	c.reg.A = c.mem[c.getDE()]
	return 1
}

func movBB(c *CPU) uint16 {
	return 1
}

func movBC(c *CPU) uint16 {
	c.reg.B = c.reg.C
	return 1
}

func movBD(c *CPU) uint16 {
	c.reg.B = c.reg.D
	return 1
}

func movBE(c *CPU) uint16 {
	c.reg.B = c.reg.E
	return 1
}

func movBH(c *CPU) uint16 {
	c.reg.B = c.reg.H
	return 1
}

func movBL(c *CPU) uint16 {
	c.reg.B = c.reg.L
	return 1
}

func movBM(c *CPU) uint16 {
	offset := c.getHL()
	c.reg.B = c.mem[offset]
	return 1
}

func movBA(c *CPU) uint16 {
	c.reg.B = c.reg.A
	return 1
}

func movCB(c *CPU) uint16 {
	c.reg.C = c.reg.B
	return 1
}

func movCC(c *CPU) uint16 {
	return 1
}

func movCD(c *CPU) uint16 {
	c.reg.C = c.reg.D
	return 1
}

func movCE(c *CPU) uint16 {
	c.reg.C = c.reg.E
	return 1
}

func movCH(c *CPU) uint16 {
	c.reg.C = c.reg.H
	return 1
}

func movCL(c *CPU) uint16 {
	c.reg.C = c.reg.L
	return 1
}

func movCM(c *CPU) uint16 {
	offset := c.getHL()
	c.reg.C = c.mem[offset]
	return 1
}

func movCA(c *CPU) uint16 {
	c.reg.C = c.reg.A
	return 1
}

func movDB(c *CPU) uint16 {
	c.reg.D = c.reg.B
	return 1
}

func movDC(c *CPU) uint16 {
	c.reg.D = c.reg.C
	return 1
}

func movDD(c *CPU) uint16 {
	return 1
}

func movDE(c *CPU) uint16 {
	c.reg.D = c.reg.E
	return 1
}

func movDH(c *CPU) uint16 {
	c.reg.D = c.reg.H
	return 1
}

func movDL(c *CPU) uint16 {
	c.reg.D = c.reg.L
	return 1
}

func movDM(c *CPU) uint16 {
	offset := c.getHL()
	c.reg.D = c.mem[offset]
	return 1
}

func movDA(c *CPU) uint16 {
	c.reg.D = c.reg.A
	return 1
}

func movEB(c *CPU) uint16 {
	c.reg.E = c.reg.B
	return 1
}

func movEC(c *CPU) uint16 {
	c.reg.E = c.reg.C
	return 1
}

func movED(c *CPU) uint16 {
	c.reg.E = c.reg.D
	return 1
}

func movEE(c *CPU) uint16 {
	return 1
}

func movEH(c *CPU) uint16 {
	c.reg.E = c.reg.H
	return 1
}

func movEL(c *CPU) uint16 {
	c.reg.E = c.reg.L
	return 1
}

func movEM(c *CPU) uint16 {
	offset := c.getHL()
	c.reg.E = c.mem[offset]
	return 1
}

func movEA(c *CPU) uint16 {
	c.reg.E = c.reg.A
	return 1
}

func movHB(c *CPU) uint16 {
	c.reg.H = c.reg.B
	return 1
}

func movHC(c *CPU) uint16 {
	c.reg.H = c.reg.C
	return 1
}

func movHD(c *CPU) uint16 {
	c.reg.H = c.reg.D
	return 1
}

func movHE(c *CPU) uint16 {
	c.reg.H = c.reg.E
	return 1
}

func movHH(c *CPU) uint16 {
	return 1
}

func movHL(c *CPU) uint16 {
	c.reg.H = c.reg.L
	return 1
}

func movHM(c *CPU) uint16 {
	offset := c.getHL()
	c.reg.H = c.mem[offset]
	return 1
}

func movHA(c *CPU) uint16 {
	c.reg.H = c.reg.A
	return 1
}

func movLB(c *CPU) uint16 {
	c.reg.L = c.reg.B
	return 1
}

func movLC(c *CPU) uint16 {
	c.reg.L = c.reg.C
	return 1
}

func movLD(c *CPU) uint16 {
	c.reg.L = c.reg.D
	return 1
}

func movLE(c *CPU) uint16 {
	c.reg.L = c.reg.E
	return 1
}

func movLH(c *CPU) uint16 {
	c.reg.L = c.reg.H
	return 1
}

func movLL(c *CPU) uint16 {
	return 1
}

func movLM(c *CPU) uint16 {
	offset := c.getHL()
	c.reg.L = c.mem[offset]
	return 1
}

func movLA(c *CPU) uint16 {
	c.reg.L = c.reg.A
	return 1
}

func movAB(c *CPU) uint16 {
	c.reg.A = c.reg.B
	return 1
}

func movAC(c *CPU) uint16 {
	c.reg.A = c.reg.C
	return 1
}

func movAD(c *CPU) uint16 {
	c.reg.A = c.reg.D
	return 1
}

func movAE(c *CPU) uint16 {
	c.reg.A = c.reg.E
	return 1
}

func movAH(c *CPU) uint16 {
	c.reg.A = c.reg.H
	return 1
}

func movAL(c *CPU) uint16 {
	c.reg.A = c.reg.L
	return 1
}

func movAM(c *CPU) uint16 {
	offset := c.getHL()
	c.reg.A = c.mem[offset]
	return 1
}

func movAA(c *CPU) uint16 {
	return 1
}

func movMB(c *CPU) uint16 {
	c.mem[c.getHL()] = c.reg.B
	return 1
}

func movMC(c *CPU) uint16 {
	c.mem[c.getHL()] = c.reg.C
	return 1
}

func movMD(c *CPU) uint16 {
	c.mem[c.getHL()] = c.reg.D
	return 1
}

func movME(c *CPU) uint16 {
	c.mem[c.getHL()] = c.reg.E
	return 1
}

func movMH(c *CPU) uint16 {
	c.mem[c.getHL()] = c.reg.H
	return 1
}

func movML(c *CPU) uint16 {
	c.mem[c.getHL()] = c.reg.L
	return 1
}

func movMA(c *CPU) uint16 {
	c.mem[c.getHL()] = c.reg.A
	return 1
}

func sta(c *CPU) uint16 {
	c.mem[c.getNextTwoBytes()] = c.reg.A
	return 3
}

func staxB(c *CPU) uint16 {
	c.mem[c.getBC()] = c.reg.A
	return 1
}

func staxD(c *CPU) uint16 {
	c.mem[c.getDE()] = c.reg.A
	return 1
}

func anaB(c *CPU) uint16 {
	c.andAccumulator(c.reg.B)
	return 1
}

func anaC(c *CPU) uint16 {
	c.andAccumulator(c.reg.C)
	return 1
}

func anaD(c *CPU) uint16 {
	c.andAccumulator(c.reg.D)
	return 1
}

func anaE(c *CPU) uint16 {
	c.andAccumulator(c.reg.E)
	return 1
}

func anaH(c *CPU) uint16 {
	c.andAccumulator(c.reg.H)
	return 1
}

func anaL(c *CPU) uint16 {
	c.andAccumulator(c.reg.L)
	return 1
}

func anaA(c *CPU) uint16 {
	c.andAccumulator(c.reg.A)
	return 1
}

func anaM(c *CPU) uint16 {
	offset := c.getHL()
	c.andAccumulator(c.mem[offset])
	return 1
}

func ani(c *CPU) uint16 {
	c.andAccumulator(c.mem[c.pc+1])
	return 2
}

func oraB(c *CPU) uint16 {
	c.orAccumulator(c.reg.B)
	return 1
}

func oraC(c *CPU) uint16 {
	c.orAccumulator(c.reg.C)
	return 1
}

func oraD(c *CPU) uint16 {
	c.orAccumulator(c.reg.D)
	return 1
}

func oraE(c *CPU) uint16 {
	c.orAccumulator(c.reg.E)
	return 1
}

func oraH(c *CPU) uint16 {
	c.orAccumulator(c.reg.H)
	return 1
}

func oraL(c *CPU) uint16 {
	c.orAccumulator(c.reg.L)
	return 1
}

func oraA(c *CPU) uint16 {
	c.orAccumulator(c.reg.A)
	return 1
}

func oraM(c *CPU) uint16 {
	offset := c.getHL()
	c.orAccumulator(c.mem[offset])
	return 1
}

func ori(c *CPU) uint16 {
	c.orAccumulator(c.mem[c.pc+1])
	return 2
}

func xraB(c *CPU) uint16 {
	c.xorAccumulator(c.reg.B)
	return 1
}

func xraC(c *CPU) uint16 {
	c.xorAccumulator(c.reg.C)
	return 1
}

func xraD(c *CPU) uint16 {
	c.xorAccumulator(c.reg.D)
	return 1
}

func xraE(c *CPU) uint16 {
	c.xorAccumulator(c.reg.E)
	return 1
}

func xraH(c *CPU) uint16 {
	c.xorAccumulator(c.reg.H)
	return 1
}

func xraL(c *CPU) uint16 {
	c.xorAccumulator(c.reg.L)
	return 1
}

func xraA(c *CPU) uint16 {
	c.xorAccumulator(c.reg.A)
	return 1
}

func xraM(c *CPU) uint16 {
	offset := c.getHL()
	c.xorAccumulator(c.mem[offset])
	return 1
}

func xri(c *CPU) uint16 {
	c.xorAccumulator(c.mem[c.pc+1])
	return 2
}

func cmpB(c *CPU) uint16 {
	c.compareAccumulator(c.reg.B)
	return 1
}

func cmpC(c *CPU) uint16 {
	c.compareAccumulator(c.reg.C)
	return 1
}

func cmpD(c *CPU) uint16 {
	c.compareAccumulator(c.reg.D)
	return 1
}

func cmpE(c *CPU) uint16 {
	c.compareAccumulator(c.reg.E)
	return 1
}

func cmpH(c *CPU) uint16 {
	c.compareAccumulator(c.reg.H)
	return 1
}

func cmpL(c *CPU) uint16 {
	c.compareAccumulator(c.reg.L)
	return 1
}

func cmpA(c *CPU) uint16 {
	c.compareAccumulator(c.reg.A)
	return 1
}

func cmpM(c *CPU) uint16 {
	offset := c.getHL()
	c.compareAccumulator(c.mem[offset])
	return 1
}

func cpi(c *CPU) uint16 {
	c.compareAccumulator(c.mem[c.pc+1])
	return 2
}

func in(c *CPU) uint16 {
	return 2
}

func out(c *CPU) uint16 {
	return 2
}

func ei(c *CPU) uint16 {
	c.intEnable = 1
	return 1
}

func di(c *CPU) uint16 {
	c.intEnable = 0
	return 1
}

func pushB(c *CPU) uint16 {
	c.mem[c.sp-1] = c.reg.B
	c.mem[c.sp-2] = c.reg.C
	c.sp -= 2
	return 1
}

func pushD(c *CPU) uint16 {
	c.mem[c.sp-1] = c.reg.D
	c.mem[c.sp-2] = c.reg.E
	c.sp -= 2
	return 1
}

func pushH(c *CPU) uint16 {
	c.mem[c.sp-1] = c.reg.H
	c.mem[c.sp-2] = c.reg.L
	c.sp -= 2
	return 1
}

func pushPSW(c *CPU) uint16 {
	c.mem[c.sp-1] = c.reg.A
	psw := (c.flags.Z | (c.flags.S << 1) | (c.flags.P << 2) | (c.flags.CY << 3) | (c.flags.AC << 4))
	c.mem[c.sp-2] = psw
	c.sp -= 2
	return 1
}

func popB(c *CPU) uint16 {
	c.reg.C = c.mem[c.sp]
	c.reg.B = c.mem[c.sp+1]
	c.sp += 2
	return 1
}

func popD(c *CPU) uint16 {
	c.reg.E = c.mem[c.sp]
	c.reg.D = c.mem[c.sp+1]
	c.sp += 2
	return 1
}

func popH(c *CPU) uint16 {
	c.reg.L = c.mem[c.sp]
	c.reg.H = c.mem[c.sp+1]
	c.sp += 2
	return 1
}

func popPSW(c *CPU) uint16 {
	c.reg.A = c.mem[c.sp+1]
	psw := c.mem[c.sp]
	if (psw & 0x01) == 0x01 {
		c.flags.Z = 1
	} else {
		c.flags.Z = 0
	}
	if (psw & 0x02) == 0x02 {
		c.flags.S = 1
	} else {
		c.flags.S = 0
	}
	if (psw & 0x04) == 0x04 {
		c.flags.P = 1
	} else {
		c.flags.P = 0
	}
	if (psw & 0x08) == 0x05 {
		c.flags.CY = 1
	} else {
		c.flags.CY = 0
	}
	if (psw & 0x10) == 0x10 {
		c.flags.AC = 1
	} else {
		c.flags.AC = 0
	}
	c.sp += 2
	return 1
}

func lhld(c *CPU) uint16 {
	c.reg.L = c.mem[c.getNextTwoBytes()]
	c.reg.H = c.mem[c.getNextTwoBytes()+1]
	return 3
}

func shld(c *CPU) uint16 {
	c.mem[c.getNextTwoBytes()] = c.reg.L
	c.mem[c.getNextTwoBytes()+1] = c.reg.H
	return 3
}

func xchg(c *CPU) uint16 {
	H := c.reg.H
	D := c.reg.D
	L := c.reg.L
	E := c.reg.E
	c.reg.D = H
	c.reg.H = D
	c.reg.L = E
	c.reg.E = L
	return 1
}

func xthl(c *CPU) uint16 {
	H := c.reg.H
	L := c.reg.L
	sp1 := c.mem[c.sp]
	sp2 := c.mem[c.sp+1]
	c.mem[c.sp] = L
	c.mem[c.sp+1] = H
	c.reg.H = sp2
	c.reg.L = sp1
	return 1
}

func sphl(c *CPU) uint16 {
	c.sp = c.getHL()
	return 1
}

func pchl(c *CPU) uint16 {
	c.pc = c.getHL()
	return 0
}

func rlc(c *CPU) uint16 {
	c.flags.CY = c.reg.A >> 7
	c.reg.A = (c.reg.A << 1) | c.flags.CY
	return 1
}

func rrc(c *CPU) uint16 {
	c.flags.CY = c.reg.A & 1
	c.reg.A = (c.reg.A >> 1) | (c.flags.CY << 7)
	return 1
}

func ral(c *CPU) uint16 {
	cy := c.flags.CY
	c.flags.CY = c.reg.A >> 7
	c.reg.A = (c.reg.A << 1) | cy
	return 1
}

func rar(c *CPU) uint16 {
	cy := c.flags.CY
	c.flags.CY = c.reg.A & 1
	c.reg.A = (c.reg.A >> 1) | (cy << 7)
	return 1
}

func stc(c *CPU) uint16 {
	c.flags.CY = 1
	return 1
}

func cmc(c *CPU) uint16 {
	c.flags.CY ^= 1
	return 1
}

func cma(c *CPU) uint16 {
	c.reg.A ^= 255
	return 1
}

func hlt(c *CPU) uint16 {
	c.halt = true
	return 1
}

func rst0(c *CPU) uint16 {
	callRst(c, 0x00)
	return 0
}

func rst1(c *CPU) uint16 {
	callRst(c, 0x08)
	return 0
}

func rst2(c *CPU) uint16 {
	callRst(c, 0x10)
	return 0
}

func rst3(c *CPU) uint16 {
	callRst(c, 0x18)
	return 0
}

func rst4(c *CPU) uint16 {
	callRst(c, 0x20)
	return 0
}

func rst5(c *CPU) uint16 {
	callRst(c, 0x28)
	return 0
}

func rst6(c *CPU) uint16 {
	callRst(c, 0x30)
	return 0
}

func rst7(c *CPU) uint16 {
	callRst(c, 0x38)
	return 0
}
