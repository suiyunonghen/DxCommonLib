package goid

import (
	"reflect"
	"unsafe"
)

const GoId = "goid"

type eface struct {
	typ, val unsafe.Pointer
}

//go:nosplit
func runtime_convT2E_hack(_type, elem unsafe.Pointer) eface {
	return eface{
		typ: _type,
		val:  elem,
	}
}

var offset int64

func init() {
	offset = func() int64 {
		g := getG()
		if f, ok := reflect.TypeOf(g).FieldByName(GoId); ok {
			return int64(f.Offset)
		}
		panic("can not find g.goid field")
	}()

}

func getG() interface{}

// GetGoID  return current goroutine id.
func GetGoID() int64
