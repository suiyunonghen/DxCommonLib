/*
从fasthttp中变更过来的GoRoutine池
Autor: 不得闲
QQ:75492895
*/
package DxCommonLib

import (
	_ "go.uber.org/automaxprocs"
	"runtime"
	"sync"
	"time"
)

type (
	GWorkerFunc func(data ...interface{})
	GWorkers    struct {
		mustStop           bool
		fMaxWorkersCount   int //能同时存在的最多线程个数
		workersCount       int
		fMaxWorkerIdleTime time.Duration //线程能空闲的最长时间，超过这个时间了会回收掉这个线程
		ready              []*workerChan //准备好的空闲线程
		lock               sync.Mutex
		workerChanPool     sync.Pool //工作者
		fStopChan          chan struct{}
	}

	workerChan struct {
		lastUseTime time.Time
		fOwner      *GWorkers
		fCurTask    chan defTaskRunner //ITaskRunner
	}

	defTaskRunner struct {
		runFunc GWorkerFunc
		runArgs []interface{}
	}
)

func (workers *GWorkers) Start() {
	if workers.fStopChan != nil {
		panic("BUG: GWorkers already started")
	}
	workers.fStopChan = make(chan struct{})
	stopCh := workers.fStopChan
	workers.workerChanPool.New = func() interface{} {
		return &workerChan{
			fOwner:   workers,
			fCurTask: make(chan defTaskRunner, workerChanCap),
		}
	}
	go func() {
		var scratch []*workerChan
		for {
			select {
			case <-stopCh:
				return
			case <-After(workers.fMaxWorkerIdleTime):
				//执行上了
				workers.clean(&scratch) //定时执行清理回收线程
			}
		}
	}()
}

func (workers *GWorkers) Stop() {
	if workers.fStopChan == nil {
		panic("BUG: GWorkers wasn't started")
	}
	close(workers.fStopChan)
	workers.fStopChan = nil

	// Stop all the workers waiting for incoming connections.
	// Do not wait for busy workers - they will stop after
	// serving the connection and noticing wp.mustStop = true.
	workers.lock.Lock()
	ready := workers.ready
	l := len(ready)
	for i := 0; i < l; i++ {
		ready[i].fCurTask <- defTaskRunner{}
		ready[i] = nil
	}
	workers.ready = ready[:0]
	workers.mustStop = true
	workers.lock.Unlock()
}

var workerChanCap = func() int {
	// Use blocking workerChan if GOMAXPROCS=1.
	// This immediately switches Serve to WorkerFunc, which results
	// in higher performance (under go1.5 at least).
	if runtime.GOMAXPROCS(0) == 1 {
		return 0
	}

	// Use non-blocking workerChan if GOMAXPROCS>1,
	// since otherwise the Serve caller (Acceptor) may lag accepting
	// new connections if WorkerFunc is CPU-bound.
	return 1
}()

func (workers *GWorkers) clean(scratch *[]*workerChan) {
	// Clean least recently used workers if they didn't serve connections
	// for more than maxIdleWorkerDuration.
	criticalTime := time.Now().Add(-workers.fMaxWorkerIdleTime)
	workers.lock.Lock()
	ready := workers.ready
	n := len(ready)
	//超过设定的最大空闲时间的，就解雇掉
	// Use binary-search algorithm to find out the index of the least recently worker which can be cleaned up.
	l, r, mid := 0, n-1, 0
	for l <= r {
		mid = (l + r) / 2
		if criticalTime.After(workers.ready[mid].lastUseTime) {
			l = mid + 1
		} else {
			r = mid - 1
		}
	}
	if r == -1 {
		workers.lock.Unlock()
		return
	}
	i := r
	*scratch = append((*scratch)[:0], ready[:i+1]...)
	m := copy(ready, ready[i+1:])
	for i = m; i < n; i++ {
		ready[i] = nil
	}
	workers.ready = ready[:m]
	workers.lock.Unlock()

	// Notify obsolete workers to stop.
	// This notification must be outside the wp.lock, since ch.ch
	// may be blocking and may consume a lot of time if many workers
	// are located on non-local CPUs.
	tmp := *scratch
	for i = 0; i < len(tmp); i++ {
		tmp[i].fCurTask <- defTaskRunner{}
		tmp[i] = nil
	}
}

func (workers *GWorkers) getCh() *workerChan {
	var ch *workerChan
	createWorker := false

	workers.lock.Lock()
	ready := workers.ready
	n := len(ready) - 1
	if n < 0 {
		if workers.workersCount < workers.fMaxWorkersCount {
			createWorker = true
			workers.workersCount++
		}
	} else {
		ch = ready[n]
		ready[n] = nil
		workers.ready = ready[:n]
	}
	workers.lock.Unlock()

	if ch == nil {
		if !createWorker {
			return nil
		}
		vch := workers.workerChanPool.Get()
		ch = vch.(*workerChan)
		go func() {
			workers.workerFunc(ch)
			workers.workerChanPool.Put(vch)
		}()
	}
	return ch
}

func (workers *GWorkers) release(ch *workerChan) bool {
	ch.lastUseTime = time.Now()
	workers.lock.Lock()
	if workers.mustStop {
		workers.lock.Unlock()
		return false
	}
	workers.ready = append(workers.ready, ch)
	workers.lock.Unlock()
	return true
}

func (workers *GWorkers) workerFunc(ch *workerChan) {
	//waitTimes := workers.fMaxWorkerIdleTime + time.Second * 5
	for {
		curTask := <-ch.fCurTask
		if curTask.runFunc == nil {
			break
		}
		curTask.runFunc(curTask.runArgs...)
		if !workers.release(ch) {
			break
		}
		/*select {
		case curTask := <-ch.fCurTask:
			if curTask.runFunc == nil {
				break
			}
			curTask.runFunc(curTask.runArgs...)
			if !workers.release(ch) {
				break
			}
		case <-After(waitTimes):
			//这么长时间都没有等到信号，是不是已经跪了
			//删除这个长期占有的channel
			workers.lock.Lock()
			n := len(workers.ready)
			for i := n - 1;i>=0;i--{
				if workers.ready[i] == ch{
					if i == n - 1{
						workers.ready = workers.ready[:i]
					}else{
						workers.ready = append(workers.ready[:i],workers.ready[i+1:]...)
					}
					workers.workersCount--
					break
				}
			}
			workers.lock.Unlock()
			return
		}*/
	}
	workers.lock.Lock()
	workers.workersCount--
	workers.lock.Unlock()
}

func (workers *GWorkers) PostFunc(routineFunc GWorkerFunc, params ...interface{}) bool {
	wch := workers.getCh()
	if wch != nil {
		wch.fCurTask <- defTaskRunner{
			runFunc: routineFunc,
			runArgs: params,
		}
		return true
	}
	return false
}

// MustPostFunc 必然投递
func (workers *GWorkers) MustPostFunc(routineFunc GWorkerFunc, params ...interface{}) {
	for {
		wch := workers.getCh()
		if wch != nil {
			wch.fCurTask <- defTaskRunner{
				runFunc: routineFunc,
				runArgs: params,
			}
			return
		}
		runtime.Gosched()
	}
}

// MustRunAsync 必须异步执行到
func (workers *GWorkers) MustRunAsync(routineFunc GWorkerFunc, params ...interface{}) {
	for i := 0; i < 10; i++ {
		wch := workers.getCh()
		if wch != nil {
			wch.fCurTask <- defTaskRunner{
				runFunc: routineFunc,
				runArgs: params,
			}
			return
		}
		runtime.Gosched()
	}
	go routineFunc(params...)
}

func (workers *GWorkers) TryPostAndRun(routineFunc GWorkerFunc, params ...interface{}) {
	for idx := 0; idx < 10; idx++ {
		wch := workers.getCh()
		if wch != nil {
			wch.fCurTask <- defTaskRunner{
				runFunc: routineFunc,
				runArgs: params,
			}
			return
		}
		runtime.Gosched()
	}
	routineFunc(params...)
	return
}

func NewWorkers(maxGoroutinesAmount int, maxGoroutineIdleDuration time.Duration) *GWorkers {
	gp := new(GWorkers)
	if maxGoroutinesAmount <= 0 {
		gp.fMaxWorkersCount = 512 * 1024
	} else {
		gp.fMaxWorkersCount = maxGoroutinesAmount
	}
	if maxGoroutineIdleDuration <= 0 {
		gp.fMaxWorkerIdleTime = 10 * time.Second
	} else {
		gp.fMaxWorkerIdleTime = maxGoroutineIdleDuration
	}
	gp.Start()
	return gp
}

var defWorkers *GWorkers

func ResetDefaultWorker(maxGoroutinesAmount int, maxGoroutineIdleDuration time.Duration) {
	if defWorkers != nil {
		defWorkers.Stop()
	}
	defWorkers = NewWorkers(maxGoroutinesAmount, maxGoroutineIdleDuration)
}

func PostFunc(routineFunc GWorkerFunc, params ...interface{}) bool {
	if defWorkers == nil {
		defWorkers = NewWorkers(0, 0)
	}
	return defWorkers.PostFunc(routineFunc, params...)
}

func MustPostFunc(routineFunc GWorkerFunc, params ...interface{}) {
	if defWorkers == nil {
		defWorkers = NewWorkers(0, 0)
	}
	defWorkers.MustPostFunc(routineFunc, params...)
}

func TryPostAndRun(routineFunc GWorkerFunc, params ...interface{}) {
	if defWorkers == nil {
		defWorkers = NewWorkers(0, 0)
	}
	defWorkers.TryPostAndRun(routineFunc, params...)
}

// MustRunAsync 必须异步执行到
func MustRunAsync(routineFunc GWorkerFunc, params ...interface{}) {
	if defWorkers == nil {
		defWorkers = NewWorkers(0, 0)
	}
	defWorkers.MustRunAsync(routineFunc, params...)
}

func StopWorkers() {
	if defWorkers != nil {
		defWorkers.Stop()
	}
}
