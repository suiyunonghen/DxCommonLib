package DxCommonLib

import (
	"fmt"
	"testing"
)

func TestStringList(t *testing.T) {
	var lst GStringList
	for i := 0; i < 1000; i++ {
		lst.Add("测试不得闲,adfadfadsfadf1")
	}
	lst.SaveToFile("d:\\t.txt")
	//fmt.Println(lst.Text())
	lst.Clear()
	fmt.Println(lst.Text())
	lst.LoadFromFile("d:\\t.txt")
	fmt.Println(lst.Count())
	fmt.Println(lst.Strings(2))

	lst.Clear()
	lst.AddPair("Name1", "不得闲")
	lst.AddPair("Age", "20")
	fmt.Println(lst.Text())
	fmt.Println(lst.ValueByName("Name1"))
	fmt.Println(lst.ValueFromIndex(1))
	fmt.Println(lst.IndexOfName("Age"))
}
