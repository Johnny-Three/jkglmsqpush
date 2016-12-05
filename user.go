package main

import (
	"errors"
	"fmt"
	"time"
)

type Userinfo struct {
	Userid    int
	Starttime int64
	Chufang   *Finshday
}

//传入参数，userid
func (u *Userinfo) New(uid int, st int64) (ui *Userinfo, err error) {

	if (st-57600)%86400 != 0 {
		return nil, fmt.Errorf("[%d],时间值错误", st)
	}

	u.Starttime = st
	var tmp Finshday
	if err := tmp.Build(u.Starttime); err != nil {
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

func GetTimestamp(date string) (timestamp int64) {
	tm, _ := time.ParseInLocation("2006-01-02", date, time.Local)
	timestamp = tm.Unix()
	return timestamp
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
