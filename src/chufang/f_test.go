package chufang

import (
	"sort"
	"sync"
	"testing"
	"time"
)

func TestDatecheckvalid(t *testing.T) {
	type args struct {
		date int64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{"test1", args{1478793600}, true},
		{"test2", args{1478793601}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Datecheckvalid(tt.args.date); got != tt.want {
				t.Errorf("Datecheckvalid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFinshday_Count(t *testing.T) {
	type fields struct {
		starttime int64
		endtime   int64
		lock      sync.RWMutex
		statemap  map[int64]int8
	}
	tests := []struct {
		name   string
		fields fields
		want   int8
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Finshday{
				starttime: tt.fields.starttime,
				endtime:   tt.fields.endtime,
				lock:      tt.fields.lock,
				statemap:  tt.fields.statemap,
			}
			if got := f.Count(); got != tt.want {
				t.Errorf("Finshday.Count() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFinshday_Set(t *testing.T) {

	var err error
	f := &Finshday{
		starttime: 1,
		endtime:   1,
		statemap:  make(map[int64]int8),
	}
	//1479571200  - 11.20
	//1479657600  - 11.21
	err = f.Build(1479571200)
	if err == nil {
		//格式无错误，查看是否==30
		t.Logf("Finshday.Build() length = %v,want length = %v", len(f.statemap), 30)
	} else {
		t.Errorf("Finshday.Build() error = %v", err)
	}
	//预期错误
	err = f.Set(1479657599, 1)
	if err == nil {
		t.Logf("Finshday.Count() length = %v,want length = %v", f.Count(), 2)
	} else {
		t.Errorf("Finshday.Set() error = %v", err)
	}
	//预期总数为1
	err = f.Set(1479571200, 1)
	if err == nil {
		t.Logf("Finshday.Count() length = %v,want length = %v", f.Count(), 1)
	} else {
		t.Errorf("Finshday.Set() error = %v", err)
	}
	//预期总数为2
	err = f.Set(1479657600, 1)
	if err == nil {
		t.Logf("Finshday.Count() length = %v,want length = %v", f.Count(), 2)
	} else {
		t.Errorf("Finshday.Set() error = %v", err)
	}
	//预期总数为1
	err = f.Set(1479657600, 0)
	if err == nil {
		t.Logf("Finshday.Count() length = %v,want length = %v", f.Count(), 1)
	} else {
		t.Errorf("Finshday.Set() error = %v", err)
	}
}

func TestFinshday_Build(t *testing.T) {

	//1479744000
	type fields struct {
		starttime int64
		endtime   int64
		statemap  map[int64]int8
		lock      sync.RWMutex
	}
	type args struct {
		date int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		len    int8
	}{
		{
			name: "测试时间格式",
			fields: fields{
				starttime: 1,
				endtime:   1,
				statemap:  make(map[int64]int8),
			},
			args: args{1479484800},
			len:  30,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Finshday{
				starttime: tt.fields.starttime,
				endtime:   tt.fields.endtime,
				lock:      tt.fields.lock,
				statemap:  tt.fields.statemap,
			}
			if err := f.Build(tt.args.date); err != nil {
				t.Errorf("Finshday.Build() error = %v", err)
			}
			if err := f.Build(tt.args.date); err == nil {
				t.Logf("Finshday.Build() length = %v,want length = %v", len(f.statemap), 30)
			}

			var keys []int
			for k := range f.statemap {
				keys = append(keys, int(k))
			}
			sort.Ints(keys)
			// To perform the opertion you want
			for _, k := range keys {
				ts := time.Unix(int64(k), 0).Format("2006-01-02 15:04:05 PM")
				t.Logf("[%v,%v]", ts, f.statemap[int64(k)])
			}

		})
	}
}

func TestFinshday_Rebuild(t *testing.T) {

	var err error
	f := &Finshday{
		starttime: 1,
		endtime:   1,
		statemap:  make(map[int64]int8),
	}
	//1479484800  - 11.19
	//1479571200  - 11.20
	//1479657600  - 11.21
	//1479744000  - 11.22
	//1481644800  - 12.14
	_ = f.Build(1479484800)
	_ = f.Set(1479484800, 1)
	_ = f.Set(1479571200, 1)
	_ = f.Set(1479657600, 1)
	_ = f.Set(1481644800, 1)
	t.Logf("==count of statemap = %v", f.Count())
	err = f.Rebuild(1479484800)
	if err == nil {
		t.Logf("Finshday.Rebuild() length = %v,want length = %v", len(f.statemap), 30)
	} else {
		t.Errorf("Finshday.Set() error = %v", err)
	}
	// 12-15
	err = f.Rebuild(1479571200)
	t.Log("f.Startime is ", f.starttime)
	if err == nil {
		t.Logf("Finshday.Rebuild() length = %v,want length = %v", len(f.statemap), 30)
	} else {
		t.Errorf("Finshday.Set() error = %v", err)
	}
	t.Logf("==count of statemap = %v", f.Count())

	var keys []int
	for k := range f.statemap {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	// To perform the opertion you want
	for _, k := range keys {
		ts := time.Unix(int64(k), 0).Format("2006-01-02 15:04:05 PM")
		t.Logf("[%v,%v]", ts, f.statemap[int64(k)])
	}
}

func TestFinshday_Changeeveryday(t *testing.T) {

	var err error
	f := &Finshday{
		starttime: 1,
		endtime:   1,
		statemap:  make(map[int64]int8),
	}
	//1477929600  - 11.01
	//1479484800  - 11.19
	//1479571200  - 11.20
	//1479657600  - 11.21
	//1481644800  - 12.14
	_ = f.Build(1477929600)
	_ = f.Set(1479484800, 1)
	_ = f.Set(1479571200, 1)
	_ = f.Set(1479657600, 1)
	_ = f.Set(1481644800, 1)

	var keys []int
	for k := range f.statemap {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	// To perform the opertion you want
	for _, k := range keys {
		ts := time.Unix(int64(k), 0).Format("2006-01-02 15:04:05 PM")
		t.Logf("[%v,%v]", ts, f.statemap[int64(k)])
	}
	t.Logf("endtime = %v,%v", f.endtime, time.Unix(f.endtime, 0).Format("2006-01-02 15:04:05 PM"))

	err = f.Changeeveryday(1482076800)
	if err == nil {
		t.Logf("Finshday.Changeeveryday() length = %v,want length = %v", len(f.statemap), 30)
	} else {
		t.Errorf("Finshday.Changeeveryday() error = %v", err)
	}

	keys = []int{}
	for k := range f.statemap {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	// To perform the opertion you want
	for _, k := range keys {
		ts := time.Unix(int64(k), 0).Format("2006-01-02 15:04:05 PM")
		t.Logf("[%v,%v]", ts, f.statemap[int64(k)])
	}

	t.Logf("count=%v", f.Count())
}

func Test_main(t *testing.T) {

}
