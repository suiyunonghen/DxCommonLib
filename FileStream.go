package DxCommonLib

import (
	"os"
)

type (
	FileCodeMode  uint8			//文件格式

	FileOpenMode	int		//文件打开方式
)



const(
	File_Code_Unknown FileCodeMode = iota
	File_Code_Utf8
	File_Code_Utf16BE
	File_Code_Utf16LE
	File_Code_GBK
)

const(
	FMCreate FileOpenMode = FileOpenMode(os.O_CREATE | os.O_WRONLY | os.O_TRUNC)
	FMOpenRead FileOpenMode = FileOpenMode(os.O_RDONLY)
	FMOpenWrite FileOpenMode = FileOpenMode(os.O_WRONLY | os.O_APPEND)
	FMOpenReadWrite FileOpenMode = FileOpenMode(os.O_RDWR | os.O_APPEND)

)

type GFileStream struct {
	fcacheSize		uint16
	fCache			[]byte
	file			*os.File
}

func (stream *GFileStream)Close()  {
	stream.file.Close()
}


func NewFileStream(fileName string,openMode FileOpenMode) (*GFileStream,error)  {
	file,err := os.OpenFile(fileName,int(openMode),0660)
	if err != nil{
		return nil,err
	}
	stream := &GFileStream{}
	stream.file = file
	return stream,nil
}


