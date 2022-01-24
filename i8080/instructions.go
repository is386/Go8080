package i8080

var (
	INSTRUCTIONS = map[uint8]func(*Emulator) uint16{
		0x00: noOp,
		0x01: lxiB,
		0x02: staxB,
		0x03: inxB,
		0x04: inrB,
		0x05: dcrB,
		0x06: mviB,
		0x07: unimplemented,
		0x08: unimplemented,
		0x09: dadB,
		0x0a: ldaxB,
		0x0b: dcxB,
		0x0c: inrC,
		0x0d: dcrC,
		0x0e: mviC,
		0x0f: rrc,
		0x10: unimplemented,
		0x11: lxiD,
		0x12: staxD,
		0x13: inxD,
		0x14: inrD,
		0x15: dcrD,
		0x16: mviD,
		0x17: unimplemented,
		0x18: unimplemented,
		0x19: dadD,
		0x1a: ldaxD,
		0x1b: dcxD,
		0x1c: inrE,
		0x1d: dcrE,
		0x1e: mviE,
		0x1f: unimplemented,
		0x20: unimplemented,
		0x21: lxiH,
		0x22: shld,
		0x23: inxH,
		0x24: inrH,
		0x25: dcrH,
		0x26: mviH,
		0x27: daa,
		0x28: unimplemented,
		0x29: dadH,
		0x2a: lhld,
		0x2b: dcxH,
		0x2c: inrL,
		0x2d: dcrL,
		0x2e: mviL,
		0x2f: unimplemented,
		0x30: unimplemented,
		0x31: lxiSP,
		0x32: sta,
		0x33: inxSP,
		0x34: inrM,
		0x35: dcrM,
		0x36: mviM,
		0x37: unimplemented,
		0x38: unimplemented,
		0x39: dadSP,
		0x3a: lda,
		0x3b: dcxSP,
		0x3c: inrA,
		0x3d: dcrA,
		0x3e: mviA,
		0x3f: unimplemented,
		0x40: movBB,
		0x41: movBC,
		0x42: movBD,
		0x43: movBE,
		0x44: movBH,
		0x45: movBL,
		0x46: movBM,
		0x47: movBA,
		0x48: movCB,
		0x49: movCC,
		0x4a: movCD,
		0x4b: movCE,
		0x4c: movCH,
		0x4d: movCL,
		0x4e: movCM,
		0x4f: movCA,
		0x50: movDB,
		0x51: movDC,
		0x52: movDD,
		0x53: movDE,
		0x54: movDH,
		0x55: movDL,
		0x56: movDM,
		0x57: movDA,
		0x58: movEB,
		0x59: movEC,
		0x5a: movED,
		0x5b: movEE,
		0x5c: movEH,
		0x5d: movEL,
		0x5e: movEM,
		0x5f: movEA,
		0x60: movHB,
		0x61: movHC,
		0x62: movHD,
		0x63: movHE,
		0x64: movHH,
		0x65: movHL,
		0x66: movHM,
		0x67: movHA,
		0x68: movLB,
		0x69: movLC,
		0x6a: movLD,
		0x6b: movLE,
		0x6c: movLH,
		0x6d: movLL,
		0x6e: movLM,
		0x6f: movLA,
		0x70: movMB,
		0x71: movMC,
		0x72: movMD,
		0x73: movME,
		0x74: movMH,
		0x75: movML,
		0x76: unimplemented,
		0x77: movMA,
		0x78: movAB,
		0x79: movAC,
		0x7a: movAD,
		0x7b: movAE,
		0x7c: movAH,
		0x7d: movAL,
		0x7e: movAM,
		0x7f: movAA,
		0x80: addB,
		0x81: addC,
		0x82: addD,
		0x83: addE,
		0x84: addH,
		0x85: addL,
		0x86: addM,
		0x87: addA,
		0x88: adcB,
		0x89: adcC,
		0x8a: adcD,
		0x8b: adcE,
		0x8c: adcH,
		0x8d: adcL,
		0x8e: adcM,
		0x8f: adcA,
		0x90: subB,
		0x91: subC,
		0x92: subD,
		0x93: subE,
		0x94: subH,
		0x95: subL,
		0x96: subM,
		0x97: subA,
		0x98: sbbB,
		0x99: sbbC,
		0x9a: sbbD,
		0x9b: sbbE,
		0x9c: sbbH,
		0x9d: sbbL,
		0x9e: sbbM,
		0x9f: sbbA,
		0xa0: anaB,
		0xa1: anaC,
		0xa2: anaD,
		0xa3: anaE,
		0xa4: anaH,
		0xa5: anaL,
		0xa6: anaM,
		0xa7: anaA,
		0xa8: xraB,
		0xa9: xraC,
		0xaa: xraD,
		0xab: xraE,
		0xac: xraH,
		0xad: xraL,
		0xae: xraM,
		0xaf: xraA,
		0xb0: oraB,
		0xb1: oraC,
		0xb2: oraD,
		0xb3: oraE,
		0xb4: oraH,
		0xb5: oraL,
		0xb6: oraM,
		0xb7: oraA,
		0xb8: cmpB,
		0xb9: cmpC,
		0xba: cmpD,
		0xbb: cmpE,
		0xbc: cmpH,
		0xbd: cmpL,
		0xbe: cmpM,
		0xbf: cmpA,
		0xc0: rnz,
		0xc1: popB,
		0xc2: jnz,
		0xc3: jmp,
		0xc4: cnz,
		0xc5: pushB,
		0xc6: adi,
		0xc7: unimplemented,
		0xc8: rz,
		0xc9: ret,
		0xca: jz,
		0xcb: unimplemented,
		0xcc: cz,
		0xcd: call,
		0xce: aci,
		0xcf: unimplemented,
		0xd0: rnc,
		0xd1: popD,
		0xd2: jnc,
		0xd3: out,
		0xd4: cnc,
		0xd5: pushD,
		0xd6: sui,
		0xd7: unimplemented,
		0xd8: rc,
		0xd9: unimplemented,
		0xda: jc,
		0xdb: in,
		0xdc: cc,
		0xdd: unimplemented,
		0xde: sbi,
		0xdf: unimplemented,
		0xe0: rpo,
		0xe1: popH,
		0xe2: jpo,
		0xe3: xthl,
		0xe4: cpo,
		0xe5: pushH,
		0xe6: ani,
		0xe7: unimplemented,
		0xe8: rpe,
		0xe9: unimplemented,
		0xea: jpe,
		0xeb: xchg,
		0xec: cpe,
		0xed: unimplemented,
		0xee: xri,
		0xef: unimplemented,
		0xf0: rp,
		0xf1: popPSW,
		0xf2: jp,
		0xf3: di,
		0xf4: cp,
		0xf5: pushPSW,
		0xf6: ori,
		0xf7: unimplemented,
		0xf8: rm,
		0xf9: unimplemented,
		0xfa: jm,
		0xfb: ei,
		0xfc: cm,
		0xfd: unimplemented,
		0xfe: cpi,
		0xff: unimplemented,
	}
)
