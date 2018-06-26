package DxCommonLib

import (
	"os"
	"io"
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
	FMOpenWrite FileOpenMode = FileOpenMode(os.O_WRONLY)// | os.O_APPEND)
	FMOpenReadWrite FileOpenMode = FileOpenMode(os.O_RDWR)// | os.O_APPEND)

)

type GFileStream struct {
	fCache			[]byte
	fcacheSize		uint
	fbufferStart	int		//缓存在文件中开始的位置
	fbufferEnd		int		//缓存的结束位置
	fbufferRW		int
	fmodified		bool
	file			*os.File
}

func (stream *GFileStream)Close()  {
	stream.FlushBuffer()
	stream.file.Close()
}

func (stream *GFileStream)FlushBuffer() error {
	if stream.fmodified{
		_,err := stream.file.Write(stream.fCache[:stream.fbufferRW])
		if err != nil{
			return err
		}
		stream.fbufferStart = int(stream.FilePosition())
		stream.fbufferRW = 0
		bfend,_ := stream.file.Read(stream.fCache)
		stream.fbufferEnd = bfend + stream.fbufferStart
		stream.fmodified = false
	}
	return nil
}

func (stream *GFileStream)Read(buffer []byte)int  {
	stream.FlushBuffer()
	bflen := len(buffer)
	if bflen == 0{
		return 0
	}
	if stream.fbufferStart < 0{ //未处理，重新设定位置读取
		if uint(bflen) >= stream.fcacheSize{
			rln,_ := stream.file.Read(buffer)
			return rln
		}else{
			stream.fbufferStart = int(stream.FilePosition())
			bfend,_ := stream.file.Read(stream.fCache)
			stream.fbufferEnd = bfend
		}
	}
	hasleave := stream.fbufferEnd - stream.fbufferStart - stream.fbufferRW
	if hasleave >= bflen{
		rln := copy(buffer,stream.fCache[stream.fbufferRW:stream.fbufferRW + bflen])
		stream.fbufferRW += rln
		return rln
	}
	copy(buffer,stream.fCache[stream.fbufferRW:])
	if uint(stream.fbufferEnd - stream.fbufferStart) < stream.fcacheSize{
		return int(hasleave)
	}
	bflen -= hasleave
	if uint(bflen) >= stream.fcacheSize{
		rln,err := stream.file.Read(buffer[hasleave:])
		if err != nil{
			return rln + int(hasleave)
		}
		stream.fbufferStart = int(stream.FilePosition())
		stream.fbufferRW = 0
		bfend,_ := stream.file.Read(stream.fCache)
		stream.fbufferEnd = bfend + stream.fbufferStart
		return rln + int(hasleave)
	}else{
		stream.fbufferStart = stream.fbufferEnd
		stream.fbufferRW = 0
		bfend,_ := stream.file.Read(stream.fCache)
		stream.fbufferEnd = bfend + stream.fbufferStart
		hasread := int(hasleave)
		hasread += stream.Read(buffer[hasread:])
		return hasread
	}
}

func (stream *GFileStream)Position()int  {
	if stream.fbufferStart < 0{
		return int(stream.FilePosition())
	}else{
		return stream.fbufferStart + stream.fbufferRW
	}
}

func (stream *GFileStream)SetPosition(ps int) error {
	if stream.fbufferStart < 0{
		_,err := stream.file.Seek(int64(ps),io.SeekStart)
		return err
	}else if ps < stream.fbufferStart || ps >= stream.fbufferEnd {
		_,err := stream.file.Seek(int64(ps),io.SeekStart)
		if err != nil{
			return err
		}
		stream.fbufferStart = ps
		stream.fbufferRW = 0
		bfend,_ := stream.file.Read(stream.fCache)
		stream.fbufferEnd = bfend + stream.fbufferStart
	}else{
		stream.fbufferRW = ps - stream.fbufferStart
	}
	return nil
}

func (stream *GFileStream)FilePosition()int64  {
	pos,_ := stream.file.Seek(0,io.SeekCurrent)
	return pos
}


func (stream *GFileStream)Write(data []byte)(int,error)  {
	datalen := len(data)
	if datalen == 0{
		return 0,nil
	}
	if !stream.fmodified{ //如果之前一直是读取的，先将文件指针移动到Position位置,然后写入到缓存中去
		stream.file.Seek(int64(stream.Position()),io.SeekStart)
	}
	stream.fbufferStart = -1
	if stream.fbufferRW == 0{ //直接一步写入
		if uint(datalen) >= stream.fcacheSize{
			return stream.file.Write(data)
		}
		wln := copy(stream.fCache[:datalen],data)
		stream.fbufferRW = wln
		stream.fmodified = true
		return wln,nil
	}
	wln := 0
	canCachelen := int(stream.fcacheSize) - stream.fbufferRW
	if datalen <= canCachelen{
		wln = copy(stream.fCache[stream.fbufferRW:stream.fbufferRW+datalen],data)
		stream.fbufferRW += wln
		stream.fmodified = true
		data = nil
	}else{
		wln = copy(stream.fCache[stream.fbufferRW:],data[:canCachelen])
		stream.fbufferRW += wln
		stream.fmodified = true
		data = data[canCachelen:]
	}
	if stream.fbufferRW == int(stream.fcacheSize){
		_,err := stream.file.Write(stream.fCache)
		if err != nil{
			return wln,err
		}
		stream.fbufferRW = 0
	}
	if data != nil{
		wln2,err := stream.Write(data)
		return wln+wln2,err
	}
	return wln,nil
}

func (stream *GFileStream)Size()int64  {
	pos,_ := stream.file.Seek(0,io.SeekCurrent)
	endpos,_ := stream.file.Seek(0,io.SeekEnd)
	stream.file.Seek(pos,io.SeekStart)
	return endpos
}

func NewFileStream(fileName string,openMode FileOpenMode,bufferSize int) (*GFileStream,error)  {
	file,err := os.OpenFile(fileName,int(openMode),0660)
	if err != nil{
		return nil,err
	}
	stream := &GFileStream{}
	stream.file = file
	if bufferSize <= 0{
		bufferSize = 32768
	}
	stream.fcacheSize = uint(bufferSize)
	stream.fCache = make([]byte,bufferSize)
	stream.fbufferStart = -1 //未读取到缓存
	return stream,nil
}


