package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/bitly/go-simplejson"
)

var MsgChan chan DownloadMsg

type DownloadMsg struct {
	Userid    int   `json:"userid"`
	Starttime int64 `json:"starttime"`
}

func Decode(msg string) error {

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

func init() {
	MsgChan = make(chan DownloadMsg, 16)
}
