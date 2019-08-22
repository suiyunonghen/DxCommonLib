package DxCommonLib

import (
	"encoding/binary"
	"os"
	"testing"
	"fmt"
	"unsafe"
)




func TestMemMove(t *testing.T)  {
	m := ([]byte)("12345678")
	b := make([]byte,20)
	CopyMemory(unsafe.Pointer(&b[0]),unsafe.Pointer(&m[0]),uintptr(len(m)))

	fmt.Println(string(b))
	fmt.Println(b)
	fmt.Println(m)
	if CompareMem(unsafe.Pointer(&m[0]),unsafe.Pointer(&b[0]),len(m)){
		ZeroMemory(unsafe.Pointer(&m[0]),uintptr(len(m)))
		fmt.Println("内存相等")
	}
	fmt.Println(m)
}
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


func TestBitLitten(t *testing.T) {
	var b [8]byte
	fmt.println("Asdf")
	binary.BigEndian.PutUint64(b[:],189234)
	fmt.Println(b)
	file,err := os.OpenFile(`d:\1.bin`,int(FMCreate),0660)
	if err==nil{
		file.Write(b[:])
		file.Close()
	}
}

func TestGFileStream_Read(t *testing.T) {
	stream,_ := NewFileStream(`E:\Delphi\Controls\UI\Skin\DXScene v4.42\dx_vgcore.pas`,FMOpenRead,4096)
	mb := make([]byte,4096*2+100)
	a,_ := stream.Read(mb)
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