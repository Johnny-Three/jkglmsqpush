package main

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

type Finshday struct {
	starttime int64 //处方下载日期
	endtime   int64 //处方下载日期 + 30 天
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

//查找处方完成天数
func (f *Finshday) Count() int {

	f.lock.RLock()
	defer f.lock.RUnlock()

	num := 0
	for _, value := range f.statemap {

		if value == 1 {
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
	for i := 0; i < 30; i++ {
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
	if len(f.statemap) != 30 {
		return f.Build(date)
	}

	//如果下载时间超前，不予变更;
	if date <= f.starttime {

		return nil
	}
	fmt.Printf("downloadtime[%d],endtime[%d]", date, f.endtime)
	//如果下载时间迟后，重做map
	if date > f.endtime {
		f.statemap = make(map[int64]int8)
		for i := 0; i < 30; i++ {
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
	return nil
}

//时间变化需要自动变化，每天0点自动运行，传参当前天
func (f *Finshday) Changeeveryday(date int64) error {

	f.lock.Lock()
	defer f.lock.Unlock()
	//如果当天0点的time比starttime要小，说明尚无需开始计算
	if date < f.starttime {
		return nil
	}

	if date-f.endtime != 86400 {
		return errors.New("传入日期有误，需要在原末尾天后补一")
	}

	f.starttime = f.starttime + 86400
	f.endtime = f.endtime + 86400
	tmpmap := make(map[int64]int8)
	// To store the keys in slice in sorted order
	var keys []int
	for k := range f.statemap {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	// To perform the opertion you want
	for _, k := range keys[1:] {
		tmpmap[int64(k)] = f.statemap[int64(k)]
	}
	//当天时间写入map
	tmpmap[f.endtime] = 0
	f.statemap = tmpmap

	return nil
}
