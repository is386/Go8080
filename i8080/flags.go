package i8080

type Flags struct {
	Z  uint8
	S  uint8
	P  uint8
	CY uint8
	AC uint8
}

func NewFlags() *Flags {
	return &Flags{Z: 1, S: 1, P: 1, CY: 1, AC: 1}
}
