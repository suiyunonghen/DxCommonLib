package DxCommonLib

import (
	"unsafe"
)

//用来存放位
type  DxBits struct{
	buffer	[]byte
	fsize	uint				//存放的bit位数是多少
}

func (bt *DxBits)Count()uint  {
	return bt.fsize
}

//选中的位的个数
func (bt *DxBits)CheckedCount()uint  {
	if bt.fsize == 0{
		return 0
	}
	result := uint(0)
	for _,v := range bt.buffer{
		for i := 0;i<8;i++{
			realv := byte(1 << uint(i))
			if v & realv == realv{
				result++
			}
		}
	}
	return result
}

func (bt *DxBits)ReSet(bsize uint)  {
	if bsize == 0{
		bt.buffer = nil
		bt.fsize = 0
		return
	}
	if bt.fsize == bsize{
		return
	}
	bt.fsize = bsize
	buflen := bt.fsize / 8
	if bt.fsize % 8 != 0{
		buflen += 1
	}
	bt.buffer = make([]byte,buflen)
}

func (bt *DxBits)Clear()  {
	for i := 0;i<len(bt.buffer);i++{
		bt.buffer[i] = 0
	}
}

func (bt *DxBits)ReSetByInt32(v int32)  {
	bt.ReSet(32)
	*(*int32)(unsafe.Pointer(&bt.buffer[0])) = v
}

func (bt *DxBits)ReSetByInt64(v int64)  {
	bt.ReSet(64)
	*(*int64)(unsafe.Pointer(&bt.buffer[0])) = v
}

func (bt *DxBits)Bits(index uint)bool  {
	if index < bt.fsize{
		btindex := index / 8
		index = index % 8
		realv := byte(1 << uint(index))
		return bt.buffer[btindex] & realv == realv
	}
	return false
}

func (bt *DxBits)SetBits(index uint,v bool)  {
	if index < bt.fsize{
		btindex := index / 8
		index = index % 8
		if v {
			realv := byte(1 << uint(index))
			bt.buffer[btindex] = bt.buffer[btindex] | realv
		}else{
			realv := byte(1 << uint(index))
			realv = ^realv
			bt.buffer[btindex] = bt.buffer[btindex] & realv
		}
	}
}

//将指定的位取反，如果指定的位为-1，则将全部的位各自取反，1变0,0变1
func (bt *DxBits)NotBits(index int){
	if index < 0{
		for idx,v := range bt.buffer{
			for i := 0;i<8;i++{
				realv := byte(1 << uint(i))
				if v & realv == realv{
					realv = ^realv
					v = v & realv
				}else{
					v = v | realv
				}
			}
			bt.buffer[idx] = v
		}
	}else{
		btindex := index / 8
		index = index % 8
		realv := byte(1 << uint(index))
		oldBitValid := bt.buffer[btindex] & realv == realv
		if oldBitValid{
			realv = ^realv
			bt.buffer[btindex] = bt.buffer[btindex] & realv
		}else{
			bt.buffer[btindex] = bt.buffer[btindex] | realv
		}
	}
}

func (bt *DxBits)AsUInt64()uint64  {
	bflen := len(bt.buffer)
	if bflen > 0{
		result := uint(bt.buffer[0])
		for i := 1;i<bflen;i++{
			if i == 8{
				break
			}
			result = result | uint(bt.buffer[i]) << uint(8 * i)
		}
		return uint64(result)
	}
	return 0
}

func (bt *DxBits)AsInt64()int64  {
	return int64(bt.AsUInt64())
}

func (bt *DxBits)AsUInt32()uint32  {
	bflen := len(bt.buffer)
	if bflen > 0{
		result := uint(bt.buffer[0])
		for i := 1;i<bflen;i++{
			if i == 4{
				break
			}
			result = result | uint(bt.buffer[i]) << uint(8 * i)
		}
		return uint32(result)
	}
	return 0
}

func (bt *DxBits)AsInt32()int32  {
	return int32(bt.AsUInt32())
}