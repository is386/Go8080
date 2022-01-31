package i8080

import (
	"io/ioutil"
	"os"
)

var (
	CYCLES = [256]int{
		04, 10, 07, 05, 05, 05, 07, 04, 04, 10, 07, 05, 05, 05, 07, 04,
		04, 10, 07, 05, 05, 05, 07, 04, 04, 10, 07, 05, 05, 05, 07, 04,
		04, 10, 16, 05, 05, 05, 07, 04, 04, 10, 16, 05, 05, 05, 07, 04,
		04, 10, 13, 05, 10, 10, 10, 04, 04, 10, 13, 05, 05, 05, 07, 04,
		05, 05, 05, 05, 05, 05, 07, 05, 05, 05, 05, 05, 05, 05, 07, 05,
		05, 05, 05, 05, 05, 05, 07, 05, 05, 05, 05, 05, 05, 05, 07, 05,
		05, 05, 05, 05, 05, 05, 07, 05, 05, 05, 05, 05, 05, 05, 07, 05,
		07, 07, 07, 07, 07, 07, 07, 07, 05, 05, 05, 05, 05, 05, 07, 05,
		04, 04, 04, 04, 04, 04, 07, 04, 04, 04, 04, 04, 04, 04, 07, 04,
		04, 04, 04, 04, 04, 04, 07, 04, 04, 04, 04, 04, 04, 04, 07, 04,
		04, 04, 04, 04, 04, 04, 07, 04, 04, 04, 04, 04, 04, 04, 07, 04,
		04, 04, 04, 04, 04, 04, 07, 04, 04, 04, 04, 04, 04, 04, 07, 04,
		05, 10, 10, 10, 11, 11, 07, 11, 05, 10, 10, 10, 11, 17, 07, 11,
		05, 10, 10, 10, 11, 11, 07, 11, 05, 10, 10, 10, 11, 17, 07, 11,
		05, 10, 10, 18, 11, 11, 07, 11, 05, 05, 10, 04, 11, 17, 07, 11,
		05, 10, 10, 04, 11, 11, 07, 11, 05, 05, 10, 04, 11, 17, 07, 11}
)

type CPU struct {
	mem             [64 * 1024]uint8
	reg             *Registers
	flags           *Flags
	pc, sp          uint16
	romMax, ramMax  uint32
	cyc             int
	intEnabled      bool
	portIn, portOut func(uint8)
}

func NewCPU(pc uint16, romMax uint32, ramMax uint32, portIn func(uint8), portOut func(uint8)) *CPU {
	return &CPU{reg: &Registers{}, flags: &Flags{},
		ramMax: ramMax, romMax: romMax,
		pc: pc, portIn: portIn, portOut: portOut}
}

func (c *CPU) GetMemory() []uint8 {
	return c.mem[:]
}

func (c *CPU) GetRegisters() *Registers {
	return c.reg
}

func (c *CPU) GetPC() uint16 {
	return c.pc
}

func (c *CPU) GetSP() uint16 {
	return c.sp
}

func (c *CPU) GetCycles() int {
	return c.cyc
}

func (c *CPU) SubtractCycles(cyc int) {
	c.cyc -= cyc
}

func (c *CPU) IsInterrupted() bool {
	return c.intEnabled
}

func (c *CPU) GetAF() uint16 {
	f := uint8(0)
	f |= c.flags.S << 7
	f |= c.flags.Z << 6
	f |= c.flags.AC << 4
	f |= c.flags.P << 2
	f |= 1 << 1
	f |= c.flags.CY << 0
	return (uint16(c.reg.A) << 8) | uint16(f)
}

func (c *CPU) GetBC() uint16 {
	return (uint16(c.reg.B) << 8) | uint16(c.reg.C)
}

func (c *CPU) GetDE() uint16 {
	return (uint16(c.reg.D) << 8) | uint16(c.reg.E)
}

func (c *CPU) GetHL() uint16 {
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

func (c *CPU) LoadRom(filename string) {
	rom, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(rom); i++ {
		c.mem[c.pc+uint16(i)] = rom[i]
	}
}

func (c *CPU) Write(addr uint16, val uint8) {
	if uint32(addr) >= c.romMax && uint32(addr) < c.ramMax {
		c.mem[addr] = val
	}
}

func (c *CPU) read(addr uint16) uint8 {
	return c.mem[addr]
}

func (c *CPU) getNextByte() uint8 {
	b := c.read(c.pc)
	c.pc += 1
	return b
}

func (c *CPU) getNextTwoBytes() uint16 {
	a := c.getNextByte()
	b := c.getNextByte()
	return (uint16(b) << 8) | uint16(a)
}

func (c *CPU) fetch() uint8 {
	return c.getNextByte()
}

func (c *CPU) decode(opcode uint8) func(*CPU) {
	return INSTRUCTIONS[opcode]
}

func (c *CPU) Execute() {
	opcode := c.fetch()
	c.cyc += CYCLES[opcode]
	instr := c.decode(opcode)
	instr(c)
}

func (c *CPU) Interrupt(vector uint16) {
	c.push(c.pc)
	c.intEnabled = false
	c.pc = vector
}

func (c *CPU) setZSP(val uint8) {
	c.setZero(uint16(val))
	c.setSign(uint16(val))
	c.setParity(uint16(val))
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
	ones := uint16(0)
	for i := 0; i < 8; i++ {
		ones += ((val >> i) & 1)
	}
	if (ones % 2) == 0 {
		c.flags.P = 1
	} else {
		c.flags.P = 0
	}
}

func (c *CPU) setCarry(val uint16) {
	if val > 0xff {
		c.flags.CY = 1
	} else {
		c.flags.CY = 0
	}
}

func flip(val uint8) uint8 {
	if val == 1 {
		return 0
	} else {
		return 1
	}
}

func (c *CPU) add(val uint8, cy uint8) {
	ans := uint16(c.reg.A) + uint16(val) + uint16(cy)
	c.setZSP(uint8(ans))
	c.setCarry(ans)
	if ((uint16(c.reg.A) ^ uint16(val) ^ ans) & 0x10) > 0 {
		c.flags.AC = 1
	} else {
		c.flags.AC = 0
	}
	c.reg.A = uint8(ans)
}

func (c *CPU) sub(val uint8, cy uint8) {
	cy = flip(cy)
	c.add(^val, cy)
	c.flags.CY = flip(c.flags.CY)
}

func (c *CPU) inr(val uint8) uint8 {
	val++
	c.setZSP(val)
	if (val & 0xf) == 0 {
		c.flags.AC = 1
	} else {
		c.flags.AC = 0
	}
	return val
}

func (c *CPU) dcr(val uint8) uint8 {
	val--
	c.setZSP(val)
	if (val & 0xf) == 0xf {
		c.flags.AC = 0
	} else {
		c.flags.AC = 1
	}
	return val
}

func (c *CPU) dad(val uint16) {
	ans := uint32(c.GetHL()) + uint32(val)
	c.setHL(uint16(ans))
	if ((ans >> 16) & 1) == 1 {
		c.flags.CY = 1
	} else {
		c.flags.CY = 0
	}
}

func (c *CPU) and(val uint8) {
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

func (c *CPU) xor(val uint8) {
	ans := uint16(c.reg.A ^ val)
	c.setZSP(uint8(ans))
	c.flags.CY = 0
	c.flags.AC = 0
	c.reg.A = uint8(ans)
}

func (c *CPU) or(val uint8) {
	ans := uint16(c.reg.A | val)
	c.setZSP(uint8(ans))
	c.flags.CY = 0
	c.flags.AC = 0
	c.reg.A = uint8(ans)
}

func (c *CPU) cmp(val uint8) {
	ans := uint16(c.reg.A) - uint16(val)
	c.setZSP(uint8(ans))
	c.setCarry(ans)
	if (^(uint16(c.reg.A) ^ ans ^ uint16(val)) & 0x10) > 0 {
		c.flags.AC = 1
	} else {
		c.flags.AC = 0
	}
}

func (c *CPU) push(val uint16) {
	c.Write(c.sp-1, uint8(val>>8))
	c.Write(c.sp-2, uint8(val&0xff))
	c.sp -= 2
}

func (c *CPU) pop() uint16 {
	c.sp += 2
	return ((uint16(c.read(c.sp-1)) << 8) | uint16(c.read(c.sp-2)))
}

// Data Transfer Group

func movBB(c *CPU) {
}

func movBC(c *CPU) {
	c.reg.B = c.reg.C
}

func movBD(c *CPU) {
	c.reg.B = c.reg.D
}

func movBE(c *CPU) {
	c.reg.B = c.reg.E
}

func movBH(c *CPU) {
	c.reg.B = c.reg.H
}

func movBL(c *CPU) {
	c.reg.B = c.reg.L
}

func movBM(c *CPU) {
	c.reg.B = c.read(c.GetHL())
}

func movBA(c *CPU) {
	c.reg.B = c.reg.A
}

func movCB(c *CPU) {
	c.reg.C = c.reg.B
}

func movCC(c *CPU) {}

func movCD(c *CPU) {
	c.reg.C = c.reg.D
}

func movCE(c *CPU) {
	c.reg.C = c.reg.E
}

func movCH(c *CPU) {
	c.reg.C = c.reg.H
}

func movCL(c *CPU) {
	c.reg.C = c.reg.L
}

func movCM(c *CPU) {
	c.reg.C = c.read(c.GetHL())
}

func movCA(c *CPU) {
	c.reg.C = c.reg.A
}

func movDB(c *CPU) {
	c.reg.D = c.reg.B
}

func movDC(c *CPU) {
	c.reg.D = c.reg.C
}

func movDD(c *CPU) {}

func movDE(c *CPU) {
	c.reg.D = c.reg.E
}

func movDH(c *CPU) {
	c.reg.D = c.reg.H
}

func movDL(c *CPU) {
	c.reg.D = c.reg.L
}

func movDM(c *CPU) {
	c.reg.D = c.read(c.GetHL())
}

func movDA(c *CPU) {
	c.reg.D = c.reg.A
}

func movEB(c *CPU) {
	c.reg.E = c.reg.B
}

func movEC(c *CPU) {
	c.reg.E = c.reg.C
}

func movED(c *CPU) {
	c.reg.E = c.reg.D
}

func movEE(c *CPU) {}

func movEH(c *CPU) {
	c.reg.E = c.reg.H
}

func movEL(c *CPU) {
	c.reg.E = c.reg.L
}

func movEM(c *CPU) {
	c.reg.E = c.read(c.GetHL())
}

func movEA(c *CPU) {
	c.reg.E = c.reg.A
}

func movHB(c *CPU) {
	c.reg.H = c.reg.B
}

func movHC(c *CPU) {
	c.reg.H = c.reg.C
}

func movHD(c *CPU) {
	c.reg.H = c.reg.D
}

func movHE(c *CPU) {
	c.reg.H = c.reg.E
}

func movHH(c *CPU) {}

func movHL(c *CPU) {
	c.reg.H = c.reg.L
}

func movHM(c *CPU) {
	c.reg.H = c.read(c.GetHL())
}

func movHA(c *CPU) {
	c.reg.H = c.reg.A
}

func movLB(c *CPU) {
	c.reg.L = c.reg.B
}

func movLC(c *CPU) {
	c.reg.L = c.reg.C
}

func movLD(c *CPU) {
	c.reg.L = c.reg.D
}

func movLE(c *CPU) {
	c.reg.L = c.reg.E
}

func movLH(c *CPU) {
	c.reg.L = c.reg.H
}

func movLL(c *CPU) {}

func movLM(c *CPU) {
	c.reg.L = c.read(c.GetHL())
}

func movLA(c *CPU) {
	c.reg.L = c.reg.A
}

func movAB(c *CPU) {
	c.reg.A = c.reg.B
}

func movAC(c *CPU) {
	c.reg.A = c.reg.C
}

func movAD(c *CPU) {
	c.reg.A = c.reg.D
}

func movAE(c *CPU) {
	c.reg.A = c.reg.E
}

func movAH(c *CPU) {
	c.reg.A = c.reg.H
}

func movAL(c *CPU) {
	c.reg.A = c.reg.L
}

func movAM(c *CPU) {
	c.reg.A = c.read(c.GetHL())
}

func movAA(c *CPU) {}

func movMB(c *CPU) {
	c.Write(c.GetHL(), c.reg.B)
}

func movMC(c *CPU) {
	c.Write(c.GetHL(), c.reg.C)
}

func movMD(c *CPU) {
	c.Write(c.GetHL(), c.reg.D)
}

func movME(c *CPU) {
	c.Write(c.GetHL(), c.reg.E)
}

func movMH(c *CPU) {
	c.Write(c.GetHL(), c.reg.H)
}

func movML(c *CPU) {
	c.Write(c.GetHL(), c.reg.L)
}

func movMA(c *CPU) {
	c.Write(c.GetHL(), c.reg.A)
}

func mviB(c *CPU) {
	c.reg.B = c.getNextByte()
}

func mviC(c *CPU) {
	c.reg.C = c.getNextByte()
}

func mviD(c *CPU) {
	c.reg.D = c.getNextByte()
}

func mviE(c *CPU) {
	c.reg.E = c.getNextByte()
}

func mviH(c *CPU) {
	c.reg.H = c.getNextByte()
}

func mviL(c *CPU) {
	c.reg.L = c.getNextByte()
}

func mviA(c *CPU) {
	c.reg.A = c.getNextByte()
}

func mviM(c *CPU) {
	c.Write(c.GetHL(), c.getNextByte())
}

func lxiB(c *CPU) {
	c.setBC(c.getNextTwoBytes())
}

func lxiD(c *CPU) {
	c.setDE(c.getNextTwoBytes())
}

func lxiH(c *CPU) {
	c.setHL(c.getNextTwoBytes())
}

func lxiSP(c *CPU) {
	c.sp = c.getNextTwoBytes()
}

func lda(c *CPU) {
	c.reg.A = c.read(c.getNextTwoBytes())
}

func sta(c *CPU) {
	c.Write(c.getNextTwoBytes(), c.reg.A)
}

func lhld(c *CPU) {
	addr := c.getNextTwoBytes()
	c.reg.L = c.read(addr)
	c.reg.H = c.read(addr + 1)
}

func shld(c *CPU) {
	addr := c.getNextTwoBytes()
	c.Write(addr, c.reg.L)
	c.Write(addr+1, c.reg.H)
}

func ldaxB(c *CPU) {
	c.reg.A = c.read(c.GetBC())
}

func ldaxD(c *CPU) {
	c.reg.A = c.read(c.GetDE())
}

func staxB(c *CPU) {
	c.Write(c.GetBC(), c.reg.A)
}

func staxD(c *CPU) {
	c.Write(c.GetDE(), c.reg.A)
}

func xchg(c *CPU) {
	c.reg.H, c.reg.D = c.reg.D, c.reg.H
	c.reg.L, c.reg.E = c.reg.E, c.reg.L
}

// Arithmetic Group

func addB(c *CPU) {
	c.add(c.reg.B, 0)
}

func addC(c *CPU) {
	c.add(c.reg.C, 0)
}

func addD(c *CPU) {
	c.add(c.reg.D, 0)
}

func addE(c *CPU) {
	c.add(c.reg.E, 0)
}

func addH(c *CPU) {
	c.add(c.reg.H, 0)
}

func addL(c *CPU) {
	c.add(c.reg.L, 0)
}

func addM(c *CPU) {
	c.add(c.read(c.GetHL()), 0)
}

func addA(c *CPU) {
	c.add(c.reg.A, 0)
}

func adi(c *CPU) {
	c.add(c.getNextByte(), 0)
}

func adcB(c *CPU) {
	c.add(c.reg.B, c.flags.CY)
}

func adcC(c *CPU) {
	c.add(c.reg.C, c.flags.CY)
}

func adcD(c *CPU) {
	c.add(c.reg.D, c.flags.CY)
}

func adcE(c *CPU) {
	c.add(c.reg.E, c.flags.CY)
}

func adcH(c *CPU) {
	c.add(c.reg.H, c.flags.CY)
}

func adcL(c *CPU) {
	c.add(c.reg.L, c.flags.CY)
}

func adcA(c *CPU) {
	c.add(c.reg.A, c.flags.CY)
}

func adcM(c *CPU) {
	c.add(c.read(c.GetHL()), c.flags.CY)
}

func aci(c *CPU) {
	c.add(c.getNextByte(), c.flags.CY)
}

func subB(c *CPU) {
	c.sub(c.reg.B, 0)
}

func subC(c *CPU) {
	c.sub(c.reg.C, 0)
}

func subD(c *CPU) {
	c.sub(c.reg.D, 0)
}

func subE(c *CPU) {
	c.sub(c.reg.E, 0)
}

func subH(c *CPU) {
	c.sub(c.reg.H, 0)
}

func subL(c *CPU) {
	c.sub(c.reg.L, 0)
}

func subA(c *CPU) {
	c.sub(c.reg.A, 0)
}

func subM(c *CPU) {
	c.sub(c.read(c.GetHL()), 0)
}

func sui(c *CPU) {
	c.sub(c.getNextByte(), 0)
}

func sbbB(c *CPU) {
	c.sub(c.reg.B, c.flags.CY)
}

func sbbC(c *CPU) {
	c.sub(c.reg.C, c.flags.CY)
}

func sbbD(c *CPU) {
	c.sub(c.reg.D, c.flags.CY)
}

func sbbE(c *CPU) {
	c.sub(c.reg.E, c.flags.CY)
}

func sbbH(c *CPU) {
	c.sub(c.reg.H, c.flags.CY)
}

func sbbL(c *CPU) {
	c.sub(c.reg.L, c.flags.CY)
}

func sbbA(c *CPU) {
	c.sub(c.reg.A, c.flags.CY)
}

func sbbM(c *CPU) {
	c.sub(c.read(c.GetHL()), c.flags.CY)
}

func sbi(c *CPU) {
	c.sub(c.getNextByte(), c.flags.CY)
}

func inrB(c *CPU) {
	c.reg.B = c.inr(c.reg.B)
}

func inrC(c *CPU) {
	c.reg.C = c.inr(c.reg.C)
}

func inrD(c *CPU) {
	c.reg.D = c.inr(c.reg.D)
}

func inrE(c *CPU) {
	c.reg.E = c.inr(c.reg.E)
}

func inrH(c *CPU) {
	c.reg.H = c.inr(c.reg.H)
}

func inrL(c *CPU) {
	c.reg.L = c.inr(c.reg.L)
}

func inrA(c *CPU) {
	c.reg.A = c.inr(c.reg.A)
}

func inrM(c *CPU) {
	c.Write(c.GetHL(), c.inr(c.read(c.GetHL())))
}

func dcrB(c *CPU) {
	c.reg.B = c.dcr(c.reg.B)
}

func dcrC(c *CPU) {
	c.reg.C = c.dcr(c.reg.C)
}

func dcrD(c *CPU) {
	c.reg.D = c.dcr(c.reg.D)
}

func dcrE(c *CPU) {
	c.reg.E = c.dcr(c.reg.E)
}

func dcrH(c *CPU) {
	c.reg.H = c.dcr(c.reg.H)
}

func dcrL(c *CPU) {
	c.reg.L = c.dcr(c.reg.L)
}

func dcrA(c *CPU) {
	c.reg.A = c.dcr(c.reg.A)
}

func dcrM(c *CPU) {
	c.Write(c.GetHL(), c.dcr(c.read(c.GetHL())))
}

func inxB(c *CPU) {
	c.setBC(c.GetBC() + 1)
}

func inxD(c *CPU) {
	c.setDE(c.GetDE() + 1)
}

func inxH(c *CPU) {
	c.setHL(c.GetHL() + 1)
}

func inxSP(c *CPU) {
	c.sp += 1
}

func dcxB(c *CPU) {
	c.setBC(c.GetBC() - 1)
}

func dcxD(c *CPU) {
	c.setDE(c.GetDE() - 1)
}

func dcxH(c *CPU) {
	c.setHL(c.GetHL() - 1)
}

func dcxSP(c *CPU) {
	c.sp -= 1
}

func dadB(c *CPU) {
	c.dad(c.GetBC())
}

func dadD(c *CPU) {
	c.dad(c.GetDE())
}

func dadH(c *CPU) {
	c.dad(c.GetHL())
}

func dadSP(c *CPU) {
	c.dad(c.sp)
}

func daa(c *CPU) {
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

	c.add(uint8(correction), 0)
	c.flags.CY = cy
}

// Logical Group

func anaB(c *CPU) {
	c.and(c.reg.B)
}

func anaC(c *CPU) {
	c.and(c.reg.C)
}

func anaD(c *CPU) {
	c.and(c.reg.D)
}

func anaE(c *CPU) {
	c.and(c.reg.E)
}

func anaH(c *CPU) {
	c.and(c.reg.H)
}

func anaL(c *CPU) {
	c.and(c.reg.L)
}

func anaA(c *CPU) {
	c.and(c.reg.A)
}

func anaM(c *CPU) {
	c.and(c.read(c.GetHL()))
}

func ani(c *CPU) {
	c.and(c.getNextByte())
}

func oraB(c *CPU) {
	c.or(c.reg.B)
}

func oraC(c *CPU) {
	c.or(c.reg.C)
}

func oraD(c *CPU) {
	c.or(c.reg.D)
}

func oraE(c *CPU) {
	c.or(c.reg.E)
}

func oraH(c *CPU) {
	c.or(c.reg.H)
}

func oraL(c *CPU) {
	c.or(c.reg.L)
}

func oraA(c *CPU) {
	c.or(c.reg.A)
}

func oraM(c *CPU) {
	c.or(c.read(c.GetHL()))
}

func ori(c *CPU) {
	c.or(c.getNextByte())
}

func xraB(c *CPU) {
	c.xor(c.reg.B)
}

func xraC(c *CPU) {
	c.xor(c.reg.C)
}

func xraD(c *CPU) {
	c.xor(c.reg.D)
}

func xraE(c *CPU) {
	c.xor(c.reg.E)
}

func xraH(c *CPU) {
	c.xor(c.reg.H)
}

func xraL(c *CPU) {
	c.xor(c.reg.L)
}

func xraA(c *CPU) {
	c.xor(c.reg.A)
}

func xraM(c *CPU) {
	c.xor(c.read(c.GetHL()))
}

func xri(c *CPU) {
	c.xor(c.getNextByte())
}

func cmpB(c *CPU) {
	c.cmp(c.reg.B)
}

func cmpC(c *CPU) {
	c.cmp(c.reg.C)
}

func cmpD(c *CPU) {
	c.cmp(c.reg.D)
}

func cmpE(c *CPU) {
	c.cmp(c.reg.E)
}

func cmpH(c *CPU) {
	c.cmp(c.reg.H)
}

func cmpL(c *CPU) {
	c.cmp(c.reg.L)
}

func cmpA(c *CPU) {
	c.cmp(c.reg.A)
}

func cmpM(c *CPU) {
	c.cmp(c.read(c.GetHL()))
}

func cpi(c *CPU) {
	c.cmp(c.getNextByte())
}

func rlc(c *CPU) {
	c.flags.CY = c.reg.A >> 7
	c.reg.A = (c.reg.A << 1) | c.flags.CY
}

func rrc(c *CPU) {
	c.flags.CY = c.reg.A & 1
	c.reg.A = (c.reg.A >> 1) | (c.flags.CY << 7)
}

func ral(c *CPU) {
	cy := c.flags.CY
	c.flags.CY = c.reg.A >> 7
	c.reg.A = (c.reg.A << 1) | cy
}

func rar(c *CPU) {
	cy := c.flags.CY
	c.flags.CY = c.reg.A & 1
	c.reg.A = (c.reg.A >> 1) | (cy << 7)
}

func cma(c *CPU) {
	c.reg.A ^= 255
}

func cmc(c *CPU) {
	c.flags.CY ^= 1
}

func stc(c *CPU) {
	c.flags.CY = 1
}

// Branch Group

func jmp(c *CPU) {
	c.pc = c.getNextTwoBytes()
}

func jmpCond(c *CPU, cond bool) {
	if cond {
		jmp(c)
	} else {
		c.pc += 2
	}
}

func jnz(c *CPU) {
	jmpCond(c, c.flags.Z == 0)
}

func jz(c *CPU) {
	jmpCond(c, c.flags.Z == 1)
}

func jnc(c *CPU) {
	jmpCond(c, c.flags.CY == 0)
}

func jc(c *CPU) {
	jmpCond(c, c.flags.CY == 1)
}

func jpo(c *CPU) {
	jmpCond(c, c.flags.P == 0)
}

func jpe(c *CPU) {
	jmpCond(c, c.flags.P == 1)
}

func jp(c *CPU) {
	jmpCond(c, c.flags.S == 0)
}

func jm(c *CPU) {
	jmpCond(c, c.flags.S == 1)
}

func ret(c *CPU) {
	c.pc = c.pop()
}

func retCond(c *CPU, cond bool) {
	if cond {
		ret(c)
		c.cyc += 6
	}
}

func rnz(c *CPU) {
	retCond(c, c.flags.Z == 0)
}

func rz(c *CPU) {
	retCond(c, c.flags.Z == 1)
}

func rnc(c *CPU) {
	retCond(c, c.flags.CY == 0)
}

func rc(c *CPU) {
	retCond(c, c.flags.CY == 1)
}

func rpo(c *CPU) {
	retCond(c, c.flags.P == 0)
}

func rpe(c *CPU) {
	retCond(c, c.flags.P == 1)
}

func rp(c *CPU) {
	retCond(c, c.flags.S == 0)
}

func rm(c *CPU) {
	retCond(c, c.flags.S == 1)
}

func call(c *CPU) {
	addr := c.getNextTwoBytes()
	c.push(c.pc)
	c.pc = addr
}

func callCond(c *CPU, cond bool) {
	if cond {
		call(c)
		c.cyc += 6
	} else {
		c.pc += 2
	}
}

func cnz(c *CPU) {
	callCond(c, c.flags.Z == 0)
}

func cz(c *CPU) {
	callCond(c, c.flags.Z == 1)
}

func cnc(c *CPU) {
	callCond(c, c.flags.CY == 0)
}

func cc(c *CPU) {
	callCond(c, c.flags.CY == 1)
}

func cpo(c *CPU) {
	callCond(c, c.flags.P == 0)
}

func cpe(c *CPU) {
	callCond(c, c.flags.P == 1)
}

func cp(c *CPU) {
	callCond(c, c.flags.S == 0)
}

func cm(c *CPU) {
	callCond(c, c.flags.S == 1)
}

func callRst(c *CPU, addr uint16) {
	call(c)
	c.pc = addr
}

func rst0(c *CPU) {
	callRst(c, 0x00)
}

func rst1(c *CPU) {
	callRst(c, 0x08)
}

func rst2(c *CPU) {
	callRst(c, 0x10)
}

func rst3(c *CPU) {
	callRst(c, 0x18)
}

func rst4(c *CPU) {
	callRst(c, 0x20)
}

func rst5(c *CPU) {
	callRst(c, 0x28)
}

func rst6(c *CPU) {
	callRst(c, 0x30)
}

func rst7(c *CPU) {
	callRst(c, 0x38)
}

func pchl(c *CPU) {
	c.pc = c.GetHL()
}

// Stack Group

func pushB(c *CPU) {
	c.push(c.GetBC())
}

func pushD(c *CPU) {
	c.push(c.GetDE())
}

func pushH(c *CPU) {
	c.push(c.GetHL())
}

func pushPSW(c *CPU) {
	psw := uint8(0)
	psw |= c.flags.Z << 6
	psw |= c.flags.S << 7
	psw |= c.flags.P << 2
	psw |= c.flags.CY << 0
	psw |= c.flags.AC << 4
	psw |= 1 << 1
	c.push((uint16(c.reg.A) << 8) | uint16(psw))
}

func popB(c *CPU) {
	c.setBC(c.pop())
}

func popD(c *CPU) {
	c.setDE(c.pop())
}

func popH(c *CPU) {
	c.setHL(c.pop())
}

func popPSW(c *CPU) {
	af := c.pop()
	c.reg.A = uint8(af >> 8)
	psw := uint8(af & 0xff)
	c.flags.Z = (psw >> 6) & 1
	c.flags.S = (psw >> 7) & 1
	c.flags.P = (psw >> 2) & 1
	c.flags.CY = (psw >> 0) & 1
	c.flags.AC = (psw >> 4) & 1
}

func xthl(c *CPU) {
	sp1 := c.read(c.sp)
	sp2 := c.read(c.sp + 1)
	c.Write(c.sp, c.reg.L)
	c.Write(c.sp+1, c.reg.H)
	c.reg.H = sp2
	c.reg.L = sp1
}

func sphl(c *CPU) {
	c.sp = c.GetHL()
}

// IO and Machine Control Group

func in(c *CPU) {
	c.portIn(c.getNextByte())
}

func out(c *CPU) {
	c.portOut(c.getNextByte())
}

func ei(c *CPU) {
	c.intEnabled = true
}

func di(c *CPU) {
	c.intEnabled = false
}

func hlt(c *CPU) {
	os.Exit(0)
}

func noOp(c *CPU) {}
