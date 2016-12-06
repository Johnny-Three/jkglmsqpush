package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
)

var MsgChan chan DownloadMsg

type DownloadMsg struct {
	Userid    int   `json:"userid"`
	Starttime int64 `json:"starttime"`
}

var UserWalkDataChan chan UserWalkdaysStruct

type UserWalkdaysStruct struct {
	Uid      int
	Walkdays []WalkDayData
}

type WalkDayData struct {
	Daydata          int
	Hourdata         []int
	Chufangid        int
	Chufangfinish    int
	Chufangtotal     int
	Faststepnum      int
	Effecitvestepnum int
	WalkDate         int64
	Timestamp        int64
}

func Slice_Atoi(strArr []string) ([]int, error) {
	// NOTE:  Read Arr as Slice as you like
	var str string // O
	var i int      // O
	var err error  // O

	iArr := make([]int, 0, len(strArr))
	for _, str = range strArr {
		i, err = strconv.Atoi(str)
		if err != nil {
			return nil, err // O
		}
		iArr = append(iArr, i)
	}
	return iArr, nil
}

func DecodeMT(msg string) error {

	js, err := simplejson.NewJson([]byte(msg))
	if err != nil {
		errback := fmt.Sprintf("decode json error the error msg is %s", err.Error())
		return errors.New(errback)
	}

	userid := js.Get("userid").MustInt()
	downloadtime := js.Get("starttime").MustInt64()
	downloadtime = GetTimestamp(time.Unix(downloadtime, 0).Format("2006-01-02"))
	downloadmsg := DownloadMsg{userid, downloadtime}
	MsgChan <- downloadmsg

	return nil
}

func DecodeMS(msg string) error {

	js, err := simplejson.NewJson([]byte(msg))
	if err != nil {
		errback := fmt.Sprintf("decode json error the error msg is %s", err.Error())
		return errors.New(errback)
	}

	var wd WalkDayData
	walkdays := []WalkDayData{}
	userwalkdata := UserWalkdaysStruct{}

	userid := js.Get("userid").MustInt()
	wd.Timestamp = js.Get("timestamp").MustInt64()
	arr, _ := js.Get("walkdays").Array()

	for index, _ := range arr {

		walkdate := js.Get("walkdays").GetIndex(index).Get("walkdate").MustInt64()
		wd.WalkDate = walkdate

		var err0 error
		walkhour := js.Get("walkdays").GetIndex(index).Get("walkhour").MustString()
		wd.Hourdata, err0 = Slice_Atoi(strings.Split(walkhour, ","))
		if err0 == nil {

			if len(wd.Hourdata) != 24 {
				errback := fmt.Sprintf("uid %d walkdate %d get wrong hourdata %v format", userid, walkdate, wd.Hourdata)
				return errors.New(errback)
			}
		}

		wd.Daydata = js.Get("walkdays").GetIndex(index).Get("walktotal").MustInt()
		wd.Faststepnum = js.Get("walkdays").GetIndex(index).Get("fast").MustInt()
		wd.Effecitvestepnum = js.Get("walkdays").GetIndex(index).Get("effective").MustInt()
		s_recipe := js.Get("walkdays").GetIndex(index).Get("recipe").MustString()
		i_recipe, err1 := Slice_Atoi(strings.Split(s_recipe, ","))
		if err1 == nil {

			if len(i_recipe) != 3 {
				errback := fmt.Sprintf("uid %d walkdate %d get wrong recipe %v format", userid, walkdate, i_recipe)
				return errors.New(errback)
			}
		}
		//no problem .. then assign the chufang related value..
		wd.Chufangid = i_recipe[0]
		wd.Chufangfinish = i_recipe[1]
		wd.Chufangtotal = i_recipe[2]

		//用户此次上传的数据消息存储在MAP中..
		walkdays = append(walkdays, wd)

	}

	userwalkdata.Uid = userid
	userwalkdata.Walkdays = walkdays
	UserWalkDataChan <- userwalkdata
	return nil
}

func init() {
	MsgChan = make(chan DownloadMsg, 16)
	UserWalkDataChan = make(chan UserWalkdaysStruct, 16)
}
