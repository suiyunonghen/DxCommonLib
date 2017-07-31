/*
GoRoutine池
Autor: 不得闲
QQ:75492895
*/
package DxCommonLib

import (
	"fmt"
	"sync"
	"time"
)

type (
	//任务
	ITaskRunner interface {
		Run()
	}

	GWorkerFunc func(data ...interface{})
	GWorkers    struct {
		workerPool chan *GWorker
		tasks      chan ITaskRunner
		fstopchan  chan struct{}
		maxcount   int
		mincount   int
		wg         sync.WaitGroup
	}

	GWorker struct {
		fOwner   *GWorkers
		fcurTask chan ITaskRunner
	}

	defTaskRunner struct {
		runfunc GWorkerFunc
		runargs []interface{}
	}
)

func (runner *defTaskRunner) Run() {
	runner.runfunc(runner.runargs...)
}

func (worker *GWorker) Execute() {
	//执行任务
	for {
		select {
		case task := <-worker.fcurTask:
			task.Run()
			worker.fOwner.wg.Done()
			//执行完成，回收
			select {
			case worker.fOwner.workerPool <- worker:
			default:
				return
			}
		case <-After(8 * time.Second): //8秒没有任务做，回收这个goroutine
			if len(worker.fOwner.workerPool) > worker.fOwner.mincount {
				return
			}
		case <-worker.fOwner.fstopchan:
			fmt.Println("exit worker execute")
			return
		}
	}
}

func (workers *GWorkers) Post(task ITaskRunner) {
	workers.wg.Add(1)
	workers.tasks <- task
}

func (workers *GWorkers) PostFunc(taskfunc GWorkerFunc, args ...interface{}) {
	defrunner := new(defTaskRunner)
	defrunner.runfunc = taskfunc
	defrunner.runargs = make([]interface{}, len(args))
	for idx, v := range args {
		defrunner.runargs[idx] = v
	}
	workers.Post(defrunner)
}

func (workers *GWorkers) Stop(waitAllOk bool) {
	if waitAllOk {
		workers.wg.Wait()
	}
	close(workers.fstopchan)
}

func NewWorkers(initworkercount, maxworkercount int) *GWorkers {
	result := new(GWorkers)
	result.workerPool = make(chan *GWorker, maxworkercount)
	result.tasks = make(chan ITaskRunner, maxworkercount)
	result.maxcount = maxworkercount
	result.mincount = initworkercount
	result.fstopchan = make(chan struct{})
	for i := 0; i < initworkercount; i++ {
		worker := new(GWorker)
		worker.fOwner = result
		worker.fcurTask = make(chan ITaskRunner)
		go worker.Execute()
		result.workerPool <- worker
	}
	go func(workers *GWorkers) {
		for {
			select {
			case task := <-workers.tasks:
				//然后查找一个有效的Worker
				select {
				case worker := <-workers.workerPool:
					worker.fcurTask <- task
				default:
					worker := new(GWorker)
					worker.fOwner = result
					worker.fcurTask = make(chan ITaskRunner)
					go worker.Execute()
					worker.fcurTask <- task
				}
			case <-workers.fstopchan:
				return
			}
		}
	}(result)
	return result
}
