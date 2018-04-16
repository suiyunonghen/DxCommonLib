/*
仿Delphi的通用类库
GStringList类似于TStringList
Autor: 不得闲
QQ:75492895
*/
package DxCommonLib

import (
	"fmt"
	"os"
	"strings"
	"io"
	"bufio"
)

type (
	IStrings interface {
		Count() int
		Strings(index int) string
		SetStrings(index int, str string)
		Text() string
		SetText(text string)
		LoadFromFile(fileName string)
		SaveToFile(fileName string)
		Add(str string)
		Insert(Index int, str string)
		Delete(index int)
		AddStrings(strs IStrings)
		AddSlice(strs []string)
		Clear()
		IndexOf(str string) int

		AddPair(Name, Value string)
		IndexOfName(Name string) int
		ValueFromIndex(index int) string
		ValueByName(Name string) string
		Names(Index int) string
		AsSlice() []string
	}

	LineBreakMode byte

	GStringList struct {
		strings           []string
		LineBreak         LineBreakMode
		UnknownCodeUseGbk bool //未知编码的时候采用GBK编码打开
	}
)

const (
	LBK_CRLF LineBreakMode = iota
	LBK_CR
	LBK_LF
)

func (lst *GStringList) LineBreakStr() string {
	switch lst.LineBreak {
	case LBK_CRLF:
		return "\r\n"
	case LBK_CR:
		return "\r"
	case LBK_LF:
		return "\n"
	default:
		return "\r\n"
	}
}

func (lst *GStringList) Count() int {
	if lst.strings == nil {
		return 0
	}
	return len(lst.strings)
}

func (lst *GStringList) Strings(index int) string {
	if index >= 0 && index < len(lst.strings) {
		return lst.strings[index]
	}
	return ""
}

func (lst *GStringList) SetStrings(index int, str string) {
	if lst.strings == nil {
		lst.strings = make([]string, 0, 64)
	}
	if index >= 0 && index < len(lst.strings) {
		lst.strings[index] = str
	}
}

func (lst *GStringList) Text() string {
	if lst.Count() == 0 {
		return ""
	}
	return strings.Join(lst.strings, lst.LineBreakStr())
}

func (lst *GStringList) SetText(text string) {
	lst.strings = strings.Split(text, lst.LineBreakStr())
}

func (lst *GStringList) LoadFromFile(fileName string) {
	if finfo, err := os.Stat(fileName); err == nil && !finfo.IsDir() {
		if file, err := os.Open(fileName); err == nil {
			//先判定文件格式，是否为utf8或者utf16，去掉BOM
			if lst.strings == nil{
				lst.strings = make([]string,0,100)
			}else{
				lst.strings = lst.strings[:0]
			}
			defer file.Close()
			var bt [3]byte
			filecodeType := File_Code_Unknown
			if _,err := file.Read(bt[:3]);err==nil{
				if bt[0] == 0xFF && bt[1] == 0xFE { //UTF-16(Little Endian)
					file.Seek(-1,io.SeekCurrent)
					filecodeType = File_Code_Utf16LE
				}else if bt[0] == 0xFE && bt[1] == 0xFF{ //UTF-16(big Endian)
					file.Seek(-1,io.SeekCurrent)
					filecodeType = File_Code_Utf16BE
				}else if bt[0] == 0xEF && bt[1] == 0xBB && bt[2] == 0xBF { //UTF-8
					filecodeType = File_Code_Utf8
				}else{
					file.Seek(-3,io.SeekCurrent)
				}
			}
			reader := bufio.NewReader(file)
			for{
				line,err := reader.ReadBytes('\n')
				if filecodeType == File_Code_Utf16LE{ //小端结尾多一个空白的0标记
					reader.ReadByte()
				}
				if err == nil || err == io.EOF{
					linelen := len(line)
					if linelen > 2{
						if line[linelen-2] == '\r'{
							line = line[:linelen - 2]
						}else if line[linelen - 1] == '\n'{
							line = line[:linelen-1]
						}
					}
					if linelen>0{
						switch filecodeType {
						case File_Code_Utf8:
							lst.Add(FastByte2String(line))
						case File_Code_Utf16LE,File_Code_Utf16BE:
							lst.Add(UTF16Byte2string(line,filecodeType == File_Code_Utf16BE))
						case File_Code_GBK,File_Code_Unknown:
							if tmpbytes, err := GBK2Utf8(line); err == nil {
								lst.Add(FastByte2String(tmpbytes))
							}else{
								lst.Add(FastByte2String(line))
							}
						}
					}else{
						lst.Add("")
					}
					if err != nil{
						return
					}
				}else{
					return
				}
			}
		}
	}
}

func (lst *GStringList) SaveToFile(fileName string) {
	//文件要先写入UTF8的BOM
	if file, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC, 0644); err == nil {
		count := lst.Count()
		if count > 0 {
			file.Write([]byte{0xEF, 0xBB, 0xBF})
			//写入内容
			//file.Write(FastString2Byte(lst.Text()))
			for i := 0;i<count;i++{
				file.WriteString(lst.strings[i])
				if i < count - 1{
					file.WriteString(lst.LineBreakStr())
				}
			}
		}
		file.Close()
	}
}

func (lst *GStringList) Add(str string) {
	if lst.strings == nil {
		lst.strings = make([]string, 0, 64)
	}
	lst.strings = append(lst.strings, str)
}

func (lst *GStringList) Insert(Index int, str string) {
	if lst.strings == nil {
		lst.strings = make([]string, 0, 64)
	}
	if Index >= 0 && Index < len(lst.strings) {
		temp := append([]string{}, lst.strings[Index:]...)
		lst.strings = append(lst.strings[0:Index], str)
		lst.strings = append(lst.strings, temp...)
	} else {
		lst.Add(str)
	}
}
func (lst *GStringList) Delete(index int) {
	if lst.Count() > 0 && index >= 0 && index < len(lst.strings) {
		lst.strings = append(lst.strings[0:index], lst.strings[index+1:]...)
	}
}

func (lst *GStringList) AddStrings(strs IStrings) {
	for idx := 0; idx < strs.Count(); idx++ {
		lst.Add(strs.Strings(idx))
	}
}

func (lst *GStringList) AddSlice(strs []string) {
	if lst.strings != nil {
		lst.strings = append(lst.strings, strs...)
	} else {
		lst.strings = strs
	}
}
func (lst *GStringList) Clear() {
	if lst.strings != nil {
		lst.strings = lst.strings[:0]
	}
}
func (lst *GStringList) IndexOf(str string) int {
	if lst.Count() > 0 {
		for idx, v := range lst.strings {
			if v == str {
				return idx
			}
		}
	}
	return -1
}

func (lst *GStringList) AddPair(Name, Value string) {
	lst.strings = append(lst.strings, fmt.Sprintf("%s=%s", Name, Value))
}
func (lst *GStringList) IndexOfName(Name string) int {
	if lst.Count() > 0 {
		for idx, v := range lst.strings {
			if eidx := strings.IndexByte(v, '='); eidx > 0 {
				bt := []byte(v)
				if string(bt[:eidx]) == Name {
					return idx
				}
			}
		}
	}
	return -1
}

func (lst *GStringList) ValueFromIndex(index int) string {
	if lst.Count() == 0 {
		return ""
	}
	if index >= 0 && index < len(lst.strings) {
		str := lst.strings[index]
		if idx := strings.IndexByte(str, '='); idx > 0 {
			bt := []byte(str)
			return FastByte2String(bt[idx+1:])
		}
		return ""
	}
	return ""
}

func (lst *GStringList) ValueByName(Name string) string {
	if lst.Count() > 0 {
		for _, v := range lst.strings {
			if eidx := strings.IndexByte(v, '='); eidx > 0 {
				bt := []byte(v)
				if string(bt[:eidx]) == Name {
					return FastByte2String(bt[eidx+1:])
				}
			}
		}
	}
	return ""
}
func (lst *GStringList) Names(Index int) string {
	if lst.Count() == 0 {
		return ""
	}
	if Index >= 0 && Index < len(lst.strings) {
		str := lst.strings[Index]
		if idx := strings.IndexByte(str, '='); idx > 0 {
			bt := []byte(str)
			return FastByte2String(bt[:idx])
		}
		return ""
	}
	return ""
}
func (lst *GStringList) AsSlice() []string {
	return lst.strings
}
