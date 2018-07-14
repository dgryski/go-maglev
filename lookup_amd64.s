#include "textflag.h"

TEXT ·lookup(SB),NOSPLIT,$0
	MOVQ t+0(FP), DI         // DI := &t
	MOVQ key+8(FP), AX       // AX := key
	MOVQ 24(DI), BX          // BX := &t.assignments[0]
	MOVQ 32(DI), DX          // DX := len(t.assignments)
	MOVQ DX, SI
	XORL DX, DX
	DIVQ SI                  // DX := key % len(t.assignments) (i.e. partition)
	MOVWLZX (BX)(DX*2), DX   // DX = t.assignments[partition]
	MOVWQSX	DX, DX
	MOVQ (DI), BX            // BX := &t.names[0]
	SHLQ $4, DX              // DX *= 16 (i.e. 16 bytes per string element within t.names)
	MOVAPS (BX)(DX*1), X0    // X0 := t.names[DX]
	MOVUPS X0, name+24(FP)   // name := X0
	RET
	