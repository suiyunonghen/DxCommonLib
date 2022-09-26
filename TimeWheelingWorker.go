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
		notifyCount   int32           //要通知多少个
		wheelCount    int32           //需要轮询多少圈触发
		curWheelIndex int32           //当前轮询的圈索引
		slotTask      []defTaskRunner //能执行的任务信息
		mu            sync.Mutex
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
func NewTimeWheelWorker(interval time.Duration, slotBlockCount int32) *TimeWheelWorker {
	if interval < minTickerInterval {
		interval = minTickerInterval
	}
	result := new(TimeWheelWorker)
	result.interval = interval
	result.quitchan = make(chan struct{})
	result.slockcount = slotBlockCount
	result.maxTimeout = interval * time.Duration(slotBlockCount)
	result.timeslocks = make([]*slotRecord, slotBlockCount)
	result.ticker = time.NewTicker(interval)
	go result.run()
	return result
}

func (worker *TimeWheelWorker) run() {
	runAfter := func(data ...interface{}) {
		slotTaskRunner := data[0].([]defTaskRunner)
		for i := range slotTaskRunner{
			slotTaskRunner[i].runFunc(slotTaskRunner[i].runArgs...)
		}
	}
	for {
		select {
		case <-worker.ticker.C:
			var slotTaskRunner []defTaskRunner
			worker.Lock()
			lastRec := worker.timeslocks[worker.curindex]
			if lastRec != nil {
				var firstRec *slotRecord
				for {
					curRec := lastRec.next
					lastRec.curWheelIndex++
					if lastRec.curWheelIndex >= lastRec.wheelCount { //到时间了，释放掉
						/*for{
							select{
							case <-worker.After(20):
								//这个After就可能会被丢弃，所以实际的通知数量可能不会有设定的个数大小
							case <-chan2:
							default:
							}
						}*/
						//通知多少次，实际的通知次数可能会比这个设定的次数小
						notifyCount := int(atomic.SwapInt32(&lastRec.notifyCount, 0))
						for i := 0; i < notifyCount; i++ {
							select {
							case lastRec.notifychan <- struct{}{}:
								//通知成功
							default:
								break
							}
						}
						slotTaskRunner = append(slotTaskRunner,lastRec.slotTask...)
						for i := range lastRec.slotTask {
							lastRec.slotTask[i].runFunc = nil
							lastRec.slotTask[i].runArgs = nil
						}
						lastRec.next = nil
						lastRec.slotTask = lastRec.slotTask[:0]
						lastRec.wheelCount = 0
						lastRec.curWheelIndex = 0
						worker.recordPool.Put(lastRec)
					} else if firstRec == nil {
						firstRec = lastRec //插入的时候就直接按照wheelCount大小排序了，只用增加一个个的序号就行了
						for curRec != nil {
							curRec.curWheelIndex++ //圈索引增加
							curRec = curRec.next
						}
						break
					}
					if curRec == nil {
						break
					}
					lastRec = curRec
				}
				worker.timeslocks[worker.curindex] = firstRec
			}
			worker.curindex++
			if worker.curindex == worker.slockcount {
				worker.curindex = 0
			}
			worker.Unlock()
			if len(slotTaskRunner) > 0{
				MustRunAsync(runAfter,slotTaskRunner)
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
		result.slotTask = make([]defTaskRunner, 0, 8)
	}
	result.curWheelIndex = 0
	result.notifyCount = 0
	result.wheelCount = wheelcount
	result.next = nil
	return result
}

func (worker *TimeWheelWorker) after(d time.Duration) *slotRecord {
	index := int32(d / worker.interval)     //触发多少次到
	wheelCount := index / worker.slockcount //轮询多少圈
	if index%worker.slockcount > 0 {
		wheelCount++
	}
	if index > 0 {
		index--
	}
	worker.Lock()
	index = (worker.curindex + index) % worker.slockcount
	rec := worker.timeslocks[index]
	if rec == nil {
		rec = worker.getRecord(wheelCount)
		worker.timeslocks[index] = rec
	} else { //查找对应的位置
		var last *slotRecord = nil
		for {
			curRec := rec.next
			if wheelCount < rec.wheelCount {
				//在当前圈前面，时间靠前插入
				curRec = worker.getRecord(wheelCount)
				curRec.next = rec
				if last == nil {
					worker.timeslocks[index] = curRec
				} else {
					last.next = curRec
				}
				rec = curRec
				break
			} else if wheelCount == rec.wheelCount { //已经存在，直接退出
				break
			} else if curRec == nil {
				curRec = worker.getRecord(wheelCount) //链接一个新的
				rec.next = curRec
				rec = curRec
				break
			}
			last = rec
			rec = curRec
		}
	}
	worker.Unlock()
	return rec
}

func (worker *TimeWheelWorker) After(d time.Duration) <-chan struct{} {
	rec := worker.after(d)
	atomic.AddInt32(&rec.notifyCount, 1)
	return rec.notifychan
}

func (worker *TimeWheelWorker) AfterFunc(d time.Duration, afterFunc GWorkerFunc, data ...interface{}) {
	rec := worker.after(d)
	rec.mu.Lock()
	rec.slotTask = append(rec.slotTask, defTaskRunner{
		runFunc: afterFunc,
		runArgs: data,
	})
	rec.mu.Unlock()
}

func (worker *TimeWheelWorker) Sleep(d time.Duration) {
	<-worker.After(d)
}

func After(d time.Duration) <-chan struct{} {
	if defaultTimeWheelWorker == nil {
		defaultTimeWheelWorker = NewTimeWheelWorker(time.Millisecond*500, 7200)
	}
	return defaultTimeWheelWorker.After(d)
}

func AfterFunc(d time.Duration, aFunc GWorkerFunc) {
	if defaultTimeWheelWorker == nil {
		defaultTimeWheelWorker = NewTimeWheelWorker(time.Millisecond*500, 7200)
	}
	defaultTimeWheelWorker.AfterFunc(d, aFunc)
}

func Sleep(d time.Duration) {
	if defaultTimeWheelWorker == nil {
		defaultTimeWheelWorker = NewTimeWheelWorker(time.Millisecond*500, 7200)
	}
	defaultTimeWheelWorker.Sleep(d)
}

func ReSetDefaultTimeWheel(Chkinterval time.Duration, slotBlockCount int32) {
	if Chkinterval < minTickerInterval {
		Chkinterval = minTickerInterval
	}
	if defaultTimeWheelWorker == nil {
		defaultTimeWheelWorker = NewTimeWheelWorker(Chkinterval, slotBlockCount)
		return
	}
	if defaultTimeWheelWorker.interval != Chkinterval ||
		defaultTimeWheelWorker.slockcount != slotBlockCount {
		defaultTimeWheelWorker.Stop()

		defaultTimeWheelWorker = NewTimeWheelWorker(Chkinterval, slotBlockCount)
	}
}
