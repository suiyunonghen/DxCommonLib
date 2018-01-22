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
	"os"
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

//将drwxrwx这些转化为 FileMode
func ModePermStr2FileMode(permStr string)(result os.FileMode)  {
	result = os.ModePerm
	filemodebytes := []byte(permStr)
	bytelen := len(filemodebytes)
	istart := 0
	if len(permStr) > 9 || filemodebytes[0]=='d' || filemodebytes[0]=='l' || filemodebytes[0]=='p'{
		istart = 1
	}
	if bytelen > istart && filemodebytes[istart] == 'r'{
		result = result | 0400
	}else{
		result = result & 0377
	}
	istart+=1
	if bytelen > istart && filemodebytes[istart] == 'w'{
		result = result | 0200
	}else{
		result = result & 0577
	}
	istart+=1

	if bytelen > istart && filemodebytes[istart] == 'x'{
		result = result | 0100
	}else{
		result = result & 0677
	}
	istart+=1

	if bytelen > istart && filemodebytes[istart] == 'r'{
		result = result | 0040
	}else{
		result = result & 0737
	}
	istart+=1

	if bytelen > istart && filemodebytes[istart] == 'w'{
		result = result | 0020
	}else{
		result = result & 0757
	}
	istart+=1

	if bytelen > istart && filemodebytes[istart] == 'x'{
		result = result | 0010
	}else{
		result = result & 0767
	}
	istart+=1


	if bytelen > istart && filemodebytes[istart] == 'r'{
		result = result | 0004
	}else{
		result = result & 0773
	}
	istart+=1

	if bytelen > istart && filemodebytes[istart] == 'w'{
		result = result | 0002
	}else{
		result = result & 0775
	}
	istart+=1

	if bytelen > istart && filemodebytes[istart] == 'x'{
		result = result | 0001
	}else{
		result = result & 0776
	}


	switch filemodebytes[0] {
	case 'd': result = result | os.ModeDir
	case 'l': result = result | os.ModeSymlink
	case 'p': result = result | os.ModeNamedPipe
	}
	return
}



//2进制转到16进制
func Binary2Hex(bt []byte)string  {
	var bf bytes.Buffer
	vhex := [16]byte{'0','1','2','3','4','5','6','7','8','9','A','B','C','D','E','F'}
	for _,vb := range bt{
		bf.WriteByte(vhex[vb >> 4])
		bf.WriteByte(vhex[vb & 0xF])
	}
	return string(bf.Bytes())
}

//16进制到2进制
func Hex2Binary(hexStr string)[]byte  {
	if hexStr == ""{
		return nil
	}
	vhex := [71]byte{'0':0,'1':1,'2':2,'3':3,'4':4,'5':5,'6':6,'7':7,'8':8,'9':9,'A':10,'B':11,'C':12,'D':13,'E':14,'F':15}
	btlen := len(hexStr) / 2
	result := make([]byte,btlen)
	btlen = btlen << 1
	jidx := 0
	for i := 0;i<btlen;i += 2{
		result[jidx] = vhex[hexStr[i]] << 4 | vhex[hexStr[i+1]]
		jidx++
	}
	return result
}