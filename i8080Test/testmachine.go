package i8080Test

import (
	"fmt"

	"github.com/is386/Go8080/i8080"
)

type TestMachine struct {
	cpu        *i8080.CPU
	cycles     int
	instrCount int
	running    bool
	showDebug  bool
}

func NewTestMachine(filename string, showDebug bool) *TestMachine {
	tm := TestMachine{showDebug: showDebug, running: true}
	cpu := i8080.NewCPU(0x100, tm.portIn, tm.portOut)
	cpu.LoadRom(filename)
	cpu.Write(0x0, 0xD3)
	cpu.Write(0x1, 0x0)
	cpu.Write(0x5, 0xD3)
	cpu.Write(0x6, 0x01)
	cpu.Write(0x7, 0xC9)
	tm.cpu = cpu
	return &tm
}

func (tm *TestMachine) Run() {
	for tm.running {
		if tm.showDebug {
			tm.printState()
		}
		tm.cpu.Execute()
		tm.instrCount++
	}
	tm.cycles = tm.cpu.GetCycles()
	fmt.Println("\n\n----------------------")
	fmt.Printf("Test Completed\nInstructions: %d\nCycles: %d\n\n", tm.instrCount, tm.cycles)
}

func (tm *TestMachine) portIn(port uint8) {
}

func (tm *TestMachine) portOut(port uint8) {
	if port == 0 {
		tm.running = false
	} else if port == 1 {
		reg := tm.cpu.GetRegisters()
		if reg.C == 9 {
			fmt.Println()
			offset := tm.cpu.GetDE()
			mem := tm.cpu.GetMemory()
			str := mem[offset]
			for str != '$' {
				fmt.Printf("%c", str)
				offset += 1
				str = mem[offset]
			}
		} else if reg.C == 2 {
			fmt.Printf("%c", reg.E)
		}
	}
}

func (tm *TestMachine) printState() {
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
