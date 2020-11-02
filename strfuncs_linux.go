// +build linux darwin dragonfly freebsd netbsd openbsd
package DxCommonLib

import (
	"errors"
	"unicode/utf16"
)

var(
	EINVAL = errors.New("invalid argument")
)

func UTF16FromString(s string) ([]uint16, error) {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return nil, EINVAL
		}
	}
	return utf16.Encode([]rune(s + "\x00")), nil
}