/*
公用包
Autor: 不得闲
QQ:75492895
 */
package DxCommonLib

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"time"
	"unicode/utf16"
	"unsafe"
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
var (
	delphiFirstTime time.Time
	IsAmd64 = runtime.GOARCH == "amd64"
)

func init() {
	delphiFirstTime = time.Date(1899,12,30,0,0,0,0,time.Local)
}

//内存拷贝函数
//go:linkname CopyMemory runtime.memmove
func CopyMemory(to, from unsafe.Pointer, n uintptr)

//清空内存
//go:linkname ZeroMemory runtime.memclrNoHeapPointers
func ZeroMemory(ptr unsafe.Pointer, n uintptr)

//go:linkname memequal runtime.memequal
func memequal(a, b unsafe.Pointer, size uintptr)bool

//go:linkname memequal_varlen runtime.memequal_varlen
func memequal_varlen(a, b unsafe.Pointer)bool

//内存比较函数
func CompareMem(a,b unsafe.Pointer,size int)bool  {
	if size <= 0{
		return memequal_varlen(a,b)
	}
	return memequal(a,b,uintptr(size))
}

func ZeroByteSlice(bt []byte)  {
	btlen := len(bt)
	if btlen > 0{
		ZeroMemory(unsafe.Pointer(&bt[0]),uintptr(btlen))
	}
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

func Time2DelphiTime(t *time.Time)TDateTime  {
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



func FastBytes2Uint16s(bt []byte)[]uint16  {
	sliceHead := (*reflect.SliceHeader)(unsafe.Pointer(&bt))
	sliceHead.Len = sliceHead.Len / 2
	sliceHead.Cap = sliceHead.Cap / 2
	return *(*[]uint16)(unsafe.Pointer(sliceHead))
}

//本函数只作为强制转换使用，不可将返回的Slice再做修改处理
func FastString2Byte(str string)[]byte  {
	strHead := (*reflect.StringHeader)(unsafe.Pointer(&str))
	var sliceHead reflect.SliceHeader
	sliceHead.Len = strHead.Len
	sliceHead.Data = strHead.Data
	sliceHead.Cap = strHead.Len
	return *(*[]byte)(unsafe.Pointer(&sliceHead))
}

func FastByte2String(bt []byte)string  {
	return *(*string)(unsafe.Pointer(&bt))
}

func UTF16Byte2string(utf16bt []byte,isBigEnd bool)string  {
	if !isBigEnd{
		//判定末尾是否为换行,utf16，识别0
		btlen := len(utf16bt)
		if utf16bt[btlen-1] == 0 && utf16bt[btlen-2]=='\n'{
			utf16bt = utf16bt[:btlen-2]
			btlen -= 2
		}else if utf16bt[btlen - 1] == '\n'{
			utf16bt = utf16bt[:btlen-1]
			btlen--
		}
		if utf16bt[btlen-1] == 0 && utf16bt[btlen-2]=='\r'{
			utf16bt = utf16bt[:btlen-2]
		}else if utf16bt[btlen - 1] == '\r'{
			utf16bt = utf16bt[:btlen-1]
			btlen--
		}
		return string(utf16.Decode(FastBytes2Uint16s(utf16bt)))
	}
	arrlen := len(utf16bt) / 2
	uint16arr := make([]uint16,arrlen)
	for j,i:=0,0;j<arrlen;j,i=j+1,i+2{
		if j == arrlen-1{
			if utf16bt[i]== '\r' || utf16bt[i+1]=='\r'{
				arrlen--
				break
			}
		}else{
			uint16arr[j] = binary.BigEndian.Uint16(utf16bt[i:i+2])
		}
	}
	return string(utf16.Decode(uint16arr[:arrlen]))
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

//Date(1402384458000)
//Date(1224043200000+0800)
func ParserJsonTime(jsontime string)TDateTime  {
	bt := FastString2Byte(jsontime)
	dtflaglen := 0
	endlen := 0
	if  bytes.HasPrefix(bt,[]byte{'D','a','t','e','('}) && bytes.HasSuffix(bt,[]byte{')'}){
		dtflaglen = 5
		endlen = 1
	}else if bytes.HasPrefix(bt,[]byte{'/','D','a','t','e','('}) && bytes.HasSuffix(bt,[]byte{')','/'}){
		dtflaglen = 6
		endlen = 2
	}
	if dtflaglen > 0{
		bt = bt[dtflaglen:len(bt)-endlen]
		var(
			ms int64
			err error
		)
		endlen = 0
		idx := bytes.IndexByte(bt,'+')
		if idx < 0{
			idx = bytes.IndexByte(bt,'-')
		}else{
			endlen = 1
		}
		if idx < 0{
			str := FastByte2String(bt[:])
			if ms,err = strconv.ParseInt(str,10,64);err != nil{
				return -1
			}
			if len(str) > 9{
				ms = ms / 1000
			}
		}else{
			if endlen == 0{
				endlen = -1
			}
			str := FastByte2String(bt[:idx])
			ms,err = strconv.ParseInt(str,10,64)
			if err != nil{
				return -1
			}
			bt = bt[idx+1:]
			if len(bt) < 2{
				return -1
			}
			bt = bt[:2]
			ctz,err := strconv.Atoi(FastByte2String(bt))
			if err != nil{
				return -1
			}
			if len(str) > 9{
				ms = ms / 1000
			}
			ms += int64(ctz * 60)
		}
		ntime := time.Now()
		ns := ntime.Unix()
		ntime = ntime.Add((time.Duration(ms - ns)*time.Second))
		return Time2DelphiTime(&ntime)
	}
	return -1
}

//将内容转义成Json字符串
func EscapeJsonStr(str string) string {
	var buf bytes.Buffer
	for _,runedata := range str{
		switch runedata {
		case '\t':
			buf.WriteByte('\\')
			buf.WriteByte('t')
		case '\f':
			buf.WriteByte('\\')
			buf.WriteByte('f')
		case '\r':
			buf.WriteByte('\\')
			buf.WriteByte('r')
		case '\n':
			buf.WriteByte('\\')
			buf.WriteByte('n')
		case '\\':
			buf.WriteByte('\\')
			buf.WriteByte('\\')
		case '"':
			buf.WriteByte('\\')
			buf.WriteByte('"')
		case '\b':
			buf.WriteByte('\\')
			buf.WriteByte('b')
			/*case '\'':
				buf.WriteByte('\\')
				buf.WriteByte('\'')
			case '/':
				buf.WriteByte('\\')
				buf.WriteByte('/')*/
		default:
			if runedata < 256{
				buf.WriteByte(byte(runedata))
			}else{
				buf.Write([]byte{'\\','u'})
				var b [4]byte
				binary.BigEndian.PutUint32(b[:],uint32(runedata))
				if b[0]==0 && b[1] == 0{
					hexstr := Binary2Hex(b[2:])
					buf.WriteString(hexstr)
				}else{
					hexstr := Binary2Hex(b[0:2])
					buf.WriteString(hexstr)
					buf.Write([]byte{'\\','u'})
					hexstr = Binary2Hex(b[2:])
					buf.WriteString(hexstr)
				}
			}
		}
	}
	return FastByte2String(buf.Bytes())
}

//解码转义字符，将"\u6821\u56ed\u7f51\t02%20得闲"这类字符串，解码成正常显示的字符串
func ParserEscapeStr(bvalue []byte)string {
	blen := len(bvalue)
	i := 0
	//IsInUnicode := false
	unicodeidx := 0
	//escapein := false
	escapeType := uint8(0)  //0 normal,1 json\escapin,2 unicode escape, 3 % url escape
	var buf bytes.Buffer
	for i < blen{
		switch escapeType {
		case 1://json escapin
			escapeType = 0
			switch bvalue[i] {
			case 't':
				buf.WriteByte('\t')
			case 'f':
				buf.WriteByte('\f')
			case 'r':
				buf.WriteByte('\r')
			case 'n':
				buf.WriteByte('\n')
			case '\\':
				buf.WriteByte('\\')
			case '"':
				buf.WriteByte('"')
			case 'b':
				buf.WriteByte('\b')
				/*case '\'':
					buf.WriteByte('\'')
				case '/':
					buf.WriteByte('/')*/
			case 'u':
				escapeType = 2 // unicode decode
				unicodeidx = i
			default:
				buf.WriteByte('\\')
				buf.WriteByte(bvalue[i])
			}
		case 2: //unicode decode
			if (bvalue[i]>='0' && bvalue[i] <= '9' || bvalue[i] >='a' && bvalue[i] <= 'f' ||
				bvalue[i] >='A' && bvalue[i] <= 'F') && i - unicodeidx <= 4{
				//还是正常的Unicode字符，4个字符为一组
				//escapeType = 2
			}else{
				unicodestr := FastByte2String(bvalue[unicodeidx+1:i])
				if arune,err := strconv.ParseInt(unicodestr,16,32);err==nil{
					buf.WriteRune(rune(arune))
				}else{
					buf.Write(bvalue[unicodeidx:i])
				}
				escapeType = 0
				continue
			}
		case 3: //url escape
			for j := 0;j<3;j++{
				if (bvalue[j+i]>='0' && bvalue[j+i] <= '9' || bvalue[i+j] >='a' && bvalue[i+j] <= 'f' ||
					bvalue[j+i] >='A' && bvalue[j+i] <= 'F') && j<2{
					//还是正常的Byte字符，2个字符为一组
					//escapeType = 2
				}else{
					bytestr := FastByte2String(bvalue[i:i+j])
					if abyte,err := strconv.ParseInt(bytestr,16,32);err==nil{
						buf.WriteByte(byte(abyte))
					}else{
						buf.Write(bvalue[i-1:i+j]) //%要加上
					}
					escapeType = 0
					i += j - 1
					break
				}
			}
		default: //normal
			switch bvalue[i] {
			case '\\':
				escapeType = 1 //json escapin
			case '%':
				escapeType = 3 // url escape
			default:
				buf.WriteByte(bvalue[i])
			}
		}
		i++
	}
	switch escapeType {
	case 1:
		buf.WriteByte('\\')
	case 2:
		unicodestr := FastByte2String(bvalue[unicodeidx+1:i])
		if arune,err := strconv.ParseInt(unicodestr,16,32);err==nil{
			buf.WriteRune(rune(arune))
		}else{
			buf.Write(bvalue[unicodeidx:i])
		}
	}
	return FastByte2String(buf.Bytes())
}


//2进制转到16进制
func Binary2Hex(bt []byte)string  {
	var bf bytes.Buffer
	vhex := [16]byte{'0','1','2','3','4','5','6','7','8','9','A','B','C','D','E','F'}
	for _,vb := range bt{
		bf.WriteByte(vhex[vb >> 4])
		bf.WriteByte(vhex[vb & 0xF])
	}
	return FastByte2String(bf.Bytes())
}

//16进制到2进制
func Hex2Binary(hexStr string)[]byte  {
	if hexStr == ""{
		return nil
	}
	vhex := ['G']byte{'0':0,'1':1,'2':2,'3':3,'4':4,'5':5,'6':6,'7':7,'8':8,'9':9,'A':10,'B':11,'C':12,'D':13,'E':14,'F':15}
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