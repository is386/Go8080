package i8080Test

import (
	"testing"
)

var (
	TST8080 = "TST8080.COM"
	DEBUG   = false
)

func TestTST8080(t *testing.T) {
	cycles := 4924
	tm := NewTestMachine(TST8080, DEBUG)
	tm.Run()
	if tm.cycles != cycles {
		t.Errorf("[cycles] expected: %d, actual: %d", cycles, tm.cycles)
	}
}
