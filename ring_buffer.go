package DxCommonLib

import (
	"errors"
)

type DxRingBuffer struct {
	rawData				[]byte
	rpos,wpos			int
	markrpos			int
	markwpos			int
	isEmpty				bool
	hasMark				bool
}

func (buffer *DxRingBuffer)Capacity()int  {
	return len(buffer.rawData)
}

//数据长度
func (buffer *DxRingBuffer)DataSize()int  {
	if buffer.isEmpty{
		return 0
	}
	if buffer.wpos > buffer.rpos{
		return buffer.wpos - buffer.rpos
	}
	if buffer.wpos == buffer.rpos{
		return len(buffer.rawData)
	}
	return len(buffer.rawData) - buffer.rpos + buffer.wpos
}

//剩下的空余大小
func (buffer *DxRingBuffer)FreeSize()int  {
	if buffer.isEmpty {
		return len(buffer.rawData)
	}
	if buffer.rpos == buffer.wpos{
		return 0
	}
	if buffer.wpos < buffer.rpos{
		return buffer.rpos - buffer.wpos
	}
	return len(buffer.rawData) - buffer.wpos + buffer.rpos
}

func (buffer *DxRingBuffer)MarkReadWrite()  {
	buffer.markrpos = buffer.rpos
	buffer.markwpos = buffer.wpos
	buffer.hasMark = true
}

func (buffer *DxRingBuffer)RestoreMark()  {
	if !buffer.hasMark{
		return
	}
	buffer.rpos = buffer.markrpos
	buffer.wpos = buffer.markwpos
	buffer.hasMark = false
}

func (buffer *DxRingBuffer)Skip(n int)  {
	if buffer.isEmpty || n<=0{
		return
	}
	if buffer.wpos > buffer.rpos{
		vlen := buffer.wpos - buffer.rpos
		if vlen > n{
			buffer.rpos = buffer.rpos + n
		}else{
			buffer.isEmpty = true
			buffer.rpos = buffer.wpos
		}
		return
	}
	size := len(buffer.rawData)
	vlen := size - buffer.rpos + buffer.wpos
	if vlen > n{
		vlen = n
	}
	buffer.rpos = (buffer.rpos+vlen) % size
	buffer.isEmpty = buffer.rpos == buffer.wpos
}

func (buffer *DxRingBuffer)Read(p []byte)(n int, err error)  {
	btlen := len(p)
	if buffer.isEmpty || btlen == 0{
		return 0,nil
	}
	size := len(buffer.rawData)
	if buffer.wpos > buffer.rpos{
		n = buffer.wpos - buffer.rpos
		if n > btlen{
			n = btlen
		}
		copy(p,buffer.rawData[buffer.rpos:buffer.rpos+n])
		buffer.rpos = (buffer.rpos + n)%size
		buffer.isEmpty = buffer.rpos == buffer.wpos
		return
	}
	n = size - buffer.rpos + buffer.wpos
	if n > btlen{
		n = btlen
	}
	if buffer.rpos + n <= size{
		copy(p,buffer.rawData[buffer.rpos:buffer.rpos+n])
	}else{
		len1 := size - buffer.rpos
		copy(p,buffer.rawData[buffer.rpos:size])
		len2 := n - len1
		copy(p[len1:],buffer.rawData[:len2])
	}
	buffer.rpos = (buffer.rpos+n) % size
	buffer.isEmpty = buffer.rpos == buffer.wpos
	return n,nil
}

func (buffer *DxRingBuffer)ReadByte()(b byte,err error)  {
	if buffer.isEmpty{
		return 0,errors.New("IsEmpty")
	}
	b = buffer.rawData[buffer.rpos]
	buffer.rpos++
	if buffer.rpos == buffer.Capacity(){
		buffer.rpos = 0
	}
	buffer.isEmpty = buffer.rpos == buffer.wpos
	return b,nil
}

func (buffer *DxRingBuffer)Write(p []byte)(n int, err error)  {
	n = len(p)
	if n == 0{
		return
	}
	free := buffer.FreeSize()
	if n > free{
		buffer.malloc(n - free)
	}
	size := len(buffer.rawData)
	if buffer.wpos >= buffer.rpos{
		c1 := size - buffer.wpos
		if c1 >= n{
			copy(buffer.rawData[buffer.wpos:],p)
			buffer.wpos += n
		}else{
			copy(buffer.rawData[buffer.wpos:],p[:c1])
			c2 := n - c1
			copy(buffer.rawData[:c2],p[c1:])
			buffer.wpos = c2
		}
	}else{
		copy(buffer.rawData[buffer.wpos:],p)
		buffer.wpos += n
	}
	if buffer.wpos == buffer.rpos{
		buffer.wpos = 0
	}
	buffer.isEmpty = false
	return
}

func (buffer * DxRingBuffer)WriteByte(b byte)  {
	if buffer.FreeSize()<1{
		buffer.malloc(1)
	}
	buffer.rawData[buffer.wpos] = b
	buffer.wpos++
	if buffer.wpos == buffer.Capacity(){
		buffer.wpos = 0
	}
	buffer.isEmpty = false
}

func (buffer *DxRingBuffer)IsFull()bool  {
	return  buffer.rpos == buffer.wpos && !buffer.isEmpty
}

func (buffer *DxRingBuffer)IsEmpty()bool  {
	return buffer.isEmpty
}

func (buffer *DxRingBuffer)Reset()  {
	buffer.hasMark = false
	buffer.rpos = 0
	buffer.wpos = 0
	buffer.isEmpty = true
}

func (buffer *DxRingBuffer)malloc(cap int)  {
	newcap := len(buffer.rawData) + cap
	if newcap < 256{
		newcap = newcap * 2
	}else if newcap < 512{
		newcap = 512
	}else {
		newcap = newcap + 512
	}
	newbuf := make([]byte,newcap)
	oldlen := buffer.DataSize()
	buffer.Read(newbuf)
	buffer.rpos = 0
	buffer.wpos = oldlen
	buffer.rawData = newbuf
}

func NewRingBuffer(size int)*DxRingBuffer  {
	return &DxRingBuffer{
		rawData: make([]byte,size),
		rpos:    0,
		wpos:    0,
		isEmpty: true,
	}
}