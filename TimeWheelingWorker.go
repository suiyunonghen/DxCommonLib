// Package DxCommonLib
/*
时间轮询调度池，只用一个定时器来实现After等超时设定，默认轮渡器设定为1个小时，精度为500毫秒
如果要使用更精确的定时器，请使用NewTimeWheelWorker自己指定定时器时间，目前在我的电脑上测试来看，最精确能到2毫秒
Author: 不得闲
QQ:75492895
*/
package DxCommonLib

import (
	"sync"
	"sync/atomic"
	"time"
)

type (
	//每个槽中的记录对象
	slotRecord struct {
		notifyCount   int32         //要通知多少个
		wheelCount    int32         //需要轮询多少圈触发
		curWheelIndex int32         //当前轮询的圈索引
		notifychan    chan struct{} //通知
		next          *slotRecord   //下一个轮询点
	}

	TimeWheelWorker struct {
		sync.Mutex       //调度锁
		curindex   int32 //当前的索引
		slockcount int32
		ticker     *time.Ticker  //调度器时钟
		timeslocks []*slotRecord //时间槽
		maxTimeout time.Duration
		quitchan   chan struct{}
		interval   time.Duration
		tkfunc     func()
		recordPool sync.Pool
	}
)

var (
	defaultTimeWheelWorker *TimeWheelWorker
	minTickerInterval      time.Duration
)

func init() {
	//获取一下最准确的精度
	minTickerInterval = time.Millisecond * 2
	ticker := time.NewTicker(minTickerInterval)
	start := time.Now()
	for i := 0; i < 9; i++ {
		<-ticker.C
	}
	cur := <-ticker.C
	minTickerInterval = cur.Sub(start) / time.Duration(10)
}

// NewTimeWheelWorker
// interval指定调度的时间间隔
// slotBlockCount指定时间轮的块长度
func NewTimeWheelWorker(interval time.Duration, slotBlockCount int32, tkfunc func()) *TimeWheelWorker {
	if interval < minTickerInterval {
		interval = minTickerInterval
	}
	result := new(TimeWheelWorker)
	result.interval = interval
	result.quitchan = make(chan struct{})
	result.slockcount = slotBlockCount
	result.tkfunc = tkfunc
	result.maxTimeout = interval * time.Duration(slotBlockCount)
	result.timeslocks = make([]*slotRecord, slotBlockCount)
	result.ticker = time.NewTicker(interval)
	go result.run()
	return result
}

func (worker *TimeWheelWorker) run() {
	for {
		select {
		case <-worker.ticker.C:
			curIndex := atomic.LoadInt32(&worker.curindex)
			nextIndex := curIndex + 1
			if nextIndex == worker.slockcount {
				nextIndex = 0
			}
			atomic.StoreInt32(&worker.curindex, nextIndex)
			worker.Lock()
			lastrec := worker.timeslocks[curIndex]
			if lastrec != nil {
				var firstrec *slotRecord
				for {
					currec := lastrec.next
					lastrec.curWheelIndex++
					if lastrec.curWheelIndex >= lastrec.wheelCount { //到时间了，释放掉
						worker.freeRecord(lastrec)
					} else if firstrec == nil {
						firstrec = lastrec //插入的时候就直接按照wheelCount大小排序了，只用增加一个个的序号就行了
						for currec != nil {
							currec.curWheelIndex++ //圈索引增加
							currec = currec.next
						}
						break
					} /*else{
						firstrec.next = lastrec
					}*/
					if currec == nil {
						break
					}
					lastrec = currec
				}
				worker.timeslocks[curIndex] = firstrec
			}
			worker.Unlock()
			if worker.tkfunc != nil {
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

func (worker *TimeWheelWorker) getRecord(wheelcount int32) *slotRecord {
	var result *slotRecord
	v := worker.recordPool.Get()
	if v != nil {
		result = v.(*slotRecord)
	} else {
		result = new(slotRecord)
		result.notifychan = make(chan struct{})
	}
	result.curWheelIndex = 0
	result.wheelCount = wheelcount
	result.next = nil
	return result
}

func (worker *TimeWheelWorker) freeRecord(rec *slotRecord) {
	//通知多少次
	notifyCount := int(atomic.SwapInt32(&rec.notifyCount, 0))
	for i := 0; i < notifyCount; i++ {
		rec.notifychan <- struct{}{}
	}
	//close(rec.notifychan)
	rec.next = nil
	rec.notifychan = nil
	rec.wheelCount = 0
	rec.curWheelIndex = 0
	worker.recordPool.Put(rec)
}

func (worker *TimeWheelWorker) After(d time.Duration) <-chan struct{} {
	index := int32(d / worker.interval)     //触发多少次到
	wheelcount := index / worker.slockcount //轮询多少圈
	if index%worker.slockcount > 0 {
		wheelcount++
	}
	if index > 0 {
		index--
	}
	index = (atomic.LoadInt32(&worker.curindex) + index) % worker.slockcount
	worker.Lock()
	rec := worker.timeslocks[index]
	if rec == nil {
		rec = worker.getRecord(wheelcount)
		worker.timeslocks[index] = rec
	} else { //查找对应的位置
		var last *slotRecord = nil
		for {
			currec := rec.next
			if wheelcount < rec.wheelCount {
				currec = worker.getRecord(wheelcount)
				currec.next = rec
				if last == nil {
					worker.timeslocks[index] = currec
				} else {
					last.next = currec
				}
				rec = currec
				break
			} else if wheelcount == rec.wheelCount { //已经存在，直接退出
				break
			} else if currec == nil {
				currec = worker.getRecord(wheelcount) //链接一个新的
				rec.next = currec
				rec = currec
				break
			}
			last = rec
			rec = currec
		}
	}
	atomic.AddInt32(&rec.notifyCount, 1)
	notifychan := rec.notifychan
	worker.Unlock()
	return notifychan
}

func (worker *TimeWheelWorker) AfterFunc(d time.Duration, afunc func()) {
	select {
	case <-worker.After(d):
		afunc()
	}
}

func (worker *TimeWheelWorker) Sleep(d time.Duration) {
	select {
	case <-worker.After(d):
		return
	}
}

func After(d time.Duration) <-chan struct{} {
	if defaultTimeWheelWorker == nil {
		defaultTimeWheelWorker = NewTimeWheelWorker(time.Millisecond*500, 7200, nil)
	}
	return defaultTimeWheelWorker.After(d)
}

func AfterFunc(d time.Duration, afunc func()) {
	if defaultTimeWheelWorker == nil {
		defaultTimeWheelWorker = NewTimeWheelWorker(time.Millisecond*500, 7200, nil)
	}
	defaultTimeWheelWorker.AfterFunc(d, afunc)
}

func Sleep(d time.Duration) {
	if defaultTimeWheelWorker == nil {
		defaultTimeWheelWorker = NewTimeWheelWorker(time.Millisecond*500, 7200, nil)
	}
	defaultTimeWheelWorker.Sleep(d)
}

func ReSetDefaultTimeWheel(Chkinterval time.Duration, slotBlockCount int32, tickerfunc func()) {
	if Chkinterval < minTickerInterval {
		Chkinterval = minTickerInterval
	}
	if defaultTimeWheelWorker == nil {
		defaultTimeWheelWorker = NewTimeWheelWorker(Chkinterval, slotBlockCount, tickerfunc)
		return
	}
	if nil != tickerfunc || defaultTimeWheelWorker.interval != Chkinterval ||
		defaultTimeWheelWorker.slockcount != slotBlockCount {
		defaultTimeWheelWorker.Stop()

		defaultTimeWheelWorker = NewTimeWheelWorker(Chkinterval, slotBlockCount, tickerfunc)
	}
}
