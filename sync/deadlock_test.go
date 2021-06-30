package sync

import (
	"testing"
	"time"
)

func TestRWMutexEx_Lock(t *testing.T) {
	go func() {
		var mu RWMutexEx
		mu.Lock()
		//time.Sleep(time.Second * 35)
		mu.Unlock()
	}()
	time.Sleep(time.Second)

}
