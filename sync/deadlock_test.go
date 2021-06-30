package sync

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestRWMutexEx_Lock(t *testing.T) {
	SetDeadCheck(true,func(deadLocks ...DeadLockInfo){
		for i := range deadLocks{
			fmt.Println("可能发生死锁了：",deadLocks[i].String())
		}
	},func(format string, data ...interface{}) {
		fmt.Fprintf(os.Stdout,format,data...)
	})
	go func() {
		var mu RWMutexEx
		mu.Lock()
		time.Sleep(time.Second * 35)
		mu.Unlock()
	}()
	time.Sleep(time.Minute)

}
