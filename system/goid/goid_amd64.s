#include "textflag.h"
#include "go_tls.h"


// func getg() interface{}
TEXT ·getG(SB), NOSPLIT, $32-16
	// get runtime.g
	get_tls(CX)
	MOVQ g(CX), AX
	// get runtime.g type
	MOVQ $type·runtime·g(SB), BX

	// convert (*g) to interface{}
	MOVQ AX, 8(SP)
	MOVQ BX, 0(SP)
	CALL ·runtime_convT2E_hack(SB)
	MOVQ 16(SP), AX
	MOVQ 24(SP), BX

	// return interface{}
	MOVQ AX, ret+0(FP)
	MOVQ BX, ret+8(FP)
	RET

// func GetGoID() int64
TEXT ·GetGoID(SB), NOSPLIT, $0-8
	get_tls(CX)
	MOVQ g(CX), AX
	MOVQ ·offset(SB), BX
	LEAQ 0(AX)(BX*1), DX
	MOVQ (DX), AX
	MOVQ AX, ret+0(FP)
	RET
