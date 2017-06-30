/*
公用包
 */
package DxCommonLib

import (
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"bytes"
	"io/ioutil"
	"unsafe"
	"reflect"
)

func Ord(b bool)byte  {
	if b{
		return 1
	}
	return 0
}

func GBKString(str string)([]byte,error)  {
	reader := bytes.NewReader([]byte(str))
	O := transform.NewReader(reader, simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return nil,e
	}
	return d,nil
}

func GBK2Utf8(gbk []byte)([]byte,error){
	reader := bytes.NewReader(gbk)
	O := transform.NewReader(reader, simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return nil,e
	}
	return d,nil
}

func FastString2Byte(str string)(result []byte)  {
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&str))
	presult := (*reflect.SliceHeader)(unsafe.Pointer(&result))
	presult.Data = pstring.Data
	presult.Len = pstring.Len
	presult.Cap = pstring.Len + 10
	return
}

func FastByte2String(bt []byte)string  {
	return *(*string)(unsafe.Pointer(&bt))
}