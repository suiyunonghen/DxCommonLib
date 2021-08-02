package sync

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

func TestShareCalls_ShareCall(t *testing.T) {
	var calls ShareCalls
	var wg sync.WaitGroup
	ntime := time.Now()
	wg.Add(20)
	for i := 0;i<10;i++{
		go func(index int) {
			shareId := index % 5
			result,cacheValue,_ := calls.ShareCall(shareId, func(i ...interface{}) (interface{}, error) {
				time.Sleep(time.Second)
				return 30*index,nil
			})
			wg.Done()
			fmt.Fprintf(os.Stdout, "%d Intfunc(%d)=(%v,%v)\r\n",index,shareId,result,cacheValue)
		}(i)

		go func(index int) {
			KeyResutl,KeyCacheValue,_ := calls.ShareCallByKey("Key001", func(i ...interface{}) (interface{}, error) {
				time.Sleep(time.Second)
				return "返回的是Key",nil
			})
			wg.Done()
			fmt.Fprintf(os.Stdout,"Keyfunc(%d)=(%v,%v)\r\n",index,KeyResutl,KeyCacheValue)
		}(i)
	}

	wg.Wait()
	fmt.Println("执行时长：", time.Now().Sub(ntime))

}
