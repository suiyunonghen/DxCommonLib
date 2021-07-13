// +build windows

package system

import (
	"os"
	"syscall"
)

var(
	kernerl32 = syscall.MustLoadDLL("kernel32.dll")
	setHandleProc = kernerl32.MustFindProc("SetStdHandle")
)

func setStdHandle(stdHandle int32, handle syscall.Handle) error {
	r0, _, e1 := syscall.Syscall(setHandleProc.Addr(), 2, uintptr(stdHandle), uintptr(handle), 0)
	if r0 == 0 {
		if e1 != 0 {
			return error(e1)
		}
		return syscall.EINVAL
	}
	return nil
}


func redirectStdErr(f *os.File)error  {
	err := setStdHandle(syscall.STD_ERROR_HANDLE,syscall.Handle(f.Fd()))
	if err != nil{
		return err
	}
	os.Stderr = f
	return nil
}