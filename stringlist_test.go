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
}
