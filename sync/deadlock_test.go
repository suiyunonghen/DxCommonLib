package sync

import (
	"fmt"
	"github.com/suiyunonghen/DxCommonLib"
	"os"
	"strconv"
	"testing"
	"time"
)

func doTest()  {
	StartDeadCheck(func(deadLocks []byte){
		fmt.Println(DxCommonLib.FastByte2String(deadLocks))
	},time.Second * 10,time.Second * 30,func(panicMsg bool,format string, data ...interface{}) {
		fmt.Fprintf(os.Stdout,format,data...)
	})
	var mu RWMutexEx
	mu.LockWithMsg("测试信息:root")
	for i := 0;i<3;i++{
		go func(v int) {
			mu.LockWithMsg("测试信息:"+strconv.Itoa(v))
			time.Sleep(time.Second * 15)
			mu.Unlock()
		}(i)
	}
	time.Sleep(time.Second * 40)
	mu.Unlock()
}

func TestRWMutexEx_Lock(t *testing.T) {
	doTest()
}
