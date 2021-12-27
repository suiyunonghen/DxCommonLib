/*
  反射相关的处理
  Autor: 不得闲
  QQ:75492895
*/

package DxCommonLib

import (
	"reflect"
	"unsafe"
)

type tflag uint8
type nameOff int32 // offset to a name
type typeOff int32 // offset to an *rtype
type textOff int32 // offset from top of text section
type rtype struct {
	size       uintptr
	ptrdata    uintptr // number of bytes in the type that can contain pointers
	hash       uint32  // hash of type; avoids computation in hash tables
	tflag      tflag   // extra type information flags
	align      uint8   // alignment of variable with this type
	fieldAlign uint8   // alignment of struct field with this type
	kind       uint8   // enumeration for C
	// function for comparing objects of this type
	// (ptr to object A, ptr to object B) -> ==?
	equal     func(unsafe.Pointer, unsafe.Pointer) bool
	gcdata    *byte   // garbage collection data
	str       nameOff // string form
	ptrToThis typeOff // type for pointer to this type, may be zero
}

const (
	kindDirectIface = 1 << 5
	kindGCProg      = 1 << 6 // Type.gc points to GC program
	kindMask        = (1 << 5) - 1
)

func (t *rtype) Kind() reflect.Kind { return reflect.Kind(t.kind & kindMask) }

type MapType struct {
	rtype
	key    *rtype // map key type
	elem   *rtype // map element (value) type
	bucket *rtype // internal bucket structure
	// function for hashing keys (ptr to key, seed) -> hash
	hasher     func(unsafe.Pointer, uintptr) uintptr
	keysize    uint8  // size of key slot
	valuesize  uint8  // size of value slot
	bucketsize uint16 // size of bucket
	flags      uint32
}

func (mapType *MapType)KeyKind()reflect.Kind  {
	return mapType.key.Kind()
}

func (mapType *MapType)ValueKind()reflect.Kind  {
	return mapType.elem.Kind()
}

func (mapType *MapType)ValueSliceType()*SliceType  {
	if mapType.elem.Kind() == reflect.Slice{
		return (*SliceType)(unsafe.Pointer(mapType.elem))
	}
	return nil
}

func (mapType *MapType)ValueMapType()*MapType  {
	if mapType.elem.Kind() == reflect.Map{
		return (*MapType)(unsafe.Pointer(mapType.elem))
	}
	return nil
}


func (mapType *MapType)CreateValueMap()*reflect.Value  {
	if mapType.elem.Kind() == reflect.Map{
		valueType := (*MapType)(unsafe.Pointer(mapType.elem))
		switch valueType.KeyKind(){
		case reflect.String:
			return _createStringMap(valueType.ValueKind())
		case reflect.Int:
			return _createIntMap(valueType.ValueKind())
		case reflect.Int16:
			return _createInt16Map(valueType.ValueKind())
		case reflect.Int8:
			return _createInt8Map(valueType.ValueKind())
		case reflect.Int32:
			return _createInt32Map(valueType.ValueKind())
		case reflect.Int64:
			return _createInt64Map(valueType.ValueKind())
		case reflect.Uint:
			return _createUintMap(valueType.ValueKind())
		case reflect.Uint16:
			return _createUint16Map(valueType.ValueKind())
		case reflect.Uint8:
			return _createUint8Map(valueType.ValueKind())
		case reflect.Uint32:
			return _createUint32Map(valueType.ValueKind())
		case reflect.Uint64:
			return _createUint64Map(valueType.ValueKind())
		case reflect.Interface:
			return _createInterfaceMap(valueType.ValueKind())
		case reflect.Uintptr:
			return _createUintptrMap(valueType.ValueKind())
		}
	}
	return nil
}

func _createUintptrMap(valueKind reflect.Kind)*reflect.Value  {
	switch valueKind {
	case reflect.String:
		vMap := reflect.ValueOf(make(map[uintptr]string,8))
		return &vMap
	case reflect.Interface:
		vMap := reflect.ValueOf(make(map[uintptr]interface{},8))
		return &vMap
	case reflect.Int:
		vMap := reflect.ValueOf(make(map[uintptr]int,8))
		return &vMap
	case reflect.Int8:
		vMap := reflect.ValueOf(make(map[uintptr]int8,8))
		return &vMap
	case reflect.Int16:
		vMap := reflect.ValueOf(make(map[uintptr]int16,8))
		return &vMap
	case reflect.Int32:
		vMap := reflect.ValueOf(make(map[uintptr]int32,8))
		return &vMap
	case reflect.Int64:
		vMap := reflect.ValueOf(make(map[uintptr]int64,8))
		return &vMap
	case reflect.Bool:
		vMap := reflect.ValueOf(make(map[uintptr]bool,8))
		return &vMap
	case reflect.Float64:
		vMap := reflect.ValueOf(make(map[uintptr]float64,8))
		return &vMap
	case reflect.Float32:
		vMap := reflect.ValueOf(make(map[uintptr]float32,8))
		return &vMap
	case reflect.Uint:
		vMap := reflect.ValueOf(make(map[uintptr]uint,8))
		return &vMap
	case reflect.Uint8:
		vMap := reflect.ValueOf(make(map[uintptr]uint8,8))
		return &vMap
	case reflect.Uint16:
		vMap := reflect.ValueOf(make(map[uintptr]uint16,8))
		return &vMap
	case reflect.Uint32:
		vMap := reflect.ValueOf(make(map[uintptr]uint32,8))
		return &vMap
	case reflect.Uint64:
		vMap := reflect.ValueOf(make(map[uintptr]uint64,8))
		return &vMap
	case reflect.Uintptr:
		vMap := reflect.ValueOf(make(map[uintptr]uintptr,8))
		return &vMap
	}
	return nil
}

func _createInterfaceMap(valueKind reflect.Kind)*reflect.Value  {
	switch valueKind {
	case reflect.String:
		vMap := reflect.ValueOf(make(map[interface{}]string,8))
		return &vMap
	case reflect.Interface:
		vMap := reflect.ValueOf(make(map[interface{}]interface{},8))
		return &vMap
	case reflect.Int:
		vMap := reflect.ValueOf(make(map[interface{}]int,8))
		return &vMap
	case reflect.Int8:
		vMap := reflect.ValueOf(make(map[interface{}]int8,8))
		return &vMap
	case reflect.Int16:
		vMap := reflect.ValueOf(make(map[interface{}]int16,8))
		return &vMap
	case reflect.Int32:
		vMap := reflect.ValueOf(make(map[interface{}]int32,8))
		return &vMap
	case reflect.Int64:
		vMap := reflect.ValueOf(make(map[interface{}]int64,8))
		return &vMap
	case reflect.Bool:
		vMap := reflect.ValueOf(make(map[interface{}]bool,8))
		return &vMap
	case reflect.Float64:
		vMap := reflect.ValueOf(make(map[interface{}]float64,8))
		return &vMap
	case reflect.Float32:
		vMap := reflect.ValueOf(make(map[interface{}]float32,8))
		return &vMap
	case reflect.Uint:
		vMap := reflect.ValueOf(make(map[interface{}]uint,8))
		return &vMap
	case reflect.Uint8:
		vMap := reflect.ValueOf(make(map[interface{}]uint8,8))
		return &vMap
	case reflect.Uint16:
		vMap := reflect.ValueOf(make(map[interface{}]uint16,8))
		return &vMap
	case reflect.Uint32:
		vMap := reflect.ValueOf(make(map[interface{}]uint32,8))
		return &vMap
	case reflect.Uint64:
		vMap := reflect.ValueOf(make(map[interface{}]uint64,8))
		return &vMap
	case reflect.Uintptr:
		vMap := reflect.ValueOf(make(map[interface{}]uintptr,8))
		return &vMap
	}
	return nil
}

func _createStringMap(valueKind reflect.Kind)*reflect.Value  {
	switch valueKind {
	case reflect.String:
		vMap := reflect.ValueOf(make(map[string]string,8))
		return &vMap
	case reflect.Interface:
		vMap := reflect.ValueOf(make(map[string]interface{},8))
		return &vMap
	case reflect.Int:
		vMap := reflect.ValueOf(make(map[string]int,8))
		return &vMap
	case reflect.Int8:
		vMap := reflect.ValueOf(make(map[string]int8,8))
		return &vMap
	case reflect.Int16:
		vMap := reflect.ValueOf(make(map[string]int16,8))
		return &vMap
	case reflect.Int32:
		vMap := reflect.ValueOf(make(map[string]int32,8))
		return &vMap
	case reflect.Int64:
		vMap := reflect.ValueOf(make(map[string]int64,8))
		return &vMap
	case reflect.Bool:
		vMap := reflect.ValueOf(make(map[string]bool,8))
		return &vMap
	case reflect.Float64:
		vMap := reflect.ValueOf(make(map[string]float64,8))
		return &vMap
	case reflect.Float32:
		vMap := reflect.ValueOf(make(map[string]float32,8))
		return &vMap
	case reflect.Uint:
		vMap := reflect.ValueOf(make(map[string]uint,8))
		return &vMap
	case reflect.Uint8:
		vMap := reflect.ValueOf(make(map[string]uint8,8))
		return &vMap
	case reflect.Uint16:
		vMap := reflect.ValueOf(make(map[string]uint16,8))
		return &vMap
	case reflect.Uint32:
		vMap := reflect.ValueOf(make(map[string]uint32,8))
		return &vMap
	case reflect.Uint64:
		vMap := reflect.ValueOf(make(map[string]uint64,8))
		return &vMap
	case reflect.Uintptr:
		vMap := reflect.ValueOf(make(map[string]uintptr,8))
		return &vMap
	}
	return nil
}

func _createInt8Map(valueKind reflect.Kind)*reflect.Value  {
	switch valueKind {
	case reflect.String:
		vMap := reflect.ValueOf(make(map[int8]string,8))
		return &vMap
	case reflect.Interface:
		vMap := reflect.ValueOf(make(map[int8]interface{},8))
		return &vMap
	case reflect.Int:
		vMap := reflect.ValueOf(make(map[int8]int,8))
		return &vMap
	case reflect.Int8:
		vMap := reflect.ValueOf(make(map[int8]int8,8))
		return &vMap
	case reflect.Int16:
		vMap := reflect.ValueOf(make(map[int8]int16,8))
		return &vMap
	case reflect.Int32:
		vMap := reflect.ValueOf(make(map[int8]int32,8))
		return &vMap
	case reflect.Int64:
		vMap := reflect.ValueOf(make(map[int8]int64,8))
		return &vMap
	case reflect.Bool:
		vMap := reflect.ValueOf(make(map[int8]bool,8))
		return &vMap
	case reflect.Float64:
		vMap := reflect.ValueOf(make(map[int8]float64,8))
		return &vMap
	case reflect.Float32:
		vMap := reflect.ValueOf(make(map[int8]float32,8))
		return &vMap
	case reflect.Uint:
		vMap := reflect.ValueOf(make(map[int8]uint,8))
		return &vMap
	case reflect.Uint8:
		vMap := reflect.ValueOf(make(map[int8]uint8,8))
		return &vMap
	case reflect.Uint16:
		vMap := reflect.ValueOf(make(map[int8]uint16,8))
		return &vMap
	case reflect.Uint32:
		vMap := reflect.ValueOf(make(map[int8]uint32,8))
		return &vMap
	case reflect.Uint64:
		vMap := reflect.ValueOf(make(map[int8]uint64,8))
		return &vMap
	case reflect.Uintptr:
		vMap := reflect.ValueOf(make(map[int8]uintptr,8))
		return &vMap
	}
	return nil
}

func _createInt32Map(valueKind reflect.Kind)*reflect.Value  {
	switch valueKind {
	case reflect.String:
		vMap := reflect.ValueOf(make(map[int32]string,8))
		return &vMap
	case reflect.Interface:
		vMap := reflect.ValueOf(make(map[int32]interface{},8))
		return &vMap
	case reflect.Int:
		vMap := reflect.ValueOf(make(map[int32]int,8))
		return &vMap
	case reflect.Int8:
		vMap := reflect.ValueOf(make(map[int32]int8,8))
		return &vMap
	case reflect.Int16:
		vMap := reflect.ValueOf(make(map[int32]int16,8))
		return &vMap
	case reflect.Int32:
		vMap := reflect.ValueOf(make(map[int32]int32,8))
		return &vMap
	case reflect.Int64:
		vMap := reflect.ValueOf(make(map[int32]int64,8))
		return &vMap
	case reflect.Bool:
		vMap := reflect.ValueOf(make(map[int32]bool,8))
		return &vMap
	case reflect.Float64:
		vMap := reflect.ValueOf(make(map[int32]float64,8))
		return &vMap
	case reflect.Float32:
		vMap := reflect.ValueOf(make(map[int32]float32,8))
		return &vMap
	case reflect.Uint:
		vMap := reflect.ValueOf(make(map[int32]uint,8))
		return &vMap
	case reflect.Uint8:
		vMap := reflect.ValueOf(make(map[int32]uint8,8))
		return &vMap
	case reflect.Uint16:
		vMap := reflect.ValueOf(make(map[int32]uint16,8))
		return &vMap
	case reflect.Uint32:
		vMap := reflect.ValueOf(make(map[int32]uint32,8))
		return &vMap
	case reflect.Uint64:
		vMap := reflect.ValueOf(make(map[int32]uint64,8))
		return &vMap
	case reflect.Uintptr:
		vMap := reflect.ValueOf(make(map[int32]uintptr,8))
		return &vMap
	}
	return nil
}

func _createInt64Map(valueKind reflect.Kind)*reflect.Value  {
	switch valueKind {
	case reflect.String:
		vMap := reflect.ValueOf(make(map[int64]string,8))
		return &vMap
	case reflect.Interface:
		vMap := reflect.ValueOf(make(map[int64]interface{},8))
		return &vMap
	case reflect.Int:
		vMap := reflect.ValueOf(make(map[int64]int,8))
		return &vMap
	case reflect.Int8:
		vMap := reflect.ValueOf(make(map[int64]int8,8))
		return &vMap
	case reflect.Int16:
		vMap := reflect.ValueOf(make(map[int64]int16,8))
		return &vMap
	case reflect.Int32:
		vMap := reflect.ValueOf(make(map[int64]int32,8))
		return &vMap
	case reflect.Int64:
		vMap := reflect.ValueOf(make(map[int64]int64,8))
		return &vMap
	case reflect.Bool:
		vMap := reflect.ValueOf(make(map[int64]bool,8))
		return &vMap
	case reflect.Float64:
		vMap := reflect.ValueOf(make(map[int64]float64,8))
		return &vMap
	case reflect.Float32:
		vMap := reflect.ValueOf(make(map[int64]float32,8))
		return &vMap
	case reflect.Uint:
		vMap := reflect.ValueOf(make(map[int64]uint,8))
		return &vMap
	case reflect.Uint8:
		vMap := reflect.ValueOf(make(map[int64]uint8,8))
		return &vMap
	case reflect.Uint16:
		vMap := reflect.ValueOf(make(map[int64]uint16,8))
		return &vMap
	case reflect.Uint32:
		vMap := reflect.ValueOf(make(map[int64]uint32,8))
		return &vMap
	case reflect.Uint64:
		vMap := reflect.ValueOf(make(map[int64]uint64,8))
		return &vMap
	case reflect.Uintptr:
		vMap := reflect.ValueOf(make(map[int64]uintptr,8))
		return &vMap
	}
	return nil
}

func _createIntMap(valueKind reflect.Kind)*reflect.Value  {
	switch valueKind {
	case reflect.String:
		vMap := reflect.ValueOf(make(map[int]string,8))
		return &vMap
	case reflect.Interface:
		vMap := reflect.ValueOf(make(map[int]interface{},8))
		return &vMap
	case reflect.Int:
		vMap := reflect.ValueOf(make(map[int]int,8))
		return &vMap
	case reflect.Int8:
		vMap := reflect.ValueOf(make(map[int]int8,8))
		return &vMap
	case reflect.Int16:
		vMap := reflect.ValueOf(make(map[int]int16,8))
		return &vMap
	case reflect.Int32:
		vMap := reflect.ValueOf(make(map[int]int32,8))
		return &vMap
	case reflect.Int64:
		vMap := reflect.ValueOf(make(map[int]int64,8))
		return &vMap
	case reflect.Bool:
		vMap := reflect.ValueOf(make(map[int]bool,8))
		return &vMap
	case reflect.Float64:
		vMap := reflect.ValueOf(make(map[int]float64,8))
		return &vMap
	case reflect.Float32:
		vMap := reflect.ValueOf(make(map[int]float32,8))
		return &vMap
	case reflect.Uint:
		vMap := reflect.ValueOf(make(map[int]uint,8))
		return &vMap
	case reflect.Uint8:
		vMap := reflect.ValueOf(make(map[int]uint8,8))
		return &vMap
	case reflect.Uint16:
		vMap := reflect.ValueOf(make(map[int]uint16,8))
		return &vMap
	case reflect.Uint32:
		vMap := reflect.ValueOf(make(map[int]uint32,8))
		return &vMap
	case reflect.Uint64:
		vMap := reflect.ValueOf(make(map[int]uint64,8))
		return &vMap
	case reflect.Uintptr:
		vMap := reflect.ValueOf(make(map[int]uintptr,8))
		return &vMap
	}
	return nil
}

func _createInt16Map(valueKind reflect.Kind)*reflect.Value  {
	switch valueKind {
	case reflect.String:
		vMap := reflect.ValueOf(make(map[int16]string,8))
		return &vMap
	case reflect.Interface:
		vMap := reflect.ValueOf(make(map[int16]interface{},8))
		return &vMap
	case reflect.Int:
		vMap := reflect.ValueOf(make(map[int16]int,8))
		return &vMap
	case reflect.Int8:
		vMap := reflect.ValueOf(make(map[int16]int8,8))
		return &vMap
	case reflect.Int16:
		vMap := reflect.ValueOf(make(map[int16]int16,8))
		return &vMap
	case reflect.Int32:
		vMap := reflect.ValueOf(make(map[int16]int32,8))
		return &vMap
	case reflect.Int64:
		vMap := reflect.ValueOf(make(map[int16]int64,8))
		return &vMap
	case reflect.Bool:
		vMap := reflect.ValueOf(make(map[int16]bool,8))
		return &vMap
	case reflect.Float64:
		vMap := reflect.ValueOf(make(map[int16]float64,8))
		return &vMap
	case reflect.Float32:
		vMap := reflect.ValueOf(make(map[int16]float32,8))
		return &vMap
	case reflect.Uint:
		vMap := reflect.ValueOf(make(map[int16]uint,8))
		return &vMap
	case reflect.Uint8:
		vMap := reflect.ValueOf(make(map[int16]uint8,8))
		return &vMap
	case reflect.Uint16:
		vMap := reflect.ValueOf(make(map[int16]uint16,8))
		return &vMap
	case reflect.Uint32:
		vMap := reflect.ValueOf(make(map[int16]uint32,8))
		return &vMap
	case reflect.Uint64:
		vMap := reflect.ValueOf(make(map[int16]uint64,8))
		return &vMap
	case reflect.Uintptr:
		vMap := reflect.ValueOf(make(map[int16]uintptr,8))
		return &vMap
	}
	return nil
}


func _createUint8Map(valueKind reflect.Kind)*reflect.Value  {
	switch valueKind {
	case reflect.String:
		vMap := reflect.ValueOf(make(map[uint8]string,8))
		return &vMap
	case reflect.Interface:
		vMap := reflect.ValueOf(make(map[uint8]interface{},8))
		return &vMap
	case reflect.Int:
		vMap := reflect.ValueOf(make(map[uint8]int,8))
		return &vMap
	case reflect.Int8:
		vMap := reflect.ValueOf(make(map[uint8]int8,8))
		return &vMap
	case reflect.Int16:
		vMap := reflect.ValueOf(make(map[uint8]int16,8))
		return &vMap
	case reflect.Int32:
		vMap := reflect.ValueOf(make(map[uint8]int32,8))
		return &vMap
	case reflect.Int64:
		vMap := reflect.ValueOf(make(map[uint8]int64,8))
		return &vMap
	case reflect.Bool:
		vMap := reflect.ValueOf(make(map[uint8]bool,8))
		return &vMap
	case reflect.Float64:
		vMap := reflect.ValueOf(make(map[uint8]float64,8))
		return &vMap
	case reflect.Float32:
		vMap := reflect.ValueOf(make(map[uint8]float32,8))
		return &vMap
	case reflect.Uint:
		vMap := reflect.ValueOf(make(map[uint8]uint,8))
		return &vMap
	case reflect.Uint8:
		vMap := reflect.ValueOf(make(map[uint8]uint8,8))
		return &vMap
	case reflect.Uint16:
		vMap := reflect.ValueOf(make(map[uint8]uint16,8))
		return &vMap
	case reflect.Uint32:
		vMap := reflect.ValueOf(make(map[uint8]uint32,8))
		return &vMap
	case reflect.Uint64:
		vMap := reflect.ValueOf(make(map[uint8]uint64,8))
		return &vMap
	case reflect.Uintptr:
		vMap := reflect.ValueOf(make(map[uint8]uintptr,8))
		return &vMap
	}
	return nil
}

func _createUint32Map(valueKind reflect.Kind)*reflect.Value  {
	switch valueKind {
	case reflect.String:
		vMap := reflect.ValueOf(make(map[uint32]string,8))
		return &vMap
	case reflect.Interface:
		vMap := reflect.ValueOf(make(map[uint32]interface{},8))
		return &vMap
	case reflect.Int:
		vMap := reflect.ValueOf(make(map[uint32]int,8))
		return &vMap
	case reflect.Int8:
		vMap := reflect.ValueOf(make(map[uint32]int8,8))
		return &vMap
	case reflect.Int16:
		vMap := reflect.ValueOf(make(map[uint32]int16,8))
		return &vMap
	case reflect.Int32:
		vMap := reflect.ValueOf(make(map[uint32]int32,8))
		return &vMap
	case reflect.Int64:
		vMap := reflect.ValueOf(make(map[uint32]int64,8))
		return &vMap
	case reflect.Bool:
		vMap := reflect.ValueOf(make(map[uint32]bool,8))
		return &vMap
	case reflect.Float64:
		vMap := reflect.ValueOf(make(map[uint32]float64,8))
		return &vMap
	case reflect.Float32:
		vMap := reflect.ValueOf(make(map[uint32]float32,8))
		return &vMap
	case reflect.Uint:
		vMap := reflect.ValueOf(make(map[uint32]uint,8))
		return &vMap
	case reflect.Uint8:
		vMap := reflect.ValueOf(make(map[uint32]uint8,8))
		return &vMap
	case reflect.Uint16:
		vMap := reflect.ValueOf(make(map[uint32]uint16,8))
		return &vMap
	case reflect.Uint32:
		vMap := reflect.ValueOf(make(map[uint32]uint32,8))
		return &vMap
	case reflect.Uint64:
		vMap := reflect.ValueOf(make(map[uint32]uint64,8))
		return &vMap
	case reflect.Uintptr:
		vMap := reflect.ValueOf(make(map[uint32]uintptr,8))
		return &vMap
	}
	return nil
}

func _createUint64Map(valueKind reflect.Kind)*reflect.Value  {
	switch valueKind {
	case reflect.String:
		vMap := reflect.ValueOf(make(map[uint64]string,8))
		return &vMap
	case reflect.Interface:
		vMap := reflect.ValueOf(make(map[uint64]interface{},8))
		return &vMap
	case reflect.Int:
		vMap := reflect.ValueOf(make(map[uint64]int,8))
		return &vMap
	case reflect.Int8:
		vMap := reflect.ValueOf(make(map[uint64]int8,8))
		return &vMap
	case reflect.Int16:
		vMap := reflect.ValueOf(make(map[uint64]int16,8))
		return &vMap
	case reflect.Int32:
		vMap := reflect.ValueOf(make(map[uint64]int32,8))
		return &vMap
	case reflect.Int64:
		vMap := reflect.ValueOf(make(map[uint64]int64,8))
		return &vMap
	case reflect.Bool:
		vMap := reflect.ValueOf(make(map[uint64]bool,8))
		return &vMap
	case reflect.Float64:
		vMap := reflect.ValueOf(make(map[uint64]float64,8))
		return &vMap
	case reflect.Float32:
		vMap := reflect.ValueOf(make(map[uint64]float32,8))
		return &vMap
	case reflect.Uint:
		vMap := reflect.ValueOf(make(map[uint64]uint,8))
		return &vMap
	case reflect.Uint8:
		vMap := reflect.ValueOf(make(map[uint64]uint8,8))
		return &vMap
	case reflect.Uint16:
		vMap := reflect.ValueOf(make(map[uint64]uint16,8))
		return &vMap
	case reflect.Uint32:
		vMap := reflect.ValueOf(make(map[uint64]uint32,8))
		return &vMap
	case reflect.Uint64:
		vMap := reflect.ValueOf(make(map[uint64]uint64,8))
		return &vMap
	case reflect.Uintptr:
		vMap := reflect.ValueOf(make(map[uint64]uintptr,8))
		return &vMap
	}
	return nil
}

func _createUintMap(valueKind reflect.Kind)*reflect.Value  {
	switch valueKind {
	case reflect.String:
		vMap := reflect.ValueOf(make(map[uint]string,8))
		return &vMap
	case reflect.Interface:
		vMap := reflect.ValueOf(make(map[uint]interface{},8))
		return &vMap
	case reflect.Int:
		vMap := reflect.ValueOf(make(map[uint]int,8))
		return &vMap
	case reflect.Int8:
		vMap := reflect.ValueOf(make(map[uint]int8,8))
		return &vMap
	case reflect.Int16:
		vMap := reflect.ValueOf(make(map[uint]int16,8))
		return &vMap
	case reflect.Int32:
		vMap := reflect.ValueOf(make(map[uint]int32,8))
		return &vMap
	case reflect.Int64:
		vMap := reflect.ValueOf(make(map[uint]int64,8))
		return &vMap
	case reflect.Bool:
		vMap := reflect.ValueOf(make(map[uint]bool,8))
		return &vMap
	case reflect.Float64:
		vMap := reflect.ValueOf(make(map[uint]float64,8))
		return &vMap
	case reflect.Float32:
		vMap := reflect.ValueOf(make(map[uint]float32,8))
		return &vMap
	case reflect.Uint:
		vMap := reflect.ValueOf(make(map[uint]uint,8))
		return &vMap
	case reflect.Uint8:
		vMap := reflect.ValueOf(make(map[uint]uint8,8))
		return &vMap
	case reflect.Uint16:
		vMap := reflect.ValueOf(make(map[uint]uint16,8))
		return &vMap
	case reflect.Uint32:
		vMap := reflect.ValueOf(make(map[uint]uint32,8))
		return &vMap
	case reflect.Uint64:
		vMap := reflect.ValueOf(make(map[uint]uint64,8))
		return &vMap
	case reflect.Uintptr:
		vMap := reflect.ValueOf(make(map[uint]uintptr,8))
		return &vMap
	}
	return nil
}

func _createUint16Map(valueKind reflect.Kind)*reflect.Value  {
	switch valueKind {
	case reflect.String:
		vMap := reflect.ValueOf(make(map[uint16]string,8))
		return &vMap
	case reflect.Interface:
		vMap := reflect.ValueOf(make(map[uint16]interface{},8))
		return &vMap
	case reflect.Int:
		vMap := reflect.ValueOf(make(map[uint16]int,8))
		return &vMap
	case reflect.Int8:
		vMap := reflect.ValueOf(make(map[uint16]int8,8))
		return &vMap
	case reflect.Int16:
		vMap := reflect.ValueOf(make(map[uint16]int16,8))
		return &vMap
	case reflect.Int32:
		vMap := reflect.ValueOf(make(map[uint16]int32,8))
		return &vMap
	case reflect.Int64:
		vMap := reflect.ValueOf(make(map[uint16]int64,8))
		return &vMap
	case reflect.Bool:
		vMap := reflect.ValueOf(make(map[uint16]bool,8))
		return &vMap
	case reflect.Float64:
		vMap := reflect.ValueOf(make(map[uint16]float64,8))
		return &vMap
	case reflect.Float32:
		vMap := reflect.ValueOf(make(map[uint16]float32,8))
		return &vMap
	case reflect.Uint:
		vMap := reflect.ValueOf(make(map[uint16]uint,8))
		return &vMap
	case reflect.Uint8:
		vMap := reflect.ValueOf(make(map[uint16]uint8,8))
		return &vMap
	case reflect.Uint16:
		vMap := reflect.ValueOf(make(map[uint16]uint16,8))
		return &vMap
	case reflect.Uint32:
		vMap := reflect.ValueOf(make(map[uint16]uint32,8))
		return &vMap
	case reflect.Uint64:
		vMap := reflect.ValueOf(make(map[uint16]uint64,8))
		return &vMap
	case reflect.Uintptr:
		vMap := reflect.ValueOf(make(map[uint16]uintptr,8))
		return &vMap
	}
	return nil
}

type SliceType struct {
	rtype
	elem *rtype // slice element type
}

func (stype *SliceType)ValueKind()reflect.Kind  {
	return stype.elem.Kind()
}

func (stype *SliceType)SliceValueType()*SliceType  {
	if stype.elem.Kind() == reflect.Slice{
		return (*SliceType)(unsafe.Pointer(stype.elem))
	}
	return nil
}

func (stype *SliceType)MapValueType()*MapType  {
	if stype.elem.Kind() == reflect.Map{
		return (*MapType)(unsafe.Pointer(stype.elem))
	}
	return nil
}

/*func (stype *SliceType)StructValueType()*MapType  {
	if stype.elem.Kind() == reflect.Map{
		return (*MapType)(unsafe.Pointer(stype.elem))
	}
	return nil
}*/


type ptrType struct {
	rtype
	elem *rtype // pointer element (pointed at) type
}

type chanType struct {
	rtype
	elem *rtype  // channel element type
	dir  uintptr // channel direction (ChanDir)
}

/*type interfaceType struct {
	rtype
	pkgPath name      // import path
	methods []imethod // sorted by hash
}*/

/*
reflect.Value结构
*/
type reflectValue struct {
	// typ holds the type of the value represented by a Value.
	typ *rtype

	// Pointer-valued data or, if flagIndir is set, pointer to data.
	// Valid when either flagIndir is set or typ.pointers() is true.
	ptr unsafe.Pointer

	// flag holds metadata about the value.
	// The lowest bits are flag bits:
	//	- flagStickyRO: obtained via unexported not embedded field, so read-only
	//	- flagEmbedRO: obtained via unexported embedded field, so read-only
	//	- flagIndir: val holds a pointer to the data
	//	- flagAddr: v.CanAddr is true (implies flagIndir)
	//	- flagMethod: v is a method value.
	// The next five bits give the Kind of the value.
	// This repeats typ.Kind() except for method values.
	// The remaining 23+ bits give a method number for method values.
	// If flag.kind() != Func, code can assume that flagMethod is unset.
	// If ifaceIndir(typ), code can assume that flagIndir is set.
	uintptr

	// A method value represents a curried method invocation
	// like r.Read for some receiver r. The typ+val+flag bits describe
	// the receiver r, but the flag's Kind bits say Func (methods are
	// functions), and the top bits of the flag give the method number
	// in r's type's method table.
}

func MapKVType(maprefv *reflect.Value) *MapType{
	if maprefv.Kind() == reflect.Map {
		refv := (*reflectValue)(unsafe.Pointer(maprefv))
		return (*MapType)(unsafe.Pointer(refv.typ))
	}
	return nil
}


func IsSimpleCopyKind(kind reflect.Kind)bool  {
	switch kind {
	case reflect.String, reflect.Int, reflect.Int64, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint64, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Float32, reflect.Float64, reflect.Bool:
			return true
	default:
		return false
	}
}


func SliceValueType(sliceV *reflect.Value) *SliceType {
	if sliceV.Kind() == reflect.Slice{
		refv := (*reflectValue)(unsafe.Pointer(sliceV))
		return (*SliceType)(unsafe.Pointer(refv.typ))
	}
	return nil
}
