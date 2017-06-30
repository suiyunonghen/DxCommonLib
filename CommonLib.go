package DxCommonLib

import (
	"strings"
	"os"
	"fmt"
)

type(
	IStrings interface {
		Count()int
		Strings(index int)string
		SetStrings(index int,str string)
		Text()string
		SetText(text string)
		LoadFromFile(fileName string)
		SaveToFile(fileName string)
		Add(str string)
		Insert(Index int,str string)
		Delete(index int)
		AddStrings(strs IStrings)
		AddSlice(strs []string)
		Clear()
		IndexOf(str string) int

		AddPair(Name,Value string)
		IndexOfName(Name string) int
		ValueFromIndex(index int) string
		ValueByName(Name string) string
		Names(Index int) string
		AsSlice()[]string
	}

	GStringList struct{
		strings 	[]string
		LineBreak	string
	}
)

func (lst *GStringList) Count()int {
	if lst.strings == nil{
		return 0
	}
	return len(lst.strings)
}


func (lst *GStringList)Strings(index int)string{
	if index >= 0 && index < len(lst.strings){
		return lst.strings[index]
	}
	return ""
}

func (lst *GStringList)SetStrings(index int,str string){
	if lst.strings == nil{
		lst.strings = make([]string,0,64)
	}
	if index >= 0 && index < len(lst.strings){
		lst.strings[index] = str
	}
}

func (lst *GStringList)Text()string{
	if lst.Count()==0{
		return ""
	}
	return  strings.Join(lst.strings,lst.LineBreak)
}

func (lst *GStringList)SetText(text string){
	lst.strings = strings.Split(text,lst.LineBreak)
}

func (lst *GStringList)LoadFromFile(fileName string){
	if finfo,err := os.Stat(fileName);err == nil && !finfo.IsDir(){
		if file,err := os.Open(fileName);err == nil{
			databytes := make([]byte,finfo.Size())
			file.Read(databytes)
			isUtf8 := databytes[0] == 0xEF && databytes[1] == 0xBB && databytes[2] == 0xBF
			if isUtf8{
				databytes = databytes[3:]
			}
			file.Close()
			if !isUtf8{
				if tmpbytes,err := GBK2Utf8(databytes);err == nil{
					databytes = tmpbytes
				}
			}
			lst.strings = strings.Split(FastByte2String(databytes),lst.LineBreak)
		}
	}
}

func (lst *GStringList)SaveToFile(fileName string)  {
	//文件要先写入UTF8的BOM
	if file,err := os.OpenFile(fileName,os.O_CREATE | os.O_TRUNC,0644);err == nil{
		if lst.Count() > 0{
			file.Write([]byte{0xEF,0xBB,0xBF})
			//写入内容
			file.Write(FastString2Byte(lst.Text()))
		}
		file.Close()
	}
}

func (lst *GStringList)Add(str string){
	if lst.strings == nil{
		lst.strings = make([]string,0,64)
	}
	lst.strings = append(lst.strings,str)
}

func (lst *GStringList)Insert(Index int,str string){
	if lst.strings == nil{
		lst.strings = make([]string,0,64)
	}
	if Index >= 0 && Index < len(lst.strings){
		temp := append([]string{},lst.strings[Index:]...)
		lst.strings = append(lst.strings[0:Index],str)
		lst.strings = append(lst.strings,temp...)
	}else{
		lst.Add(str)
	}
}
func (lst *GStringList)Delete(index int){
	if lst.Count() > 0 && index >= 0 && index < len(lst.strings){
		lst.strings = append(lst.strings[0:index],lst.strings[index+1:]...)
	}
}

func (lst *GStringList)AddStrings(strs IStrings){
	for idx := 0;idx < strs.Count();idx++{
		lst.Add(strs.Strings(idx))
	}
}

func (lst *GStringList)AddSlice(strs []string){
	if lst.strings != nil{
		lst.strings = append(lst.strings,strs...)
	}else{
		lst.strings = strs
	}
}
func (lst *GStringList)Clear(){
	if lst.strings != nil{
		lst.strings = lst.strings[:0]
	}
}
func (lst *GStringList)IndexOf(str string)int{
	if lst.Count() > 0{
		for idx,v := range lst.strings{
			if v == str{
				return idx
			}
		}
	}
	return -1
}

func (lst *GStringList)AddPair(Name,Value string){
	lst.strings = append(lst.strings,fmt.Sprintf("%s=%s",Name,Value))
}
func (lst *GStringList)IndexOfName(Name string)int{
	if lst.Count() > 0{
		for idx,v := range lst.strings{
			if eidx := strings.IndexByte(v,'=');idx > 0{
				bt := FastString2Byte(v)
				if FastByte2String(bt[:eidx]) == Name {
					return idx
				}
			}
		}
	}
	return -1
}

func (lst *GStringList)ValueFromIndex(index int)string{
	if lst.Count() == 0{
		return ""
	}
	if index >= 0 && index < len(lst.strings){
		str := lst.strings[index]
		if idx := strings.IndexByte(str,'=');idx > 0{
			bt := FastString2Byte(str)
			return FastByte2String(bt[idx+1:])
		}
		return ""
	}
	return ""
}

func (lst *GStringList)ValueByName(Name string)string{
	if lst.Count() > 0{
		for idx,v := range lst.strings{
			if eidx := strings.IndexByte(v,'=');idx > 0{
				bt := FastString2Byte(v)
				if FastByte2String(bt[:eidx]) == Name {
					return FastByte2String(bt[eidx+1:])
				}
			}
		}
	}
	return ""
}
func (lst *GStringList)Names(Index int) string{
	if lst.Count() == 0{
		return ""
	}
	if Index >= 0 && Index < len(lst.strings){
		str := lst.strings[Index]
		if idx := strings.IndexByte(str,'=');idx > 0{
			bt := FastString2Byte(str)
			return FastByte2String(bt[:idx])
		}
		return ""
	}
	return ""
}
func (lst *GStringList)AsSlice()[]string{
	return lst.strings
}

