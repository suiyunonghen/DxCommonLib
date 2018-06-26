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

func TestGFileStream_Read(t *testing.T) {
	stream,_ := NewFileStream(`E:\Delphi\Controls\UI\Skin\DXScene v4.42\dx_vgcore.pas`,FMOpenRead,4096)
	mb := make([]byte,4096*2+100)
	a := stream.Read(mb)
	fmt.Println(a)
	fmt.Println(string(mb))
	fmt.Println(stream.FilePosition())
	fmt.Println(stream.Position())

	stream.Read(mb)
	fmt.Println(stream.FilePosition())

	fmt.Println(string(mb))
	stream.Read(mb)
	fmt.Println(string(mb))
	stream.Read(mb)
	fmt.Println(string(mb))
}

func TestGFileStream_Write(t *testing.T) {
	stream,_ := NewFileStream(`E:\1.txt`,FMOpenReadWrite,4096)
	stream.Write([]byte("测试不得闲"))
	stream.Close()
}