package DxCommonLib

import (
	"fmt"
	"testing"
	"time"
)

func worktest(data ...interface{}) {
	if len(data) < 3 {
		Sleep(time.Second)
	}
	fmt.Println(data)
}

func TestWorker(t *testing.T) {
	work := NewWorkers(100, 0)
	for i := 0; i <= 10; i++ {
		work.PostFunc(worktest, 3, 4)
		work.PostFunc(worktest, 3, 4, 6, 786, 867)
		work.PostFunc(worktest, 345, 34, "$56456")
	}
	<-After(time.Second * 12)
	fmt.Println("准备关闭")
	work.Stop()
}
