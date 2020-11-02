/*
公用包
Autor: 不得闲
QQ:75492895
 */
package DxCommonLib

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"
)

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

func Ord(x bool) uint8 {
	// Avoid branches. In the SSA compiler, this compiles to
	// exactly what you would want it to.
	return *(*uint8)(unsafe.Pointer(&x))
}


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

func StringFromUtf8Pointer(utf8Addr uintptr,maxlen int)string  {
	if utf8Addr == 0 {
		return ""
	}
	for i := 0; i< maxlen;i++{
		mb := (*byte)(unsafe.Pointer(uintptr(uint(utf8Addr)+uint(i))))
		if *mb==0{
			resultb := make([]byte,i)
			CopyMemory(unsafe.Pointer(&resultb[0]),unsafe.Pointer(utf8Addr),uintptr(i))
			return string(resultb)
		}
	}
	return ""
}

func StringFromUtf16Pointer(utf16Addr uintptr,maxlen int)string  {
	if utf16Addr == 0 {
		return ""
	}
	for i := 0; i< maxlen;i++{
		mb := (*uint16)(unsafe.Pointer(uintptr(uint(utf16Addr)+uint(i*2))))
		if *mb==0{
			resultb := make([]uint16,i)
			CopyMemory(unsafe.Pointer(&resultb[0]),unsafe.Pointer(utf16Addr),uintptr(i*2))
			return string(utf16.Decode(resultb))
		}
	}
	return ""
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


func String2Utf16Byte(s string)([]byte,error)  {
	bt,err := UTF16FromString(s)
	if err != nil{
		return nil,err
	}
	btlen := len(bt) * 2
	result := make([]byte,btlen)
	CopyMemory(unsafe.Pointer(&result[0]),unsafe.Pointer(&bt[0]),uintptr(btlen))
	return result,nil
}

func FastString2Utf16Byte(s string)([]byte,error)  {
	bt,err := UTF16FromString(s)
	if err != nil{
		return nil,err
	}
	strHead := (*reflect.SliceHeader)(unsafe.Pointer(&bt))
	var sliceHead reflect.SliceHeader
	sliceHead.Len = strHead.Len * 2
	sliceHead.Data = strHead.Data
	sliceHead.Cap = strHead.Cap * 2
	return *(*[]byte)(unsafe.Pointer(&sliceHead)),nil
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


//将内容转义成Json字符串
func EscapeJsonStr(str string,EscapeUnicode bool) string {
	return FastByte2String(EscapeJsonbyte(str,EscapeUnicode,nil))
}


func EscapeJsonbyte(str string,EscapeUnicode bool,dst []byte) []byte {
	dstlen := len(dst)
	strlen := len(str)
	if dstlen == 0{
		dst = make([]byte,0,strlen * 2)
	}
	for _,runedata := range str{
		switch runedata {
		case '\a':
			dst = append(dst, '\\','a')
		case '\b':
			dst = append(dst, '\\','b')
		case '\f':
			dst = append(dst, '\\','f')
		case '\n':
			dst = append(dst, '\\','n')
		case '\r':
			dst = append(dst, '\\','r')
		case '\t':
			dst = append(dst, '\\','t')
		case '\v':
			dst = append(dst, '\\','v')
		case '\\':
			dst = append(dst, '\\','\\')
		case '"':
			dst = append(dst, '\\','"')
		case '\'':
			dst = append(dst, '\\','\'')
		default:
			if EscapeUnicode{
				switch {
				case runedata < utf8.RuneSelf:
					dst = append(dst,byte(runedata))
				case runedata < ' ':
					dst = append(dst, '\\','x')
					dst = append(dst, vhex[byte(runedata)>>4],vhex[byte(runedata)&0xF])
				case runedata > utf8.MaxRune:
					runedata = 0xFFFD
					fallthrough
				case runedata < 0x10000:
					dst = append(dst, `\u`...)
					for s := 12; s >= 0; s -= 4 {
						dst = append(dst, vhex[runedata>>uint(s)&0xF])
					}
				default:
					dst = append(dst, `\U`...)
					for s := 28; s >= 0; s -= 4 {
						dst = append(dst, vhex[runedata>>uint(s)&0xF])
					}
				}
			}else{
				if runedata < utf8.RuneSelf {
					dst = append(dst,byte(runedata))
				}else{
					l := len(dst)
					dst = append(dst,0,0,0,0)
					n := utf8.EncodeRune(dst[l:l+4],runedata)
					dst = dst[:l+n]
				}
			}
		}
	}
	return dst
}

func UnEscapeStr(bvalue []byte)[]byte {
	buf := make([]byte,0,256)
	blen := len(bvalue)
	i := 0
	unicodeidx := 0
	escapeType := uint8(0)  //0 normal,1 json\escapin,2 unicode escape, 3 % url escape
	for i < blen{
		switch escapeType {
		case 1://json escapin
			escapeType = 0
			switch bvalue[i] {
			case 'a':
				buf = append(buf,'\a')
			case 'b':
				buf = append(buf,'\b')
			case 'f':
				buf = append(buf,'\f')
			case 'n':
				buf = append(buf,'\n')
			case 'r':
				buf = append(buf,'\r')
			case 't':
				buf = append(buf,'\t')
			case 'v':
				buf = append(buf,'\v')
			case '\\':
				buf = append(buf,'\\')
			case '"':
				buf = append(buf,'"')
			case '\'':
				buf = append(buf,'\'')
			case '/':
				buf = append(buf,'/')
			case 'u':
				escapeType = 2 // unicode decode
				unicodeidx = i
			default:
				buf = append(buf,'\\',bvalue[i])
			}
		case 2: //unicode decode
			if (bvalue[i]>='0' && bvalue[i] <= '9' || bvalue[i] >='a' && bvalue[i] <= 'f' ||
				bvalue[i] >='A' && bvalue[i] <= 'F') && i - unicodeidx <= 4{
				//还是正常的Unicode字符，4个字符为一组
				//escapeType = 2
			}else{
				unicodestr := FastByte2String(bvalue[unicodeidx+1:i])
				if arune,err := strconv.ParseInt(unicodestr,16,32);err==nil{
					l := len(buf)
					buf = append(buf,0,0,0,0)
					runelen := utf8.EncodeRune(buf[l:l+4],rune(arune))
					buf = buf[:l+runelen]
				}else{
					buf = append(buf,bvalue[unicodeidx:i]...)
				}
				escapeType = 0
				continue
			}
		case 3: //url escape
			for j := 0;j<3;j++{
				curidx := j+i
				if (curidx < blen) && (bvalue[curidx]>='0' && bvalue[curidx] <= '9' || bvalue[curidx] >='a' && bvalue[curidx] <= 'f' ||
					bvalue[curidx] >='A' && bvalue[curidx] <= 'F') && j<2{
					//还是正常的Byte字符，2个字符为一组
					//escapeType = 2
				}else{
					if j < 2{
						buf = append(buf,bvalue[i-1:curidx]...)//%要加上
					}else{
						bytestr := FastByte2String(bvalue[i:curidx])
						if abyte,err := strconv.ParseInt(bytestr,16,32);err==nil{
							if abyte < 0x20 || (abyte>='0' && abyte <= '9' || abyte >='a' && abyte <= 'f' ||
								abyte >='A' && abyte <= 'F') {
								buf = append(buf,bvalue[i-1:curidx]...)//%要加上
							}else{
								buf = append(buf,byte(abyte))
							}
						}else{
							buf = append(buf,bvalue[i-1:curidx]...)//%要加上
						}
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
				buf = append(buf,bvalue[i])
			}
		}
		i++
	}
	switch escapeType {
	case 1:
		buf = append(buf,'\\')
	case 2:
		unicodestr := FastByte2String(bvalue[unicodeidx+1:i])
		if arune,err := strconv.ParseInt(unicodestr,16,32);err==nil{
			l := len(buf)
			buf = append(buf,0,0,0,0)
			runelen := utf8.EncodeRune(buf[l:l+4],rune(arune))
			buf = buf[:l+runelen]
		}else{
			buf = append(buf,bvalue[unicodeidx:i]...)
		}
	case 3:
		buf = append(buf,'%')
	}
	return buf
}


//解码转义字符，将"\u6821\u56ed\u7f51\t02%20得闲"这类字符串，解码成正常显示的字符串
func ParserEscapeStr(bvalue []byte)string {
	return FastByte2String(UnEscapeStr(bvalue))
}

//Date(1402384458000)
//Date(1224043200000+0800)
func ParserJsonTime2Go(jsontime string)time.Time  {
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
				return time.Time{}
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
				return time.Time{}
			}
			bt = bt[idx+1:]
			if len(bt) < 2{
				return time.Time{}
			}
			bt = bt[:2]
			ctz,err := strconv.Atoi(FastByte2String(bt))
			if err != nil{
				return time.Time{}
			}
			if len(str) > 9{
				ms = ms / 1000
			}
			ms += int64(ctz * 60)
		}
		ntime := time.Now()
		ns := ntime.Unix()
		ntime = ntime.Add((time.Duration(ms - ns)*time.Second))
		return ntime
	}
	return time.Time{}
}

var(
	vhex = [16]byte{'0','1','2','3','4','5','6','7','8','9','A','B','C','D','E','F'}
)

//2进制转到16进制
func Binary2Hex(bt []byte,dst []byte)[]byte  {
	l := len(dst)
	if len(dst) == 0{
		dst = make([]byte,0,l * 2)
	}
	for _,vb := range bt{
		dst = append(dst,vhex[vb >> 4],vhex[vb & 0xF])
	}
	return dst
}

func Bin2Hex(bt []byte)string  {
	return FastByte2String(Binary2Hex(bt,nil))
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

//From github.com/valyala/fastjson/tree/master/fastfloat
func StrToIntDef(vstr string,defv int64)int64  {
	if v,err := ParseInt64(vstr);err != nil{
		return defv
	}else{
		return v
	}
}

func ParseInt64(s string) (int64, error) {
	if len(s) == 0 {
		return 0, errors.New("cannot parse int64 from empty string")
	}
	i := uint(0)
	minus := s[0] == '-'
	if minus {
		i++
		if i >= uint(len(s)) {
			return 0, fmt.Errorf("cannot parse int64 from %q", s)
		}
	}

	d := int64(0)
	j := i
	for i < uint(len(s)) {
		if s[i] >= '0' && s[i] <= '9' {
			d = d*10 + int64(s[i]-'0')
			i++
			if i > 18 {
				// The integer part may be out of range for int64.
				// Fall back to slow parsing.
				dd, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return 0, err
				}
				return dd, nil
			}
			continue
		}
		break
	}
	if i <= j {
		return 0, fmt.Errorf("cannot parse int64 from %q", s)
	}
	if i < uint(len(s)) {
		// Unparsed tail left.
		return 0, fmt.Errorf("unparsed tail left after parsing int64 form %q: %q", s, s[i:])
	}
	if minus {
		d = -d
	}
	return d, nil
}

func ParseFloat(s string) (float64, error) {
	if len(s) == 0 {
		return 0, fmt.Errorf("cannot parse float64 from empty string")
	}
	i := uint(0)
	minus := s[0] == '-'
	if minus {
		i++
		if i >= uint(len(s)) {
			return 0, fmt.Errorf("cannot parse float64 from %q", s)
		}
	}

	d := uint64(0)
	j := i
	for i < uint(len(s)) {
		if s[i] >= '0' && s[i] <= '9' {
			d = d*10 + uint64(s[i]-'0')
			i++
			if i > 18 {
				// The integer part may be out of range for uint64.
				// Fall back to slow parsing.
				f, err := strconv.ParseFloat(s, 64)
				if err != nil && !math.IsInf(f, 0) {
					return 0, err
				}
				return f, nil
			}
			continue
		}
		break
	}
	if i <= j {
		ss := s[i:]
		if strings.HasPrefix(ss, "+") {
			ss = ss[1:]
		}
		if strings.EqualFold(ss, "inf") {
			if minus {
				return -inf, nil
			}
			return inf, nil
		}
		if strings.EqualFold(ss, "nan") {
			return nan, nil
		}
		return 0, fmt.Errorf("unparsed tail left after parsing float64 from %q: %q", s, ss)
	}
	f := float64(d)
	if i >= uint(len(s)) {
		// Fast path - just integer.
		if minus {
			f = -f
		}
		return f, nil
	}

	if s[i] == '.' {
		// Parse fractional part.
		i++
		if i >= uint(len(s)) {
			return 0, fmt.Errorf("cannot parse fractional part in %q", s)
		}
		k := i
		for i < uint(len(s)) {
			if s[i] >= '0' && s[i] <= '9' {
				d = d*10 + uint64(s[i]-'0')
				i++
				if i-j >= uint(len(float64pow10)) {
					// The mantissa is out of range. Fall back to standard parsing.
					f, err := strconv.ParseFloat(s, 64)
					if err != nil && !math.IsInf(f, 0) {
						return 0, fmt.Errorf("cannot parse mantissa in %q: %s", s, err)
					}
					return f, nil
				}
				continue
			}
			break
		}
		if i < k {
			return 0, fmt.Errorf("cannot find mantissa in %q", s)
		}
		// Convert the entire mantissa to a float at once to avoid rounding errors.
		f = float64(d) / float64pow10[i-k]
		if i >= uint(len(s)) {
			// Fast path - parsed fractional number.
			if minus {
				f = -f
			}
			return f, nil
		}
	}
	if s[i] == 'e' || s[i] == 'E' {
		// Parse exponent part.
		i++
		if i >= uint(len(s)) {
			return 0, fmt.Errorf("cannot parse exponent in %q", s)
		}
		expMinus := false
		if s[i] == '+' || s[i] == '-' {
			expMinus = s[i] == '-'
			i++
			if i >= uint(len(s)) {
				return 0, fmt.Errorf("cannot parse exponent in %q", s)
			}
		}
		exp := int16(0)
		j := i
		for i < uint(len(s)) {
			if s[i] >= '0' && s[i] <= '9' {
				exp = exp*10 + int16(s[i]-'0')
				i++
				if exp > 300 {
					// The exponent may be too big for float64.
					// Fall back to standard parsing.
					f, err := strconv.ParseFloat(s, 64)
					if err != nil && !math.IsInf(f, 0) {
						return 0, fmt.Errorf("cannot parse exponent in %q: %s", s, err)
					}
					return f, nil
				}
				continue
			}
			break
		}
		if i <= j {
			return 0, fmt.Errorf("cannot parse exponent in %q", s)
		}
		if expMinus {
			exp = -exp
		}
		f *= math.Pow10(int(exp))
		if i >= uint(len(s)) {
			if minus {
				f = -f
			}
			return f, nil
		}
	}
	return 0, fmt.Errorf("cannot parse float64 from %q", s)
}

//From github.com/valyala/fastjson/tree/master/fastfloat
func StrToUintDef(vstr string,defv uint64)uint64 {
	vlen := uint(len(vstr))
	if vlen == 0 {
		return defv
	}
	if vlen > 18{
		if dd, err := strconv.ParseUint(vstr, 10, 64);err != nil {
			return defv
		}else{
			return dd
		}
	}
	i := uint(0)
	d := uint64(0)
	j := i
	for i < vlen {
		if vstr[i] >= '0' && vstr[i] <= '9' {
			d = d*10 + uint64(vstr[i]-'0')
			i++
			continue
		}
		break
	}
	if i <= j || i < vlen {
		return defv
	}
	return d
}

var (
	inf = math.Inf(1)
	nan = math.NaN()
)

var float64pow10 = [...]float64{
	1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16,1e17,1e18,1e19,1e20,1e21,1e22,1e123,1e124,
}
//github.com/valyala/fastjson/tree/master/fastfloat
func StrToFloatDef(s string,defv float64) float64 {
	vlen := uint(len(s))
	if vlen == 0 {
		return defv
	}
	i := uint(0)
	minus := s[0] == '-'
	if minus {
		i++
		if i >= vlen {
			return defv
		}
	}

	d := uint64(0)
	j := i
	for i < vlen {
		if s[i] >= '0' && s[i] <= '9' {
			d = d*10 + uint64(s[i]-'0')
			i++
			if i > 18 {
				// The integer part may be out of range for uint64.
				// Fall back to slow parsing.
				f, err := strconv.ParseFloat(s, 64)
				if err != nil && !math.IsInf(f, 0) {
					return defv
				}
				return f
			}
			continue
		}
		break
	}
	if i <= j {
		ss := s[i:]
		if strings.HasPrefix(ss, "+") {
			ss = ss[1:]
		}
		if strings.EqualFold(ss, "inf") {
			if minus {
				return -inf
			}
			return inf
		}
		if strings.EqualFold(ss, "nan") {
			return nan
		}
		return defv
	}
	f := float64(d)
	if i >= uint(len(s)) {
		// Fast path - just integer.
		if minus {
			f = -f
		}
		return f
	}

	if s[i] == '.' {
		// Parse fractional part.
		i++
		if i >= uint(len(s)) {
			return defv
		}
		k := i
		for i < uint(len(s)) {
			if s[i] >= '0' && s[i] <= '9' {
				d = d*10 + uint64(s[i]-'0')
				i++
				if i-j >= uint(len(float64pow10)) {
					// The mantissa is out of range. Fall back to standard parsing.
					f, err := strconv.ParseFloat(s, 64)
					if err != nil && !math.IsInf(f, 0) {
						return defv
					}
					return f
				}
				continue
			}
			break
		}
		if i < k {
			return defv
		}
		// Convert the entire mantissa to a float at once to avoid rounding errors.
		f = float64(d) / float64pow10[i-k]
		if i >= uint(len(s)) {
			// Fast path - parsed fractional number.
			if minus {
				f = -f
			}
			return f
		}
	}
	if s[i] == 'e' || s[i] == 'E' {
		// Parse exponent part.
		i++
		if i >= uint(len(s)) {
			return defv
		}
		expMinus := false
		if s[i] == '+' || s[i] == '-' {
			expMinus = s[i] == '-'
			i++
			if i >= uint(len(s)) {
				return defv
			}
		}
		exp := int16(0)
		j := i
		for i < uint(len(s)) {
			if s[i] >= '0' && s[i] <= '9' {
				exp = exp*10 + int16(s[i]-'0')
				i++
				if exp > 300 {
					// The exponent may be too big for float64.
					// Fall back to standard parsing.
					f, err := strconv.ParseFloat(s, 64)
					if err != nil && !math.IsInf(f, 0) {
						return defv
					}
					return f
				}
				continue
			}
			break
		}
		if i <= j {
			return defv
		}
		if expMinus {
			exp = -exp
		}
		f *= math.Pow10(int(exp))
		if i >= uint(len(s)) {
			if minus {
				f = -f
			}
			return f
		}
	}
	return defv
}