package util

import "time"

func GetTimestamp(date string) (timestamp int64) {
	tm, _ := time.ParseInLocation("2006-01-02", date, time.Local)
	timestamp = tm.Unix()
	return timestamp
}
