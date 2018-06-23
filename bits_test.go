package DxCommonLib

import (
	"testing"
	"fmt"
)

func TestTDxBits_Bits(t *testing.T) {
	bits := DxBits{}
	bits.ReSetByInt32(255)
	fmt.Println(bits.Bits(3))
	fmt.Println(bits.AsInt32())
	bits.SetBits(3,true)
	fmt.Println(bits.AsInt32())
	fmt.Println(bits.Bits(3))

	bits.ReSetByInt32(-1)
	bits.NotBits(-1)
	fmt.Println(bits.AsInt32())
}