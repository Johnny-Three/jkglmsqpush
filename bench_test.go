package main

import "testing"

func Benchmark_Datecheckvalid(b *testing.B) {
	//use b.N for looping
	for i := 0; i < b.N; i++ {
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
			b.Run(tt.name, func(b *testing.B) {
				if got := Datecheckvalid(tt.args.date); got != tt.want {
					b.Errorf("Datecheckvalid() = %v, want %v", got, tt.want)
				}
			})
		}
	}
}

// 测试并发效率
func Benchmark_Datecheckvalid_Parallel(b *testing.B) {

	b.RunParallel(func(pb *testing.PB) {
		type args struct {
			date int64
		}
		rw := args{1478793600}
		for pb.Next() {
			if got := Datecheckvalid(rw.date); !got {
				b.Errorf("Datecheckvalid() = %v", got)
			}
		}
	})
}

func Benchmark_Set(b *testing.B) {
	var err error
	f := &Finshday{
		starttime: 1,
		endtime:   1,
		statemap:  make(map[int64]int8),
	}

	//1479571200  - 11.20
	//1479657600  - 11.21
	err = f.Build(1479571200)
	if err != nil {
		//格式无错误，查看是否==30
		b.Errorf("Finshday.Build() error = %v", err)
	}
	//use b.N for looping
	for i := 0; i < b.N; i++ {
		err = f.Set(1479657600, 1)
		if err != nil {
			b.Logf("Finshday.Set() error = %v", err)
		}
	}
}

// 测试并发效率
func Benchmark_Set_Parallel(b *testing.B) {

	b.RunParallel(func(pb *testing.PB) {
		var err error
		f := &Finshday{
			starttime: 1,
			endtime:   1,
			statemap:  make(map[int64]int8),
		}

		//1479571200  - 11.20
		//1479657600  - 11.21
		err = f.Build(1479571200)
		if err != nil {
			//格式无错误，查看是否==30
			b.Errorf("Finshday.Build() error = %v", err)
		}

		for pb.Next() {

			err = f.Set(1479657600, 1)
			if err != nil {
				b.Logf("Finshday.Set() error = %v", err)
			}
		}
	})
}

func Benchmark_Changeeverday(b *testing.B) {

	var err error
	f := &Finshday{
		starttime: 1,
		endtime:   1,
		statemap:  make(map[int64]int8),
	}

	//use b.N for looping
	for i := 0; i < b.N; i++ {

		//1479571200  - 11.01
		err = f.Build(1477929600)
		if err != nil {
			//格式无错误，查看是否==30
			b.Errorf("Finshday.Build() error = %v", err)
		}
		//12.01
		err = f.Changeeveryday(1480521600)
		if err != nil {
			b.Logf("Finshday.Changeeveryday() error = %v", err)
		}
	}
}

func Benchmark_Changeeverday_Parallel(b *testing.B) {

	b.RunParallel(func(pb *testing.PB) {

		var err error
		f := &Finshday{
			starttime: 1,
			endtime:   1,
			statemap:  make(map[int64]int8),
		}

		for pb.Next() {

			//1479571200  - 11.01
			err = f.Build(1477929600)
			if err != nil {
				//格式无错误，查看是否==30
				b.Errorf("Finshday.Build() error = %v", err)
			}

			//12.01
			err = f.Changeeveryday(1480521600)
			if err != nil {
				b.Logf("Finshday.Changeeveryday() error = %v", err)
			}
		}

	})
}

func Benchmark_Count(b *testing.B) {

	f := &Finshday{
		starttime: 1,
		endtime:   1,
		statemap:  make(map[int64]int8),
	}
	//1479571200  - 11.20
	//1479657600  - 11.21
	_ = f.Build(1479571200)
	//预期总数为1
	_ = f.Set(1479571200, 1)
	//预期总数为2
	_ = f.Set(1479657600, 1)

	//use b.N for looping
	for i := 0; i < b.N; i++ {
		//1479571200  - 11.01
		_ = f.Count()
	}
}
