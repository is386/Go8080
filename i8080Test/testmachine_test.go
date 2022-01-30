package i8080Test

import (
	"testing"
)

var (
	TST8080  = "roms/TST8080.COM"
	i8080PRE = "roms/8080PRE.COM"
	CPUTEST  = "roms/CPUTEST.COM"
	i8080EXM = "roms/8080EXM.COM"
	DEBUG    = false
)

func TestTST8080(t *testing.T) {
	cycles := 4924
	tm := NewTestMachine(TST8080, DEBUG)
	tm.Run()
	if tm.cycles != cycles {
		t.Errorf("[cycles] expected: %d, actual: %d", cycles, tm.cycles)
	}
}

func Test8080PRE(t *testing.T) {
	cycles := 7817
	tm := NewTestMachine(i8080PRE, DEBUG)
	tm.Run()
	if tm.cycles != cycles {
		t.Errorf("[cycles] expected: %d, actual: %d", cycles, tm.cycles)
	}
}

func TestCPUTEST(t *testing.T) {
	cycles := 255653383
	tm := NewTestMachine(CPUTEST, DEBUG)
	tm.Run()
	if tm.cycles != cycles {
		t.Errorf("[cycles] expected: %d, actual: %d", cycles, tm.cycles)
	}
}

func Test8080EXM(t *testing.T) {
	cycles := 23803381171
	tm := NewTestMachine(i8080EXM, DEBUG)
	tm.Run()
	if tm.cycles != cycles {
		t.Errorf("[cycles] expected: %d, actual: %d", cycles, tm.cycles)
	}
}
