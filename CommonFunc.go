/*
公用包
Autor: 不得闲
QQ:75492895
 */
package DxCommonLib

import (
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"bytes"
	"io/ioutil"
	"unsafe"
	"time"
	"unicode/utf16"
	"reflect"
)

type(
	TDateTime  float64
)

const (
	MinsPerHour = 60
	MinsPerDay = 24 * MinsPerHour
	SecsPerDay = MinsPerDay * 60
	MSecsPerDay = SecsPerDay * 1000
)
var delphiFirstTime time.Time

func init() {
	delphiFirstTime = time.Date(1899,12,30,0,0,0,0,time.Local)
}


/*
从Delphi日期转为Go日期格式
Delphi的日期规则为到1899-12-30号的天数+当前的毫秒数/一天的总共毫秒数集合
 */
func (date TDateTime)ToTime()time.Time  {
	mDay := time.Duration(date)
	ms := (date - TDateTime(mDay)) * TDateTime(MSecsPerDay)
	return delphiFirstTime.Add(mDay*time.Hour*24 + time.Duration(ms)*time.Millisecond)
}

func (date *TDateTime)WrapTime2Self(t time.Time)  {
	days := t.Sub(delphiFirstTime) / (time.Hour * 24)
	y,m,d := t.Date()
	nowdate := time.Date(y,m,d,0,0,0,0,time.Local)
	times := float64(t.Sub(nowdate))/float64(time.Hour*24)
	*date = TDateTime(float64(days) + times)
}

func Time2DelphiTime(t time.Time)TDateTime  {
	days := t.Sub(delphiFirstTime) / (time.Hour * 24)
	y,m,d := t.Date()
	nowdate := time.Date(y,m,d,0,0,0,0,time.Local)
	times := float64(t.Sub(nowdate))/float64(time.Hour*24)
	return TDateTime(float64(days) + times)
}

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

func PcharLen(dstr uintptr)int  {
	if dstr == 0{
		return 0
	}
	ptr := unsafe.Pointer(dstr)
	for i := 0; ; i++ {
		if 0 == *(*uint16)(ptr) {
			return int(i)
		}
		ptr = unsafe.Pointer(uintptr(ptr) + 2)
	}
	return 0
}

func DelphiPcharLen(dstr uintptr)(result int32)  {
	//Delphi字符串的地址的-4地址位置为长度
	if dstr == 0{
		return 0
	}
	result = *(*int32)(unsafe.Pointer(dstr - 4))
	return
}

//将常规的pchar返回到string
func Pchar2String(pcharstr uintptr)string  {
	if pcharstr == 0{
		return ""
	}
	ptr := unsafe.Pointer(pcharstr)
	gbt := make([]uint16,0,255)
	for i := 0; ; i++ {
		if 0 == *(*uint16)(ptr) {
			break
		}
		gbt = append(gbt,*(*uint16)(ptr))
		ptr = unsafe.Pointer(uintptr(ptr) + 2)
	}
	return string(utf16.Decode(gbt))
}

//快速转换，不修改的
func FastPchar2String(pcharstr uintptr)string  {
	if pcharstr==0{
		return ""
	}
	s := new(reflect.SliceHeader)
	s.Data = pcharstr
	s.Len = PcharLen(pcharstr)
	s.Cap = s.Len
	return string(utf16.Decode(*(*[] uint16)(unsafe.Pointer(s))))
}

//将Delphi的Pchar转换到string,Unicode
func DelphiPchar2String(dstr uintptr)string  {
	if dstr == 0{
		return ""
	}
	ptr := unsafe.Pointer(dstr)
	gbt := make([]uint16,DelphiPcharLen(dstr))
	for i := 0; ; i++ {
		if 0 == *(*uint16)(ptr) {
			break
		}
		gbt[i] = *(*uint16)(ptr)
		ptr = unsafe.Pointer(uintptr(ptr) + 2)
	}
	return string(utf16.Decode(gbt))
}

func FastDelphiPchar2String(pcharstr uintptr)string  {
	if pcharstr==0{
		return ""
	}
	s := new(reflect.SliceHeader)
	s.Data = pcharstr
	s.Len = int(DelphiPcharLen(pcharstr)*2)
	s.Cap = s.Len
	return string(utf16.Decode(*(*[] uint16)(unsafe.Pointer(s))))
}

//本函数只作为强制转换使用，不可将返回的Slice再做修改处理
func FastString2Byte(str string)[]byte  {
	return *(*[]byte)(unsafe.Pointer(&str))
}

func FastByte2String(bt []byte)string  {
	return *(*string)(unsafe.Pointer(&bt))
}

