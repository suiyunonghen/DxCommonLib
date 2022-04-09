package DxCommonLib

import (
	"bytes"
	"errors"
	"runtime"
	"strconv"
	"time"
)

type (
	TDateTime float64
)

const (
	MinsPerHour = 60
	MinsPerDay  = 24 * MinsPerHour
	SecsPerDay  = MinsPerDay * 60
	MSecsPerDay = SecsPerDay * 1000
)

var (
	delphiFirstTime        time.Time
	IsAmd64                = runtime.GOARCH == "amd64"
	ErrInvalidJsonDateTime = errors.New("invalidate json dateTime format")
)

func init() {
	delphiFirstTime = time.Date(1899, 12, 30, 0, 0, 0, 0, time.Local)
}

/*ToTime
从Delphi日期转为Go日期格式
Delphi的日期规则为到1899-12-30号的天数+当前的毫秒数/一天的总共毫秒数集合
*/
func (date TDateTime) ToTime() time.Time {
	mDay := time.Duration(date)
	ms := (date - TDateTime(mDay)) * TDateTime(MSecsPerDay)
	return delphiFirstTime.Add(mDay*time.Hour*24 + time.Duration(ms)*time.Millisecond)
}

func (date *TDateTime) WrapTime2Self(t time.Time) {
	days := t.Sub(delphiFirstTime) / (time.Hour * 24)
	y, m, d := t.Date()
	nowdate := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	times := float64(t.Sub(nowdate)) / float64(time.Hour*24)
	*date = TDateTime(float64(days) + times)
}

func Time2DelphiTime(t time.Time) TDateTime {
	if t.IsZero() {
		return 0
	}
	days := t.Sub(delphiFirstTime) / (time.Hour * 24)
	y, m, d := t.Date()
	nowdate := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	times := float64(t.Sub(nowdate)) / float64(time.Hour*24)
	return TDateTime(float64(days) + times)
}

//ParserJsonTime
//Date(1402384458000)
//Date(1224043200000+0800)
func ParserJsonTime(jsontime string) (time.Time, error) {
	bt := FastString2Byte(jsontime)
	dtflaglen := 0
	endlen := 0
	if bytes.HasPrefix(bt, []byte{'D', 'a', 't', 'e', '('}) && bytes.HasSuffix(bt, []byte{')'}) {
		dtflaglen = 5
		endlen = 1
	} else if bytes.HasPrefix(bt, []byte{'/', 'D', 'a', 't', 'e', '('}) && bytes.HasSuffix(bt, []byte{')', '/'}) {
		dtflaglen = 6
		endlen = 2
	}
	if dtflaglen > 0 {
		bt = bt[dtflaglen : len(bt)-endlen]
		var (
			ms  int64
			err error
		)
		endlen = 0
		idx := bytes.IndexByte(bt, '+')
		if idx < 0 {
			idx = bytes.IndexByte(bt, '-')
		} else {
			endlen = 1
		}
		if idx < 0 {
			str := FastByte2String(bt[:])
			if ms, err = strconv.ParseInt(str, 10, 64); err != nil {
				return time.Time{}, ErrInvalidJsonDateTime
			}
			if len(str) > 9 {
				ms = ms / 1000
			}
		} else {
			if endlen == 0 {
				endlen = -1
			}
			str := FastByte2String(bt[:idx])
			ms, err = strconv.ParseInt(str, 10, 64)
			if err != nil {
				return time.Time{}, err
			}
			bt = bt[idx+1:]
			if len(bt) < 2 {
				return time.Time{}, ErrInvalidJsonDateTime
			}
			bt = bt[:2]
			ctz, err := strconv.Atoi(FastByte2String(bt))
			if err != nil {
				return time.Time{}, ErrInvalidJsonDateTime
			}
			if len(str) > 9 {
				ms = ms / 1000
			}
			ms += int64(ctz * 60)
		}
		ntime := time.Now()
		ns := ntime.Unix()
		ntime = ntime.Add((time.Duration(ms-ns) * time.Second))
		return ntime, nil
	}
	return time.Time{}, ErrInvalidJsonDateTime
}
