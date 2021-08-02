package sync

import (
	"github.com/suiyunonghen/DxCommonLib"
	"github.com/suiyunonghen/DxCommonLib/system"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 256)
	},
}

type mutexStruct struct {
	locking *LockInfo		//当前锁定的信息
}

func (mutex mutexStruct)caller(offset int) (string, string) {
	//0是当前位置,1是上一个位置Lock,2是调用点位置
	pc, file, line, ok := runtime.Caller(2 + offset)
	if !ok {
		return "", ""
	}
	buffer := bufferPool.Get().([]byte)
	idx := strings.LastIndexByte(file, '/')
	if idx == -1 {
		buffer = append(buffer[:0], file...)
	} else {
		idx = strings.LastIndexByte(file[:idx], '/')
		if idx == -1 {
			buffer = append(buffer[:0], file...)
		} else {
			buffer = append(buffer[:0], file[idx+1:]...)
		}
	}
	funcName := runtime.FuncForPC(pc).Name()
	buffer = append(buffer, ':')
	buffer = strconv.AppendInt(buffer, int64(line), 10)
	result := string(buffer)
	buffer = buffer[:0]
	bufferPool.Put(buffer)
	idx = strings.IndexByte(funcName, '.')
	if idx > 0 {
		funcName = funcName[idx+1:]
	}
	return result, funcName
}

type LockStyle	uint8

const(
	LckUnLock	LockStyle = iota
	LckLockBlock					//LockWait
	LckLocking						//Locking
)

// 锁定信息

type LockInfo struct {
	IsRLock			bool			//是否是RLock
	LockStyle		LockStyle
	CheckIndex		byte			//检查了多少次
	StartTime		time.Time
	GoRoutine		int64
	LockMsg			string
	Caller			string			//调用Lock的位置
	CallerFunc		string			//调用Lock的函数
	BeforeCaller	string			//前一步调用
	BeforeFunc		string			//前一步调用
	Owner			*mutexStruct
}

var(
	lockPool = sync.Pool{
		New: func() interface{}{
			return &LockInfo{}
		},
	}
)

func freeLockInfo(lckInfo *LockInfo)  {
	lckInfo.GoRoutine = 0
	lckInfo.Owner.locking = nil
	lckInfo.Owner = nil
	lckInfo.Caller = ""
	lckInfo.LockMsg = ""
	lckInfo.CheckIndex = 0
	lckInfo.CallerFunc = ""
	lockPool.Put(lckInfo)
}

func (lckInfo *LockInfo)String()string {
	if lckInfo.IsRLock{
		if lckInfo.LockStyle == LckUnLock{
			return lckInfo.Caller+"."+lckInfo.CallerFunc+".RUnlock"
		}
		return lckInfo.Caller+"."+lckInfo.CallerFunc+".RLock"
	}
	if lckInfo.LockStyle == LckUnLock{
		return lckInfo.Caller+"."+lckInfo.CallerFunc+".UnLock"
	}
	return lckInfo.Caller+"."+lckInfo.CallerFunc+".Lock"
}

// 扩展支持检查死锁

type RWMutexEx struct {
	mutexStruct
	sync.RWMutex
}

func (mutex *RWMutexEx)Lock()  {
	mutex.LockWithMsg("")
}

func (mutex *RWMutexEx)LockWithMsg(lockMsg string)  {
	if atomic.LoadInt32(&deadCheck) == 0{
		mutex.RWMutex.Lock()
		return
	}
	//先获取当前的位置
	gid := system.GetRoutineId()
	caller,method := mutex.caller(0)
	before,beforeMethod := mutex.caller(1) //调用位置
	lckInfo := lockPool.Get().(*LockInfo)
	lckInfo.IsRLock = false
	lckInfo.LockStyle = LckLockBlock	//lockWait
	lckInfo.StartTime = time.Now()
	lckInfo.GoRoutine = gid
	lckInfo.LockMsg = lockMsg
	lckInfo.Caller = caller
	lckInfo.Owner = &mutex.mutexStruct
	lckInfo.CallerFunc = method
	lckInfo.BeforeFunc = beforeMethod
	lckInfo.BeforeCaller = before
	lockChan <- lckInfo
	mutex.RWMutex.Lock()
	lckInfo = lockPool.Get().(*LockInfo)
	lckInfo.IsRLock = false
	lckInfo.LockStyle = LckLocking
	lckInfo.StartTime = time.Now()
	lckInfo.GoRoutine = gid
	lckInfo.LockMsg = lockMsg
	lckInfo.Caller = caller
	lckInfo.Owner = &mutex.mutexStruct
	lckInfo.CallerFunc = method
	lckInfo.BeforeFunc = beforeMethod
	lckInfo.BeforeCaller = before
	lockChan <- lckInfo
}

func (mutex *RWMutexEx)RLock()  {
	mutex.RLockWithMsg("")
}

func (mutex *RWMutexEx)RLockWithMsg(lockMsg string)  {
	if atomic.LoadInt32(&deadCheck) == 0{
		mutex.RWMutex.RLock()
		return
	}
	gid := system.GetRoutineId()
	caller,method := mutex.caller(0)
	before,beforeMethod := mutex.caller(1) //调用位置
	lckInfo := lockPool.Get().(*LockInfo)
	lckInfo.IsRLock = true
	lckInfo.LockStyle = LckLockBlock	//lockWait
	lckInfo.StartTime = time.Now()
	lckInfo.GoRoutine = gid
	lckInfo.Caller = caller
	lckInfo.LockMsg = lockMsg
	lckInfo.Owner = &mutex.mutexStruct
	lckInfo.CallerFunc = method
	lckInfo.BeforeFunc = beforeMethod
	lckInfo.BeforeCaller = before
	lockChan <- lckInfo

	mutex.RWMutex.RLock()

	lckInfo = lockPool.Get().(*LockInfo)
	lckInfo.IsRLock = true
	lckInfo.LockStyle = LckLocking
	lckInfo.StartTime = time.Now()
	lckInfo.GoRoutine = gid
	lckInfo.LockMsg = lockMsg
	lckInfo.Caller = caller
	lckInfo.Owner = &mutex.mutexStruct
	lckInfo.CallerFunc = method
	lckInfo.BeforeFunc = beforeMethod
	lckInfo.BeforeCaller = before
	lockChan <- lckInfo
}

func (mutex *RWMutexEx)Unlock()  {
	if atomic.LoadInt32(&deadCheck) == 0{
		mutex.RWMutex.Unlock()
		return
	}
	gid := system.GetRoutineId()
	caller,method := mutex.caller(0)

	lckInfo := lockPool.Get().(*LockInfo)
	lckInfo.IsRLock = false
	lckInfo.LockStyle = LckUnLock
	lckInfo.StartTime = time.Now()
	lckInfo.GoRoutine = gid
	lckInfo.Caller = caller
	lckInfo.Owner = &mutex.mutexStruct
	lckInfo.CallerFunc = method
	lockChan <- lckInfo
	mutex.RWMutex.Unlock()
}

func (mutex *RWMutexEx)RUnlock()  {
	if atomic.LoadInt32(&deadCheck) == 0{
		mutex.RWMutex.RUnlock()
		return
	}
	gid := system.GetRoutineId()
	caller,method := mutex.caller(0)

	lckInfo := lockPool.Get().(*LockInfo)
	lckInfo.IsRLock = true
	lckInfo.LockStyle = LckUnLock
	lckInfo.StartTime = time.Now()
	lckInfo.GoRoutine = gid
	lckInfo.Caller = caller
	lckInfo.Owner = &mutex.mutexStruct
	lckInfo.CallerFunc = method
	lockChan <- lckInfo
	mutex.RWMutex.RUnlock()
}

type MutexEx struct {
	mutexStruct
	sync.Mutex
}

func (mutex *MutexEx)Lock()  {
	mutex.LockWithMsg("")
}

func (mutex *MutexEx)LockWithMsg(lockMsg string)  {
	if atomic.LoadInt32(&deadCheck) == 0{
		mutex.Mutex.Lock()
		return
	}
	//先获取当前的位置
	gid := system.GetRoutineId()
	caller,method := mutex.caller(0)
	before,beforeMethod := mutex.caller(1) //调用位置

	lckInfo := lockPool.Get().(*LockInfo)
	lckInfo.IsRLock = false
	lckInfo.LockStyle = LckLockBlock
	lckInfo.StartTime = time.Now()
	lckInfo.GoRoutine = gid
	lckInfo.Caller = caller
	lckInfo.LockMsg = lockMsg
	lckInfo.Owner = &mutex.mutexStruct
	lckInfo.CallerFunc = method
	lckInfo.BeforeFunc = beforeMethod
	lckInfo.BeforeCaller = before
	lockChan <- lckInfo

	mutex.Mutex.Lock()

	lckInfo = lockPool.Get().(*LockInfo)
	lckInfo.IsRLock = false
	lckInfo.LockStyle = LckLocking
	lckInfo.StartTime = time.Now()
	lckInfo.GoRoutine = gid
	lckInfo.LockMsg = lockMsg
	lckInfo.Caller = caller
	lckInfo.Owner = &mutex.mutexStruct
	lckInfo.CallerFunc = method
	lckInfo.BeforeFunc = beforeMethod
	lckInfo.BeforeCaller = before
	lockChan <- lckInfo
}

func (mutex *MutexEx)Unlock()  {
	if atomic.LoadInt32(&deadCheck) == 0{
		mutex.Mutex.Unlock()
		return
	}
	gid := system.GetRoutineId()
	caller,method := mutex.caller(0)

	lckInfo := lockPool.Get().(*LockInfo)
	lckInfo.IsRLock = false
	lckInfo.LockStyle = LckUnLock
	lckInfo.StartTime = time.Now()
	lckInfo.GoRoutine = gid
	lckInfo.Caller = caller
	lckInfo.Owner = &mutex.mutexStruct
	lckInfo.CallerFunc = method
	lockChan <- lckInfo
	mutex.Mutex.Unlock()
}

var lockChan chan  *LockInfo
var deadCheck int32
var quitDeadCheck chan struct{}


func StartDeadCheck(lockNotify func(deadLockInfo []byte),checkInterval,maxLockInterval time.Duration,debugLog func(panicMsg bool, format string,data ...interface{}))  {
	if debugLog != nil{
		debugLog(false,"启动DeadLock检查")
	}
	atomic.StoreInt32(&deadCheck,1)
	quitDeadCheck = make(chan struct{})
	go checkDeadLock(quitDeadCheck,checkInterval,maxLockInterval,lockNotify,debugLog)
}

func StopDeadCheck()  {
	atomic.StoreInt32(&deadCheck,0)
	if quitDeadCheck != nil{
		close(quitDeadCheck)
		quitDeadCheck = nil
	}
}

func init()  {
	lockChan = make(chan *LockInfo,256)
}

func checkDeadLock(quit chan struct{},checkInterval,maxLockInterval time.Duration,lockNotify func(deadLockInfo []byte),debugLog func(panicMsg bool,format string,data ...interface{}))  {
	var lockWaits sync.Map		//一般锁都会在全局用到，所以删除的概率比较小
	locking := make([]*LockInfo,0,1024)
	if checkInterval < time.Second * 5{
		checkInterval = time.Second * 5
	}
	buffer := make([]byte,0,2048)
	willPanic := false
	tk := DxCommonLib.After(checkInterval)
	for{
		select{
		case <-quit:
			return
		case lckInfo := <-lockChan:
			switch lckInfo.LockStyle {
			case LckLockBlock:
				var lockWait []*LockInfo
				if wait,ok := lockWaits.Load(lckInfo.Owner);!ok{
					lockWait = make([]*LockInfo,0,1024)
				}else{
					lockWait = wait.([]*LockInfo)
					for i := range lockWait{
						if lockWait[i].GoRoutine == lckInfo.GoRoutine{
							//同一个goroutine多次执行锁定
							if debugLog != nil{
								debugLog(true,"同一goroutine锁定1：%s,当前重入锁定2：%s",lockWait[i].String(),lckInfo.String())
							}
							return
						}
					}
				}
				lockWait = append(lockWait,lckInfo)
				lockWaits.Store(lckInfo.Owner,lockWait)
			case LckLocking:
				//将上次的LockWait的释放掉
				if lckInfo.Owner.locking != nil{
					if debugLog != nil{
						debugLog(true,"已经锁定：%s,当前重入锁定：%s",lckInfo.Owner.locking.String(),lckInfo.String())
						return
					}
				}
				if wait,ok := lockWaits.Load(lckInfo.Owner);ok{
					var lastLock *LockInfo
					lockWait := wait.([]*LockInfo)
					l := len(lockWait)
					for i := 0;i < l;i++{
						lastLock = lockWait[i]
						if lastLock.GoRoutine == lckInfo.GoRoutine{
							lockWait[i] = nil
							lockPool.Put(lastLock)
							if i == l - 1{
								lockWait = lockWait[:l-1]
							}else{
								lockWait = append(lockWait[:i],lockWait[i+1:]...)
							}
							lockWaits.Store(lckInfo.Owner,lockWait)
							break
						}
					}
					lckInfo.Owner.locking = lckInfo
					locking = append(locking,lckInfo)
				}else{
					freeLockInfo(lckInfo)
				}
			case LckUnLock:
				//释放Locking,一般Lock和UnLock成对出现,都是在同一个goroutine中，所以 goroutine是一致的
				l := len(locking)
				var foundLock *LockInfo
				for i := 0;i < l;i++{
					if locking[i].GoRoutine == lckInfo.GoRoutine && locking[i].Owner == lckInfo.Owner && locking[i].IsRLock == lckInfo.IsRLock{
						foundLock = locking[i]
						locking[i] = nil
						freeLockInfo(lckInfo)
						if i == l - 1{
							locking = locking[:l-1]
						}else{
							locking = append(locking[:i],locking[i+1:]...)
						}
						break
					}
				}
				if foundLock == nil && debugLog != nil{
					debugLog(true,"未找到在同一goroutine的成对释放%s",lckInfo.String())
				}
			}
		case <-tk:
			//检查一次锁信息
			buffer = buffer[:0]
			now := time.Now()
			for i := range locking{
				lockTimes := now.Sub(locking[i].StartTime)
				if lockTimes >= checkInterval{
					willPanic = lockTimes >= maxLockInterval
					locking[i].CheckIndex = locking[i].CheckIndex + 1
					buffer = append(buffer,"触发次数:"...)
					buffer = strconv.AppendInt(buffer,int64(locking[i].CheckIndex),10)
					buffer = append(buffer,",锁定时长:"...)
					buffer = append(buffer,lockTimes.String()...)
					buffer = append(buffer,",GoRoutine:"...)
					buffer = strconv.AppendInt(buffer,locking[i].GoRoutine,10)
					buffer = append(buffer,",锁定位置:"...)
					buffer = append(buffer,locking[i].Caller...)
					buffer = append(buffer,'.')
					buffer = append(buffer,locking[i].CallerFunc...)
					buffer = append(buffer," from "...)
					buffer = append(buffer,locking[i].BeforeCaller...)
					buffer = append(buffer,'.')
					buffer = append(buffer,locking[i].BeforeFunc...)

					buffer = append(buffer,",调用消息："...)
					buffer = append(buffer,locking[i].LockMsg...)
					if waits,ok := lockWaits.Load(locking[i].Owner);ok{
						buffer = append(buffer,"\r\n堵塞内容：\r\n"...)
						lockWait := waits.([]*LockInfo)
						for waitIndex := range lockWait{
							buffer = append(buffer,"堵塞时长："...)
							buffer = append(buffer,now.Sub(lockWait[waitIndex].StartTime).String()...)
							buffer = append(buffer,",GoRoutine:"...)
							buffer = strconv.AppendInt(buffer,lockWait[waitIndex].GoRoutine,10)
							buffer = append(buffer,",等待位置:"...)
							buffer = append(buffer,lockWait[waitIndex].Caller...)
							buffer = append(buffer,'.')
							buffer = append(buffer,lockWait[waitIndex].CallerFunc...)

							buffer = append(buffer," from "...)
							buffer = append(buffer,lockWait[waitIndex].BeforeCaller...)
							buffer = append(buffer,'.')
							buffer = append(buffer,lockWait[waitIndex].BeforeFunc...)

							buffer = append(buffer,",调用消息："...)
							buffer = append(buffer,lockWait[waitIndex].LockMsg...)
							buffer = append(buffer,[]byte{'\r','\n'}...)
						}
					}
				}
			}
			if willPanic{

				panic(DxCommonLib.FastByte2String(buffer))
			}
			if len(buffer) > 0{
				if lockNotify != nil{
					lockNotify(buffer)
				}else if debugLog != nil{
					debugLog(false,DxCommonLib.FastByte2String(buffer))
				}
			}
			tk = DxCommonLib.After(checkInterval)
		}
	}
}