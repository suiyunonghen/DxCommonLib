package system

import (
	"github.com/suiyunonghen/DxCommonLib"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

func getg() unsafe.Pointer

var (
	goroutinePrefix = "goroutine "
	gidPos          = getGidPos()
)

func slowID() int64 {
	buf := [32]byte{}

	// Parse the 4707 out of "goroutine 4707 ["
	str := strings.TrimPrefix(
		string(buf[:runtime.Stack(buf[:], false)]),
		goroutinePrefix,
	)

	if lastPos := strings.IndexByte(str, ' '); lastPos > 0 {
		if id, err := DxCommonLib.ParseInt64(str[:lastPos]); err == nil {
			return id
		}
	}
	return 0
}

var getGidPos = func() func() int {
	fnFindGidPos := func(start int) int {
		if currGID, ptr := slowID(), getg(); currGID > 0 && uintptr(ptr) > 0 {
			pos := start
			for pos < 4096 {
				if *(*int64)(unsafe.Pointer(uintptr(ptr) + uintptr(pos))) == currGID {
					return pos
				}
				pos++
			}
		}
		return -1
	}

	fnCheckPos := func(pos int) bool {
		var wg sync.WaitGroup
		checkCount := int32(32)
		wg.Add(int(checkCount))
		checkIndex := int32(0)
		for i := 0; i < int(checkCount); i++ {
			DxCommonLib.PostFunc(func(data ...interface{}) {
				if fnFindGidPos(pos) == pos {
					atomic.AddInt32(&checkIndex,1)
				}
				wg.Done()
			})
		}
		wg.Wait()
		return atomic.LoadInt32(&checkIndex) == checkCount
	}

	return func() int {
		pos := fnFindGidPos(0)
		for pos != -1 {
			if fnCheckPos(pos) {
				break
			} else {
				pos = fnFindGidPos(pos + 1)
			}
		}
		return pos
	}
}()

// GetRoutineId get current goroutine id
func GetRoutineId() int64 {
	if gidPos >= 0 {
		return *(*int64)(unsafe.Pointer(uintptr(getg()) + uintptr(gidPos)))
	}

	return slowID()
}
