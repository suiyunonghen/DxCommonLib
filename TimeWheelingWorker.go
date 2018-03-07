/*
时间轮询调度池，只用一个定时器来实现After等超时设定，默认轮渡器设定为1个小时，精度为500毫秒
Autor: 不得闲
QQ:75492895
*/
package DxCommonLib

import (
	"sync"
	"time"
	"sync/atomic"
)

type (
	TimeWheelWorker struct {
		sync.Mutex                 //调度锁
		ticker     *time.Ticker    //调度器时钟
		timeslocks []chan struct{} //时间槽
		slockcount int
		maxTimeout time.Duration
		quitchan   chan struct{}
		curindex   int //当前的索引
		interval   time.Duration
		tkfunc		func()
	}
)

var (
	defaultTimeWheelWorker *TimeWheelWorker
    coarseTime 				atomic.Value		//存放的是当前的实际时间
)

//interval指定调度的时间间隔
//slotBlockCount指定时间轮的块长度
func NewTimeWheelWorker(interval time.Duration, slotBlockCount int,tkfunc func()) *TimeWheelWorker {
	result := new(TimeWheelWorker)
	result.interval = interval
	result.quitchan = make(chan struct{})
	result.slockcount = slotBlockCount
	result.tkfunc = tkfunc
	result.maxTimeout = interval * time.Duration(slotBlockCount)
	result.timeslocks = make([]chan struct{}, slotBlockCount)
	result.ticker = time.NewTicker(interval)
	go result.run()
	return result
}

func (worker *TimeWheelWorker) run() {
	for {
		select {
		case <-worker.ticker.C:
			//执行定时操作
			//获取当前的时间槽数据
			worker.Lock()
			lastC := worker.timeslocks[worker.curindex]
			worker.timeslocks[worker.curindex] = make(chan struct{})
			worker.curindex = (worker.curindex + 1) % worker.slockcount
			worker.Unlock()
			if lastC != nil {
				close(lastC)
			}
			if worker.tkfunc!=nil{
				worker.tkfunc()
			}
		case <-worker.quitchan:
			worker.ticker.Stop()
			return
		}
	}
}

func (worker *TimeWheelWorker) Stop() {
	close(worker.quitchan)
}

func (worker *TimeWheelWorker) After(d time.Duration) <-chan struct{} {
	if d >= worker.maxTimeout {
		panic("timeout too much, over maxtimeout")
	}
	index := int(d / worker.interval)
	if index > 0 {
		index--
	}
	worker.Lock()
	index = (worker.curindex + index) % worker.slockcount
	b := worker.timeslocks[index]
	if b == nil {
		b = make(chan struct{})
		worker.timeslocks[index] = b
	}
	worker.Unlock()
	return b
}

func (worker *TimeWheelWorker)AfterFunc(d time.Duration,afunc func())  {
	select{
	case <-worker.After(d):
		afunc()
	}
}

func (worker *TimeWheelWorker) Sleep(d time.Duration) {
	select{
	case <-worker.After(d):
		return
	}
}

func After(d time.Duration) <-chan struct{} {
	return defaultTimeWheelWorker.After(d)
}

func AfterFunc(d time.Duration,afunc func()) {
	defaultTimeWheelWorker.AfterFunc(d,afunc)
}

func Sleep(d time.Duration) {
	defaultTimeWheelWorker.Sleep(d)
}

func init()  {
	defaultTimeWheelWorker = NewTimeWheelWorker(time.Millisecond*500, 7200, func() {
		t := time.Now().Truncate(time.Millisecond*500)
		coarseTime.Store(&t)
	})
	t := time.Now().Truncate(time.Millisecond*500)
	coarseTime.Store(&t)
}

func CoarseTimeNow() time.Time {
	tp := coarseTime.Load().(*time.Time)
	return *tp
}

