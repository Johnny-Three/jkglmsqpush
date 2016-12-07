package user

import (
	"errors"
	"fmt"
	"time"
	M "wbproject/jkglmsgpush/src/chufang"
	. "wbproject/jkglmsgpush/src/util"
)

type Userinfo struct {
	Userid    int
	Starttime int64 //处方下载时间
	Chufang   *M.Finshday
}

func (u *Userinfo) CompareDate(min, max int64) (yes bool, s, e int64) {

	st := u.Chufang.GetStarttime()
	ed := u.Chufang.GetEndtime()

	if min > ed || max < st {
		return false, 0, 0
	}

	if max <= ed {
		e = max
	} else {
		e = ed
	}

	if min <= st {
		s = st
	} else {
		s = min
	}

	return true, s, e
}

//传入参数，userid
func (u *Userinfo) New(uid int, st, st1 int64) (ui *Userinfo, err error) {

	if (st-57600)%86400 != 0 {
		return nil, fmt.Errorf("[%d],时间值错误", st)
	}

	var tmp M.Finshday
	if err := tmp.Build(st1); err != nil {
		return nil, err
	}

	return &Userinfo{
		Userid:    uid,
		Starttime: st,
		Chufang:   &tmp,
	}, nil
}

//更改starttime
func (u *Userinfo) ModifyStarttime(st int64) error {

	if (st-57600)%86400 == 0 {
		u.Starttime = st
		return nil
	}
	return errors.New("时间值错误")
}

//查看用户从下载日期到当前日期经过了多少天
func (u *Userinfo) TimeAfterStart() (error, int) {

	ct := GetTimestamp(time.Unix(time.Now().Unix(), 0).Format("2006-01-02"))
	//fmt.Printf("now:%d,starttime:%d,userid:%d\n", ct, u.Starttime, u.Userid)
	if (ct-u.Starttime)%86400 != 0 {
		return errors.New("时间值错误"), -1
	}
	return nil, int((ct - u.Starttime) / 86400)
}
