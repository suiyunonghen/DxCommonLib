/*
从fasthttp中变更过来的GoRoutine池
Autor: 不得闲
QQ:75492895
*/
package DxCommonLib

import (
	"sync"
	"runtime"
	"time"
)

type (
	//任务
	ITaskRunner interface {
		Run()
	}

	GWorkerFunc func(data ...interface{})
	GWorkers    struct {
		fMaxWorkersCount		int				//能同时存在的最多线程个数
		fMaxWorkerIdleTime		time.Duration	//线程能空闲的最长时间，超过这个时间了会回收掉这个线程
		ready 					[]*workerChan   //准备好的空闲线程
		lock           	   		sync.Mutex
		workerChanPool 			sync.Pool		//工作者
		deftaskrunnerPool		sync.Pool		//默认任务
		mustStop				bool
		fstopchan  				chan struct{}
		workersCount 			int
	}

	workerChan struct {
		lastUseTime		time.Time
		fOwner			*GWorkers
		fcurTask		chan ITaskRunner
	}

	defTaskRunner struct {
		runfunc GWorkerFunc
		runargs []interface{}
	}
)

func (runner *defTaskRunner) Run() {
	runner.runfunc(runner.runargs...)
}

func (workers *GWorkers) Start() {
	if workers.fstopchan != nil {
		panic("BUG: GWorkers already started")
	}
	workers.fstopchan = make(chan struct{})
	stopCh := workers.fstopchan
	workers.workerChanPool.New = func() interface{} {
		return &workerChan{
			fOwner:workers,
			fcurTask: make(chan ITaskRunner, workerChanCap),
		}
	}
	go func() {
		var scratch []*workerChan
		for {
			select {
			case <-stopCh:
				return
			case <-After(workers.fMaxWorkerIdleTime):
				workers.clean(&scratch)//定时执行清理回收线程
			}
		}
	}()
}


func (workers *GWorkers) Stop() {
	if workers.fstopchan == nil {
		panic("BUG: GWorkers wasn't started")
	}
	close(workers.fstopchan)
	workers.fstopchan = nil

	// Stop all the workers waiting for incoming connections.
	// Do not wait for busy workers - they will stop after
	// serving the connection and noticing wp.mustStop = true.
	workers.lock.Lock()
	ready := workers.ready
	l := len(ready)
	for i:=0;i<l;i++{
		ready[i].fcurTask <- nil
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
	maxIdleWorkerDuration := workers.fMaxWorkerIdleTime

	// Clean least recently used workers if they didn't serve connections
	// for more than maxIdleWorkerDuration.
	criticalTime := time.Now().Add(-maxIdleWorkerDuration)
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
	i := r
	if i == -1 {
		workers.lock.Unlock()
		return
	}

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
	l = len(tmp)
	for i:=0;i< l;i++{
		tmp[i].fcurTask <- nil
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
	for curTask := range ch.fcurTask {
		if curTask == nil {
			break
		}
		curTask.Run() //执行
		//回收curtask
		switch runner := curTask.(type) {
		case *defTaskRunner:
			workers.deftaskrunnerPool.Put(runner)
		}
		if !workers.release(ch) {
			break
		}
	}
	workers.lock.Lock()
	workers.workersCount--
	workers.lock.Unlock()
}

func (workers *GWorkers)PostFunc(routineFunc GWorkerFunc,params ...interface{})bool  {
	wch := workers.getCh()
	if wch != nil {
		taskrunner := workers.deftaskrunnerPool.Get().(*defTaskRunner)
		taskrunner.runfunc = routineFunc
		taskrunner.runargs = params
		wch.fcurTask <- taskrunner
		return true
	}
	return false
}

//必须异步执行到
func (workers *GWorkers)MustRunAsync(routineFunc GWorkerFunc,params ...interface{})  {
	for idx := 0;idx<=4;idx++{
		if workers.PostFunc(routineFunc,params...){
			return
		}
		runtime.Gosched()
	}
	go routineFunc(params...)
}

func (workers *GWorkers)TryPostAndRun(routineFunc GWorkerFunc,params ...interface{})bool  {
	if workers.PostFunc(routineFunc,params...){
		return true
	}
	routineFunc(params...)
	return false
}

func (workers *GWorkers)Post(runner ITaskRunner)bool  {
	wch := workers.getCh()
	if wch != nil {
		wch.fcurTask <- runner
		return true
	}
	return false
}

func NewWorkers(maxGoroutinesAmount int, maxGoroutineIdleDuration time.Duration) *GWorkers {
	gp := new(GWorkers)
	gp.deftaskrunnerPool.New = func() interface{} {
		return new(defTaskRunner)
	}
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

func ResetDefaultWorker(maxGoroutinesAmount int, maxGoroutineIdleDuration time.Duration)  {
	if defWorkers != nil{
		defWorkers.Stop()
	}
	defWorkers = NewWorkers(maxGoroutinesAmount,maxGoroutineIdleDuration)
}

func PostFunc(routineFunc GWorkerFunc,params ...interface{})bool  {
	if defWorkers == nil{
		defWorkers = NewWorkers(0,0)
	}
	return defWorkers.PostFunc(routineFunc,params...)
}

func TryPostAndRun(routineFunc GWorkerFunc,params ...interface{})bool  {
	if defWorkers == nil{
		defWorkers = NewWorkers(0,0)
	}
	return defWorkers.TryPostAndRun(routineFunc,params...)
}

//必须异步执行到
func MustRunAsync(routineFunc GWorkerFunc,params ...interface{})  {
	if defWorkers == nil{
		defWorkers = NewWorkers(0,0)
	}
	defWorkers.MustRunAsync(routineFunc,params...)
}

func Post(runner ITaskRunner)  {
	if defWorkers == nil{
		defWorkers = NewWorkers(0,0)
	}
	defWorkers.Post(runner)
}

func StopWorkers()  {
	if defWorkers != nil{
		defWorkers.Stop()
	}
}