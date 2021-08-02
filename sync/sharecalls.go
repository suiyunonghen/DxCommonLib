/**
并发共享函数调用返回的库
 */
package sync

import (
	"sync"
)

type funcCall struct {
	shareID		int
	shareKey	string
	callErr	error
	result	interface{}
	sync.WaitGroup
}

type ShareCalls struct {
	fCalls		[]*funcCall
	keyCalls	[]*funcCall
	sync.RWMutex
}

func (calls *ShareCalls)ShareCall(shareID int,fn func(...interface{})(interface{},error),params ...interface{})(result interface{},fromShareCache bool,err error){
	calls.Lock()
	for i := 0;i<len(calls.fCalls);i++{
		if calls.fCalls[i].shareID == shareID{
			c := calls.fCalls[i]
			calls.Unlock()
			c.Wait()
			return c.result, true,c.callErr
		}
	}
	if cap(calls.fCalls) == 0{
		calls.fCalls = make([]*funcCall,0,16)
	}
	call := &funcCall{
		shareID: shareID,
	}
	call.Add(1)
	calls.fCalls = append(calls.fCalls,call)
	calls.Unlock()

	//执行函数
	call.result,call.callErr = fn(params...)

	//执行完毕
	calls.Lock()
	for i := 0;i<len(calls.fCalls);i++{
		if calls.fCalls[i].shareID == shareID{
			calls.fCalls[i] = nil
			calls.fCalls = append(calls.fCalls[:i],calls.fCalls[i+1:]...)
			break
		}
	}
	calls.Unlock()

	call.Done()
	return call.result,false,call.callErr
}

func (calls *ShareCalls)ShareCallByKey(sharekey string,fn func(...interface{})(interface{},error),params ...interface{})(result interface{},fromShareCache bool,err error){
	calls.Lock()
	for i := 0;i<len(calls.keyCalls);i++{
		if calls.keyCalls[i].shareKey == sharekey{
			c := calls.keyCalls[i]
			calls.Unlock()
			c.Wait()
			return c.result, true,c.callErr
		}
	}
	if cap(calls.keyCalls) == 0{
		calls.keyCalls = make([]*funcCall,0,16)
	}
	call := &funcCall{
		shareKey: sharekey,
	}
	call.Add(1)
	calls.keyCalls = append(calls.keyCalls,call)
	calls.Unlock()
	//执行函数
	call.result,call.callErr = fn(params...)

	//执行完毕
	calls.Lock()
	for i := 0;i<len(calls.keyCalls);i++{
		if calls.keyCalls[i].shareKey == sharekey{
			calls.keyCalls[i] = nil
			calls.keyCalls = append(calls.keyCalls[:i],calls.keyCalls[i+1:]...)
			break
		}
	}
	calls.Unlock()

	call.Done()
	return call.result,false,call.callErr
}