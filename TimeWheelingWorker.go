/*
时间轮询调度池，只用一个定时器来实现After等超时设定，默认轮渡器设定为1个小时，精度为500毫秒
Autor: 不得闲
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
		curWheelIndex		int			//当前轮询的索引
		wheelCount			int			//需要轮询多少次触发
		notifychan			chan struct{} //通知
		next				*slotRecord	//下一个轮询点
	}

	TimeWheelWorker struct {
		sync.Mutex                 //调度锁
		ticker     *time.Ticker    //调度器时钟
		timeslocks []*slotRecord   //时间槽
		slockcount int
		maxTimeout time.Duration
		quitchan   chan struct{}
		curindex   int //当前的索引
		interval   time.Duration
		tkfunc		func()
		recordPool	sync.Pool
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
	result.timeslocks = make([]*slotRecord, slotBlockCount)
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
			lastrec := worker.timeslocks[worker.curindex]
			if lastrec != nil{
				var firstrec *slotRecord
				for{
					currec := lastrec.next
					lastrec.curWheelIndex++
					if lastrec.curWheelIndex >= lastrec.wheelCount{ //断开
						worker.freeRecord(lastrec)
					}else if firstrec == nil{
						firstrec = lastrec //插入的时候就直接按照wheelCount大小排序了，只用增加一个个的序号就行了
						for currec != nil{
							currec.curWheelIndex++
							currec = currec.next
						}
						break
					}/*else{
						firstrec.next = lastrec
					}*/
					if currec == nil{
						break
					}
					lastrec = currec
				}
				worker.timeslocks[worker.curindex] = firstrec
			}
			worker.curindex = (worker.curindex + 1) % worker.slockcount
			worker.Unlock()
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

func (worker *TimeWheelWorker)getRecord(wheelcount int)*slotRecord  {
	var result *slotRecord
	v := worker.recordPool.Get()
	if v!=nil{
		result = v.(*slotRecord)
	}else{
		result = new(slotRecord)
	}
	result.curWheelIndex = 0
	result.wheelCount = wheelcount
	result.notifychan = make(chan struct{})
	result.next = nil
	return result
}

func (worker *TimeWheelWorker)freeRecord(rec *slotRecord)  {
	rec.next = nil
	close(rec.notifychan)
	rec.notifychan = nil
	rec.wheelCount = 0
	rec.curWheelIndex = 0
	worker.recordPool.Put(rec)
}

func (worker *TimeWheelWorker) After(d time.Duration) <-chan struct{} {
	index := int(d / worker.interval) //触发多少次到
	wheelcount := int(index / worker.slockcount)
	if index % worker.slockcount > 0{
		wheelcount++
	}
	if index > 0 {
		index--
	}
	worker.Lock()
	index = (worker.curindex + index) % worker.slockcount
	rec := worker.timeslocks[index]
	if rec == nil {
		rec = worker.getRecord(wheelcount)
		worker.timeslocks[index] = rec
	}else{ //查找对应的位置
		var last *slotRecord=nil
		for{
			currec := rec.next
			if wheelcount < rec.wheelCount{
				currec = worker.getRecord(wheelcount)
				currec.next = rec
				if last == nil{
					worker.timeslocks[index] = currec
				}else{
					last.next = currec
				}
				rec = currec
				break
			}else if wheelcount == rec.wheelCount{ //已经存在，直接退出
				break
			}else if currec == nil{
				currec = worker.getRecord(wheelcount) //链接一个新的
				rec.next = currec
				rec = currec
				break
			}
			last = rec
			rec = currec
		}
	}
	worker.Unlock()
	return rec.notifychan
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

func ReSetDefaultTimeWheel(Chkinterval time.Duration,slotBlockCount int){
	if defaultTimeWheelWorker.interval != Chkinterval  ||
		defaultTimeWheelWorker.slockcount != slotBlockCount{
			defaultTimeWheelWorker.Stop()
			defaultTimeWheelWorker = NewTimeWheelWorker(Chkinterval, slotBlockCount, func() {
				t := time.Now().Truncate(Chkinterval)
				coarseTime.Store(&t)
			})
	}
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

