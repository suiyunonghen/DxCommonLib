package sync

import (
	"bytes"
	"fmt"
	"github.com/suiyunonghen/DxCommonLib"
	"os"
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

type mutexStruct struct {}

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
	lockWaitId		int64
	StartTime		time.Time
	GoRoutine		uint64
	Caller			string			//调用Lock的位置
	CallerFunc		string			//调用Lock的函数
	Owner			*mutexStruct
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

func getGID() uint64 {
	b := make([]byte, 64)
	//runtime.Stack将调用其的go程的调用栈踪迹格式化后写入到buf中并返回写入的字节数。
	// 若all为true，函数会在写入当前go程的踪迹信息后，将其它所有go程的调用栈踪迹都格式化写入到buf中。
	b = b[:runtime.Stack(b, false)]
	// 去除前缀(注意此处goroutine后的空格)
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	// 截取ID
	b = b[:bytes.IndexByte(b, ' ')]
	// 类型转换
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

// 扩展支持检查死锁

type RWMutexEx struct {
	mutexStruct
	sync.RWMutex
}

func (mutex *RWMutexEx)Lock()  {
	if atomic.LoadInt32(&deadCheck) == 0{
		mutex.RWMutex.Lock()
		return
	}
	//先获取当前的位置
	lockWaitId := atomic.AddInt64(&lckId,1)
	gid := getGID()
	caller,method := mutex.caller(0)
	lockChan <- LockInfo{
		IsRLock: false,
		LockStyle: LckLockBlock,	//lockWait
		lockWaitId: lockWaitId,
		StartTime: time.Now(),
		GoRoutine: gid,
		Caller:	caller,
		Owner: &mutex.mutexStruct,
		CallerFunc: method,
	}
	mutex.RWMutex.Lock()
	lockChan <- LockInfo{
		IsRLock: false,
		LockStyle: LckLocking,
		lockWaitId: lockWaitId,
		StartTime: time.Now(),
		GoRoutine: gid,
		Caller:	caller,
		Owner: &mutex.mutexStruct,
		CallerFunc: method,
	}
}

func (mutex *RWMutexEx)RLock()  {
	if atomic.LoadInt32(&deadCheck) == 0{
		mutex.RWMutex.RLock()
		return
	}
	lockWaitId := atomic.AddInt64(&lckId,1)
	gid := getGID()
	caller,method := mutex.caller(0)
	lockChan <- LockInfo{
		IsRLock: true,
		LockStyle: LckLockBlock,	//lockWait
		lockWaitId: lockWaitId,
		StartTime: time.Now(),
		GoRoutine: gid,
		Caller:	caller,
		Owner: &mutex.mutexStruct,
		CallerFunc: method,
	}
	mutex.RWMutex.RLock()
	lockChan <- LockInfo{
		IsRLock: true,
		LockStyle: LckLocking,
		lockWaitId: -1,
		StartTime: time.Now(),
		GoRoutine: gid,
		Caller:	caller,
		Owner: &mutex.mutexStruct,
		CallerFunc: method,
	}
}

func (mutex *RWMutexEx)Unlock()  {
	if atomic.LoadInt32(&deadCheck) == 0{
		mutex.RWMutex.Unlock()
		return
	}
	gid := getGID()
	caller,method := mutex.caller(0)
	lockChan <- LockInfo{
		IsRLock: false,
		LockStyle: LckUnLock,
		lockWaitId: -1,
		StartTime: time.Now(),
		GoRoutine: gid,
		Caller:	caller,
		Owner: &mutex.mutexStruct,
		CallerFunc: method,
	}
	mutex.RWMutex.Unlock()
}

func (mutex *RWMutexEx)RUnlock()  {
	if atomic.LoadInt32(&deadCheck) == 0{
		mutex.RWMutex.RUnlock()
		return
	}
	gid := getGID()
	caller,method := mutex.caller(0)
	lockChan <- LockInfo{
		IsRLock: true,
		LockStyle: LckUnLock,
		lockWaitId: -1,
		StartTime: time.Now(),
		GoRoutine: gid,
		Caller:	caller,
		Owner: &mutex.mutexStruct,
		CallerFunc: method,
	}
	mutex.RWMutex.RUnlock()
}

type MutexEx struct {
	mutexStruct
	sync.Mutex
}

func (mutex *MutexEx)Lock()  {
	if atomic.LoadInt32(&deadCheck) == 0{
		mutex.Mutex.Lock()
		return
	}
	//先获取当前的位置
	lockWaitId := atomic.AddInt64(&lckId,1)
	gid := getGID()
	caller,method := mutex.caller(0)
	lockChan <- LockInfo{
		IsRLock: false,
		LockStyle: LckLockBlock,	//lockWait
		lockWaitId: lockWaitId,
		StartTime: time.Now(),
		GoRoutine: gid,
		Caller:	caller,
		Owner: &mutex.mutexStruct,
		CallerFunc: method,
	}
	mutex.Mutex.Lock()
	lockChan <- LockInfo{
		IsRLock: false,
		LockStyle: LckLocking,
		lockWaitId: lockWaitId,
		StartTime: time.Now(),
		GoRoutine: gid,
		Caller:	caller,
		Owner: &mutex.mutexStruct,
		CallerFunc: method,
	}
}

func (mutex *MutexEx)Unlock()  {
	if atomic.LoadInt32(&deadCheck) == 0{
		mutex.Mutex.Unlock()
		return
	}
	gid := getGID()
	caller,method := mutex.caller(0)
	lockChan <- LockInfo{
		IsRLock: false,
		LockStyle: LckUnLock,
		lockWaitId: -1,
		StartTime: time.Now(),
		GoRoutine: gid,
		Caller:	caller,
		Owner: &mutex.mutexStruct,
		CallerFunc: method,
	}
	mutex.Mutex.Unlock()
}


type DeadLockInfo struct {
	LockTimes	time.Duration
	LockInfo
}

func (lockInf *DeadLockInfo)String()string  {
	return fmt.Sprintf("GoRoutine=%d 锁定时长%s,死锁位置：%s %s",lockInf.GoRoutine, lockInf.LockTimes.String(), lockInf.Caller,lockInf.CallerFunc)
}

var lockChan chan  LockInfo
var lckId int64
var deadLockNotify func(deadLocks ...DeadLockInfo)
var debugLog func(format string,data ...interface{})
var deadCheck int32
var quitDeadCheck chan struct{}
func SetDeadCheck(deadChecked bool,lockNotify func(deadLocks ...DeadLockInfo),log func(format string,data ...interface{}))  {
	debugLog = log
	if deadChecked{
		if debugLog != nil{
			debugLog("启动DeadLock检查")
		}
		deadLockNotify = lockNotify
		atomic.StoreInt32(&deadCheck,1)
		quitDeadCheck = make(chan struct{})
		go checkDeadLock(quitDeadCheck)
	}else{
		if debugLog != nil {
			debugLog("关闭DeadLock检查")
		}
		atomic.StoreInt32(&deadCheck,0)
		if quitDeadCheck != nil{
			close(quitDeadCheck)
			quitDeadCheck = nil
		}
	}
}

func init()  {
	lockChan = make(chan LockInfo,256)
	SetDeadCheck(false,nil, func(format string, data ...interface{}) {
		fmt.Fprintf(os.Stdout,format,data...)
	})
}

func checkDeadLock(quit chan struct{})  {
	lockWait := make([]LockInfo,0,1024)
	locking := make([]LockInfo,0,1024)
	tk := DxCommonLib.After(time.Second * 30)
	for{
		select{
		case <-quit:
			return
		case lckInfo := <-lockChan:
			switch lckInfo.LockStyle {
			case LckLockBlock:
				lockWait = append(lockWait,lckInfo)
			case LckLocking:
				//将上次的LockWait的释放掉
				l := len(lockWait)
				for i := 0;i < l;i++{
					if lockWait[i].lockWaitId == lckInfo.lockWaitId{
						if i == l - 1{
							lockWait = lockWait[:l-1]
						}else{
							lockWait = append(lockWait[:i],lockWait[i+1:]...)
						}
						break
					}
				}
				locking = append(locking,lckInfo)
			case LckUnLock:
				//释放Locking,一般Lock和UnLock成对出现,都是在同一个goroutine中，所以 goroutine是一致的
				l := len(locking)
				var foundLock LockInfo
				for i := 0;i < l;i++{
					if locking[i].GoRoutine == lckInfo.GoRoutine && locking[i].Owner == lckInfo.Owner && locking[i].IsRLock == lckInfo.IsRLock{
						foundLock = locking[i]
						if i == l - 1{
							locking = locking[:l-1]
						}else{
							locking = append(locking[:i],locking[i+1:]...)
						}
						break
					}
				}
				DxCommonLib.PostFunc(func(data ...interface{}) {
					if foundLock.Owner == nil && debugLog != nil{
						debugLog("未找到在同一goroutine的成对释放%s",lckInfo.String())
					}
				})
			}
		case <-tk:
			//检查一次锁信息
			lockTimeouts := make([]DeadLockInfo,0,len(locking))
			now := time.Now()
			for i := range locking{
				lockTimes := now.Sub(locking[i].StartTime)
				if lockTimes >= time.Second * 30{
					locking[i].CheckIndex = locking[i].CheckIndex + 1
					lockTimeouts = append(lockTimeouts,DeadLockInfo{
						LockInfo: locking[i],
						LockTimes: lockTimes,
					})
				}
			}
			//显示一下其他锁等待的时长
			if len(lockTimeouts) > 0{
				if deadLockNotify != nil{
					DxCommonLib.PostFunc(func(data ...interface{}) {
						deadLockNotify(lockTimeouts...)
					})
				}else if debugLog != nil{
					DxCommonLib.PostFunc(func(data ...interface{}) {
						for i := range lockTimeouts{
							debugLog("可能已经死锁：%s\r\n",lockTimeouts[i].String())
						}
					})
				}
			}
			tk = DxCommonLib.After(time.Second * 30)
		}
	}
}