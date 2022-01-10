package i8080

import (
	"github.com/is386/GoCHIP/chip8"
)

// some instrs use the next 2-3 bytes as input
// ABCDEHL registers, PC, and stack
// BC, DE, and HL are pairs
// 16-bit addresses

/* Memory Map
ROM
$0000-$07FF: invaders.h
$0800-$0FFF: invaders.g
$1000-$17FF: invaders.f
$1800-$1FFF: invaders.e

RAM
$2000-$23FF: work
$2400-$3FFF: video
$4000 onwards: mirror
*/

type Emulator struct {
	memory    [64 * 1024]uint8
	registers *Registers
	pc        uint16
	stack     chip8.Stack
	flags     *Flags
}

func NewEmulator() *Emulator {
	return &Emulator{flags: NewFlags()}
}

func (e *Emulator) fetch() uint8 {
	return e.memory[e.pc]
}

func (e *Emulator) decode(opcode uint8) func(*Emulator) uint16 {
	return INSTRUCTIONS[opcode]
}

func (e *Emulator) Execute() {
	opcode := e.fetch()
	instr := e.decode(opcode)
	e.pc += instr(e)
}

func noOperation(e *Emulator) uint16 {
	return 1
}

func loadRegisterPairImmediate(e *Emulator) uint16 {
	e.registers.C = e.memory[e.pc+1]
	e.registers.B = e.memory[e.pc+2]
	return 3
}
