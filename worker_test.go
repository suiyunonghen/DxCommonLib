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
	fmt.Println(data...)
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

func TestTimeWheelWorker_After(t *testing.T) {

	mm := NewTimeWheelWorker(time.Millisecond*2, 6000) //目前只能精确到2毫秒,低于2毫秒，就不行了
	start := time.Now()
	c1 := mm.After(time.Second * 18)
	c2 := mm.After(time.Second * 8)
	c3 := mm.After(time.Second * 18)
	c4 := mm.After(time.Millisecond * 500)
	c11 := mm.After(time.Millisecond * 500)
	c12 := mm.After(time.Millisecond * 500)

	go func() {
		select {
		case <-c11:
			fmt.Println("C11触发：", time.Now().Sub(start))
		}
	}()

	go func() {
		select {
		case <-c12:
			fmt.Println("C12触发：", time.Now().Sub(start))
		}
	}()

	c5 := mm.After(time.Second * 2)
	go func() {
		select {
		case <-c2:
			fmt.Println("C2触发：", time.Now().Sub(start))
		}
	}()

	go func() {
		select {
		case <-c4:
			fmt.Println("C4触发：", time.Now().Sub(start))
		}
	}()

	go func() {
		select {
		case <-c5:
			fmt.Println("C5触发：", time.Now().Sub(start))
		}
	}()

	go func() {
		select {
		case <-c3:
			fmt.Println("C3触发：", time.Now().Sub(start))
		}
	}()

	go func() {
		select {
		case <-c1:
			fmt.Println("C1触发：", time.Now().Sub(start))
		}
	}()

	<-c1
}
