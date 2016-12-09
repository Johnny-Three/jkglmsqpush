package chufang

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
	. "wbproject/jkglmsgpush/src/util"
)

type Finshday struct {
	starttime int64 //处方下载日期
	endtime   int64 //处方下载日期 + 31 天
	lock      sync.RWMutex
	statemap  map[int64]int8
}

func Datecheckvalid(date int64) bool {

	//东八区，需要减掉8个小时
	if (date-57600)%86400 == 0 {
		return true
	}
	return false
}

func (f *Finshday) ToString() string {

	f.lock.RLock()
	defer f.lock.RUnlock()

	var keys []int
	for k := range f.statemap {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	var buffer bytes.Buffer
	buffer.WriteString("\n")
	for i := 0; i < len(keys); i++ {

		t := time.Unix(int64(keys[i]), 0).Format("2006-01-02")
		s := f.statemap[int64(keys[i])]
		buffer.WriteString(fmt.Sprintf("[%s,%d,%d]\n", t, keys[i], s))
	}
	return buffer.String()
}

//查找处方完成天数
func (f *Finshday) Count() int {

	f.lock.RLock()
	defer f.lock.RUnlock()

	ct := GetTimestamp(time.Unix(time.Now().Unix(), 0).Format("2006-01-02"))
	//排除当天..
	num := 0
	for k, v := range f.statemap {

		if k == ct {
			continue
		}
		if v == 1 {
			num++
		}
	}
	return num
}

//重置某天任务完成
func (f *Finshday) Set(date int64, value int8) error {

	f.lock.Lock()
	defer f.lock.Unlock()

	if value < 0 || value > 1 {

		return fmt.Errorf("value值Expected 1 or 0 ,but 传入 %v", value)
	}

	if !Datecheckvalid(date) {
		return errors.New("日期格式错误，需为某天的0点数据,timestamp格式")
	}
	// 查找键值是否存在
	if _, ok := f.statemap[date]; ok {
		f.statemap[date] = value
		return nil
	}
	return errors.New("对应日期的数据不存在")
}

//返回开始时间
func (f *Finshday) GetStarttime() int64 {
	return f.starttime
}

//返回结束时间
func (f *Finshday) GetEndtime() int64 {
	return f.endtime
}

//初始化重置
func (f *Finshday) Build(date int64) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	if !Datecheckvalid(date) {
		return errors.New("日期格式错误，需为某天的0点数据,timestamp格式")
	}
	f.starttime, f.endtime = date, date
	f.statemap = make(map[int64]int8)
	for i := 0; i < 31; i++ {
		f.endtime = f.starttime + int64(i*86400)
		f.statemap[f.endtime] = 0
	}
	return nil
}

//starttime变更，重新rebuild底层map
func (f *Finshday) Rebuild(date int64) error {

	f.lock.Lock()
	defer f.lock.Unlock()

	if !Datecheckvalid(date) {
		return errors.New("日期格式错误，需为某天的0点数据,timestamp格式")
	}

	//如果尚未build，则build之
	if len(f.statemap) != 31 {
		return f.Build(date)
	}

	//如果下载时间超前，不予变更;
	if date <= f.starttime {

		return nil
	}

	//如果下载时间迟后，重做map
	if date > f.endtime {
		f.statemap = make(map[int64]int8)
		for i := 0; i < 31; i++ {
			f.endtime = f.starttime + int64(i*86400)
			f.statemap[f.endtime] = 0
		}
		return nil
	}
	//下载时间在某个区域内，需要部分拷贝
	si := (date - f.starttime) / 86400
	tmpmap := make(map[int64]int8)
	// To store the keys in slice in sorted order
	var keys []int
	for k := range f.statemap {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	// To perform the opertion you want
	for _, k := range keys[si:] {
		tmpmap[int64(k)] = f.statemap[int64(k)]
		f.endtime = int64(k)
	}
	for i := si; i > 0; i-- {
		tmpmap[f.endtime+i*86400] = 0
	}
	f.statemap = tmpmap
	f.starttime = date
	return nil
}

//时间变化需要自动变化，每天0点自动运行，传参当前天
func (f *Finshday) Changeeveryday(date int64) error {

	f.lock.Lock()
	defer f.lock.Unlock()
	//如果当天0点的time比starttime要小，说明尚无需开始计算
	if date-f.endtime <= 0 {
		return nil
	}

	if i := (date - f.endtime) % 86400; i != 0 {
		return fmt.Errorf("date:[%d]格式错误", date)
	}

	i := (date - f.endtime) / 86400

	tmpmap := make(map[int64]int8)
	// To store the keys in slice in sorted order
	var keys []int
	for k := range f.statemap {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	// To perform the opertion you want
	for _, k := range keys[i:] {
		tmpmap[int64(k)] = f.statemap[int64(k)]
	}

	//修正开始和结束时间
	f.starttime = f.starttime + i*86400
	var c int64
	for c = 1; c <= i; c++ {
		tmpmap[f.endtime+c*86400] = 0
	}

	f.statemap = tmpmap
	return nil
}
