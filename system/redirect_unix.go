// +build linux darwin dragonfly freebsd netbsd openbsd

package system

import (
	"os"
	"syscall"
)

func redirectStdErr(f *os.File) error {
	return syscall.Dup2(int(f.Fd()),int(os.Stderr.Fd()))
}
