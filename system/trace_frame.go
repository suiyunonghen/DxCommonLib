package system

import (
	"bytes"
	"runtime"
	"strconv"
	"strings"
)

func TraceFrame2Buffer(frameC int, buffer *bytes.Buffer) {
	if frameC < 1 {
		return
	}
	for i := 1; i < frameC+1; i++ {
		if pc, file, line, ok := runtime.Caller(i); !ok {
			break
		} else {
			if i > 1 {
				buffer.Write([]byte{'\r', '\n'})
			}
			idx := strings.LastIndexByte(file, '/')
			if idx == -1 {
				buffer.WriteString(file)
			} else {
				idx = strings.LastIndexByte(file[:idx], '/')
				if idx == -1 {
					buffer.WriteString(file)
				} else {
					buffer.WriteString(file[idx+1:])
				}
			}
			buffer.WriteByte(':')
			buffer.WriteString(strconv.Itoa(line))
			funcName := runtime.FuncForPC(pc).Name()
			idx = strings.IndexByte(funcName, '.')
			if idx > 0 {
				funcName = funcName[idx+1:]
			}
			buffer.WriteByte('.')
			buffer.WriteString(funcName)
		}
	}
}

func TraceFrame(frameC int, bt []byte) []byte {
	if frameC < 1 {
		return nil
	}
	for i := 1; i < frameC+1; i++ {
		if pc, file, line, ok := runtime.Caller(i); !ok {
			break
		} else {
			if i > 1 {
				bt = append(bt, '\r', '\n')
			}
			idx := strings.LastIndexByte(file, '/')
			if idx == -1 {
				//buffer.WriteString(file)
				bt = append(bt, file...)
			} else {
				idx = strings.LastIndexByte(file[:idx], '/')
				if idx == -1 {
					bt = append(bt, file...)
				} else {
					bt = append(bt, file[idx+1:]...)
				}
			}
			bt = append(bt, ':')
			bt = append(bt, strconv.Itoa(line)...)
			funcName := runtime.FuncForPC(pc).Name()
			idx = strings.IndexByte(funcName, '.')
			if idx > 0 {
				funcName = funcName[idx+1:]
			}
			bt = append(bt, '.')
			bt = append(bt, funcName...)
		}
	}
	return bt
}
