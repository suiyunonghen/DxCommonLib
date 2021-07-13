package system

import "os"

func RedirectPanic(f *os.File)error  {
	return redirectStdErr(f)
}
