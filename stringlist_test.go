package DxCommonLib

import (
	"fmt"
	"syscall"
	"testing"
	"unsafe"
)


func TestStringList(t *testing.T) {
	fmt.Println(Ord(true),Ord(false))
	var lst GStringList
	for i := 0; i < 1000; i++ {
		lst.Add("测试不得闲,adfadfadsfadf1")
	}
	lst.SaveToFile("d:\\t.txt")
	lst.Clear()
	fmt.Println(lst.Text())
	lst.LoadFromFile("d:\\t.txt")
	fmt.Println(lst.Count())
	fmt.Println(lst.Text())
	fmt.Println(lst.Strings(2))

	lst.Clear()
	lst.AddPair("Name1", "不得闲")
	lst.AddPair("Age", "20")
	fmt.Println(lst.Text())
	fmt.Println(lst.ValueByName("Name1"))
	fmt.Println(lst.ValueFromIndex(1))
	fmt.Println(lst.IndexOfName("Age"))
}

func TestStringFromUtf8Pointer(t *testing.T) {
	str := "不得闲测试"
	utf16ptr, _ := syscall.UTF16PtrFromString(str)
	fmt.Println(StringFromUtf16Pointer(uintptr(unsafe.Pointer(utf16ptr)), 1024))
}