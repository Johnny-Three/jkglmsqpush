package enviroment

import (
	"database/sql"
	"fmt"

	seelog "github.com/cihub/seelog"

	"errors"

	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql"
)

type ConfigFile struct {
	LogDes        string
	Wanbudatabase string
	Hmpdatabase   string
	Nsqaddress    string
	Nsqtopic1     string
	Nsqtopic2     string
	Starttime     string
	Sendtime      string
	Message       []string
	Listenport    string
	Debug         string
	Wenjuan1      []int
	Wenjuan2      []int
}

type Config struct {
	Db1        *sql.DB
	Db2        *sql.DB
	Condition1 []int
	Condition2 []int
	Nsqaddress string
	Nsqtopic1  string
	Nsqtopic2  string
	Starttime  string
	Sendtime   string
	Message    []string
	Listenport string
	Debug      string
	LogDes     string
	Err        error
}

var Logger seelog.LoggerInterface

func loadAppConfig(appConfig string) {

	logger, err := seelog.LoggerFromConfigAsBytes([]byte(appConfig))
	if err != nil {
		fmt.Println(err)
		return
	}
	UseLogger(logger)
}

// DisableLog disables all library log output
func DisableLog() {
	Logger = seelog.Disabled
}

// UseLogger uses a specified seelog.LoggerInterface to output library log.
// Use this func if you are using Seelog logging system in your app.
func UseLogger(newLogger seelog.LoggerInterface) {
	Logger = newLogger
}

//EnvBuild需要正确的解析文件并且初始化DB和Redis的连接。。
func EnvBuild(filepath string) Config {

	var tmp ConfigFile
	var conf Config

	if _, err := toml.DecodeFile(filepath, &tmp); err != nil {
		conf.Err = err
		return conf
	}
	conf.Starttime = tmp.Starttime
	conf.Sendtime = tmp.Sendtime
	conf.Message = tmp.Message
	conf.Nsqaddress = tmp.Nsqaddress
	conf.Nsqtopic1 = tmp.Nsqtopic1
	conf.Nsqtopic2 = tmp.Nsqtopic2
	conf.Listenport = tmp.Listenport
	conf.Debug = tmp.Debug
	conf.Condition1 = tmp.Wenjuan1
	if len(conf.Condition1) != 4 {
		conf.Err = errors.New("condition1 len should be 4")
		return conf
	}
	conf.Condition2 = tmp.Wenjuan2
	if len(conf.Condition2) != 3 {
		conf.Err = errors.New("condition2 len should be 3")
		return conf
	}
	if conf.Condition2[0] >= conf.Condition2[1] {
		conf.Err = errors.New("condition2 param1 should less than param2")
		return conf
	}
	conf.LogDes = tmp.LogDes

	//open db1
	db1, err := sql.Open("mysql", tmp.Wanbudatabase)
	fmt.Println("db1", tmp.Wanbudatabase)
	//defer db1.Close()
	db1.SetMaxOpenConns(50)
	db1.SetMaxIdleConns(10)
	db1.Ping()

	if err != nil {
		conf.Err = err
		return conf
	}

	conf.Db1 = db1

	//open db2
	db2, err := sql.Open("mysql", tmp.Hmpdatabase)
	fmt.Println("db2", tmp.Hmpdatabase)
	//defer db1.Close()
	db2.SetMaxOpenConns(50)
	db2.SetMaxIdleConns(10)
	db2.Ping()
	if err != nil {
		conf.Err = err
		return conf
	}

	conf.Db2 = db2

	DisableLog()
	loadAppConfig(conf.LogDes)

	conf.Err = nil
	return conf
}
