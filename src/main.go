package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	. "wbproject/jkglmsgpush/src/business"
	. "wbproject/jkglmsgpush/src/enviroment"
	httpserver "wbproject/jkglmsgpush/src/httpserver"
	. "wbproject/jkglmsgpush/src/nsq"
	. "wbproject/jkglmsgpush/src/user"

	"github.com/jasonlvhit/gocron"
)

var users Users
var version = "1.0.0PR10"
var filepath = "../etc/config.toml"
var modulename = "jkglmsgpush"

func main() {

	args := os.Args
	if len(args) == 2 && (args[1] == "-v") {
		fmt.Println("看好了兄弟，现在的版本是【", version, "】，可别弄错了")
	} else {

		fmt.Println("运行环境初始化进行中...")
		//初始化运行环境
		c := EnvBuild(filepath)
		if c.Err != nil {
			panic(c.Err)
		}
		fmt.Println("运行环境初始化完毕...")

		start := time.Now()
		Logger.Infof("begin time is %s", start.Format("2006-01-02 15:04:05"))
		var wg sync.WaitGroup
		//从HMP库中构建users..
		if err := users.BuildFromDb(&wg, c.Db1, c.Db2); err != nil {
			Logger.Critical(err)
		}
		wg.Wait()
		elapsed := time.Since(start)
		Logger.Info("Load db person query total time:", elapsed)

		fmt.Println("你好，用户的总数是：=====", len(users.Sl))
		//启动处理更改下载时间事件..
		go users.ModifyUsersStarttime(MsgChan)
		go users.ModifyUsersChuFangStatus(UserWalkDataChan)

		go func(users *Users, c *Config) {
			//启动接口
			httpserver.StoreUsers(users)
			router := httpserver.NewRouter()
			Logger.Info("report service starting...")
			err := http.ListenAndServe(":"+c.Listenport, router)
			if err != nil {
				Logger.Critical("report service failed")
			}
		}(&users, &c)

		//Consumer运行，消费下载消息..
		go func() {
			//对接NSQ，消费下载消息
			consumer, err := NewConsummer(c.Nsqtopic1, modulename)
			if err != nil {
				panic(err)
			}
			err = ConsumerRun(consumer, c.Nsqtopic1, c.Nsqaddress)
			if err != nil {
				panic(err)
			}
		}()
		//Consumer运行，消费上传消息..
		go func() {
			//对接NSQ，消费上传消息
			consumer, err := NewConsummer(c.Nsqtopic2, modulename)
			if err != nil {
				panic(err)
			}
			err = ConsumerRun(consumer, c.Nsqtopic2, c.Nsqaddress)
			if err != nil {
				panic(err)
			}
		}()

		//debug on 立刻执行
		if strings.EqualFold(c.Debug, "on") {

			TaskWenjuan1(&users, &c)
			TaskWenjuan2(&users, &c)

			for {
				time.Sleep(100 * time.Second)
			}
		} else if strings.EqualFold(c.Debug, "off") {

			//0点1分触发处方完成率滚动任务
			gocron.Every(1).Day().At("00:01").Do(TaskGundong, &users)

			//在指定时间触发固定问卷任务和处方完成率任务
			gocron.Every(1).Day().At(c.Sendtime).Do(TaskWenjuan1, &users, &c)
			gocron.Every(1).Day().At(c.Sendtime).Do(TaskWenjuan2, &users, &c)
			// function Start start all the pending jobs
			<-gocron.Start()
		}
	}
}
